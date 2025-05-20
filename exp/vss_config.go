package exp

import (
	"time"
)

// 센서 카테고리 정의
const (
	SensorTypeHigh   = "high_frequency"   // 고빈도 센서 (10-50ms 업데이트)
	SensorTypeMedium = "medium_frequency" // 중빈도 센서 (100-500ms 업데이트)
	SensorTypeLow    = "low_frequency"    // 저빈도 센서 (1000ms+ 업데이트)
	ActuatorTypeHigh = "high_variable"    // 높은 변동성 액추에이터
	ActuatorTypeLow  = "low_variable"     // 낮은 변동성 액추에이터
	AttributeType    = "attribute"        // 거의 변경되지 않는 속성값
)

//=============================================================================
// 시나리오 및 차량 데이터 구조체
//=============================================================================

// VehicleData는 차량의 센서, 액추에이터, 속성 값을 저장합니다
type VehicleData struct {
	SensorsHighFreq  map[string]*SensorData // 고빈도 업데이트 센서 (GPS, 속도 등)
	SensorsMedFreq   map[string]*SensorData // 중간 빈도 업데이트 센서 (온도, 배터리 등)
	SensorsLowFreq   map[string]*SensorData // 저빈도 업데이트 센서 (연료, 타이어 압력 등)
	ActuatorsHighVar map[string]*SensorData // 높은 변동성 액추에이터 (브레이크, 가속 등)
	ActuatorsLowVar  map[string]*SensorData // 낮은 변동성 액추에이터 (창문, 도어 등)
	Attributes       map[string]*SensorData // 거의 변경되지 않는 속성값 (차량 ID, 모델 등)
}

// SensorData는 개별 센서/액추에이터 데이터를 저장합니다
type SensorData struct {
	Path           string             // 센서 경로
	Value          interface{}        // 현재 값
	Type           string             // 값 타입 (number, string, boolean)
	Timestamp      int64              // 마지막 업데이트 타임스탬프
	UpdateInterval int                // 업데이트 간격 (밀리초)
	ChangePattern  string             // 변경 패턴 (constant, linear, sinusoidal, random, triggered)
	ChangeParams   map[string]float64 // 변경 패턴 파라미터
	ChangeCounter  int                // 변경 카운터 (시간 경과 추적용)
}

// VSSSensor는 VSS 기반 차량 센서, 액추에이터 또는 속성을 정의합니다
type VSSSensor struct {
	Path           string             // VSS 경로
	DefaultValue   interface{}        // 기본값
	Type           string             // 값 타입 (number, string, boolean)
	Category       string             // 센서 카테고리 (high_frequency, medium_frequency, 등)
	UpdateInterval int                // 업데이트 간격 (밀리초)
	Description    string             // 설명
	Unit           string             // 단위 (km/h, celsius 등)
	Min            float64            // 최소값 (숫자 타입인 경우)
	Max            float64            // 최대값 (숫자 타입인 경우)
	ChangePatterns map[string]string  // 시나리오별 변경 패턴
	ChangeParams   map[string]float64 // 변경 패턴 매개변수
}

//=============================================================================
// VSS 센서 정의
//=============================================================================

