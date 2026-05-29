package csustkit

import (
	"context"
	"errors"
	"testing"
)

type fakeCampusCardElectricityAPI struct {
	buildingsCalls   int
	roomsCalls       int
	electricityCalls int
	buildings        []Building
	rooms            map[string][]Room
}

func (f *fakeCampusCardElectricityAPI) GetBuildings(ctx context.Context, campus Campus) ([]Building, error) {
	f.buildingsCalls++
	return f.buildings, nil
}

func (f *fakeCampusCardElectricityAPI) GetRooms(ctx context.Context, building Building) ([]Room, error) {
	f.roomsCalls++
	return f.rooms[building.ID], nil
}

func (f *fakeCampusCardElectricityAPI) GetElectricity(ctx context.Context, room Room) (float64, error) {
	f.electricityCalls++
	return 12.5, nil
}

func TestElectricityClientCachesBuildingsAndRooms(t *testing.T) {
	building := Building{Name: "至诚轩1栋", ID: "building-1", Campus: CampusYuntang}
	room := Room{Name: "101", ID: "room-101", Building: building}
	api := &fakeCampusCardElectricityAPI{
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
