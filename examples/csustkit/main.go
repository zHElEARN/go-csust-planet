package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/zHElEARN/go-csust-planet/utils/csustkit"
)

func main() {
	campusName := flag.String("campus", "云塘", "校区名称：云塘或金盆岭")
	buildingName := flag.String("building", "", "楼栋名称")
	roomName := flag.String("room", "", "房间名称或ID")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Printf("[WARN] 未找到 .env 文件，将尝试直接使用系统环境变量")
	}

	username := os.Getenv("CSUST_AUTHSERVER_USERNAME")
	password := os.Getenv("CSUST_AUTHSERVER_PASSWORD")
	if username == "" || password == "" {
		log.Fatal("缺少 CSUST_AUTHSERVER_USERNAME 或 CSUST_AUTHSERVER_PASSWORD")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := csustkit.NewClient()
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	fmt.Println("获取 SSO 登录表单...")
	form, err := client.SSO().GetLoginForm(ctx)
	if err != nil {
		log.Fatalf("获取 SSO 登录表单失败: %v", err)
	}

	fmt.Println("登录 SSO...")
	if err := client.SSO().Login(ctx, form, username, password, ""); err != nil {
		log.Fatalf("登录 SSO 失败: %v", err)
	}

	fmt.Println("登录 CampusCard...")
	ticket, err := client.SSO().LoginToCampusCard(ctx)
	if err != nil {
		log.Fatalf("从 SSO 登录 CampusCard 失败: %v", err)
	}
	if err := client.CampusCard().SyncToken(ctx, ticket); err != nil {
		log.Fatalf("同步 CampusCard token 失败: %v", err)
	}

	campus, err := selectCampus(*campusName)
	if err != nil {
		log.Fatal(err)
	}

	buildings, err := client.CampusCard().GetBuildings(ctx, campus)
	if err != nil {
		log.Fatalf("获取楼栋列表失败: %v", err)
	}
	fmt.Printf("%s楼栋数量: %d\n", campus.DisplayName, len(buildings))
	building, err := selectBuilding(buildings, *buildingName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("使用楼栋: %s (%s)\n", building.Name, building.ID)

	rooms, err := client.CampusCard().GetRooms(ctx, building)
	if err != nil {
		log.Fatalf("获取房间列表失败: %v", err)
	}
	fmt.Printf("%s房间数量: %d\n", building.Name, len(rooms))
	room, err := selectRoom(rooms, *roomName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("使用房间: %s (%s)\n", room.Name, room.ID)

	electricity, err := client.CampusCard().GetElectricity(ctx, room)
	if err != nil {
		log.Fatalf("获取电量失败: %v", err)
	}
	fmt.Printf("当前剩余电量: %.2f\n", electricity)
}

func selectCampus(name string) (csustkit.Campus, error) {
	for _, campus := range csustkit.Campuses() {
		if campus.Name == name || campus.DisplayName == name {
			return campus, nil
		}
	}
	return csustkit.Campus{}, fmt.Errorf("未知校区: %s", name)
}

func selectBuilding(buildings []csustkit.Building, name string) (csustkit.Building, error) {
	if len(buildings) == 0 {
		return csustkit.Building{}, fmt.Errorf("楼栋列表为空")
	}
	if name == "" {
		return buildings[0], nil
	}
	for _, building := range buildings {
		if building.Name == name || building.ID == name {
			return building, nil
		}
	}
	return csustkit.Building{}, fmt.Errorf("未找到楼栋: %s", name)
}

func selectRoom(rooms []csustkit.Room, name string) (csustkit.Room, error) {
	if len(rooms) == 0 {
		return csustkit.Room{}, fmt.Errorf("房间列表为空")
	}
	if name == "" {
		return rooms[0], nil
	}
	for _, room := range rooms {
		if room.Name == name || room.ID == name {
			return room, nil
		}
	}
	return csustkit.Room{}, fmt.Errorf("未找到房间: %s", name)
}
