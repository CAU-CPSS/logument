package exp

import (
	"math/rand"
	"strings"
)

// 주차 조작 시나리오 설정
func setupParkingManeuverScenario(vehicle *VehicleData) {
	// 기본 설정
	setupUrbanTrafficScenario(vehicle)

	// 주차 특화 설정 일부 오버라이드
	// 속도가 낮고 스티어링 변화가 큼
	for path, sensor := range vehicle.SensorsHighFreq {
		if strings.Contains(path, "Speed") {
			sensor.Value = rand.Float64() * 10.0
			sensor.ChangePattern = "constant_with_noise"
			sensor.ChangeParams = map[string]float64{
				"baseline": 5.0 + rand.Float64()*5.0,
				"noise":    0.5 + rand.Float64()*1.0,
			}
		} else if strings.Contains(path, "SteeringWheel") {
			sensor.ChangePattern = "sinusoidal"
			sensor.ChangeParams = map[string]float64{
				"amplitude": 30.0 + rand.Float64()*20.0,
				"period":    2000.0 + rand.Float64()*1000.0,
				"baseline":  0.0,
			}
		}
	}

	// 높은 변동성 액추에이터 (브레이크, 기어 등)도 더 자주 변경
	for path, actuator := range vehicle.ActuatorsHighVar {
		if strings.Contains(path, "Brake") {
			actuator.ChangePattern = "random_walk"
			actuator.ChangeParams = map[string]float64{
				"step_size": 3.0 + rand.Float64()*7.0,
				"min":       0.0,
				"max":       100.0,
			}
		} else if strings.Contains(path, "Transmission") {
			actuator.ChangePattern = "random_selection"
			actuator.ChangeParams = map[string]float64{
				"num_choices": 3.0, // P, R, D만 사용
			}
		}
	}
}

// 긴급 제동 시나리오 설정
func setupEmergencyBrakingScenario(vehicle *VehicleData) {
	// 기본 설정
	setupUrbanTrafficScenario(vehicle)

	// 긴급 제동 특화 설정
	// 속도가 급격히 감소하고 브레이크가 급격히 증가
	for path, sensor := range vehicle.SensorsHighFreq {
		if strings.Contains(path, "Speed") {
			sensor.Value = 50.0 + rand.Float64()*30.0
			sensor.ChangePattern = "linear"
			sensor.ChangeParams = map[string]float64{
				"slope": -0.01 - rand.Float64()*0.02,
				"min":   0.0,
				"max":   80.0,
			}
		} else if strings.Contains(path, "Acceleration") {
			sensor.Value = 0.0
			sensor.ChangePattern = "triggered"
			sensor.ChangeParams = map[string]float64{
				"threshold":      0.8,
				"reset_interval": 15000.0,
				"min_value":      -30.0,
				"max_value":      5.0,
			}
		}
	}

	// 브레이크 패턴 (급격한 증가)
	for path, actuator := range vehicle.ActuatorsHighVar {
		if strings.Contains(path, "Brake") {
			actuator.Value = 20.0 + rand.Float64()*10.0
			actuator.ChangePattern = "linear"
			actuator.ChangeParams = map[string]float64{
				"slope": 0.05 + rand.Float64()*0.1,
				"min":   0.0,
				"max":   100.0,
			}
		} else if strings.Contains(path, "ThrottlePosition") {
			actuator.Value = 30.0 + rand.Float64()*20.0
			actuator.ChangePattern = "linear"
			actuator.ChangeParams = map[string]float64{
				"slope": -0.05 - rand.Float64()*0.1,
				"min":   0.0,
				"max":   100.0,
			}
		}
	}
}
