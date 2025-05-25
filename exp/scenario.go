package exp

import (
	"math/rand"
	"strings"
	"time"
)

//=============================================================================
// Vehicle Data Initialization Function
//=============================================================================

// InitializeVehicleData initializes vehicle data according to the specified scenario
func InitializeVehicleData(scenario string) *VehicleData {
	vehicle := &VehicleData{
		SensorsHighFreq:  make(map[string]*SensorData),
		SensorsMedFreq:   make(map[string]*SensorData),
		SensorsLowFreq:   make(map[string]*SensorData),
		ActuatorsHighVar: make(map[string]*SensorData),
		ActuatorsLowVar:  make(map[string]*SensorData),
		Attributes:       make(map[string]*SensorData),
	}

	// Iterate through all sensors and initialize
	for path, vssSensor := range VSSSensors {
		// Select change pattern according to the scenario
		changePattern := "constant"
		if pattern, exists := vssSensor.ChangePatterns[scenario]; exists {
			changePattern = pattern
		}

		// Create sensor data
		sensorData := &SensorData{
			Path:           path,
			Value:          vssSensor.DefaultValue,
			Type:           vssSensor.Type,
			Timestamp:      time.Now().UnixNano(),
			UpdateInterval: vssSensor.UpdateInterval,
			ChangePattern:  changePattern,
			ChangeParams:   make(map[string]float64),
		}

		// Set default parameters based on the pattern
		switch changePattern {
		case "random_walk":
			sensorData.ChangeParams["step_size"] = (vssSensor.Max - vssSensor.Min) * 0.01
			sensorData.ChangeParams["min"] = vssSensor.Min
			sensorData.ChangeParams["max"] = vssSensor.Max
		case "linear":
			if scenario == "battery_charging" && path == "Vehicle.Powertrain.TractionBattery.StateOfCharge.Current" {
				sensorData.ChangeParams["slope"] = 0.1 // Battery increase during charging
			} else if scenario == "battery_charging" && path == "Vehicle.Powertrain.TractionBattery.Charging.TimeRemaining" {
				sensorData.ChangeParams["slope"] = -0.5 // Decrease charging time
			} else {
				sensorData.ChangeParams["slope"] = -0.04 // Default decrease
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

		// Add to the appropriate map based on the category
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

	// Apply scenario-specific settings
	applyScenarioSettings(vehicle, scenario)

	return vehicle
}

// applyScenarioSettings applies scenario-specific settings to vehicle
func applyScenarioSettings(vehicle *VehicleData, scenario string) {
	// For mixed_scenario, use urban_traffic as the base pattern
	actualScenario := scenario
	if scenario == "mixed_scenario" {
		actualScenario = "urban_traffic"
	}

	// First, update all sensors with their scenario-specific patterns from VSSSensors
	updateAllSensorPatterns(vehicle, actualScenario)

	switch scenario {
	case "urban_traffic":
		applyUrbanTrafficSettings(vehicle)
	case "highway_cruising":
		applyHighwayCruisingSettings(vehicle)
	case "battery_charging":
		if batterySensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.StateOfCharge.Current"]; ok {
			batterySensor.Value = 20.0 + rand.Float64()*20.0 // 20-40% SOC
		}
		applyBatteryChargingSettings(vehicle)
	case "mixed_scenario":
		// For mixed scenario, start with urban_traffic settings
		applyUrbanTrafficSettings(vehicle)
	}
}

// updateAllSensorPatterns updates all sensors with their scenario-specific patterns
func updateAllSensorPatterns(vehicle *VehicleData, scenario string) {
	// Update all sensor groups
	for _, sensorGroup := range []map[string]*SensorData{
		vehicle.SensorsHighFreq,
		vehicle.SensorsMedFreq,
		vehicle.SensorsLowFreq,
		vehicle.ActuatorsHighVar,
		vehicle.ActuatorsLowVar,
		vehicle.Attributes,
	} {
		for path, sensor := range sensorGroup {
			if vssSensor, exists := VSSSensors[path]; exists {
				// Update pattern if defined for this scenario
				if pattern, patternExists := vssSensor.ChangePatterns[scenario]; patternExists {
					sensor.ChangePattern = pattern

					// Reset parameters for the new pattern
					sensor.ChangeParams = make(map[string]float64)

					// Set default parameters based on the new pattern
					switch pattern {
					case "random_walk":
						sensor.ChangeParams["step_size"] = (vssSensor.Max - vssSensor.Min) * 0.01
						sensor.ChangeParams["min"] = vssSensor.Min
						sensor.ChangeParams["max"] = vssSensor.Max

					case "linear":
						if scenario == "battery_charging" && path == "Vehicle.Powertrain.TractionBattery.StateOfCharge.Current" {
							sensor.ChangeParams["slope"] = 0.1 // Battery increase during charging
						} else if scenario == "battery_charging" && path == "Vehicle.Powertrain.TractionBattery.Charging.TimeRemaining" {
							sensor.ChangeParams["slope"] = -0.5 // Decrease charging time
						} else if strings.Contains(path, "Battery") && (scenario == "urban_traffic" || scenario == "highway_cruising") {
							// Battery discharge patterns
							if scenario == "urban_traffic" {
								sensor.ChangeParams["slope"] = -0.1
							} else {
								sensor.ChangeParams["slope"] = -0.2 // Faster discharge on highway
							}
						} else {
							sensor.ChangeParams["slope"] = -0.04 // Default decrease
						}
						sensor.ChangeParams["min"] = vssSensor.Min
						sensor.ChangeParams["max"] = vssSensor.Max

					case "sinusoidal":
						sensor.ChangeParams["amplitude"] = (vssSensor.Max - vssSensor.Min) * 0.1
						sensor.ChangeParams["period"] = 5000.0
						sensor.ChangeParams["baseline"] = (vssSensor.Min + vssSensor.Max) * 0.5

					case "constant_with_noise":
						if val, ok := sensor.Value.(float64); ok {
							sensor.ChangeParams["baseline"] = val
						} else {
							sensor.ChangeParams["baseline"] = vssSensor.DefaultValue.(float64)
						}
						sensor.ChangeParams["noise"] = (vssSensor.Max - vssSensor.Min) * 0.01

					case "toggle":
						sensor.ChangeParams["toggle_probability"] = 0.05

					case "constant":
						// No parameters needed for constant

					default:
						// Keep existing parameters for unknown patterns
					}
				}
			}
		}
	}
}

//=============================================================================
// Scenario-specific configuration functions
//=============================================================================

// applyUrbanTrafficSettings applies additional settings for the urban traffic scenario
func applyUrbanTrafficSettings(vehicle *VehicleData) {
	// Urban traffic scenario - additional customizations beyond pattern updates

	// Speed settings (set specific initial values and parameters)
	if speedSensor, ok := vehicle.SensorsHighFreq["Vehicle.Speed"]; ok {
		speedSensor.Value = 40.0 + rand.Float64()*20.0
		if speedSensor.ChangePattern == "random_walk" {
			speedSensor.ChangeParams["step_size"] = 2.0
			speedSensor.ChangeParams["min"] = 0.0
			speedSensor.ChangeParams["max"] = 70.0
		}
	}

	// Location sensor (almost straight path) - customized linear progression
	if latSensor, ok := vehicle.SensorsHighFreq["Vehicle.CurrentLocation.Latitude"]; ok {
		if latSensor.ChangePattern == "random_walk" {
			latSensor.ChangeParams["step_size"] = 1.0
		}
	}

	if lonSensor, ok := vehicle.SensorsHighFreq["Vehicle.CurrentLocation.Longitude"]; ok {
		if lonSensor.ChangePattern == "random_walk" {
			lonSensor.ChangeParams["step_size"] = 1.0
		}
	}

	// Steering settings (frequent direction changes)
	if steeringSensor, ok := vehicle.SensorsHighFreq["Vehicle.Chassis.SteeringWheel.Angle"]; ok {
		if steeringSensor.ChangePattern == "sinusoidal" {
			steeringSensor.ChangeParams["amplitude"] = 15.0
			steeringSensor.ChangeParams["period"] = 10000.0
			steeringSensor.ChangeParams["baseline"] = 0.0
		}
	}

	// Acceleration settings (irregular acceleration/deceleration)
	if accelSensor, ok := vehicle.SensorsHighFreq["Vehicle.Acceleration.Longitudinal"]; ok {
		if accelSensor.ChangePattern == "random_walk" {
			accelSensor.ChangeParams["step_size"] = 0.5
			accelSensor.ChangeParams["min"] = -3.0
			accelSensor.ChangeParams["max"] = 2.0
		}
	}

	// Brake settings (frequent use)
	if brakeSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Brake.PedalPosition"]; ok {
		if brakeSensor.ChangePattern == "random_walk" {
			brakeSensor.ChangeParams["step_size"] = 5.0
		}
	}

	// Accelerator pedal settings
	if accelPedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Accelerator.PedalPosition"]; ok {
		if accelPedalSensor.ChangePattern == "random_walk" {
			accelPedalSensor.ChangeParams["step_size"] = 3.0
		}
	}

	// Turn signal settings (frequent use in urban environments)
	if turnLeftSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Left.IsActive"]; ok {
		if turnLeftSensor.ChangePattern == "toggle" {
			turnLeftSensor.ChangeParams["toggle_probability"] = 0.02
		}
	}

	if turnRightSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Right.IsActive"]; ok {
		if turnRightSensor.ChangePattern == "toggle" {
			turnRightSensor.ChangeParams["toggle_probability"] = 0.02
		}
	}
}

