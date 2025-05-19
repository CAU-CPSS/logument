package exp

import (
	"math/rand"
	"strings"
)

// 교통 체증 시나리오 설정
func setupTrafficJamScenario(vehicle *VehicleData) {
	// 기본 설정
	setupUrbanTrafficScenario(vehicle)

	// 교통 체증 특화 설정 일부 오버라이드
	// 속도가 낮고 브레이크 사용이 많음
	for path, sensor := range vehicle.SensorsHighFreq {
		if strings.Contains(path, "Speed") {
			sensor.Value = rand.Float64() * 20.0
			sensor.ChangePattern = "random_walk"
			sensor.ChangeParams = map[string]float64{
				"step_size": 0.5 + rand.Float64()*1.5,
				"min":       0.0,
				"max":       30.0,
			}
		}
	}

	// 브레이크와 액셀 패턴 (잦은 변경)
	for path, actuator := range vehicle.ActuatorsHighVar {
		if strings.Contains(path, "Brake") {
			actuator.ChangePattern = "sinusoidal"
			actuator.ChangeParams = map[string]float64{
				"amplitude": 40.0 + rand.Float64()*30.0,
				"period":    5000.0 + rand.Float64()*3000.0,
				"baseline":  50.0,
			}
		} else if strings.Contains(path, "ThrottlePosition") {
			actuator.ChangePattern = "sinusoidal"
			actuator.ChangeParams = map[string]float64{
				"amplitude": 30.0 + rand.Float64()*20.0,
				"period":    5000.0 + rand.Float64()*3000.0,
				"baseline":  30.0,
				"phase":     2500.0, // 브레이크와 반대 위상
			}
		}
	}
}
