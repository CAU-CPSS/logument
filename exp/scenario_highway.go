package exp

import (
	"math/rand"
	"strings"
)

// 고속도로 주행 시나리오 설정
func setupHighwayCruisingScenario(vehicle *VehicleData) {
	// 기본 설정과 유사하지만, 속도가 더 높고 안정적이며 변경이 적음
	setupUrbanTrafficScenario(vehicle)

	// 고속도로 특화 설정 일부 오버라이드
	// 고속도로에서는 속도가 높고 안정적, 스티어링 변화 적음
	for path, sensor := range vehicle.SensorsHighFreq {
		if strings.Contains(path, "Speed") {
			sensor.Value = 80.0 + rand.Float64()*40.0
			sensor.ChangePattern = "constant_with_noise"
			sensor.ChangeParams = map[string]float64{
				"baseline": 100.0 + rand.Float64()*20.0,
				"noise":    0.2 + rand.Float64()*1.0,
			}
		} else if strings.Contains(path, "SteeringWheel") {
			sensor.ChangePattern = "constant_with_noise"
			sensor.ChangeParams = map[string]float64{
				"baseline": 0.0,
				"noise":    0.5 + rand.Float64()*1.0,
			}
		}
	}
}
