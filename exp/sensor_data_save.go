package exp

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// KeySensorPaths is a list of key sensor paths to track for graph visualization
var KeySensorPaths = []string{
	"Vehicle.CurrentLocation.Latitude",
	"Vehicle.CurrentLocation.Longitude",
	"Vehicle.Speed",
	"Vehicle.Powertrain.TractionBattery.StateOfCharge.Current",
}

// SensorDataRecorder records key sensor data into CSV files
type SensorDataRecorder struct {
	OutputDir  string                 // Output directory
	Scenario   string                 // Scenario name
	CSVFiles   map[string]*os.File    // Files for each sensor
	CSVWriters map[string]*csv.Writer // CSV writers for each sensor
	StartTime  time.Time              // Experiment start time
}

// NewSensorDataRecorder creates a new SensorDataRecorder instance
func NewSensorDataRecorder(baseDir, scenario string) (*SensorDataRecorder, error) {
	// Create a directory for the scenario
	sensorDataDir := filepath.Join(baseDir, scenario, "sensor_data")
	if err := os.MkdirAll(sensorDataDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("Failed to create sensor data directory: %v", err)
	}

	recorder := &SensorDataRecorder{
		OutputDir:  sensorDataDir,
		Scenario:   scenario,
		CSVFiles:   make(map[string]*os.File),
		CSVWriters: make(map[string]*csv.Writer),
		StartTime:  time.Now(),
	}

	// Create a CSV file for each key sensor
	for _, path := range KeySensorPaths {
		// Extract the last part of the path to use as the file name
		shortName := filepath.Base(path)
		fileName := filepath.Join(sensorDataDir, shortName+".csv")

		file, err := os.Create(fileName)
		if err != nil {
			// Close all open files in case of an error
			recorder.Close()
			return nil, fmt.Errorf("Failed to create CSV file %s: %v", fileName, err)
		}

		recorder.CSVFiles[path] = file
		writer := csv.NewWriter(file)
		recorder.CSVWriters[path] = writer

		// Write header
		writer.Write([]string{"TimestampMs", "SimulationTimeMs", "Value"})
		writer.Flush()
	}

	return recorder, nil
}

// RecordSensorData records a sensor data point
func (r *SensorDataRecorder) RecordSensorData(simulationTimeMs int64, vehicle *VehicleData) error {
	// For each key sensor
	for _, path := range KeySensorPaths {
		// Extract sensor value
		var sensorValue any
		var found bool

		// Check the category of the sensor and extract its value
		if sensor, ok := vehicle.SensorsHighFreq[path]; ok {
			sensorValue = sensor.Value
			found = true
		} else if sensor, ok := vehicle.SensorsMedFreq[path]; ok {
			sensorValue = sensor.Value
			found = true
		} else if sensor, ok := vehicle.SensorsLowFreq[path]; ok {
			sensorValue = sensor.Value
			found = true
		} else if sensor, ok := vehicle.ActuatorsHighVar[path]; ok {
			sensorValue = sensor.Value
			found = true
		} else if sensor, ok := vehicle.ActuatorsLowVar[path]; ok {
			sensorValue = sensor.Value
			found = true
		} else if sensor, ok := vehicle.Attributes[path]; ok {
			sensorValue = sensor.Value
			found = true
		}

		if !found {
			continue // Sensor not found
		}

		// Write to CSV
		if writer, ok := r.CSVWriters[path]; ok {
			// Timestamp, simulation time, value
			now := time.Now()
			elapsedMs := now.Sub(r.StartTime).Milliseconds()

			writer.Write([]string{
				fmt.Sprintf("%d", elapsedMs),
				fmt.Sprintf("%d", simulationTimeMs),
				fmt.Sprintf("%v", sensorValue),
			})

			// Periodically flush the buffer
			if simulationTimeMs%1000 == 0 {
				writer.Flush()
			}
		}
	}

	return nil
}

// Close closes all CSV files
func (r *SensorDataRecorder) Close() {
	// Flush all CSV buffers
	for _, writer := range r.CSVWriters {
		writer.Flush()
	}

	// Close all files
	for _, file := range r.CSVFiles {
		file.Close()
	}

	fmt.Printf("Sensor data recording completed: %s\n", r.OutputDir)
}