// VSSSensors는 차량 신호 표준(VSS) 기반 센서 정의 맵입니다
var VSSSensors = map[string]*VSSSensor{
	// 위치 센서 (고빈도)
	"Vehicle.CurrentLocation.Latitude": {
		Path:           "Vehicle.CurrentLocation.Latitude",
		DefaultValue:   37.7749,
		Type:           "number",
		Category:       SensorTypeHigh,
		UpdateInterval: 50,
		Description:    "차량의 현재 위도",
		Unit:           "degrees",
		Min:            -90.0,
		Max:            90.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "random_walk",
			"battery_charging": "constant",
		},
	},
	"Vehicle.CurrentLocation.Longitude": {
		Path:           "Vehicle.CurrentLocation.Longitude",
		DefaultValue:   -122.4194,
		Type:           "number",
		Category:       SensorTypeHigh,
		UpdateInterval: 50,
		Description:    "차량의 현재 경도",
		Unit:           "degrees",
		Min:            -180.0,
		Max:            180.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "random_walk",
			"battery_charging": "constant",
		},
	},
	"Vehicle.CurrentLocation.Altitude": {
		Path:           "Vehicle.CurrentLocation.Altitude",
		DefaultValue:   10.0,
		Type:           "number",
		Category:       SensorTypeHigh,
		UpdateInterval: 100,
		Description:    "차량의 현재 고도",
		Unit:           "m",
		Min:            -500.0,
		Max:            9000.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// 속도 센서 (고빈도)
	"Vehicle.Speed": {
		Path:           "Vehicle.Speed",
		DefaultValue:   0.0,
		Type:           "number",
		Category:       SensorTypeHigh,
		UpdateInterval: 20,
		Description:    "차량의 현재 속도",
		Unit:           "km/h",
		Min:            0.0,
		Max:            220.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// 가속도 센서 (고빈도)
	"Vehicle.Acceleration.Longitudinal": {
		Path:           "Vehicle.Acceleration.Longitudinal",
		DefaultValue:   0.0,
		Type:           "number",
		Category:       SensorTypeHigh,
		UpdateInterval: 20,
		Description:    "차량의 종방향 가속도",
		Unit:           "m/s²",
		Min:            -20.0,
		Max:            20.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "sinusoidal",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},
	"Vehicle.Acceleration.Lateral": {
		Path:           "Vehicle.Acceleration.Lateral",
		DefaultValue:   0.0,
		Type:           "number",
		Category:       SensorTypeHigh,
		UpdateInterval: 20,
		Description:    "차량의 횡방향 가속도",
		Unit:           "m/s²",
		Min:            -20.0,
		Max:            20.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "sinusoidal",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// 조향각 센서 (고빈도)
	"Vehicle.Chassis.SteeringWheel.Angle": {
		Path:           "Vehicle.Chassis.SteeringWheel.Angle",
		DefaultValue:   0.0,
		Type:           "number",
		Category:       SensorTypeHigh,
		UpdateInterval: 25,
		Description:    "현재 스티어링 휠 각도",
		Unit:           "degrees",
		Min:            -180.0,
		Max:            180.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// 엔진 속도 (중빈도)
	"Vehicle.Powertrain.CombustionEngine.Speed": {
		Path:           "Vehicle.Powertrain.CombustionEngine.Speed",
		DefaultValue:   800.0,
		Type:           "number",
		Category:       SensorTypeMedium,
		UpdateInterval: 100,
		Description:    "엔진 회전 속도",
		Unit:           "rpm",
		Min:            0.0,
		Max:            8000.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// 배터리 센서 (중빈도)
	"Vehicle.Powertrain.TractionBattery.StateOfCharge.Current": {
		Path:           "Vehicle.Powertrain.TractionBattery.StateOfCharge.Current",
		DefaultValue:   80.0,
		Type:           "number",
		Category:       SensorTypeMedium,
		UpdateInterval: 200,
		Description:    "배터리 충전 상태",
		Unit:           "percent",
		Min:            0.0,
		Max:            100.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "linear",
			"highway_cruising": "linear",
			"battery_charging": "linear",
		},
	},

	"Vehicle.Powertrain.TractionBattery.Temperature": {
		Path:           "Vehicle.Powertrain.TractionBattery.Temperature",
		DefaultValue:   25.0,
		Type:           "number",
		Category:       SensorTypeMedium,
		UpdateInterval: 500,
		Description:    "배터리 온도",
		Unit:           "celsius",
		Min:            -20.0,
		Max:            80.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "constant_with_noise",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "linear",
		},
	},

	// 연료 상태 (저빈도)
	"Vehicle.Powertrain.FuelSystem.Level": {
		Path:           "Vehicle.Powertrain.FuelSystem.Level",
		DefaultValue:   80.0,
		Type:           "number",
		Category:       SensorTypeLow,
		UpdateInterval: 5000,
		Description:    "연료 탱크 레벨",
		Unit:           "percent",
		Min:            0.0,
		Max:            100.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "linear",
			"highway_cruising": "linear",
			"battery_charging": "constant",
		},
	},

	// 타이어 센서 (저빈도)
	"Vehicle.Chassis.Axle.Row1.Wheel.Left.Tire.Pressure": {
		Path:           "Vehicle.Chassis.Axle.Row1.Wheel.Left.Tire.Pressure",
		DefaultValue:   2.2,
		Type:           "number",
		Category:       SensorTypeLow,
		UpdateInterval: 10000,
		Description:    "왼쪽 앞바퀴 타이어 공기압",
		Unit:           "bar",
		Min:            0.0,
		Max:            4.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "constant_with_noise",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},
	"Vehicle.Chassis.Axle.Row1.Wheel.Right.Tire.Pressure": {
		Path:           "Vehicle.Chassis.Axle.Row1.Wheel.Right.Tire.Pressure",
		DefaultValue:   2.2,
		Type:           "number",
		Category:       SensorTypeLow,
		UpdateInterval: 10000,
		Description:    "오른쪽 앞바퀴 타이어 공기압",
		Unit:           "bar",
		Min:            0.0,
		Max:            4.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "constant_with_noise",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// 액추에이터 - 브레이크 (고변동성)
	"Vehicle.Chassis.Brake.PedalPosition": {
		Path:           "Vehicle.Chassis.Brake.PedalPosition",
		DefaultValue:   0.0,
		Type:           "number",
		Category:       ActuatorTypeHigh,
		UpdateInterval: 50,
		Description:    "브레이크 페달 위치",
		Unit:           "percent",
		Min:            0.0,
		Max:            100.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// 액추에이터 - 가속 페달 (고변동성)
	"Vehicle.Chassis.Accelerator.PedalPosition": {
		Path:           "Vehicle.Chassis.Accelerator.PedalPosition",
		DefaultValue:   0.0,
		Type:           "number",
		Category:       ActuatorTypeHigh,
		UpdateInterval: 50,
		Description:    "가속 페달 위치",
		Unit:           "percent",
		Min:            0.0,
		Max:            100.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// 액추에이터 - 헤드라이트 (저변동성)
	"Vehicle.Body.Lights.Headlight.IsActive": {
		Path:           "Vehicle.Body.Lights.Headlight.IsActive",
		DefaultValue:   false,
		Type:           "boolean",
		Category:       ActuatorTypeLow,
		UpdateInterval: 5000,
		Description:    "헤드라이트 상태",
		Unit:           "",
		ChangePatterns: map[string]string{
			"urban_traffic":    "toggle",
			"highway_cruising": "toggle",
			"battery_charging": "constant",
		},
	},

	// 액추에이터 - 방향지시등 (저변동성)
	"Vehicle.Body.Lights.TurnSignal.Left.IsActive": {
		Path:           "Vehicle.Body.Lights.TurnSignal.Left.IsActive",
		DefaultValue:   false,
		Type:           "boolean",
		Category:       ActuatorTypeLow,
		UpdateInterval: 1000,
		Description:    "왼쪽 방향지시등 상태",
		Unit:           "",
		ChangePatterns: map[string]string{
			"urban_traffic":    "toggle",
			"highway_cruising": "toggle",
			"battery_charging": "constant",
		},
	},
	"Vehicle.Body.Lights.TurnSignal.Right.IsActive": {
		Path:           "Vehicle.Body.Lights.TurnSignal.Right.IsActive",
		DefaultValue:   false,
		Type:           "boolean",
		Category:       ActuatorTypeLow,
		UpdateInterval: 1000,
		Description:    "오른쪽 방향지시등 상태",
		Unit:           "",
		ChangePatterns: map[string]string{
			"urban_traffic":    "toggle",
			"highway_cruising": "toggle",
			"battery_charging": "constant",
		},
	},

	// 속성 - 차량 식별 정보
	"Vehicle.VehicleIdentification.VIN": {
		Path:           "Vehicle.VehicleIdentification.VIN",
		DefaultValue:   "1HGCM82633A123456",
		Type:           "string",
		Category:       AttributeType,
		UpdateInterval: 86400000, // 24시간
		Description:    "차대 번호",
		Unit:           "",
		ChangePatterns: map[string]string{
			"urban_traffic":    "constant",
			"highway_cruising": "constant",
			"battery_charging": "constant",
		},
	},
	"Vehicle.VehicleIdentification.Model": {
		Path:           "Vehicle.VehicleIdentification.Model",
		DefaultValue:   "Model X",
		Type:           "string",
		Category:       AttributeType,
		UpdateInterval: 86400000, // 24시간
		Description:    "차량 모델명",
		Unit:           "",
		ChangePatterns: map[string]string{
			"urban_traffic":    "constant",
			"highway_cruising": "constant",
			"battery_charging": "constant",
		},
	},

	// 충전 시나리오 특화 센서
	"Vehicle.Powertrain.TractionBattery.Charging.IsCharging": {
		Path:           "Vehicle.Powertrain.TractionBattery.Charging.IsCharging",
		DefaultValue:   false,
		Type:           "boolean",
		Category:       SensorTypeMedium,
		UpdateInterval: 1000,
		Description:    "배터리 충전 중 여부",
		Unit:           "",
		ChangePatterns: map[string]string{
			"urban_traffic":    "constant",
			"highway_cruising": "constant",
			"battery_charging": "constant",
		},
	},
	"Vehicle.Powertrain.TractionBattery.Charging.ChargingRate": {
		Path:           "Vehicle.Powertrain.TractionBattery.Charging.ChargingRate",
		DefaultValue:   0.0,
		Type:           "number",
		Category:       SensorTypeMedium,
		UpdateInterval: 500,
		Description:    "배터리 충전 속도",
		Unit:           "kW",
		Min:            0.0,
		Max:            350.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "constant",
			"highway_cruising": "constant",
			"battery_charging": "sinusoidal",
		},
	},
	"Vehicle.Powertrain.TractionBattery.Charging.TimeRemaining": {
		Path:           "Vehicle.Powertrain.TractionBattery.Charging.TimeRemaining",
		DefaultValue:   120.0,
		Type:           "number",
		Category:       SensorTypeMedium,
		UpdateInterval: 1000,
		Description:    "충전 완료까지 남은 시간",
		Unit:           "minutes",
		Min:            0.0,
		Max:            500.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "constant",
			"highway_cruising": "constant",
			"battery_charging": "linear",
		},
	},
}

//=============================================================================
// 차량 데이터 초기화 함수
//=============================================================================

// InitializeVehicleData는 지정된 시나리오에 맞게 차량 데이터를 초기화합니다
func InitializeVehicleData(scenario string) *VehicleData {
	vehicle := &VehicleData{
		SensorsHighFreq:  make(map[string]*SensorData),
		SensorsMedFreq:   make(map[string]*SensorData),
		SensorsLowFreq:   make(map[string]*SensorData),
		ActuatorsHighVar: make(map[string]*SensorData),
		ActuatorsLowVar:  make(map[string]*SensorData),
		Attributes:       make(map[string]*SensorData),
	}

	// 모든 센서를 순회하며 초기화
	for path, vssSensor := range VSSSensors {
		// 시나리오에 맞는 변경 패턴 선택
		changePattern := "constant"
		if pattern, exists := vssSensor.ChangePatterns[scenario]; exists {
			changePattern = pattern
		}

		// 센서 데이터 생성
		sensorData := &SensorData{
			Path:           path,
			Value:          vssSensor.DefaultValue,
			Type:           vssSensor.Type,
			Timestamp:      time.Now().UnixNano(),
			UpdateInterval: vssSensor.UpdateInterval,
			ChangePattern:  changePattern,
			ChangeParams:   make(map[string]float64),
		}

		// 패턴에 따른 기본 매개변수 설정
		switch changePattern {
		case "random_walk":
			sensorData.ChangeParams["step_size"] = (vssSensor.Max - vssSensor.Min) * 0.01
			sensorData.ChangeParams["min"] = vssSensor.Min
			sensorData.ChangeParams["max"] = vssSensor.Max
		case "linear":
			if scenario == "battery_charging" && path == "Vehicle.Powertrain.TractionBattery.StateOfCharge.Current" {
				sensorData.ChangeParams["slope"] = 0.05 // 충전 중 배터리 증가
			} else if scenario == "battery_charging" && path == "Vehicle.Powertrain.TractionBattery.Charging.TimeRemaining" {
				sensorData.ChangeParams["slope"] = -0.5 // 충전 시간 감소
			} else {
				sensorData.ChangeParams["slope"] = -0.04 // 기본 감소
			}
			sensorData.ChangeParams["min"] = vssSensor.Min
			sensorData.ChangeParams["max"] = vssSensor.Max
		case "sinusoidal":
			sensorData.ChangeParams["amplitude"] = (vssSensor.Max - vssSensor.Min) * 0.1
			sensorData.ChangeParams["period"] = 5000.0
			sensorData.ChangeParams["baseline"] = (vssSensor.Min + vssSensor.Max) * 0.5
		case "constant_with_noise":
			sensorData.ChangeParams["baseline"] = vssSensor.DefaultValue.(float64)
			sensorData.ChangeParams["noise"] = (vssSensor.Max - vssSensor.Min) * 0.01
		case "toggle":
			sensorData.ChangeParams["toggle_probability"] = 0.05
		}

		// 카테고리에 따라 적절한 맵에 추가
		switch vssSensor.Category {
		case SensorTypeHigh:
			vehicle.SensorsHighFreq[path] = sensorData
		case SensorTypeMedium:
			vehicle.SensorsMedFreq[path] = sensorData
		case SensorTypeLow:
			vehicle.SensorsLowFreq[path] = sensorData
		case ActuatorTypeHigh:
			vehicle.ActuatorsHighVar[path] = sensorData
		case ActuatorTypeLow:
			vehicle.ActuatorsLowVar[path] = sensorData
		case AttributeType:
			vehicle.Attributes[path] = sensorData
		}
	}

	// 시나리오별 특화 설정 적용
	switch scenario {
	case "urban_traffic":
		applyUrbanTrafficSettings(vehicle)
	case "highway_cruising":
		applyHighwayCruisingSettings(vehicle)
	case "battery_charging":
		applyBatteryChargingSettings(vehicle)
	}

	return vehicle
}
