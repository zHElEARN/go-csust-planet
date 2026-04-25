package campuscard

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const initRetryDelay = 2 * time.Second

type buildingStore struct {
	mu        sync.RWMutex
	buildings map[string]map[string]Building
}

var (
	storeInstance     = &buildingStore{}
	initOnce          sync.Once
	supportedCampuses = map[string]Campus{
		"云塘":  CampusYuntang,
		"金盆岭": CampusJinpenling,
	}
)

func InitBuildingStoreBlocking() {
	initOnce.Do(func() {
		for {
			if err := storeInstance.loadAll(); err != nil {
				log.Printf("[ERROR] campuscard 楼栋缓存初始化失败，将在 %s 后重试: %v", initRetryDelay, err)
				time.Sleep(initRetryDelay)
				continue
			}

			log.Println("[INFO] campuscard 楼栋缓存初始化完成")
			return
		}
	})
}

func GetBuildingByCampusName(campusName, buildingName string) (Building, error) {
	storeInstance.mu.RLock()
	defer storeInstance.mu.RUnlock()

	campusMap, ok := storeInstance.buildings[campusName]
	if !ok {
		return Building{}, fmt.Errorf("未知的校区: %s", campusName)
	}

	building, ok := campusMap[buildingName]
	if !ok {
		return Building{}, fmt.Errorf("在 %s 未找到楼栋: %s", campusName, buildingName)
	}

	return building, nil
}

func (s *buildingStore) loadAll() error {
	loaded := make(map[string]map[string]Building, len(supportedCampuses))

	for campusName, campus := range supportedCampuses {
		buildings, err := GetBuildings(campus)
		if err != nil {
			return fmt.Errorf("加载[%s]校区楼栋失败: %w", campusName, err)
		}

		campusMap := make(map[string]Building, len(buildings))
		for _, building := range buildings {
			campusMap[building.Name] = building
		}
		loaded[campusName] = campusMap

		log.Printf("[INFO] [%s]校区楼栋加载完成，共计 %d 栋", campusName, len(buildings))
	}

	s.mu.Lock()
	s.buildings = loaded
	s.mu.Unlock()

	return nil
}
