package exp

import (
	"math/rand"
)

//=============================================================================
// Scenario-specific configuration functions
//=============================================================================

// applyUrbanTrafficSettings applies settings for the urban traffic scenario
func applyUrbanTrafficSettings(vehicle *VehicleData) {
	// Urban traffic scenario - irregular speed, frequent direction changes

	// Speed settings (low to medium speed)
	if speedSensor, ok := vehicle.SensorsHighFreq["Vehicle.Speed"]; ok {
		speedSensor.Value = 30.0 + rand.Float64()*20.0
		speedSensor.ChangePattern = "random_walk"
		speedSensor.ChangeParams["step_size"] = 2.0
		speedSensor.ChangeParams["min"] = 0.0
		speedSensor.ChangeParams["max"] = 70.0
	}

	// Steering settings (frequent direction changes)
	if steeringSensor, ok := vehicle.SensorsHighFreq["Vehicle.Chassis.SteeringWheel.Angle"]; ok {
		steeringSensor.ChangePattern = "sinusoidal"
		steeringSensor.ChangeParams["amplitude"] = 15.0
		steeringSensor.ChangeParams["period"] = 10000.0
		steeringSensor.ChangeParams["baseline"] = 0.0
	}

	// Acceleration settings (irregular acceleration/deceleration)
	if accelSensor, ok := vehicle.SensorsHighFreq["Vehicle.Acceleration.Longitudinal"]; ok {
		accelSensor.ChangePattern = "random_walk"
		accelSensor.ChangeParams["step_size"] = 0.5
		accelSensor.ChangeParams["min"] = -3.0
		accelSensor.ChangeParams["max"] = 2.0
	}

	// Brake settings (frequent use)
	if brakeSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Brake.PedalPosition"]; ok {
		brakeSensor.ChangePattern = "random_walk"
		brakeSensor.ChangeParams["step_size"] = 5.0
		brakeSensor.ChangeParams["min"] = 0.0
		brakeSensor.ChangeParams["max"] = 80.0
	}

	// Accelerator pedal settings
	if accelPedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Accelerator.PedalPosition"]; ok {
		accelPedalSensor.ChangePattern = "random_walk"
		accelPedalSensor.ChangeParams["step_size"] = 3.0
		accelPedalSensor.ChangeParams["min"] = 0.0
		accelPedalSensor.ChangeParams["max"] = 60.0
	}

	// Battery discharge settings (consumption during urban driving)
	if batterySensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.StateOfCharge.Current"]; ok {
		batterySensor.ChangePattern = "linear"
		batterySensor.ChangeParams["slope"] = -0.1
		batterySensor.ChangeParams["min"] = 0.0
		batterySensor.ChangeParams["max"] = 100.0
	}

	// Fuel consumption settings
	if fuelSensor, ok := vehicle.SensorsLowFreq["Vehicle.Powertrain.FuelSystem.Level"]; ok {
		fuelSensor.ChangePattern = "linear"
		fuelSensor.ChangeParams["slope"] = -0.001
		fuelSensor.ChangeParams["min"] = 0.0
		fuelSensor.ChangeParams["max"] = 100.0
	}

	// Turn signal settings (frequent use in urban environments)
	if turnLeftSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Left.IsActive"]; ok {
		turnLeftSensor.ChangePattern = "toggle"
		turnLeftSensor.ChangeParams["toggle_probability"] = 0.02
	}

	if turnRightSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Right.IsActive"]; ok {
		turnRightSensor.ChangePattern = "toggle"
		turnRightSensor.ChangeParams["toggle_probability"] = 0.02
	}
}

