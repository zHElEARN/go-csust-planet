package campuscard

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

const BaseURL = "http://yktwd.csust.edu.cn:8988/web/Common/Tsm.html"

type Campus struct {
	ID          string
	DisplayName string
}

var (
	CampusYuntang    = Campus{ID: "0030000000002501", DisplayName: "云塘校区"}
	CampusJinpenling = Campus{ID: "0030000000002502", DisplayName: "金盆岭校区"}
)

type Building struct {
	ID     string
	Name   string
	Campus Campus
}

type queryElecBuildingReq struct {
	Aid         string    `json:"aid"`
	Account     string    `json:"account"`
	Area        area      `json:"area"`
	RetCode     *string   `json:"retcode,omitempty"`
	ErrMsg      *string   `json:"errmsg,omitempty"`
	BuildingTab *[]string `json:"buildingtab,omitempty"`
}

type area struct {
	Area     string `json:"area"`
	AreaName string `json:"areaname"`
}

type queryElecRoomReq struct {
	Aid      string `json:"aid"`
	Account  string `json:"account"`
	Room     room   `json:"room"`
	Floor    floor  `json:"floor"`
	Area     area   `json:"area"`
	Building bld    `json:"building"`
}

type room struct {
	RoomID string `json:"roomid"`
	Room   string `json:"room"`
}
type floor struct {
	FloorID string `json:"floorid"`
	Floor   string `json:"floor"`
}
type bld struct {
	BuildingID string `json:"buildingid"`
	Building   string `json:"building"`
}

type buildingResp struct {
	QueryElecBuilding struct {
		BuildingTab []struct {
			BuildingID string `json:"buildingid"`
			Building   string `json:"building"`
		} `json:"buildingtab"`
	} `json:"query_elec_building"`
}

type roomResp struct {
	QueryElecRoomInfo struct {
		ErrMsg string `json:"errmsg"`
	} `json:"query_elec_roominfo"`
}

func GetBuildings(c Campus) ([]Building, error) {
	req := map[string]any{
		"query_elec_building": queryElecBuildingReq{
			Aid:     c.ID,
			Account: "000001",
			Area:    area{Area: c.DisplayName, AreaName: c.DisplayName},
		},
	}

	return request(req, "synjones.onecard.query.elec.building", func(resp buildingResp) ([]Building, error) {
		var results []Building
		for _, b := range resp.QueryElecBuilding.BuildingTab {
			results = append(results, Building{ID: b.BuildingID, Name: b.Building, Campus: c})
		}
		if len(results) == 0 {
			return nil, fmt.Errorf("楼栋列表为空")
		}
		return results, nil
	})
}

func GetElectricity(b Building, roomNum string) (float64, error) {
	req := map[string]any{
		"query_elec_roominfo": queryElecRoomReq{
			Aid:      b.Campus.ID,
			Account:  "000001",
			Room:     room{RoomID: roomNum, Room: roomNum},
			Area:     area{Area: b.Campus.DisplayName, AreaName: b.Campus.DisplayName},
			Building: bld{BuildingID: b.ID, Building: ""},
		},
	}

	return request(req, "synjones.onecard.query.elec.roominfo", func(resp roomResp) (float64, error) {
		re := regexp.MustCompile(`(\d+(\.\d+)?)`)
		match := re.FindString(resp.QueryElecRoomInfo.ErrMsg)
		if match == "" {
			return 0, fmt.Errorf("未能解析电费数值: %s", resp.QueryElecRoomInfo.ErrMsg)
		}
		return strconv.ParseFloat(match, 64)
	})
}

// 通用请求逻辑
func request[T any, R any](payload any, funname string, parser func(T) (R, error)) (R, error) {
	var zero R
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return zero, err
	}

	vals := url.Values{}
	vals.Set("jsondata", string(jsonData))
	vals.Set("funname", funname)
	vals.Set("json", "true")

	resp, err := http.PostForm(BaseURL, vals)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data T
	if err := json.Unmarshal(body, &data); err != nil {
		return zero, err
	}

	return parser(data)
}
