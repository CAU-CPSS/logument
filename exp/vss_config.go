package exp

import (
	"time"
)

// Sensor category definitions
const (
	SensorTypeHigh   = "high_frequency"   // High-frequency sensors (10-50ms updates)
	SensorTypeMedium = "medium_frequency" // Medium-frequency sensors (100-500ms updates)
	SensorTypeLow    = "low_frequency"    // Low-frequency sensors (1000ms+ updates)
	ActuatorTypeHigh = "high_variable"    // High variability actuators
	ActuatorTypeLow  = "low_variable"     // Low variability actuators
	AttributeType    = "attribute"        // Attributes that rarely change
)

//=============================================================================
// Scenario and vehicle data structures
//=============================================================================

// VehicleData stores the sensor, actuator, and attribute values of a vehicle
type VehicleData struct {
	SensorsHighFreq  map[string]*SensorData // High-frequency update sensors (e.g., GPS, speed)
	SensorsMedFreq   map[string]*SensorData // Medium-frequency update sensors (e.g., temperature, battery)
	SensorsLowFreq   map[string]*SensorData // Low-frequency update sensors (e.g., fuel, tire pressure)
	ActuatorsHighVar map[string]*SensorData // High variability actuators (e.g., brake, accelerator)
	ActuatorsLowVar  map[string]*SensorData // Low variability actuators (e.g., windows, doors)
	Attributes       map[string]*SensorData // Attributes that rarely change (e.g., vehicle ID, model)
}

// SensorData stores individual sensor/actuator data
type SensorData struct {
	Path           string             // Sensor path
	Value          any                // Current value
	Type           string             // Value type (number, string, boolean)
	Timestamp      int64              // Last update timestamp
	UpdateInterval int                // Update interval (milliseconds)
	ChangePattern  string             // Change pattern (constant, linear, sinusoidal, random, triggered)
	ChangeParams   map[string]float64 // Change pattern parameters
	ChangeCounter  int                // Change counter (for tracking elapsed time)
}

// VSSSensor defines a VSS-based vehicle sensor, actuator, or attribute
type VSSSensor struct {
	Path           string             // VSS path
	DefaultValue   any                // Default value
	Type           string             // Value type (number, string, boolean)
	Category       string             // Sensor category (high_frequency, medium_frequency, etc.)
	UpdateInterval int                // Update interval (milliseconds)
	Description    string             // Description
	Unit           string             // Unit (e.g., km/h, celsius)
	Min            float64            // Minimum value (for numeric types)
	Max            float64            // Maximum value (for numeric types)
	ChangePatterns map[string]string  // Change patterns for different scenarios
	ChangeParams   map[string]float64 // Change pattern parameters
}

//=============================================================================
// VSS Sensor Definitions
//=============================================================================

