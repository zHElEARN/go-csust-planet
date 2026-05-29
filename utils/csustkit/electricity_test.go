package csustkit

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

type fakeCampusCardElectricityAPI struct {
	buildingsCalls    int
	roomsCalls        int
	electricityCalls  int
	syncTokenCalls    int
	loggedIn          bool
	buildings         []Building
	rooms             map[string][]Room
	buildingsErrors   []error
	roomsErrors       []error
	electricityErrors []error
}

func (f *fakeCampusCardElectricityAPI) GetBuildings(ctx context.Context, campus Campus) ([]Building, error) {
	f.buildingsCalls++
	if len(f.buildingsErrors) > 0 {
		err := f.buildingsErrors[0]
		f.buildingsErrors = f.buildingsErrors[1:]
		return nil, err
	}
	return f.buildings, nil
}

func (f *fakeCampusCardElectricityAPI) GetRooms(ctx context.Context, building Building) ([]Room, error) {
	f.roomsCalls++
	if len(f.roomsErrors) > 0 {
		err := f.roomsErrors[0]
		f.roomsErrors = f.roomsErrors[1:]
		return nil, err
	}
	return f.rooms[building.ID], nil
}

func (f *fakeCampusCardElectricityAPI) GetElectricity(ctx context.Context, room Room) (float64, error) {
	f.electricityCalls++
	if len(f.electricityErrors) > 0 {
		err := f.electricityErrors[0]
		f.electricityErrors = f.electricityErrors[1:]
		return 0, err
	}
	return 12.5, nil
}

func (f *fakeCampusCardElectricityAPI) SyncToken(ctx context.Context, ticket string) error {
	f.syncTokenCalls++
	f.loggedIn = true
	return nil
}

func (f *fakeCampusCardElectricityAPI) IsLoggedIn(ctx context.Context) bool {
	return f.loggedIn
}

type fakeSSOAPI struct {
	loggedIn                bool
	getLoginFormCalls       int
	loginCalls              int
	loginToCampusCardCalls  int
	loginToCampusCardErrors []error
}

func (f *fakeSSOAPI) GetLoginForm(ctx context.Context) (SSOLoginForm, error) {
	f.getLoginFormCalls++
	return SSOLoginForm{PwdEncryptSalt: "salt", Execution: "execution"}, nil
}

func (f *fakeSSOAPI) Login(ctx context.Context, form SSOLoginForm, username, password, captcha string) error {
	f.loginCalls++
	f.loggedIn = true
	return nil
}

func (f *fakeSSOAPI) LoginToCampusCard(ctx context.Context) (string, error) {
	f.loginToCampusCardCalls++
	if len(f.loginToCampusCardErrors) > 0 {
		err := f.loginToCampusCardErrors[0]
		f.loginToCampusCardErrors = f.loginToCampusCardErrors[1:]
		return "", err
	}
	return "ticket", nil
}

func (f *fakeSSOAPI) IsLoggedIn(ctx context.Context) bool {
	return f.loggedIn
}

func TestElectricityClientCachesBuildingsAndRooms(t *testing.T) {
	building := Building{Name: "至诚轩1栋", ID: "building-1", Campus: CampusYuntang}
	room := Room{Name: "101", ID: "room-101", Building: building}
	api := &fakeCampusCardElectricityAPI{
		loggedIn:  true,
		buildings: []Building{building},
		rooms: map[string][]Room{
			building.ID: {room},
		},
	}
	client := newElectricityClient(api)
	ctx := context.Background()

	if err := client.ValidateRoom(ctx, "云塘", "至诚轩1栋", "101"); err != nil {
		t.Fatalf("expected first validation to succeed: %v", err)
	}
	if err := client.ValidateRoom(ctx, "云塘", "至诚轩1栋", "room-101"); err != nil {
		t.Fatalf("expected room ID validation to succeed: %v", err)
	}
	electricity, err := client.GetElectricity(ctx, "云塘校区", "building-1", "101")
	if err != nil {
		t.Fatalf("expected electricity fetch to succeed: %v", err)
	}
	if electricity != 12.5 {
		t.Fatalf("expected electricity 12.5, got %v", electricity)
	}

	if api.buildingsCalls != 1 {
		t.Fatalf("expected buildings to load once, got %d", api.buildingsCalls)
	}
	if api.roomsCalls != 1 {
		t.Fatalf("expected rooms to load once, got %d", api.roomsCalls)
	}
	if api.electricityCalls != 1 {
		t.Fatalf("expected electricity to fetch once, got %d", api.electricityCalls)
	}
}