// applyHighwayCruisingSettings applies additional settings for the highway cruising scenario
func applyHighwayCruisingSettings(vehicle *VehicleData) {
	// Highway cruising scenario - high speed, steady direction, additional customizations

	// Speed settings (high speed)
	if speedSensor, ok := vehicle.SensorsHighFreq["Vehicle.Speed"]; ok {
		speedSensor.Value = 90.0 + rand.Float64()*20.0
		if speedSensor.ChangePattern == "constant_with_noise" {
			speedSensor.ChangeParams["baseline"] = 100.0
			speedSensor.ChangeParams["noise"] = 0.5
		}
	}

	// Steering settings (stable)
	if steeringSensor, ok := vehicle.SensorsHighFreq["Vehicle.Chassis.SteeringWheel.Angle"]; ok {
		if steeringSensor.ChangePattern == "constant_with_noise" {
			steeringSensor.ChangeParams["baseline"] = 0.0
			steeringSensor.ChangeParams["noise"] = 0.2
		}
	}

	// Acceleration settings (stable)
	if accelSensor, ok := vehicle.SensorsHighFreq["Vehicle.Acceleration.Longitudinal"]; ok {
		accelSensor.Value = 0.0
		if accelSensor.ChangePattern == "constant_with_noise" {
			accelSensor.ChangeParams["baseline"] = 0.0
			accelSensor.ChangeParams["noise"] = 0.1
		}
	}

	// Location sensor (almost straight path) - customized linear progression
	if latSensor, ok := vehicle.SensorsHighFreq["Vehicle.CurrentLocation.Latitude"]; ok {
		if latSensor.ChangePattern == "linear" {
			latSensor.ChangeParams["slope"] = 0.1 // Very slight change
		}
	}

	if lonSensor, ok := vehicle.SensorsHighFreq["Vehicle.CurrentLocation.Longitude"]; ok {
		if lonSensor.ChangePattern == "linear" {
			lonSensor.ChangeParams["slope"] = 0.1 // Very slight change
		}
	}

	// Brake pedal (rarely used on highway)
	if brakePedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Brake.PedalPosition"]; ok {
		brakePedalSensor.Value = 0.0
		if brakePedalSensor.ChangePattern == "toggle" {
			brakePedalSensor.ChangeParams["toggle_probability"] = 0.01
		}
	}

	// Accelerator pedal settings (maintain constant for highway)
	if accelPedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Accelerator.PedalPosition"]; ok {
		accelPedalSensor.Value = 25.0
		if accelPedalSensor.ChangePattern == "constant_with_noise" {
			accelPedalSensor.ChangeParams["baseline"] = 25.0
			accelPedalSensor.ChangeParams["noise"] = 0.5
		}
	}

	// Turn signal settings (rarely used on highways)
	if turnLeftSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Left.IsActive"]; ok {
		if turnLeftSensor.ChangePattern == "toggle" {
			turnLeftSensor.ChangeParams["toggle_probability"] = 0.005
		}
	}

	if turnRightSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Right.IsActive"]; ok {
		if turnRightSensor.ChangePattern == "toggle" {
			turnRightSensor.ChangeParams["toggle_probability"] = 0.005
		}
	}
}

