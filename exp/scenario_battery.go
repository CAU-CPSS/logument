package exp

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// setupBatteryChargingScenario 함수는 배터리 충전 시나리오를 설정합니다.
func setupBatteryChargingScenario(vehicle *VehicleData) {
	// 기본 설정
	setupUrbanTrafficScenario(vehicle)

	// 배터리 충전 특화 설정
	// 차량은 정지해 있고 배터리 충전 관련 센서가 중요합니다

	// 속도와 가속도 센서 (0에 가까움)
	for path, sensor := range vehicle.SensorsHighFreq {
		if strings.Contains(path, "Speed") {
			sensor.Value = 0.0
			sensor.Type = "number"
			sensor.ChangePattern = "constant_with_noise"
			sensor.ChangeParams = map[string]float64{
				"baseline": 0.0,
				"noise":    0.05,
			}
		} else if strings.Contains(path, "Acceleration") {
			sensor.Value = 0.0
			sensor.Type = "number"
			sensor.ChangePattern = "constant_with_noise"
			sensor.ChangeParams = map[string]float64{
				"baseline": 0.0,
				"noise":    0.1,
			}
		} else if strings.Contains(path, "Steering") {
			sensor.Value = 0.0
			sensor.Type = "number"
			sensor.ChangePattern = "constant"
			sensor.ChangeParams = map[string]float64{}
		}
	}

	// 배터리 관련 센서 (점진적 증가)
	// 여러 개의 배터리 관련 센서 추가
	batteryPaths := []string{
		"/Vehicle/Powertrain/Battery/StateOfCharge",
		"/Vehicle/Powertrain/Battery/ChargingStatus",
		"/Vehicle/Powertrain/Battery/ChargingCurrent",
		"/Vehicle/Powertrain/Battery/ChargingVoltage",
		"/Vehicle/Powertrain/Battery/Temperature",
		"/Vehicle/Powertrain/Battery/EstimatedRange",
		"/Vehicle/Powertrain/Battery/TimeToFullCharge",
		"/Vehicle/Powertrain/Battery/ChargingPower",
		"/Vehicle/Powertrain/Battery/InstantPower",
		"/Vehicle/Powertrain/Battery/CellVoltage",
	}

	for i, path := range batteryPaths {
		updateInterval := 0
		changePattern := ""
		params := make(map[string]float64)
		var value interface{}
		valueType := "number"

		// 경로에 따라 다른 설정
		switch {
		case strings.Contains(path, "StateOfCharge"):
			// 배터리 충전 상태 (%)
			updateInterval = 1000
			changePattern = "linear"
			params["slope"] = 0.01 + rand.Float64()*0.02 // 천천히 증가
			params["min"] = 0.0
			params["max"] = 100.0
			value = 20.0 + rand.Float64()*30.0 // 시작값: 20~50%

		case strings.Contains(path, "ChargingStatus"):
			// 충전 상태 (문자열: "charging", "full", "error")
			updateInterval = 10000
			changePattern = "constant"
			value = "charging" // 문자열 값을 확실히 설정
			valueType = "string"

		case strings.Contains(path, "ChargingCurrent"):
			// 충전 전류
			updateInterval = 500
			changePattern = "sinusoidal"
			params["amplitude"] = 0.5
			params["period"] = 10000.0
			params["baseline"] = 30.0 + rand.Float64()*10.0
			value = 30.0 + rand.Float64()*10.0 // 30~40A

		case strings.Contains(path, "ChargingVoltage"):
			// 충전 전압
			updateInterval = 500
			changePattern = "constant_with_noise"
			params["baseline"] = 400.0
			params["noise"] = 2.0
			value = 400.0 + (rand.Float64()*4.0 - 2.0)

		case strings.Contains(path, "Temperature"):
			// 배터리 온도
			updateInterval = 2000
			changePattern = "linear"
			params["slope"] = 0.001 // 매우 천천히 증가
			params["min"] = 20.0
			params["max"] = 45.0
			value = 25.0 + rand.Float64()*5.0

		case strings.Contains(path, "EstimatedRange"):
			// 예상 주행 거리
			updateInterval = 5000
			changePattern = "linear"
			params["slope"] = 0.05 + rand.Float64()*0.05 // 점진적 증가
			params["min"] = 0.0
			params["max"] = 500.0
			value = 100.0 + rand.Float64()*50.0

		case strings.Contains(path, "TimeToFullCharge"):
			// 완충까지 남은 시간
			updateInterval = 30000
			changePattern = "linear"
			params["slope"] = -0.01 // 점진적 감소
			params["min"] = 0.0
			params["max"] = 180.0
			value = 120.0 + rand.Float64()*60.0 // 2~3시간

		case strings.Contains(path, "ChargingPower"):
			// 충전 전력
			updateInterval = 1000
			changePattern = "sinusoidal"
			params["amplitude"] = 0.5
			params["period"] = 15000.0
			params["baseline"] = 11.0
			value = 11.0 + (rand.Float64()*1.0 - 0.5) // 약 11kW

		case strings.Contains(path, "InstantPower"):
			// 순간 전력
			updateInterval = 200
			changePattern = "random_walk"
			params["step_size"] = 0.2
			params["min"] = 10.0
			params["max"] = 12.0
			value = 11.0 + (rand.Float64()*0.4 - 0.2)

		case strings.Contains(path, "CellVoltage"):
			// 셀 전압
			updateInterval = 5000
			changePattern = "linear"
			params["slope"] = 0.0001 // 매우 천천히 증가
			params["min"] = 3.6
			params["max"] = 4.2
			value = 3.8 + rand.Float64()*0.2
		}

		// 센서 그룹 결정 (업데이트 빈도에 따라)
		var sensorGroup map[string]*SensorData
		if updateInterval < 100 {
			sensorGroup = vehicle.SensorsHighFreq
		} else if updateInterval < 1000 {
			sensorGroup = vehicle.SensorsMedFreq
		} else {
			sensorGroup = vehicle.SensorsLowFreq
		}

		// 경로에 인덱스 추가
		fullPath := fmt.Sprintf("%s/%d", path, i/10)

		// 센서 데이터 생성
		sensorGroup[fullPath] = &SensorData{
			Path:           fullPath,
			Value:          value,
			Type:           valueType,
			Timestamp:      time.Now().UnixNano(),
			UpdateInterval: updateInterval,
			ChangePattern:  changePattern,
			ChangeParams:   params,
		}
	}

	// 충전 관련 액추에이터 (거의 변화 없음)
	for path, actuator := range vehicle.ActuatorsHighVar {
		if strings.Contains(path, "Throttle") || strings.Contains(path, "Brake") {
			actuator.Value = 0.0
			actuator.Type = "number"
			actuator.ChangePattern = "constant"
			actuator.ChangeParams = map[string]float64{}
		}
	}

	// 충전 포트 상태 액추에이터 추가
	vehicle.ActuatorsLowVar["/Vehicle/Powertrain/Battery/ChargingPortOpen"] = &SensorData{
		Path:           "/Vehicle/Powertrain/Battery/ChargingPortOpen",
		Value:          true,
		Type:           "boolean",
		Timestamp:      time.Now().UnixNano(),
		UpdateInterval: 30000, // 매우 드물게 업데이트
		ChangePattern:  "constant",
		ChangeParams:   map[string]float64{},
	}

	// 충전 모드 액추에이터 추가
	vehicle.ActuatorsLowVar["/Vehicle/Powertrain/Battery/ChargingMode"] = &SensorData{
		Path:           "/Vehicle/Powertrain/Battery/ChargingMode",
		Value:          "fast", // 문자열 값을 확실히 설정
		Type:           "string",
		Timestamp:      time.Now().UnixNano(),
		UpdateInterval: 60000, // 매우 드물게 업데이트
		ChangePattern:  "constant",
		ChangeParams:   map[string]float64{},
	}
}