func TestElectricityClientReturnsNotFoundErrors(t *testing.T) {
	building := Building{Name: "至诚轩1栋", ID: "building-1", Campus: CampusYuntang}
	api := &fakeCampusCardElectricityAPI{
		loggedIn:  true,
		buildings: []Building{building},
		rooms: map[string][]Room{
			building.ID: {},
		},
	}
	client := newElectricityClient(api)
	ctx := context.Background()

	if err := client.ValidateRoom(ctx, "未知", "至诚轩1栋", "101"); !errors.Is(err, ErrCampusNotFound) {
		t.Fatalf("expected ErrCampusNotFound, got %v", err)
	}
	if err := client.ValidateRoom(ctx, "云塘", "未知楼栋", "101"); !errors.Is(err, ErrBuildingNotFound) {
		t.Fatalf("expected ErrBuildingNotFound, got %v", err)
	}
	if err := client.ValidateRoom(ctx, "云塘", "至诚轩1栋", "101"); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestElectricityClientReloginsAndRetriesCampusCardOperationOnce(t *testing.T) {
	building := Building{Name: "至诚轩1栋", ID: "building-1", Campus: CampusYuntang}
	room := Room{Name: "101", ID: "room-101", Building: building}
	campusCard := &fakeCampusCardElectricityAPI{
		buildings:       []Building{building},
		buildingsErrors: []error{ErrCampusCardNotLoggedIn},
		rooms: map[string][]Room{
			building.ID: {room},
		},
	}
	sso := &fakeSSOAPI{loggedIn: true}
	client := newElectricityClientWithAuth(sso, campusCard, "user", "password")

	if err := client.ValidateRoom(context.Background(), "云塘", "至诚轩1栋", "101"); err != nil {
		t.Fatalf("expected validation to succeed after relogin: %v", err)
	}
	if campusCard.buildingsCalls != 2 {
		t.Fatalf("expected original operation to run twice, got %d", campusCard.buildingsCalls)
	}
	if sso.loginCalls != 0 {
		t.Fatalf("expected existing SSO session to be reused, got %d SSO logins", sso.loginCalls)
	}
	if sso.loginToCampusCardCalls != 1 || campusCard.syncTokenCalls != 1 {
		t.Fatalf("expected one CampusCard relogin, got loginToCampusCard=%d syncToken=%d", sso.loginToCampusCardCalls, campusCard.syncTokenCalls)
	}
}

func TestElectricityClientDoesNotRetryAgainWhenOperationStillNotLoggedIn(t *testing.T) {
	building := Building{Name: "至诚轩1栋", ID: "building-1", Campus: CampusYuntang}
	campusCard := &fakeCampusCardElectricityAPI{
		buildings:       []Building{building},
		buildingsErrors: []error{ErrCampusCardNotLoggedIn, ErrCampusCardNotLoggedIn},
		rooms:           map[string][]Room{},
	}
	sso := &fakeSSOAPI{loggedIn: true}
	client := newElectricityClientWithAuth(sso, campusCard, "user", "password")

	err := client.ValidateRoom(context.Background(), "云塘", "至诚轩1栋", "101")
	if !errors.Is(err, ErrCampusCardNotLoggedIn) {
		t.Fatalf("expected ErrCampusCardNotLoggedIn, got %v", err)
	}
	if campusCard.buildingsCalls != 2 {
		t.Fatalf("expected exactly one retry, got %d calls", campusCard.buildingsCalls)
	}
	if sso.loginToCampusCardCalls != 1 || campusCard.syncTokenCalls != 1 {
		t.Fatalf("expected one relogin, got loginToCampusCard=%d syncToken=%d", sso.loginToCampusCardCalls, campusCard.syncTokenCalls)
	}
}

func TestElectricityClientReloginsSSOWhenNeeded(t *testing.T) {
	building := Building{Name: "至诚轩1栋", ID: "building-1", Campus: CampusYuntang}
	room := Room{Name: "101", ID: "room-101", Building: building}
	campusCard := &fakeCampusCardElectricityAPI{
		buildings:       []Building{building},
		buildingsErrors: []error{ErrCampusCardNotLoggedIn},
		rooms: map[string][]Room{
			building.ID: {room},
		},
	}
	sso := &fakeSSOAPI{}
	client := newElectricityClientWithAuth(sso, campusCard, "user", "password")

	if err := client.ValidateRoom(context.Background(), "云塘", "至诚轩1栋", "101"); err != nil {
		t.Fatalf("expected validation to succeed after SSO relogin: %v", err)
	}
	if sso.getLoginFormCalls != 1 || sso.loginCalls != 1 {
		t.Fatalf("expected one SSO relogin, got form=%d login=%d", sso.getLoginFormCalls, sso.loginCalls)
	}
	if sso.loginToCampusCardCalls != 1 || campusCard.syncTokenCalls != 1 {
		t.Fatalf("expected one CampusCard login, got loginToCampusCard=%d syncToken=%d", sso.loginToCampusCardCalls, campusCard.syncTokenCalls)
	}
}

func TestElectricityClientReloginsSSOWhenCampusCardLoginReportsSSOExpired(t *testing.T) {
	building := Building{Name: "至诚轩1栋", ID: "building-1", Campus: CampusYuntang}
	room := Room{Name: "101", ID: "room-101", Building: building}
	campusCard := &fakeCampusCardElectricityAPI{
		buildings:       []Building{building},
		buildingsErrors: []error{ErrCampusCardNotLoggedIn},
		rooms: map[string][]Room{
			building.ID: {room},
		},
	}
	sso := &fakeSSOAPI{
		loggedIn:                true,
		loginToCampusCardErrors: []error{ErrSSONotLoggedIn},
	}
	client := newElectricityClientWithAuth(sso, campusCard, "user", "password")

	if err := client.ValidateRoom(context.Background(), "云塘", "至诚轩1栋", "101"); err != nil {
		t.Fatalf("expected validation to succeed after SSO relogin: %v", err)
	}
	if sso.loginCalls != 1 {
		t.Fatalf("expected one SSO login after SSO expiry, got %d", sso.loginCalls)
	}
	if sso.loginToCampusCardCalls != 2 || campusCard.syncTokenCalls != 1 {
		t.Fatalf("expected CampusCard login to be retried after SSO login, got loginToCampusCard=%d syncToken=%d", sso.loginToCampusCardCalls, campusCard.syncTokenCalls)
	}
}

func TestElectricityClientDoesNotRetryNonLoginError(t *testing.T) {
	upstreamErr := errors.New("campus card unavailable")
	campusCard := &fakeCampusCardElectricityAPI{
		loggedIn:        true,
		buildingsErrors: []error{upstreamErr},
		rooms:           map[string][]Room{},
	}
	sso := &fakeSSOAPI{loggedIn: true}
	client := newElectricityClientWithAuth(sso, campusCard, "user", "password")

	err := client.ValidateRoom(context.Background(), "云塘", "至诚轩1栋", "101")
	if !errors.Is(err, upstreamErr) {
		t.Fatalf("expected upstream error, got %v", err)
	}
	if campusCard.buildingsCalls != 1 {
		t.Fatalf("expected no retry for non-login error, got %d calls", campusCard.buildingsCalls)
	}
	if sso.loginToCampusCardCalls != 0 || campusCard.syncTokenCalls != 0 {
		t.Fatalf("expected no relogin, got loginToCampusCard=%d syncToken=%d", sso.loginToCampusCardCalls, campusCard.syncTokenCalls)
	}
}

func TestRoundElectricity(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{name: "round down", in: 12.344, want: 12.34},
		{name: "round up", in: 12.345, want: 12.35},
		{name: "keep two decimals", in: 12.3, want: 12.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := roundElectricity(tt.in); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCampusCardNotLoggedInErrorsAreSentinel(t *testing.T) {
	_, err := (&CampusCardHelper{}).newChargeRequest(context.Background(), nil)
	if !errors.Is(err, ErrCampusCardNotLoggedIn) {
		t.Fatalf("expected empty token to return ErrCampusCardNotLoggedIn, got %v", err)
	}

	queryResp := baseQueryResponse[any]{Code: http.StatusUnauthorized}
	if err := queryResp.errIfUnauthorized(); !errors.Is(err, ErrCampusCardNotLoggedIn) {
		t.Fatalf("expected unauthorized response to return ErrCampusCardNotLoggedIn, got %v", err)
	}
}