// applyBatteryChargingSettings applies additional settings for the battery charging scenario
func applyBatteryChargingSettings(vehicle *VehicleData) {
	// Battery charging scenario - vehicle stationary, charging in progress

	// Speed and motion sensor settings (stationary) - force values to 0
	if speedSensor, ok := vehicle.SensorsHighFreq["Vehicle.Speed"]; ok {
		speedSensor.Value = 0.0
	}

	// Steering settings (stationary)
	if steeringSensor, ok := vehicle.SensorsHighFreq["Vehicle.Chassis.SteeringWheel.Angle"]; ok {
		steeringSensor.Value = 0.0
	}

	// Acceleration settings (stationary)
	if accelSensor, ok := vehicle.SensorsHighFreq["Vehicle.Acceleration.Longitudinal"]; ok {
		accelSensor.Value = 0.0
	}

	// Brake and accelerator settings (not in use)
	if brakeSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Brake.PedalPosition"]; ok {
		brakeSensor.Value = 0.0
	}

	if accelPedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Accelerator.PedalPosition"]; ok {
		accelPedalSensor.Value = 0.0
	}

	// Battery charging settings - enhance the charging pattern
	if batterySensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.StateOfCharge.Current"]; ok {
		// Keep current value but ensure charging pattern parameters
		if batterySensor.ChangePattern == "linear" {
			batterySensor.ChangeParams["slope"] = 0.2 // Gradual increase
		}
	}

	// Battery temperature settings (slight increase during charging)
	if tempSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Temperature"]; ok {
		if tempSensor.ChangePattern == "linear" {
			tempSensor.ChangeParams["slope"] = 0.001
		}
	}

	// Charging-related sensor settings
	if chargingSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Charging.IsCharging"]; ok {
		chargingSensor.Value = true
	}

	if ratesSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Charging.ChargingRate"]; ok {
		ratesSensor.Value = 11.0 + rand.Float64()*2.0 // 11-13kW charging
		if ratesSensor.ChangePattern == "sinusoidal" {
			ratesSensor.ChangeParams["amplitude"] = 1.0
			ratesSensor.ChangeParams["period"] = 15000.0
			ratesSensor.ChangeParams["baseline"] = 11.5
		}
	}

	if timeSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Charging.TimeRemaining"]; ok {
		if currentVal, ok := timeSensor.Value.(float64); !ok || currentVal <= 0 {
			timeSensor.Value = 120.0 + rand.Float64()*60.0 // 2-3 hours if not set
		}
		if timeSensor.ChangePattern == "linear" {
			timeSensor.ChangeParams["slope"] = -0.05 // Time decreases
		}
	}
}

