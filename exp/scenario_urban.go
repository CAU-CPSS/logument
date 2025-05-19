package exp

import (
	"fmt"
	"math/rand"
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
