package exp

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// 도심 교통 시나리오 설정
func setupUrbanTrafficScenario(vehicle *VehicleData) {
	// 고빈도 센서 (10-50ms 업데이트)
	for i := 0; i < numHighFreqSensors; i++ {
		path := ""
		switch i % 5 {
		case 0:
			path = fmt.Sprintf("/Vehicle/Powertrain/Acceleration/Longitudinal/%d", i/5)
		case 1:
			path = fmt.Sprintf("/Vehicle/Chassis/Speed/%d", i/5)
		case 2:
			path = fmt.Sprintf("/Vehicle/Chassis/SteeringWheel/Angle/%d", i/5)
		case 3:
			path = fmt.Sprintf("/Vehicle/Cabin/Infotainment/Navigation/CurrentLocation/Latitude/%d", i/5)
		case 4:
			path = fmt.Sprintf("/Vehicle/Cabin/Infotainment/Navigation/CurrentLocation/Longitude/%d", i/5)
		}

		// 업데이트 간격을 10-50ms 사이로 설정
		interval := 10 + rand.Intn(41)

		// 패턴 선택
		pattern := ""
		params := make(map[string]float64)
		switch i % 4 {
		case 0:
			pattern = "sinusoidal"
			params["amplitude"] = 5.0 + rand.Float64()*10.0
			params["period"] = 5000.0 + rand.Float64()*10000.0
			params["baseline"] = 30.0 + rand.Float64()*20.0
		case 1:
			pattern = "random_walk"
			params["step_size"] = 0.1 + rand.Float64()*0.5
			params["min"] = 0.0
			params["max"] = 100.0
		case 2:
			pattern = "triggered"
			params["threshold"] = 0.2
			params["reset_interval"] = 2000.0 + rand.Float64()*5000.0
		case 3:
			pattern = "constant_with_noise"
			params["baseline"] = 50.0 + rand.Float64()*20.0
			params["noise"] = 0.5 + rand.Float64()*2.0
		}

		value := rand.Float64() * 100.0

		vehicle.SensorsHighFreq[path] = &SensorData{
			Path:           path,
			Value:          value,
			Type:           "number",
			Timestamp:      time.Now().UnixNano(),
			UpdateInterval: interval,
			ChangePattern:  pattern,
			ChangeParams:   params,
		}
	}

	// 중간 빈도 센서 (100-500ms 업데이트)
	for i := 0; i < numMedFreqSensors; i++ {
		path := ""
		switch i % 5 {
		case 0:
			path = fmt.Sprintf("/Vehicle/Powertrain/Engine/Speed/%d", i/5)
		case 1:
			path = fmt.Sprintf("/Vehicle/Body/Lights/Indicator/Left/IsActive/%d", i/5)
		case 2:
			path = fmt.Sprintf("/Vehicle/Body/Lights/Indicator/Right/IsActive/%d", i/5)
		case 3:
			path = fmt.Sprintf("/Vehicle/Cabin/Door/Row1/Left/IsOpen/%d", i/5)
		case 4:
			path = fmt.Sprintf("/Vehicle/Cabin/Seat/Row1/Pos1/IsBelted/%d", i/5)
		}

		// 업데이트 간격을 100-500ms 사이로 설정
		interval := 100 + rand.Intn(401)

		// 타입 및 초기값 설정
		var value interface{}
		var valueType string

		switch i % 3 {
		case 0:
			value = rand.Float64() * 100.0
			valueType = "number"
		case 1:
			value = rand.Intn(2) == 1
			valueType = "boolean"
		case 2:
			value = fmt.Sprintf("value_%d", i)
			valueType = "string"
		}

		// 패턴 선택
		pattern := ""
		params := make(map[string]float64)
		if valueType == "number" {
			switch i % 3 {
			case 0:
				pattern = "sinusoidal"
				params["amplitude"] = 3.0 + rand.Float64()*5.0
				params["period"] = 10000.0 + rand.Float64()*20000.0
				params["baseline"] = 50.0 + rand.Float64()*20.0
			case 1:
				pattern = "random_walk"
				params["step_size"] = 0.05 + rand.Float64()*0.2
				params["min"] = 0.0
				params["max"] = 100.0
			case 2:
				pattern = "constant_with_noise"
				params["baseline"] = 30.0 + rand.Float64()*40.0
				params["noise"] = 0.1 + rand.Float64()*1.0
			}
		} else if valueType == "boolean" {
			pattern = "toggle"
			params["toggle_probability"] = 0.01 + rand.Float64()*0.05
		} else {
			pattern = "constant"
		}

		vehicle.SensorsMedFreq[path] = &SensorData{
			Path:           path,
			Value:          value,
			Type:           valueType,
			Timestamp:      time.Now().UnixNano(),
			UpdateInterval: interval,
			ChangePattern:  pattern,
			ChangeParams:   params,
		}
	}

	// 저빈도 센서 (1000ms+ 업데이트)
	for i := 0; i < numLowFreqSensors; i++ {
		path := ""
		switch i % 5 {
		case 0:
			path = fmt.Sprintf("/Vehicle/Powertrain/FuelSystem/Level/%d", i/5)
		case 1:
			path = fmt.Sprintf("/Vehicle/Powertrain/Battery/StateOfCharge/%d", i/5)
		case 2:
			path = fmt.Sprintf("/Vehicle/Chassis/Axle/Row1/Wheel/Left/Tire/Pressure/%d", i/5)
		case 3:
			path = fmt.Sprintf("/Vehicle/Chassis/Axle/Row1/Wheel/Right/Tire/Pressure/%d", i/5)
		case 4:
			path = fmt.Sprintf("/Vehicle/Body/Trunk/IsOpen/%d", i/5)
		}

		// 업데이트 간격을 1000-5000ms 사이로 설정
		interval := 1000 + rand.Intn(4001)

		// 타입 및 초기값 설정
		var value interface{}
		var valueType string

		switch i % 3 {
		case 0:
			value = rand.Float64() * 100.0
			valueType = "number"
		case 1:
			value = rand.Intn(2) == 1
			valueType = "boolean"
		case 2:
			value = fmt.Sprintf("value_%d", i)
			valueType = "string"
		}

		// 패턴 선택
		pattern := ""
		params := make(map[string]float64)
		if valueType == "number" {
			switch i % 3 {
			case 0:
				pattern = "linear"
				params["slope"] = -0.001 - rand.Float64()*0.005
				params["min"] = 0.0
				params["max"] = 100.0
			case 1:
				pattern = "random_walk"
				params["step_size"] = 0.01 + rand.Float64()*0.05
				params["min"] = 0.0
				params["max"] = 100.0
			case 2:
				pattern = "constant_with_noise"
				params["baseline"] = 50.0 + rand.Float64()*20.0
				params["noise"] = 0.05 + rand.Float64()*0.2
			}
		} else if valueType == "boolean" {
			pattern = "toggle"
			params["toggle_probability"] = 0.001 + rand.Float64()*0.01
		} else {
			pattern = "constant"
		}

		vehicle.SensorsLowFreq[path] = &SensorData{
			Path:           path,
			Value:          value,
			Type:           valueType,
			Timestamp:      time.Now().UnixNano(),
			UpdateInterval: interval,
			ChangePattern:  pattern,
			ChangeParams:   params,
		}
	}

	// 높은 변동성 액추에이터
	for i := 0; i < numHighVarActuators; i++ {
		path := ""
		switch i % 5 {
		case 0:
			path = fmt.Sprintf("/Vehicle/Chassis/Brake/PedalPosition/%d", i/5)
		case 1:
			path = fmt.Sprintf("/Vehicle/Powertrain/Transmission/GearPosition/%d", i/5)
		case 2:
			path = fmt.Sprintf("/Vehicle/Powertrain/Engine/ThrottlePosition/%d", i/5)
		case 3:
			path = fmt.Sprintf("/Vehicle/Chassis/SteeringWheel/Angle/%d", i/5)
		case 4:
			path = fmt.Sprintf("/Vehicle/Body/Horn/IsActive/%d", i/5)
		}

		// 업데이트 간격을 50-200ms 사이로 설정
		interval := 50 + rand.Intn(151)

		// 타입 및 초기값 설정
		var value interface{}
		var valueType string

		switch i % 3 {
		case 0:
			value = rand.Float64() * 100.0
			valueType = "number"
		case 1:
			value = rand.Intn(2) == 1
			valueType = "boolean"
		case 2:
			value = fmt.Sprintf("gear_%d", rand.Intn(8))
			valueType = "string"
		}

		// 패턴 선택
		pattern := ""
		params := make(map[string]float64)
		if valueType == "number" {
			switch i % 3 {
			case 0:
				pattern = "random_walk"
				params["step_size"] = 1.0 + rand.Float64()*5.0
				params["min"] = 0.0
				params["max"] = 100.0
			case 1:
				pattern = "triggered"
				params["threshold"] = 0.3
				params["reset_interval"] = 1000.0 + rand.Float64()*2000.0
			case 2:
				pattern = "sinusoidal"
				params["amplitude"] = 20.0 + rand.Float64()*30.0
				params["period"] = 2000.0 + rand.Float64()*3000.0
				params["baseline"] = 50.0
			}
		} else if valueType == "boolean" {
			pattern = "toggle"
			params["toggle_probability"] = 0.1 + rand.Float64()*0.3
		} else {
			pattern = "random_selection"
			params["num_choices"] = 6.0
		}

		vehicle.ActuatorsHighVar[path] = &SensorData{
			Path:           path,
			Value:          value,
			Type:           valueType,
			Timestamp:      time.Now().UnixNano(),
			UpdateInterval: interval,
			ChangePattern:  pattern,
			ChangeParams:   params,
		}
	}

	// 낮은 변동성 액추에이터
	for i := 0; i < numLowVarActuators; i++ {
		path := ""
		switch i % 5 {
		case 0:
			path = fmt.Sprintf("/Vehicle/Body/Lights/Headlight/IsActive/%d", i/5)
		case 1:
			path = fmt.Sprintf("/Vehicle/Body/Lights/Brake/IsActive/%d", i/5)
		case 2:
			path = fmt.Sprintf("/Vehicle/Cabin/HVAC/Temperature/%d", i/5)
		case 3:
			path = fmt.Sprintf("/Vehicle/Cabin/HVAC/FanSpeed/%d", i/5)
		case 4:
			path = fmt.Sprintf("/Vehicle/Cabin/Sunroof/Position/%d", i/5)
		}

		// 업데이트 간격을 500-2000ms 사이로 설정
		interval := 500 + rand.Intn(1501)

		// 타입 및 초기값 설정
		var value interface{}
		var valueType string

		switch i % 3 {
		case 0:
			value = rand.Float64() * 100.0
			valueType = "number"
		case 1:
			value = rand.Intn(2) == 1
			valueType = "boolean"
		case 2:
			value = fmt.Sprintf("setting_%d", rand.Intn(5))
			valueType = "string"
		}

		// 패턴 선택
		pattern := ""
		params := make(map[string]float64)
		if valueType == "number" {
			switch i % 3 {
			case 0:
				pattern = "stepped"
				params["step_size"] = 5.0 + rand.Float64()*10.0
				params["step_interval"] = 10000.0 + rand.Float64()*20000.0
				params["min"] = 0.0
				params["max"] = 100.0
			case 1:
				pattern = "constant_with_noise"
				params["baseline"] = 50.0 + rand.Float64()*20.0
				params["noise"] = 0.1 + rand.Float64()*0.5
			case 2:
				pattern = "random_walk"
				params["step_size"] = 0.1 + rand.Float64()*0.5
				params["min"] = 0.0
				params["max"] = 100.0
			}
		} else if valueType == "boolean" {
			pattern = "toggle"
			params["toggle_probability"] = 0.01 + rand.Float64()*0.05
		} else {
			pattern = "random_selection"
			params["num_choices"] = 5.0
		}

		vehicle.ActuatorsLowVar[path] = &SensorData{
			Path:           path,
			Value:          value,
			Type:           valueType,
			Timestamp:      time.Now().UnixNano(),
			UpdateInterval: interval,
			ChangePattern:  pattern,
			ChangeParams:   params,
		}
	}

	// 속성값 (거의 변경되지 않음)
	for i := 0; i < numAttributes; i++ {
		path := ""
		switch i % 4 {
		case 0:
			path = fmt.Sprintf("/Vehicle/VehicleIdentification/VIN/%d", i/4)
		case 1:
			path = fmt.Sprintf("/Vehicle/VehicleIdentification/Model/%d", i/4)
		case 2:
			path = fmt.Sprintf("/Vehicle/VehicleIdentification/Year/%d", i/4)
		case 3:
			path = fmt.Sprintf("/Vehicle/VehicleIdentification/Color/%d", i/4)
		}

		// 업데이트 간격을 10000-30000ms 사이로 설정 (거의 변경 안됨)
		interval := 10000 + rand.Intn(20001)

		// 타입 및 초기값 설정
		var value interface{}
		var valueType string

		switch i % 3 {
		case 0:
			value = rand.Float64() * 100.0
			valueType = "number"
		case 1:
			value = rand.Intn(2) == 1
			valueType = "boolean"
		case 2:
			value = fmt.Sprintf("attribute_%d", i)
			valueType = "string"
		}

		// 패턴 선택 (대부분 상수)
		pattern := "constant"
		params := make(map[string]float64)

		vehicle.Attributes[path] = &SensorData{
			Path:           path,
			Value:          value,
			Type:           valueType,
			Timestamp:      time.Now().UnixNano(),
			UpdateInterval: interval,
			ChangePattern:  pattern,
			ChangeParams:   params,
		}
	}
}

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