// // applyUrbanTrafficSettings applies settings for the urban traffic scenario
// func applyUrbanTrafficSettings(vehicle *VehicleData) {
// 	// Urban traffic scenario - irregular speed, frequent direction changes

// 	// Speed settings (low to medium speed)
// 	if speedSensor, ok := vehicle.SensorsHighFreq["Vehicle.Speed"]; ok {
// 		speedSensor.Value = 40.0 + rand.Float64()*20.0
// 		speedSensor.ChangePattern = "random_walk"
// 		speedSensor.ChangeParams["step_size"] = 1.0
// 		speedSensor.ChangeParams["min"] = 0.0
// 		speedSensor.ChangeParams["max"] = 70.0
// 	}

// 	// Steering settings (frequent direction changes)
// 	if steeringSensor, ok := vehicle.SensorsHighFreq["Vehicle.Chassis.SteeringWheel.Angle"]; ok {
// 		steeringSensor.ChangePattern = "sinusoidal"
// 		steeringSensor.ChangeParams["amplitude"] = 15.0
// 		steeringSensor.ChangeParams["period"] = 10000.0
// 		steeringSensor.ChangeParams["baseline"] = 0.0
// 	}

// 	// Acceleration settings (irregular acceleration/deceleration)
// 	if accelSensor, ok := vehicle.SensorsHighFreq["Vehicle.Acceleration.Longitudinal"]; ok {
// 		accelSensor.ChangePattern = "random_walk"
// 		accelSensor.ChangeParams["step_size"] = 0.5
// 		accelSensor.ChangeParams["min"] = -3.0
// 		accelSensor.ChangeParams["max"] = 2.0
// 	}

// 	// Brake settings (frequent use)
// 	if brakeSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Brake.PedalPosition"]; ok {
// 		brakeSensor.ChangePattern = "random_walk"
// 		brakeSensor.ChangeParams["step_size"] = 5.0
// 		brakeSensor.ChangeParams["min"] = 0.0
// 		brakeSensor.ChangeParams["max"] = 80.0
// 	}

