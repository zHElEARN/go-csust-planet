package csustkit

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrCampusCardNotLoggedIn = errors.New("校园卡系统未登录")
	ErrSSONotLoggedIn        = errors.New("统一身份认证未登录")
	ErrCampusNotFound        = errors.New("未找到校区")
	ErrBuildingNotFound      = errors.New("未找到楼栋")
	ErrRoomNotFound          = errors.New("未找到宿舍")
)

type campusCardAPI interface {
	GetBuildings(ctx context.Context, campus Campus) ([]Building, error)
	GetRooms(ctx context.Context, building Building) ([]Room, error)
	GetElectricity(ctx context.Context, room Room) (float64, error)
	SyncToken(ctx context.Context, ticket string) error
	IsLoggedIn(ctx context.Context) bool
}

type ssoAPI interface {
	GetLoginForm(ctx context.Context) (SSOLoginForm, error)
	Login(ctx context.Context, form SSOLoginForm, username, password, captcha string) error
	LoginToCampusCard(ctx context.Context) (string, error)
	IsLoggedIn(ctx context.Context) bool
}

type ElectricityClient struct {
	sso        ssoAPI
	campusCard campusCardAPI
	username   string
	password   string

	mu        sync.RWMutex
	reloginMu sync.Mutex
	buildings map[string]map[string]Building
	rooms     map[string]map[string]Room
}

func NewAuthenticatedElectricityClient(ctx context.Context, username, password string) (*ElectricityClient, error) {
	client, err := NewClient()
	if err != nil {
		return nil, fmt.Errorf("创建校园卡客户端失败: %w", err)
	}

	electricityClient := newElectricityClientWithAuth(client.SSO(), client.CampusCard(), username, password)
	if err := electricityClient.loginSSO(ctx); err != nil {
		return nil, err
	}
	if err := electricityClient.loginCampusCard(ctx); err != nil {
		return nil, err
	}

	return electricityClient, nil
}

func newElectricityClient(campusCard campusCardAPI) *ElectricityClient {
	return newElectricityClientWithAuth(nil, campusCard, "", "")
}

func newElectricityClientWithAuth(sso ssoAPI, campusCard campusCardAPI, username, password string) *ElectricityClient {
	return &ElectricityClient{
		sso:        sso,
		campusCard: campusCard,
		username:   username,
		password:   password,
		buildings:  make(map[string]map[string]Building),
		rooms:      make(map[string]map[string]Room),
	}
}

func (c *ElectricityClient) ValidateRoom(ctx context.Context, campusName, buildingName, roomName string) error {
	_, err := c.resolveRoom(ctx, campusName, buildingName, roomName)
	return err
}

func (c *ElectricityClient) GetElectricity(ctx context.Context, campusName, buildingName, roomName string) (float64, error) {
	room, err := c.resolveRoom(ctx, campusName, buildingName, roomName)
	if err != nil {
		return 0, err
	}
	return withCampusCardAuthRetry(c, ctx, func() (float64, error) {
		return c.campusCard.GetElectricity(ctx, room)
	})
}

func (c *ElectricityClient) resolveRoom(ctx context.Context, campusName, buildingName, roomName string) (Room, error) {
	campus, err := findCampus(campusName)
	if err != nil {
		return Room{}, err
	}

	building, err := c.resolveBuilding(ctx, campus, buildingName)
	if err != nil {
		return Room{}, err
	}

	rooms, err := c.roomsForBuilding(ctx, building)
	if err != nil {
		return Room{}, err
	}
	if room, ok := rooms[roomName]; ok {
		return room, nil
	}
	return Room{}, fmt.Errorf("%w: %s", ErrRoomNotFound, roomName)
}

func (c *ElectricityClient) resolveBuilding(ctx context.Context, campus Campus, buildingName string) (Building, error) {
	buildings, err := c.buildingsForCampus(ctx, campus)
	if err != nil {
		return Building{}, err
	}
	if building, ok := buildings[buildingName]; ok {
		return building, nil
	}
	return Building{}, fmt.Errorf("%w: %s", ErrBuildingNotFound, buildingName)
}

