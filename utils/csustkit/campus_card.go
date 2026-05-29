package csustkit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	campusCardOAuthAuthorization  = "Basic bW9iaWxlX3NlcnZpY2VfcGxhdGZvcm06bW9iaWxlX3NlcnZpY2VfcGxhdGZvcm1fc2VjcmV0"
	campusCardChargeAuthorization = "Y2hhcmdlOmNoYXJnZV9zZWNyZXQ="
)

type CampusCardHelper struct {
	client *Client
	token  string
}

type Campus struct {
	Name        string
	DisplayName string
	FeeItemID   string
	CampusID    string
}

var (
	CampusYuntang    = Campus{Name: "云塘", DisplayName: "云塘校区", FeeItemID: "448", CampusID: "1"}
	CampusJinpenling = Campus{Name: "金盆岭", DisplayName: "金盆岭校区", FeeItemID: "468", CampusID: "22"}
)

type Building struct {
	Name   string
	ID     string
	Campus Campus
}

type Room struct {
	Name     string
	ID       string
	Building Building
}

func Campuses() []Campus {
	return []Campus{CampusYuntang, CampusJinpenling}
}

func (h *CampusCardHelper) SyncToken(ctx context.Context, ticket string) error {
	values := url.Values{}
	values.Set("username", ticket)
	values.Set("password", ticket)
	values.Set("grant_type", "password")
	values.Set("scope", "all")
	values.Set("loginFrom", "h5")
	values.Set("logintype", "sso")
	values.Set("device_token", "h5")
	values.Set("synAccessSource", "h5")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.client.makeURL(ServiceCampusCard, "/berserker-auth/oauth/token"), strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", campusCardOAuthAuthorization)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var tokenResp tokenResponse
	if err := h.doJSON(req, &tokenResp); err != nil {
		return fmt.Errorf("同步校园卡令牌失败: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return fmt.Errorf("同步校园卡令牌失败: 响应缺少access_token")
	}

	h.token = tokenResp.AccessToken
	return nil
}

func (h *CampusCardHelper) GetBuildings(ctx context.Context, campus Campus) ([]Building, error) {
	req, err := h.newChargeRequest(ctx, map[string]string{
		"feeitemid": campus.FeeItemID,
		"type":      "select",
		"level":     "1",
		"xiaoqu_id": campus.CampusID,
	})
	if err != nil {
		return nil, err
	}

	var queryResp baseQueryResponse[[]nameValueItem]
	if err := h.doJSON(req, &queryResp); err != nil {
		return nil, fmt.Errorf("获取楼栋列表失败: %w", err)
	}
	if err := queryResp.errIfUnauthorized(); err != nil {
		return nil, err
	}
	if queryResp.Map == nil {
		return nil, fmt.Errorf("获取楼栋列表失败: %s", queryResp.errorMessage())
	}

	buildings := make([]Building, 0, len(queryResp.Map.Data))
	for _, item := range queryResp.Map.Data {
		buildings = append(buildings, Building{Name: item.Name, ID: item.Value, Campus: campus})
	}
	return buildings, nil
}

func (h *CampusCardHelper) GetRooms(ctx context.Context, building Building) ([]Room, error) {
	req, err := h.newChargeRequest(ctx, map[string]string{
		"feeitemid":  building.Campus.FeeItemID,
		"type":       "select",
		"level":      "2",
		"xiaoqu_id":  building.Campus.CampusID,
		"loudong_id": building.ID,
	})
	if err != nil {
		return nil, err
	}

	var queryResp baseQueryResponse[[]nameValueItem]
	if err := h.doJSON(req, &queryResp); err != nil {
		return nil, fmt.Errorf("获取宿舍列表失败: %w", err)
	}
	if err := queryResp.errIfUnauthorized(); err != nil {
		return nil, err
	}
	if queryResp.Map == nil {
		return nil, fmt.Errorf("获取宿舍列表失败: %s", queryResp.errorMessage())
	}

	rooms := make([]Room, 0, len(queryResp.Map.Data))
	for _, item := range queryResp.Map.Data {
		rooms = append(rooms, Room{Name: item.Name, ID: item.Value, Building: building})
	}
	return rooms, nil
}

func (h *CampusCardHelper) GetElectricity(ctx context.Context, room Room) (float64, error) {
	req, err := h.newChargeRequest(ctx, map[string]string{
		"feeitemid":  room.Building.Campus.FeeItemID,
		"type":       "IEC",
		"level":      "3",
		"xiaoqu_id":  room.Building.Campus.CampusID,
		"loudong_id": room.Building.ID,
		"room_id":    room.ID,
	})
	if err != nil {
		return 0, err
	}

	var queryResp baseQueryResponse[roomPowerInfo]
	if err := h.doJSON(req, &queryResp); err != nil {
		return 0, fmt.Errorf("获取宿舍电量失败: %w", err)
	}
	if err := queryResp.errIfUnauthorized(); err != nil {
		return 0, err
	}
	if queryResp.Map == nil {
		return 0, fmt.Errorf("获取宿舍电量失败: %s", queryResp.errorMessage())
	}

	allValue, err := strconv.ParseFloat(queryResp.Map.Data.AllAmp, 64)
	if err != nil {
		return 0, fmt.Errorf("解析总电量失败: %w", err)
	}
	usedValue, err := strconv.ParseFloat(queryResp.Map.Data.UsedAmp, 64)
	if err != nil {
		return 0, fmt.Errorf("解析已用电量失败: %w", err)
	}
	return roundElectricity(allValue - usedValue), nil
}

func roundElectricity(value float64) float64 {
	return math.Round(value*100) / 100
}

func (h *CampusCardHelper) newChargeRequest(ctx context.Context, parameters map[string]string) (*http.Request, error) {
	if h.token == "" {
		return nil, fmt.Errorf("校园卡系统未登录")
	}

	values := url.Values{}
	for key, value := range parameters {
		values.Set(key, value)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.client.makeURL(ServiceCampusCard, "/charge/feeitem/getThirdData"), strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", campusCardChargeAuthorization)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("synjones-auth", "bearer "+h.token)
	return req, nil
}

func (h *CampusCardHelper) doJSON(req *http.Request, target any) error {
	resp, err := h.client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.Unmarshal(body, target); err != nil {
		return err
	}
	return nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

type baseQueryResponse[T any] struct {
	Message string `json:"message"`
	Msg     string `json:"msg"`
	Code    int    `json:"code"`
	Map     *struct {
		Data T `json:"data"`
	} `json:"map"`
}

func (r baseQueryResponse[T]) errIfUnauthorized() error {
	if r.Code == http.StatusUnauthorized {
		return fmt.Errorf("校园卡系统未登录")
	}
	return nil
}

func (r baseQueryResponse[T]) errorMessage() string {
	if r.Message != "" {
		return r.Message
	}
	if r.Msg != "" {
		return r.Msg
	}
	return "nil"
}

type nameValueItem struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type roomPowerInfo struct {
	RoomID     string `json:"room_id"`
	AllAmp     string `json:"allAmp"`
	CampusID   string `json:"xiaoqu_id"`
	UsedAmp    string `json:"usedAmp"`
	BuildingID string `json:"loudong_id"`
}