// 	// Accelerator pedal settings
// 	if accelPedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Accelerator.PedalPosition"]; ok {
// 		accelPedalSensor.ChangePattern = "random_walk"
// 		accelPedalSensor.ChangeParams["step_size"] = 3.0
// 		accelPedalSensor.ChangeParams["min"] = 0.0
// 		accelPedalSensor.ChangeParams["max"] = 60.0
// 	}

// 	// Battery discharge settings (consumption during urban driving)
// 	if batterySensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.StateOfCharge.Current"]; ok {
// 		batterySensor.ChangePattern = "linear"
// 		batterySensor.ChangeParams["slope"] = -0.1
// 		batterySensor.ChangeParams["min"] = 0.0
// 		batterySensor.ChangeParams["max"] = 100.0
// 	}

// 	// Fuel consumption settings
// 	if fuelSensor, ok := vehicle.SensorsLowFreq["Vehicle.Powertrain.FuelSystem.Level"]; ok {
// 		fuelSensor.ChangePattern = "linear"
// 		fuelSensor.ChangeParams["slope"] = -0.001
// 		fuelSensor.ChangeParams["min"] = 0.0
// 		fuelSensor.ChangeParams["max"] = 100.0
// 	}

// 	// Turn signal settings (frequent use in urban environments)
// 	if turnLeftSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Left.IsActive"]; ok {
// 		turnLeftSensor.ChangePattern = "toggle"
// 		turnLeftSensor.ChangeParams["toggle_probability"] = 0.02
// 	}

// 	if turnRightSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Right.IsActive"]; ok {
// 		turnRightSensor.ChangePattern = "toggle"
// 		turnRightSensor.ChangeParams["toggle_probability"] = 0.02
// 	}
// }

// // applyHighwayCruisingSettings applies settings for the highway cruising scenario
// func applyHighwayCruisingSettings(vehicle *VehicleData) {
// 	// Highway cruising scenario - high speed, steady direction

// 	// Speed settings (high speed)
// 	if speedSensor, ok := vehicle.SensorsHighFreq["Vehicle.Speed"]; ok {
// 		speedSensor.Value = 90.0 + rand.Float64()*20.0
// 		speedSensor.ChangePattern = "constant_with_noise"
// 		speedSensor.ChangeParams["baseline"] = 100.0
// 		speedSensor.ChangeParams["noise"] = 0.5
// 	}

// 	// Steering settings (stable)
// 	if steeringSensor, ok := vehicle.SensorsHighFreq["Vehicle.Chassis.SteeringWheel.Angle"]; ok {
// 		steeringSensor.ChangePattern = "constant_with_noise"
// 		steeringSensor.ChangeParams["baseline"] = 0.0
// 		steeringSensor.ChangeParams["noise"] = 0.2
// 	}

// 	// Acceleration settings (stable)
// 	if accelSensor, ok := vehicle.SensorsHighFreq["Vehicle.Acceleration.Longitudinal"]; ok {
// 		accelSensor.Value = 0.0
// 		accelSensor.ChangePattern = "constant_with_noise"
// 		accelSensor.ChangeParams["baseline"] = 0.0
// 		accelSensor.ChangeParams["noise"] = 0.1
// 	}

// 	// Location sensor (almost straight path)
// 	if latSensor, ok := vehicle.SensorsHighFreq["Vehicle.CurrentLocation.Latitude"]; ok {
// 		latSensor.ChangePattern = "linear"
// 		latSensor.ChangeParams["slope"] = 0.1 // Very slight change
// 		latSensor.ChangeParams["min"] = -180.0
// 		latSensor.ChangeParams["max"] = 180.0
// 	}

// 	if lonSensor, ok := vehicle.SensorsHighFreq["Vehicle.CurrentLocation.Longitude"]; ok {
// 		lonSensor.ChangePattern = "linear"    // Straight path
// 		lonSensor.ChangeParams["slope"] = 0.1 // Very slight change
// 		lonSensor.ChangeParams["min"] = -180.0
// 		lonSensor.ChangeParams["max"] = 180.0
// 	}

// 	if brakePedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Brake.PedalPosition"]; ok {
// 		brakePedalSensor.Value = 0.0
// 		brakePedalSensor.ChangePattern = "toggle"
// 		brakePedalSensor.ChangeParams["toggle_probability"] = 0.01
// 	}

