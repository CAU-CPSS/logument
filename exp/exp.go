package exp

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

//=============================================================================
// Constants and Global Variables
//=============================================================================

// Simulation parameters
const (
	outputDir             = "./results/patch_experiment"
	simulationDurationSec = 300 // 5-minute simulation
	simulationTimeStepMs  = 100 // Update every 100ms
	numHighFreqSensors    = 10  // Number of high-frequency sensors
	numMedFreqSensors     = 15  // Number of medium-frequency sensors
	numLowFreqSensors     = 20  // Number of low-frequency sensors
	numHighVarActuators   = 5   // Number of high-variability actuators
	numLowVarActuators    = 10  // Number of low-variability actuators
	numAttributes         = 8   // Number of attributes
)

// Scenario types
var scenarios = []string{
	"urban_traffic",    // Urban traffic
	"highway_cruising", // Highway cruising
	"battery_charging", // Battery charging
}

// JSON document (value-timestamp structure)
type JsonDoc map[string]interface{}

// TSON document (timestamp as metadata)
type TsonDoc map[string]interface{}

// JSON patch
type JsonPatch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// TSON patch
type TsonPatch struct {
	Op        string      `json:"op"`
	Path      string      `json:"path"`
	Value     interface{} `json:"value"`
	Timestamp int64       `json:"timestamp"`
}

// Experiment result
type ExperimentResult struct {
	Timestamp            int64   // Experiment timestamp
	Scenario             string  // Scenario name
	SimulationTimeMs     int64   // Simulation time (milliseconds)
	TotalUpdates         int     // Total number of updates
	ValueChanges         int     // Number of value changes
	TimestampOnlyChanges int     // Number of timestamp-only changes
	JsonPatchCount       int     // Number of JSON patches
	TsonPatchCount       int     // Number of TSON patches
	JsonPatchSize        int     // Size of JSON patches (bytes)
	TsonPatchSize        int     // Size of TSON patches (bytes)
	JsonProcessingTimeNs int64   // JSON processing time (nanoseconds)
	TsonProcessingTimeNs int64   // TSON processing time (nanoseconds)
	JsonBandwidthUsage   float64 // JSON bandwidth usage (KB/s)
	TsonBandwidthUsage   float64 // TSON bandwidth usage (KB/s)
	CumulativeTimeVec    []int64 // Cumulative time vector
	CumulativeJsonVec    []int   // Cumulative JSON patch count vector
	CumulativeTsonVec    []int   // Cumulative TSON patch count vector
}

// Timing record
type TimingRecord struct {
	UpdateIndex          int    // Update index
	TimeMs               int64  // Simulation time (milliseconds)
	SensorPath           string // Sensor path
	JsonProcessingTimeNs int64  // JSON processing time (nanoseconds)
	TsonProcessingTimeNs int64  // TSON processing time (nanoseconds)
	ValueChanged         bool   // Whether the value changed
}

//=============================================================================
// Main Experiment Function
//=============================================================================