func (c *ElectricityClient) buildingsForCampus(ctx context.Context, campus Campus) (map[string]Building, error) {
	cacheKey := campus.Name

	c.mu.RLock()
	buildings, ok := c.buildings[cacheKey]
	c.mu.RUnlock()
	if ok {
		return buildings, nil
	}

	loadedBuildings, err := withCampusCardAuthRetry(c, ctx, func() ([]Building, error) {
		return c.campusCard.GetBuildings(ctx, campus)
	})
	if err != nil {
		return nil, err
	}
	loaded := make(map[string]Building, len(loadedBuildings)*2)
	for _, building := range loadedBuildings {
		loaded[building.Name] = building
		if building.ID != "" {
			loaded[building.ID] = building
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if buildings, ok := c.buildings[cacheKey]; ok {
		return buildings, nil
	}
	c.buildings[cacheKey] = loaded
	return loaded, nil
}

func (c *ElectricityClient) roomsForBuilding(ctx context.Context, building Building) (map[string]Room, error) {
	cacheKey := buildingRoomCacheKey(building)

	c.mu.RLock()
	rooms, ok := c.rooms[cacheKey]
	c.mu.RUnlock()
	if ok {
		return rooms, nil
	}

	loadedRooms, err := withCampusCardAuthRetry(c, ctx, func() ([]Room, error) {
		return c.campusCard.GetRooms(ctx, building)
	})
	if err != nil {
		return nil, err
	}
	loaded := make(map[string]Room, len(loadedRooms)*2)
	for _, room := range loadedRooms {
		loaded[room.Name] = room
		if room.ID != "" {
			loaded[room.ID] = room
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if rooms, ok := c.rooms[cacheKey]; ok {
		return rooms, nil
	}
	c.rooms[cacheKey] = loaded
	return loaded, nil
}

func findCampus(name string) (Campus, error) {
	for _, campus := range Campuses() {
		if campus.Name == name || campus.DisplayName == name {
			return campus, nil
		}
	}
	return Campus{}, fmt.Errorf("%w: %s", ErrCampusNotFound, name)
}

func buildingRoomCacheKey(building Building) string {
	return building.Campus.CampusID + "|" + building.ID
}

func withCampusCardAuthRetry[T any](c *ElectricityClient, ctx context.Context, operation func() (T, error)) (T, error) {
	result, err := operation()
	if !errors.Is(err, ErrCampusCardNotLoggedIn) {
		return result, err
	}

	var zero T
	if loginErr := c.restoreCampusCardLogin(ctx); loginErr != nil {
		return zero, loginErr
	}
	return operation()
}

func (c *ElectricityClient) restoreCampusCardLogin(ctx context.Context) error {
	c.reloginMu.Lock()
	defer c.reloginMu.Unlock()

	if c.campusCard.IsLoggedIn(ctx) {
		return nil
	}
	if c.sso == nil {
		return fmt.Errorf("校园卡自动登录未配置")
	}

	if !c.sso.IsLoggedIn(ctx) {
		if err := c.loginSSO(ctx); err != nil {
			return err
		}
	}

	if err := c.loginCampusCard(ctx); err != nil {
		if !errors.Is(err, ErrSSONotLoggedIn) {
			return err
		}
		if err := c.loginSSO(ctx); err != nil {
			return err
		}
		return c.loginCampusCard(ctx)
	}
	return nil
}

func (c *ElectricityClient) loginSSO(ctx context.Context) error {
	if c.sso == nil {
		return fmt.Errorf("统一身份认证自动登录未配置")
	}

	form, err := c.sso.GetLoginForm(ctx)
	if err != nil {
		return fmt.Errorf("获取 SSO 登录表单失败: %w", err)
	}
	if err := c.sso.Login(ctx, form, c.username, c.password, ""); err != nil {
		return fmt.Errorf("登录 SSO 失败: %w", err)
	}
	return nil
}

func (c *ElectricityClient) loginCampusCard(ctx context.Context) error {
	if c.sso == nil {
		return fmt.Errorf("校园卡自动登录未配置")
	}

	ticket, err := c.sso.LoginToCampusCard(ctx)
	if err != nil {
		return fmt.Errorf("从 SSO 登录 CampusCard 失败: %w", err)
	}
	if err := c.campusCard.SyncToken(ctx, ticket); err != nil {
		return fmt.Errorf("同步 CampusCard token 失败: %w", err)
	}
	return nil
}