// 	// Accelerator pedal settings (maintain constant)
// 	if accelPedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Accelerator.PedalPosition"]; ok {
// 		accelPedalSensor.Value = 25.0
// 		accelPedalSensor.ChangePattern = "constant_with_noise"
// 		accelPedalSensor.ChangeParams["baseline"] = 25.0
// 		accelPedalSensor.ChangeParams["noise"] = 0.5
// 	}

// 	// Battery discharge settings (faster consumption during high-speed driving)
// 	if batterySensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.StateOfCharge.Current"]; ok {
// 		batterySensor.ChangePattern = "linear"
// 		batterySensor.ChangeParams["slope"] = -0.2
// 		batterySensor.ChangeParams["min"] = 0.0
// 		batterySensor.ChangeParams["max"] = 100.0
// 	}

// 	// Fuel consumption settings (faster consumption during high-speed driving)
// 	if fuelSensor, ok := vehicle.SensorsLowFreq["Vehicle.Powertrain.FuelSystem.Level"]; ok {
// 		fuelSensor.ChangePattern = "linear"
// 		fuelSensor.ChangeParams["slope"] = -0.002
// 		fuelSensor.ChangeParams["min"] = 0.0
// 		fuelSensor.ChangeParams["max"] = 100.0
// 	}

// 	// Turn signal settings (rarely used on highways)
// 	if turnLeftSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Left.IsActive"]; ok {
// 		turnLeftSensor.ChangePattern = "toggle"
// 		turnLeftSensor.ChangeParams["toggle_probability"] = 0.005
// 	}

// 	if turnRightSensor, ok := vehicle.ActuatorsLowVar["Vehicle.Body.Lights.TurnSignal.Right.IsActive"]; ok {
// 		turnRightSensor.ChangePattern = "toggle"
// 		turnRightSensor.ChangeParams["toggle_probability"] = 0.005
// 	}
// }

// // applyBatteryChargingSettings applies settings for the battery charging scenario
// func applyBatteryChargingSettings(vehicle *VehicleData) {
// 	// Battery charging scenario - vehicle stationary, charging in progress

// 	// Speed and motion sensor settings (stationary)
// 	if speedSensor, ok := vehicle.SensorsHighFreq["Vehicle.Speed"]; ok {
// 		speedSensor.Value = 0.0
// 		speedSensor.ChangePattern = "constant"
// 	}

// 	// Steering settings (stationary)
// 	if steeringSensor, ok := vehicle.SensorsHighFreq["Vehicle.Chassis.SteeringWheel.Angle"]; ok {
// 		steeringSensor.Value = 0.0
// 		steeringSensor.ChangePattern = "constant"
// 	}

// 	// Acceleration settings (stationary)
// 	if accelSensor, ok := vehicle.SensorsHighFreq["Vehicle.Acceleration.Longitudinal"]; ok {
// 		accelSensor.Value = 0.0
// 		accelSensor.ChangePattern = "constant"
// 	}

// 	// Brake and accelerator settings (not in use)
// 	if brakeSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Brake.PedalPosition"]; ok {
// 		brakeSensor.Value = 0.0
// 		brakeSensor.ChangePattern = "constant"
// 	}

// 	if accelPedalSensor, ok := vehicle.ActuatorsHighVar["Vehicle.Chassis.Accelerator.PedalPosition"]; ok {
// 		accelPedalSensor.Value = 0.0
// 		accelPedalSensor.ChangePattern = "constant"
// 	}

// 	// Battery charging settings
// 	if batterySensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.StateOfCharge.Current"]; ok {
// 		batterySensor.ChangePattern = "linear"
// 		batterySensor.ChangeParams["slope"] = 0.1 // Gradual increase
// 		batterySensor.ChangeParams["min"] = 0.0
// 		batterySensor.ChangeParams["max"] = 100.0
// 	}

// 	// Battery temperature settings (slight increase during charging)
// 	if tempSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Temperature"]; ok {
// 		tempSensor.ChangePattern = "linear"
// 		tempSensor.ChangeParams["slope"] = 0.001
// 		tempSensor.ChangeParams["min"] = 20.0
// 		tempSensor.ChangeParams["max"] = 45.0
// 	}

// 	// Charging-related sensor settings
// 	if chargingSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Charging.IsCharging"]; ok {
// 		chargingSensor.Value = true
// 	}

// 	if ratesSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Charging.ChargingRate"]; ok {
// 		ratesSensor.Value = 11.0 + rand.Float64()*2.0 // 11-13kW charging
// 		ratesSensor.ChangePattern = "sinusoidal"
// 		ratesSensor.ChangeParams["amplitude"] = 1.0
// 		ratesSensor.ChangeParams["period"] = 15000.0
// 		ratesSensor.ChangeParams["baseline"] = 11.5
// 	}