// RealworldScenario performs experiments based on real-world automotive scenarios
func RealworldScenario() {
	// Create output directory
	os.MkdirAll(outputDir, os.ModePerm)

	// Perform experiments for all scenarios
	for _, scenario := range scenarios {
		fmt.Printf("\n=== Starting experiment for scenario: %s ===\n", scenario)

		// Create scenario-specific result directory
		scenarioDir := filepath.Join(outputDir, scenario)
		os.MkdirAll(scenarioDir, os.ModePerm)

		// Create results file
		resultsFile, err := os.Create(filepath.Join(scenarioDir, "results.csv"))
		if err != nil {
			fmt.Printf("Error creating results file for %s: %v\n", scenario, err)
			continue
		}
		defer resultsFile.Close()

		csvWriter := csv.NewWriter(resultsFile)
		defer csvWriter.Flush()

		// Write CSV headers
		headers := []string{
			"Timestamp",
			"SimulationTimeMs",
			"TotalUpdates",
			"ValueChanges",
			"TimestampOnlyChanges",
			"JsonPatchCount",
			"TsonPatchCount",
			"JsonPatchSize",
			"TsonPatchSize",
			"JsonProcessingTimeNs",
			"TsonProcessingTimeNs",
			"JsonBandwidthUsage",
			"TsonBandwidthUsage",
			"PatchReduction",
		}
		csvWriter.Write(headers)

		// File for cumulative data
		cumulativeFile, err := os.Create(filepath.Join(scenarioDir, "cumulative.csv"))
		if err != nil {
			fmt.Printf("Error creating cumulative data file for %s: %v\n", scenario, err)
			continue
		}
		defer cumulativeFile.Close()

		cumWriter := csv.NewWriter(cumulativeFile)
		defer cumWriter.Flush()

		// Cumulative data headers
		cumWriter.Write([]string{
			"TimeMs",
			"JsonPatchCount",
			"TsonPatchCount",
		})

		// File for timing data
		timingFile, err := os.Create(filepath.Join(scenarioDir, "timing.csv"))
		if err != nil {
			fmt.Printf("Error creating timing file for %s: %v\n", scenario, err)
			continue
		}
		defer timingFile.Close()

		timingWriter := csv.NewWriter(timingFile)
		defer timingWriter.Flush()

		// Timing data headers
		timingWriter.Write([]string{
			"UpdateIndex",
			"TimeMs",
			"SensorPath",
			"JsonProcessingTimeNs",
			"TsonProcessingTimeNs",
			"ValueChanged",
			"ProcessingTimeDiffPct",
		})

		// Run the experiment
		results, timingData := runRealisticExperiment(scenario, timingWriter)
		fmt.Printf("Timing Data: %d\n", len(timingData))

		fmt.Printf("\n--- Results for scenario: %s ---\n", scenario)
		fmt.Printf("Total updates: %d\n", results.TotalUpdates)
		fmt.Printf("Value changes: %d (%.1f%%)\n",
			results.ValueChanges,
			float64(results.ValueChanges)*100/float64(results.TotalUpdates))
		fmt.Printf("Timestamp-only changes: %d (%.1f%%)\n",
			results.TimestampOnlyChanges,
			float64(results.TimestampOnlyChanges)*100/float64(results.TotalUpdates))
		fmt.Printf("JSON patches: %d\n", results.JsonPatchCount)
		fmt.Printf("TSON patches: %d\n", results.TsonPatchCount)
		fmt.Printf("Patch reduction: %.1f%%\n",
			(1-float64(results.TsonPatchCount)/float64(results.JsonPatchCount))*100)
		fmt.Printf("JSON processing time: %.2f ms\n", float64(results.JsonProcessingTimeNs)/1e6)
		fmt.Printf("TSON processing time: %.2f ms\n", float64(results.TsonProcessingTimeNs)/1e6)
		fmt.Printf("Processing time difference: %.1f%%\n",
			(1-float64(results.TsonProcessingTimeNs)/float64(results.JsonProcessingTimeNs))*100)

		// Save results to CSV
		reductionRate := (1 - float64(results.TsonPatchCount)/float64(results.JsonPatchCount)) * 100

		row := []string{
			fmt.Sprintf("%d", results.Timestamp),
			fmt.Sprintf("%d", results.SimulationTimeMs),
			fmt.Sprintf("%d", results.TotalUpdates),
			fmt.Sprintf("%d", results.ValueChanges),
			fmt.Sprintf("%d", results.TimestampOnlyChanges),
			fmt.Sprintf("%d", results.JsonPatchCount),
			fmt.Sprintf("%d", results.TsonPatchCount),
			fmt.Sprintf("%d", results.JsonPatchSize),
			fmt.Sprintf("%d", results.TsonPatchSize),
			fmt.Sprintf("%d", results.JsonProcessingTimeNs),
			fmt.Sprintf("%d", results.TsonProcessingTimeNs),
			fmt.Sprintf("%.2f", results.JsonBandwidthUsage),
			fmt.Sprintf("%.2f", results.TsonBandwidthUsage),
			fmt.Sprintf("%.2f", reductionRate),
		}
		csvWriter.Write(row)

		// Save cumulative data
		timeVector, jsonVector, tsonVector := results.CumulativeTimeVec, results.CumulativeJsonVec, results.CumulativeTsonVec
		for i := range timeVector {
			cumWriter.Write([]string{
				fmt.Sprintf("%d", timeVector[i]),
				fmt.Sprintf("%d", jsonVector[i]),
				fmt.Sprintf("%d", tsonVector[i]),
			})
		}

		fmt.Printf("\nExperiment for scenario %s completed. Results saved:\n", scenario)
		fmt.Printf("- Summary results: %s\n", filepath.Join(scenarioDir, "results.csv"))
		fmt.Printf("- Cumulative patches: %s\n", filepath.Join(scenarioDir, "cumulative.csv"))
		fmt.Printf("- Timing data: %s\n", filepath.Join(scenarioDir, "timing.csv"))
	}

	fmt.Println("\nAll experiments completed.")
}
// runRealisticExperiment performs an experiment for the given scenario
func runRealisticExperiment(scenario string, timingWriter *csv.Writer) (ExperimentResult, []TimingRecord) {
	fmt.Printf("Starting simulation for scenario '%s'...\n", scenario)

	// Create a data storage instance for this scenario
	dataStorage := NewDataStorage(outputDir, scenario)

	// Create a sensor data recorder
	sensorRecorder, err := NewSensorDataRecorder(outputDir, scenario)
	if err != nil {
		fmt.Printf("Failed to create sensor data recorder: %v\n", err)
	} else {
		defer sensorRecorder.Close()
	}

	// 1. Generate initial data based on VSS
	vehicle := InitializeVehicleData(scenario)

	// 2. Initialize JSON and TSON documents
	jsonDoc := vehicleDataToJsonDoc(vehicle)
	tsonDoc := vehicleDataToTsonDoc(vehicle)

	// Save initial documents
	if err := dataStorage.SaveInitialJSON(jsonDoc); err != nil {
		fmt.Printf("Error saving initial JSON: %v\n", err)
	}
	if err := dataStorage.SaveInitialTSON(tsonDoc); err != nil {
		fmt.Printf("Error saving initial TSON: %v\n", err)
	}

	// 3. Initialize variables for simulation results
	result := ExperimentResult{
		Timestamp:        time.Now().Unix(),
		Scenario:         scenario,
		SimulationTimeMs: simulationDurationSec * 1000,
	}

	jsonPatches := make([]JsonPatch, 0)
	tsonPatches := make([]TsonPatch, 0)

	// Initialize vectors for collecting timestamps
	timeVec := make([]int64, 0, simulationDurationSec*10) // Save every 10 seconds
	jsonPatchesVec := make([]int, 0, simulationDurationSec*10)
	tsonPatchesVec := make([]int, 0, simulationDurationSec*10)

	// Array for recording processing times
	timingRecords := make([]TimingRecord, 0)

	// Estimate the total number of updates
	totalExpectedUpdates := calculateExpectedUpdateCount(vehicle, simulationDurationSec)
	fmt.Printf("Estimated number of updates: approximately %d\n", totalExpectedUpdates)

	// 4. Run the simulation
	simulationStart := time.Now()
	jsonTotalProcessingTime := int64(0)
	tsonTotalProcessingTime := int64(0)

	// Counters for progress display
	updateCounter := 0
	valueChangeCounter := 0
	timestampChangeCounter := 0

	// Simulation time step (ms)
	for currentTimeMs := int64(0); currentTimeMs < simulationDurationSec*1000; currentTimeMs += simulationTimeStepMs {
		// Update timestamp vector (every 10 seconds)
		if currentTimeMs%(10*1000) == 0 {
			timeVec = append(timeVec, currentTimeMs)
			jsonPatchesVec = append(jsonPatchesVec, len(jsonPatches))
			tsonPatchesVec = append(tsonPatchesVec, len(tsonPatches))
		}

		// Display progress (every 10%)
		if currentTimeMs%(simulationDurationSec*100) == 0 {
			progressPct := float64(currentTimeMs) / float64(simulationDurationSec*1000) * 100
			fmt.Printf("\rSimulation progress: %.1f%% (time: %d ms, updates: %d, changes: %d)",
				progressPct, currentTimeMs, updateCounter, valueChangeCounter)
		}

		// Record major sensor data
		if sensorRecorder != nil {
			sensorRecorder.RecordSensorData(currentTimeMs, vehicle)
		}

		// Check and update each sensor
		for _, sensorGroups := range []map[string]*SensorData{
			vehicle.SensorsHighFreq,
			vehicle.SensorsMedFreq,
			vehicle.SensorsLowFreq,
			vehicle.ActuatorsHighVar,
			vehicle.ActuatorsLowVar,
			vehicle.Attributes,
		} {
			for path, sensor := range sensorGroups {
				// Check if this sensor should be updated at this time step
				if currentTimeMs%int64(sensor.UpdateInterval) == 0 {
					updateCounter++

					// Calculate new value and timestamp
					newValue := calculateNewValue(sensor, currentTimeMs, scenario)
					newTimestamp := simulationStart.UnixNano() + currentTimeMs*1000000 // ms -> ns

					// Check if the value has changed
					valueChanged := !isEqual(sensor.Value, newValue)

					if valueChanged {
						valueChangeCounter++
						sensor.Value = newValue
					} else {
						timestampChangeCounter++
					}

					sensor.Timestamp = newTimestamp

					// Start JSON processing time
					jsonStartTime := time.Now()

					// Generate JSON patch (always update timestamp, update value if changed)
					// Timestamp patch
					jsonPatches = append(jsonPatches, JsonPatch{
						Op:    "replace",
						Path:  path + "/timestamp",
						Value: newTimestamp,
					})

					// Add value patch if the value has changed
					if valueChanged {
						jsonPatches = append(jsonPatches, JsonPatch{
							Op:    "replace",
							Path:  path + "/value",
							Value: newValue,
						})
					}

					// Update JSON document
					updateJsonDoc(jsonDoc, path, newValue, newTimestamp)

					// Measure JSON processing time
					jsonProcessingTime := time.Since(jsonStartTime).Nanoseconds()
					jsonTotalProcessingTime += jsonProcessingTime

					// Start TSON processing time
					tsonStartTime := time.Now()

					// Update TSON+TestSet (generate patch only if the value has changed)
					if valueChanged {
						// Generate TSON patch
						tsonPatches = append(tsonPatches, TsonPatch{
							Op:        "replace",
							Path:      path,
							Value:     newValue,
							Timestamp: newTimestamp,
						})

						// Update TSON document
						updateTsonDoc(tsonDoc, path, newValue, newTimestamp)
					}

					// Measure TSON processing time
					tsonProcessingTime := time.Since(tsonStartTime).Nanoseconds()
					tsonTotalProcessingTime += tsonProcessingTime

					// Save processing time record
					timingRecord := TimingRecord{
						UpdateIndex:          updateCounter,
						TimeMs:               currentTimeMs,
						SensorPath:           path,
						JsonProcessingTimeNs: jsonProcessingTime,
						TsonProcessingTimeNs: tsonProcessingTime,
						ValueChanged:         valueChanged,
					}
					timingRecords = append(timingRecords, timingRecord)

					// Write processing time to CSV
					timingDiffPct := 0.0
					if jsonProcessingTime > 0 {
						timingDiffPct = (1.0 - float64(tsonProcessingTime)/float64(jsonProcessingTime)) * 100.0
					}

					timingWriter.Write([]string{
						fmt.Sprintf("%d", updateCounter),
						fmt.Sprintf("%d", currentTimeMs),
						path,
						fmt.Sprintf("%d", jsonProcessingTime),
						fmt.Sprintf("%d", tsonProcessingTime),
						fmt.Sprintf("%t", valueChanged),
						fmt.Sprintf("%.2f", timingDiffPct),
					})

					// Flush buffer periodically
					if updateCounter%100 == 0 {
						timingWriter.Flush()
					}
				}
			}
		}
	}

	// Newline (after progress display)
	fmt.Println()

	// Calculate JSON patch size
	jsonPatchBytes, _ := json.Marshal(jsonPatches)
	jsonPatchSize := len(jsonPatchBytes)

	// Calculate TSON patch size
	tsonPatchBytes, _ := json.Marshal(tsonPatches)
	tsonPatchSize := len(tsonPatchBytes)

	// Calculate bandwidth usage (KB/s)
	simulationTimeSeconds := float64(simulationDurationSec)
	jsonBandwidth := float64(jsonPatchSize) / simulationTimeSeconds / 1024
	tsonBandwidth := float64(tsonPatchSize) / simulationTimeSeconds / 1024

	// Save results
	result.TotalUpdates = updateCounter
	result.ValueChanges = valueChangeCounter
	result.TimestampOnlyChanges = timestampChangeCounter
	result.JsonPatchCount = len(jsonPatches)
	result.TsonPatchCount = len(tsonPatches)
	result.JsonPatchSize = jsonPatchSize
	result.TsonPatchSize = tsonPatchSize
	result.JsonProcessingTimeNs = jsonTotalProcessingTime
	result.TsonProcessingTimeNs = tsonTotalProcessingTime
	result.JsonBandwidthUsage = jsonBandwidth
	result.TsonBandwidthUsage = tsonBandwidth
	result.CumulativeTimeVec = timeVec
	result.CumulativeJsonVec = jsonPatchesVec
	result.CumulativeTsonVec = tsonPatchesVec

	// At the end of the simulation:
	// Save final documents
	if err := dataStorage.SaveFinalJSON(jsonDoc); err != nil {
		fmt.Printf("Error saving final JSON: %v\n", err)
	}
	if err := dataStorage.SaveFinalTSON(tsonDoc); err != nil {
		fmt.Printf("Error saving final TSON: %v\n", err)
	}

	// Save all patches
	if err := dataStorage.SaveAllJSONPatches(jsonPatches); err != nil {
		fmt.Printf("Error saving JSON patches: %v\n", err)
	}
	if err := dataStorage.SaveAllTSONPatches(tsonPatches); err != nil {
		fmt.Printf("Error saving TSON patches: %v\n", err)
	}

	fmt.Printf("Simulation completed. Total updates: %d, value changes: %d (%.1f%%)\n",
		updateCounter, valueChangeCounter,
		float64(valueChangeCounter)*100/float64(updateCounter))

	return result, timingRecords
}
