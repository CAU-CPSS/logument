package exp

import (
	"math/rand"
)

//=============================================================================
// 시나리오별 설정 함수
//=============================================================================

// applyUrbanTrafficSettings은 도심 교통 시나리오에 맞는 설정을 적용합니다
func applyUrbanTrafficSettings(vehicle *VehicleData) {
	// 도심 교통 시나리오 - 불규칙한 속도, 잦은 방향 전환
	
	// 속도 설정 (중저속)
	if speedSensor, ok := vehicle.SensorsHighFreq["Vehicle.Speed"]; ok {
		speedSensor.Value = 30.0 + rand.Float64()*20.0
		speedSensor.ChangePattern = "random_walk"
		speedSensor.ChangeParams["step_size"] = 2.0
		speedSensor.ChangeParams["min"] = 0.0
		speedSensor.ChangeParams["max"] = 70.0
	}
	
	// 스티어링 설정 (빈번한 방향 전환)
	if steeringSensor, ok := vehicle.SensorsHighFreq["Vehicle.Chassis.SteeringWheel.Angle"]; ok {
		steeringSensor.ChangePattern = "sinusoidal"
		steeringSensor.ChangeParams["amplitude"] = 15.0
		steeringSensor.ChangeParams["period"] = 10000.0
		steeringSensor.ChangeParams["baseline"] = 0.0
	}
	
	// 가속도 설정 (불규칙한 가속/감속)
	if accelSensor, ok := vehicle.SensorsHighFreq["Vehicle.Acceleration.Longitudinal"]; ok {
		accelSensor.ChangePattern = "random_walk"
		accelSensor.ChangeParams["step_size"] = 0.5
		accelSensor.ChangeParams["min"] = -3.0
		accelSensor.ChangeParams["max"] = 2.0
	}
	
	// 브레이크 설정 (자주 사용)
	if brakeSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Brake.PedalPosition"]; ok {
		brakeSensor.ChangePattern = "random_walk"
		brakeSensor.ChangeParams["step_size"] = 5.0
		brakeSensor.ChangeParams["min"] = 0.0
		brakeSensor.ChangeParams["max"] = 80.0
	}
	
	// 가속 페달 설정
	if accelPedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Accelerator.PedalPosition"]; ok {
		accelPedalSensor.ChangePattern = "random_walk"
		accelPedalSensor.ChangeParams["step_size"] = 3.0
		accelPedalSensor.ChangeParams["min"] = 0.0
		accelPedalSensor.ChangeParams["max"] = 60.0
	}
	
	// 배터리 방전 설정 (도심 주행 시 소모)
	if batterySensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.StateOfCharge.Current"]; ok {
		batterySensor.ChangePattern = "linear"
		batterySensor.ChangeParams["slope"] = -0.1
		batterySensor.ChangeParams["min"] = 0.0
		batterySensor.ChangeParams["max"] = 100.0
	}
	
	// 연료 소모 설정
	if fuelSensor, ok := vehicle.SensorsLowFreq["Vehicle.Powertrain.FuelSystem.Level"]; ok {
		fuelSensor.ChangePattern = "linear"
		fuelSensor.ChangeParams["slope"] = -0.001
		fuelSensor.ChangeParams["min"] = 0.0
		fuelSensor.ChangeParams["max"] = 100.0
	}
	
	// 방향지시등 설정 (도심 환경에서 자주 사용)
	if turnLeftSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Left.IsActive"]; ok {
		turnLeftSensor.ChangePattern = "toggle"
		turnLeftSensor.ChangeParams["toggle_probability"] = 0.02
	}
	
	if turnRightSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Right.IsActive"]; ok {
		turnRightSensor.ChangePattern = "toggle"
		turnRightSensor.ChangeParams["toggle_probability"] = 0.02
	}
}