// 	if timeSensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.Charging.TimeRemaining"]; ok {
// 		timeSensor.Value = 120.0 + rand.Float64()*60.0 // 2-3 hours
// 		timeSensor.ChangePattern = "linear"
// 		timeSensor.ChangeParams["slope"] = -0.05 // Time decreases
// 		timeSensor.ChangeParams["min"] = 0.0
// 		timeSensor.ChangeParams["max"] = 180.0
// 	}
// }

//=============================================================================
// Scenario State Management for Mixed Scenarios
//=============================================================================

// ScenarioState manages the current state of mixed scenarios
type ScenarioState struct {
	CurrentScenario  string  // Current active scenario
	PreviousScenario string  // Previous scenario (for battery_charging recovery)
	BatteryLevel     float64 // Current battery level for scenario switching
	LastSwitchTime   int64   // Time of last scenario switch (to prevent rapid switching)
	SwitchCooldown   int64   // Minimum time between switches (ms)
}

// NewScenarioState creates a new scenario state
func NewScenarioState(initialScenario string, initialBatteryLevel float64) *ScenarioState {
	return &ScenarioState{
		CurrentScenario:  initialScenario,
		PreviousScenario: initialScenario,
		BatteryLevel:     initialBatteryLevel,
		LastSwitchTime:   0,
		SwitchCooldown:   60000, // 30 seconds minimum between switches
	}
}

// UpdateScenarioState updates the scenario based on current conditions
func (state *ScenarioState) UpdateScenarioState(vehicle *VehicleData, currentTimeMs int64) bool {
	// Get current battery level
	if batterySensor, ok := vehicle.SensorsMedFreq["Vehicle.Powertrain.TractionBattery.StateOfCharge.Current"]; ok {
		if batteryValue, ok := batterySensor.Value.(float64); ok {
			state.BatteryLevel = batteryValue
		}
	}

	// Check if we can switch (cooldown period)
	if currentTimeMs-state.LastSwitchTime < state.SwitchCooldown {
		return false
	}

	oldScenario := state.CurrentScenario

	// Scenario switching logic
	switch state.CurrentScenario {
	case "urban_traffic":
		// From urban_traffic, can go to highway_cruising or battery_charging
		if state.BatteryLevel <= 10.0 {
			state.PreviousScenario = state.CurrentScenario
			state.CurrentScenario = "battery_charging"
		}
		// else if rand.Float64() < 0.01 { // Small chance to switch to highway
		// 	state.PreviousScenario = state.CurrentScenario
		// 	state.CurrentScenario = "highway_cruising"
		// }

	// case "highway_cruising":
	// 	// From highway_cruising, can go to urban_traffic or battery_charging
	// 	if state.BatteryLevel <= 10.0 {
	// 		state.PreviousScenario = state.CurrentScenario
	// 		state.CurrentScenario = "battery_charging"
	// 	} else if rand.Float64() < 0.01 { // Small chance to switch back to urban
	// 		state.PreviousScenario = state.CurrentScenario
	// 		state.CurrentScenario = "urban_traffic"
	// 	}

	case "battery_charging":
		// From battery_charging, return to previous scenario when battery is full
		if state.BatteryLevel >= 99.95 { // Round to 100%
			state.CurrentScenario = state.PreviousScenario
			// Don't update PreviousScenario here to maintain the chain
		}
	}

	// If scenario changed, update timestamp and apply new settings
	if oldScenario != state.CurrentScenario {
		state.LastSwitchTime = currentTimeMs
		applyScenarioSettingsForMixed(vehicle, state.CurrentScenario)
		return true
	}

	return false
}

// applyScenarioSettingsForMixed applies scenario settings specifically for mixed scenario transitions
func applyScenarioSettingsForMixed(vehicle *VehicleData, currentScenario string) {
	// Update all sensors with their scenario-specific patterns
	updateAllSensorPatterns(vehicle, currentScenario)

	// Apply scenario-specific customizations
	switch currentScenario {
	case "urban_traffic":
		applyUrbanTrafficSettings(vehicle)
	case "highway_cruising":
		applyHighwayCruisingSettings(vehicle)
	case "battery_charging":
		applyBatteryChargingSettings(vehicle)
	}
}