// VSSSensors is a map defining sensors based on the Vehicle Signal Specification (VSS)
var VSSSensors = map[string]*VSSSensor{
	// Location sensors (high frequency)
	"Vehicle.CurrentLocation.Latitude": {
		Path:           "Vehicle.CurrentLocation.Latitude",
		DefaultValue:   37.7749,
		Type:           "number",
		Category:       SensorTypeHigh,
		UpdateInterval: 50,
		Description:    "Current latitude of the vehicle",
		Unit:           "degrees",
		Min:            -90.0,
		Max:            90.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "linear",
			"battery_charging": "constant",
		},
	},
	"Vehicle.CurrentLocation.Longitude": {
		Path:           "Vehicle.CurrentLocation.Longitude",
		DefaultValue:   -122.4194,
		Type:           "number",
		Category:       SensorTypeHigh,
		UpdateInterval: 50,
		Description:    "Current longitude of the vehicle",
		Unit:           "degrees",
		Min:            -180.0,
		Max:            180.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "linear",
			"battery_charging": "constant",
		},
	},
	"Vehicle.CurrentLocation.Altitude": {
		Path:           "Vehicle.CurrentLocation.Altitude",
		DefaultValue:   10.0,
		Type:           "number",
		Category:       SensorTypeHigh,
		UpdateInterval: 100,
		Description:    "Current altitude of the vehicle",
		Unit:           "m",
		Min:            -500.0,
		Max:            9000.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// Speed sensor (high frequency)
	"Vehicle.Speed": {
		Path:           "Vehicle.Speed",
		DefaultValue:   0.0,
		Type:           "number",
		Category:       SensorTypeHigh,
		UpdateInterval: 20,
		Description:    "Current speed of the vehicle",
		Unit:           "km/h",
		Min:            0.0,
		Max:            220.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// Acceleration sensors (high frequency)
	"Vehicle.Acceleration.Longitudinal": {
		Path:           "Vehicle.Acceleration.Longitudinal",
		DefaultValue:   0.0,
		Type:           "number",
		Category:       SensorTypeHigh,
		UpdateInterval: 20,
		Description:    "Longitudinal acceleration of the vehicle",
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
		Description:    "Lateral acceleration of the vehicle",
		Unit:           "m/s²",
		Min:            -20.0,
		Max:            20.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "sinusoidal",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// Steering angle sensor (high frequency)
	"Vehicle.Chassis.SteeringWheel.Angle": {
		Path:           "Vehicle.Chassis.SteeringWheel.Angle",
		DefaultValue:   0.0,
		Type:           "number",
		Category:       SensorTypeHigh,
		UpdateInterval: 25,
		Description:    "Current steering wheel angle",
		Unit:           "degrees",
		Min:            -180.0,
		Max:            180.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// Engine speed (medium frequency)
	"Vehicle.Powertrain.CombustionEngine.Speed": {
		Path:           "Vehicle.Powertrain.CombustionEngine.Speed",
		DefaultValue:   800.0,
		Type:           "number",
		Category:       SensorTypeMedium,
		UpdateInterval: 100,
		Description:    "Engine rotational speed",
		Unit:           "rpm",
		Min:            0.0,
		Max:            8000.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// Battery sensors (medium frequency)
	"Vehicle.Powertrain.TractionBattery.StateOfCharge.Current": {
		Path:           "Vehicle.Powertrain.TractionBattery.StateOfCharge.Current",
		DefaultValue:   80.0,
		Type:           "number",
		Category:       SensorTypeMedium,
		UpdateInterval: 200,
		Description:    "Battery state of charge",
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
		Description:    "Battery temperature",
		Unit:           "celsius",
		Min:            -20.0,
		Max:            80.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "constant_with_noise",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "linear",
		},
	},

	// Fuel level (low frequency)
	"Vehicle.Powertrain.FuelSystem.Level": {
		Path:           "Vehicle.Powertrain.FuelSystem.Level",
		DefaultValue:   80.0,
		Type:           "number",
		Category:       SensorTypeLow,
		UpdateInterval: 5000,
		Description:    "Fuel tank level",
		Unit:           "percent",
		Min:            0.0,
		Max:            100.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "linear",
			"highway_cruising": "linear",
			"battery_charging": "constant",
		},
	},

	// Tire sensors (low frequency)
	"Vehicle.Chassis.Axle.Row1.Wheel.Left.Tire.Pressure": {
		Path:           "Vehicle.Chassis.Axle.Row1.Wheel.Left.Tire.Pressure",
		DefaultValue:   2.2,
		Type:           "number",
		Category:       SensorTypeLow,
		UpdateInterval: 10000,
		Description:    "Left front tire pressure",
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
		Description:    "Right front tire pressure",
		Unit:           "bar",
		Min:            0.0,
		Max:            4.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "constant_with_noise",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// Actuator - Brake pedal (high variability)
	"Vehicle.Chassis.Brake.PedalPosition": {
		Path:           "Vehicle.Chassis.Brake.PedalPosition",
		DefaultValue:   0.0,
		Type:           "number",
		Category:       ActuatorTypeHigh,
		UpdateInterval: 50,
		Description:    "Brake pedal position",
		Unit:           "percent",
		Min:            0.0,
		Max:            100.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "toggle",
			"battery_charging": "constant",
		},
	},

	// Actuator - Accelerator pedal (high variability)
	"Vehicle.Chassis.Accelerator.PedalPosition": {
		Path:           "Vehicle.Chassis.Accelerator.PedalPosition",
		DefaultValue:   0.0,
		Type:           "number",
		Category:       ActuatorTypeHigh,
		UpdateInterval: 50,
		Description:    "Accelerator pedal position",
		Unit:           "percent",
		Min:            0.0,
		Max:            100.0,
		ChangePatterns: map[string]string{
			"urban_traffic":    "random_walk",
			"highway_cruising": "constant_with_noise",
			"battery_charging": "constant",
		},
	},

	// Actuator - Headlights (low variability)
	"Vehicle.Body.Lights.Headlight.IsActive": {
		Path:           "Vehicle.Body.Lights.Headlight.IsActive",
		DefaultValue:   false,
		Type:           "boolean",
		Category:       ActuatorTypeLow,
		UpdateInterval: 5000,
		Description:    "Headlight status",
		Unit:           "",
		ChangePatterns: map[string]string{
			"urban_traffic":    "toggle",
			"highway_cruising": "toggle",
			"battery_charging": "constant",
		},
	},

	// Actuator - Turn signals (low variability)
	"Vehicle.Body.Lights.TurnSignal.Left.IsActive": {
		Path:           "Vehicle.Body.Lights.TurnSignal.Left.IsActive",
		DefaultValue:   false,
		Type:           "boolean",
		Category:       ActuatorTypeLow,
		UpdateInterval: 1000,
		Description:    "Left turn signal status",
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
		Description:    "Right turn signal status",
		Unit:           "",
		ChangePatterns: map[string]string{
			"urban_traffic":    "toggle",
			"highway_cruising": "toggle",
			"battery_charging": "constant",
		},
	},

	// Attributes - Vehicle identification
	"Vehicle.VehicleIdentification.VIN": {
		Path:           "Vehicle.VehicleIdentification.VIN",
		DefaultValue:   "1HGCM82633A123456",
		Type:           "string",
		Category:       AttributeType,
		UpdateInterval: 86400000, // 24 hours
		Description:    "Vehicle Identification Number (VIN)",
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
		UpdateInterval: 86400000, // 24 hours
		Description:    "Vehicle model name",
		Unit:           "",
		ChangePatterns: map[string]string{
			"urban_traffic":    "constant",
			"highway_cruising": "constant",
			"battery_charging": "constant",
		},
	},

	// Charging scenario-specific sensors
	"Vehicle.Powertrain.TractionBattery.Charging.IsCharging": {
		Path:           "Vehicle.Powertrain.TractionBattery.Charging.IsCharging",
		DefaultValue:   false,
		Type:           "boolean",
		Category:       SensorTypeMedium,
		UpdateInterval: 1000,
		Description:    "Battery charging status",
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
		Description:    "Battery charging rate",
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
		Description:    "Time remaining until charging is complete",
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
				sensorData.ChangeParams["slope"] = 0.05 // Battery increase during charging
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