// applyHighwayCruisingSettings는 고속도로 주행 시나리오에 맞는 설정을 적용합니다
func applyHighwayCruisingSettings(vehicle *VehicleData) {
	// 고속도로 주행 시나리오 - 높은 속도, 일정한 방향
	
	// 속도 설정 (고속)
	if speedSensor, ok := vehicle.SensorsHighFreq["Vehicle.Speed"]; ok {
		speedSensor.Value = 90.0 + rand.Float64()*20.0
		speedSensor.ChangePattern = "constant_with_noise"
		speedSensor.ChangeParams["baseline"] = 100.0
		speedSensor.ChangeParams["noise"] = 1.0
	}
	
	// 스티어링 설정 (안정적)
	if steeringSensor, ok := vehicle.SensorsHighFreq["Vehicle.Chassis.SteeringWheel.Angle"]; ok {
		steeringSensor.ChangePattern = "constant_with_noise"
		steeringSensor.ChangeParams["baseline"] = 0.0
		steeringSensor.ChangeParams["noise"] = 1.0
	}
	
	// 가속도 설정 (안정적)
	if accelSensor, ok := vehicle.SensorsHighFreq["Vehicle.Acceleration.Longitudinal"]; ok {
		accelSensor.Value = 0.0
		accelSensor.ChangePattern = "constant_with_noise"
		accelSensor.ChangeParams["baseline"] = 0.0
		accelSensor.ChangeParams["noise"] = 0.3
	}
	
	// 브레이크 설정 (거의 사용 안함)
	if brakeSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Brake.PedalPosition"]; ok {
		brakeSensor.Value = 0.0
		brakeSensor.ChangePattern = "constant_with_noise"
		brakeSensor.ChangeParams["baseline"] = 0.0
		brakeSensor.ChangeParams["noise"] = 1.0
	}
	
	// 가속 페달 설정 (일정하게 유지)
	if accelPedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Accelerator.PedalPosition"]; ok {
		accelPedalSensor.Value = 25.0
		accelPedalSensor.ChangePattern = "constant_with_noise"
		accelPedalSensor.ChangeParams["baseline"] = 25.0
		accelPedalSensor.ChangeParams["noise"] = 3.0
	}
	
	// 배터리 방전 설정 (고속 주행 시 더 빠르게 소모)
	if batterySensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.StateOfCharge.Current"]; ok {
		batterySensor.ChangePattern = "linear"
		batterySensor.ChangeParams["slope"] = -0.2
		batterySensor.ChangeParams["min"] = 0.0
		batterySensor.ChangeParams["max"] = 100.0
	}
	
	// 연료 소모 설정 (고속 주행 시 더 빠르게 소모)
	if fuelSensor, ok := vehicle.SensorsLowFreq["Vehicle.Powertrain.FuelSystem.Level"]; ok {
		fuelSensor.ChangePattern = "linear"
		fuelSensor.ChangeParams["slope"] = -0.002
		fuelSensor.ChangeParams["min"] = 0.0
		fuelSensor.ChangeParams["max"] = 100.0
	}
	
	// 방향지시등 설정 (고속도로에서 드물게 사용)
	if turnLeftSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Left.IsActive"]; ok {
		turnLeftSensor.ChangePattern = "toggle"
		turnLeftSensor.ChangeParams["toggle_probability"] = 0.005
	}
	
	if turnRightSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Right.IsActive"]; ok {
		turnRightSensor.ChangePattern = "toggle"
		turnRightSensor.ChangeParams["toggle_probability"] = 0.005
	}
}

// applyBatteryChargingSettings는 배터리 충전 시나리오에 맞는 설정을 적용합니다
func applyBatteryChargingSettings(vehicle *VehicleData) {
	// 배터리 충전 시나리오 - 차량 정지, 충전 진행중
	
	// 속도 및 동작 센서 설정 (정지)
	if speedSensor, ok := vehicle.SensorsHighFreq["Vehicle.Speed"]; ok {
		speedSensor.Value = 0.0
		speedSensor.ChangePattern = "constant"
	}
	
	// 스티어링 설정 (정지)
	if steeringSensor, ok := vehicle.SensorsHighFreq["Vehicle.Chassis.SteeringWheel.Angle"]; ok {
		steeringSensor.Value = 0.0
		steeringSensor.ChangePattern = "constant"
	}
	
	// 가속도 설정 (정지)
	if accelSensor, ok := vehicle.SensorsHighFreq["Vehicle.Acceleration.Longitudinal"]; ok {
		accelSensor.Value = 0.0
		accelSensor.ChangePattern = "constant"
	}
	
	// 브레이크 및 액셀 설정 (미사용)
	if brakeSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Brake.PedalPosition"]; ok {
		brakeSensor.Value = 0.0
		brakeSensor.ChangePattern = "constant"
	}
	
	if accelPedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Accelerator.PedalPosition"]; ok {
		accelPedalSensor.Value = 0.0
		accelPedalSensor.ChangePattern = "constant"
	}
	
	// 배터리 충전 설정
	if batterySensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.StateOfCharge.Current"]; ok {
		batterySensor.Value = 20.0 + rand.Float64()*20.0  // 시작 충전 상태 20-40%
		batterySensor.ChangePattern = "linear"
		batterySensor.ChangeParams["slope"] = 0.4        // 점진적 증가
		batterySensor.ChangeParams["min"] = 0.0
		batterySensor.ChangeParams["max"] = 100.0
	}
	
	// 배터리 온도 설정 (충전 중 소폭 상승)
	if tempSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Temperature"]; ok {
		tempSensor.ChangePattern = "linear"
		tempSensor.ChangeParams["slope"] = 0.001
		tempSensor.ChangeParams["min"] = 20.0
		tempSensor.ChangeParams["max"] = 45.0
	}
	
	// 충전 관련 센서 설정
	if chargingSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Charging.IsCharging"]; ok {
		chargingSensor.Value = true
	}
	
	if ratesSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Charging.ChargingRate"]; ok {
		ratesSensor.Value = 11.0 + rand.Float64()*2.0  // 11-13kW 충전
		ratesSensor.ChangePattern = "sinusoidal"
		ratesSensor.ChangeParams["amplitude"] = 1.0
		ratesSensor.ChangeParams["period"] = 15000.0
		ratesSensor.ChangeParams["baseline"] = 11.5
	}
	
	if timeSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Charging.TimeRemaining"]; ok {
		timeSensor.Value = 120.0 + rand.Float64()*60.0  // 2-3시간
		timeSensor.ChangePattern = "linear"
		timeSensor.ChangeParams["slope"] = -0.05        // 시간 감소
		timeSensor.ChangeParams["min"] = 0.0
		timeSensor.ChangeParams["max"] = 180.0
	}
}