// applyHighwayCruisingSettings applies settings for the highway cruising scenario
func applyHighwayCruisingSettings(vehicle *VehicleData) {
	// Highway cruising scenario - high speed, steady direction

	// Speed settings (high speed)
	if speedSensor, ok := vehicle.SensorsHighFreq["Vehicle.Speed"]; ok {
		speedSensor.Value = 90.0 + rand.Float64()*20.0
		speedSensor.ChangePattern = "constant_with_noise"
		speedSensor.ChangeParams["baseline"] = 100.0
		speedSensor.ChangeParams["noise"] = 0.5
	}

	// Steering settings (stable)
	if steeringSensor, ok := vehicle.SensorsHighFreq["Vehicle.Chassis.SteeringWheel.Angle"]; ok {
		steeringSensor.ChangePattern = "constant_with_noise"
		steeringSensor.ChangeParams["baseline"] = 0.0
		steeringSensor.ChangeParams["noise"] = 0.2
	}

	// Acceleration settings (stable)
	if accelSensor, ok := vehicle.SensorsHighFreq["Vehicle.Acceleration.Longitudinal"]; ok {
		accelSensor.Value = 0.0
		accelSensor.ChangePattern = "constant_with_noise"
		accelSensor.ChangeParams["baseline"] = 0.0
		accelSensor.ChangeParams["noise"] = 0.1
	}

	// Location sensor (almost straight path)
	if latSensor, ok := vehicle.SensorsHighFreq["Vehicle.CurrentLocation.Latitude"]; ok {
		latSensor.ChangePattern = "linear"
		latSensor.ChangeParams["slope"] = 0.1 // Very slight change
		latSensor.ChangeParams["min"] = -180.0
		latSensor.ChangeParams["max"] = 180.0
	}

	if lonSensor, ok := vehicle.SensorsHighFreq["Vehicle.CurrentLocation.Longitude"]; ok {
		lonSensor.ChangePattern = "linear"    // Straight path
		lonSensor.ChangeParams["slope"] = 0.1 // Very slight change
		lonSensor.ChangeParams["min"] = -180.0
		lonSensor.ChangeParams["max"] = 180.0
	}

	if brakePedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Brake.PedalPosition"]; ok {
		brakePedalSensor.Value = 0.0
		brakePedalSensor.ChangePattern = "toggle"
		brakePedalSensor.ChangeParams["toggle_probability"] = 0.01
	}

	// Accelerator pedal settings (maintain constant)
	if accelPedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Accelerator.PedalPosition"]; ok {
		accelPedalSensor.Value = 25.0
		accelPedalSensor.ChangePattern = "constant_with_noise"
		accelPedalSensor.ChangeParams["baseline"] = 25.0
		accelPedalSensor.ChangeParams["noise"] = 0.5
	}

	// Battery discharge settings (faster consumption during high-speed driving)
	if batterySensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.StateOfCharge.Current"]; ok {
		batterySensor.ChangePattern = "linear"
		batterySensor.ChangeParams["slope"] = -0.2
		batterySensor.ChangeParams["min"] = 0.0
		batterySensor.ChangeParams["max"] = 100.0
	}

	// Fuel consumption settings (faster consumption during high-speed driving)
	if fuelSensor, ok := vehicle.SensorsLowFreq["Vehicle.Powertrain.FuelSystem.Level"]; ok {
		fuelSensor.ChangePattern = "linear"
		fuelSensor.ChangeParams["slope"] = -0.002
		fuelSensor.ChangeParams["min"] = 0.0
		fuelSensor.ChangeParams["max"] = 100.0
	}

	// Turn signal settings (rarely used on highways)
	if turnLeftSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Left.IsActive"]; ok {
		turnLeftSensor.ChangePattern = "toggle"
		turnLeftSensor.ChangeParams["toggle_probability"] = 0.005
	}

	if turnRightSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Right.IsActive"]; ok {
		turnRightSensor.ChangePattern = "toggle"
		turnRightSensor.ChangeParams["toggle_probability"] = 0.005
	}
}

// applyBatteryChargingSettings applies settings for the battery charging scenario
func applyBatteryChargingSettings(vehicle *VehicleData) {
	// Battery charging scenario - vehicle stationary, charging in progress

	// Speed and motion sensor settings (stationary)
	if speedSensor, ok := vehicle.SensorsHighFreq["Vehicle.Speed"]; ok {
		speedSensor.Value = 0.0
		speedSensor.ChangePattern = "constant"
	}

	// Steering settings (stationary)
	if steeringSensor, ok := vehicle.SensorsHighFreq["Vehicle.Chassis.SteeringWheel.Angle"]; ok {
		steeringSensor.Value = 0.0
		steeringSensor.ChangePattern = "constant"
	}

	// Acceleration settings (stationary)
	if accelSensor, ok := vehicle.SensorsHighFreq["Vehicle.Acceleration.Longitudinal"]; ok {
		accelSensor.Value = 0.0
		accelSensor.ChangePattern = "constant"
	}

	// Brake and accelerator settings (not in use)
	if brakeSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Brake.PedalPosition"]; ok {
		brakeSensor.Value = 0.0
		brakeSensor.ChangePattern = "constant"
	}

	if accelPedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Accelerator.PedalPosition"]; ok {
		accelPedalSensor.Value = 0.0
		accelPedalSensor.ChangePattern = "constant"
	}

	// Battery charging settings
	if batterySensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.StateOfCharge.Current"]; ok {
		batterySensor.Value = 20.0 + rand.Float64()*20.0 // Initial charge state 20-40%
		batterySensor.ChangePattern = "linear"
		batterySensor.ChangeParams["slope"] = 0.4 // Gradual increase
		batterySensor.ChangeParams["min"] = 0.0
		batterySensor.ChangeParams["max"] = 100.0
	}

	// Battery temperature settings (slight increase during charging)
	if tempSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Temperature"]; ok {
		tempSensor.ChangePattern = "linear"
		tempSensor.ChangeParams["slope"] = 0.001
		tempSensor.ChangeParams["min"] = 20.0
		tempSensor.ChangeParams["max"] = 45.0
	}

	// Charging-related sensor settings
	if chargingSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Charging.IsCharging"]; ok {
		chargingSensor.Value = true
	}

	if ratesSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Charging.ChargingRate"]; ok {
		ratesSensor.Value = 11.0 + rand.Float64()*2.0 // 11-13kW charging
		ratesSensor.ChangePattern = "sinusoidal"
		ratesSensor.ChangeParams["amplitude"] = 1.0
		ratesSensor.ChangeParams["period"] = 15000.0
		ratesSensor.ChangeParams["baseline"] = 11.5
	}

	if timeSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Charging.TimeRemaining"]; ok {
		timeSensor.Value = 120.0 + rand.Float64()*60.0 // 2-3 hours
		timeSensor.ChangePattern = "linear"
		timeSensor.ChangeParams["slope"] = -0.05 // Time decreases
		timeSensor.ChangeParams["min"] = 0.0
		timeSensor.ChangeParams["max"] = 180.0
	}
}
