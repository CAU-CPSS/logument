package exp

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Temporal query experiment parameters
const (
	temporalOutputDir = "./results/temporal_query_experiment"
)

// Run temporal query experiments for all scenarios
func RunTemporalQueryExperiments() {
	// Create output directory
	os.MkdirAll(temporalOutputDir, os.ModePerm)

	scenarios = []string{"urban_traffic"}

	// Run experiments for each scenario
	for _, scenario := range scenarios {
		fmt.Printf("\n=== Starting temporal query experiment for scenario: %s ===\n", scenario)

		// Load TSON patches from previous experiment
		scenarioDir := filepath.Join(outputDir, scenario)

		timestampDirs, err := findLatestTimestampDir(scenarioDir)
		if err != nil {
			fmt.Printf("Error finding timestamp directory for %s: %v\n", scenario, err)
			return
		}

		tsonPatches, err := loadTSONPatches(timestampDirs)
		if err != nil {
			fmt.Printf("Error loading TSON patches for %s: %v\n", scenario, err)
			continue
		}

		// Create experiment
		experiment, err := NewTemporalQueryExperiment(timestampDirs, temporalOutputDir, scenario, tsonPatches)
		if err != nil {
			fmt.Printf("Error creating experiment for %s: %v\n", scenario, err)
			continue
		}

		// Run temporal snapshot experiment
		timestamps := generateTimestamps(tsonPatches, 61) // Generate 10 sample timestamps
		snapshotResults, err := experiment.RunTemporalSnapshotExperiment(timestamps)
		if err != nil {
			fmt.Printf("Error running temporal snapshot experiment: %v\n", err)
		} else {
			experiment.SaveResults(snapshotResults, "snapshot_results.csv")
		}

		// Run temporal track experiment
		paths := []string{
			"Vehicle.Speed",
			"Vehicle.CurrentLocation.Latitude",
			"Vehicle.CurrentLocation.Longitude",
			"Vehicle.Powertrain.TractionBattery.StateOfCharge.Current",
		}
		startTime, endTime := getTimeRange(tsonPatches)
		trackResults, err := experiment.RunTemporalTrackExperiment(paths, startTime, endTime)
		if err != nil {
			fmt.Printf("Error running temporal track experiment: %v\n", err)
		} else {
			experiment.SaveResults(trackResults, "track_results.csv")
		}

		gap := int64(10000) // 10 seconds gap
		if endTime - startTime < gap {
			panic("Time range is too small for gap")
		}

		rng := rand.New(rand.NewPCG(42, 0))

		// left ∈ [min, max-gap] (inclusive).
		left := rng.Int64N(endTime-gap-startTime+1) + startTime
		right := left + gap

		// Run event search experiment
		thresholds := []float64{10.0, 30.0, 50.0} // Different speed thresholds
		eventResults, err := experiment.RunEventSearchExperiment(
			[]string{"Vehicle.Speed"},
			left,
			right,
			thresholds,
		)
		if err != nil {
			fmt.Printf("Error running event search experiment: %v\n", err)
		} else {
			experiment.SaveResults(eventResults, "event_search_results.csv")
		}

		// Report completion
		fmt.Printf("Temporal query experiments for %s completed successfully.\n", scenario)
	}

	fmt.Println("\nAll temporal query experiments completed.")
}

// Helper function to load TSON patches from previous experiment
func loadTSONPatches(patchesDir string) ([]TsonPatch, error) {
	// Create a slice to hold all patches
	patches := make([]TsonPatch, 0)

	// Walk through the patches directory
	err := filepath.Walk(patchesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process JSON files
		if filepath.Ext(path) != ".json" {
			return nil
		}

		if filepath.Base(path) != "all_tson_patches.json" {
			return nil
		}

		// Read the file
		fileContent, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read patch file %s: %v", path, err)
		}

		// Parse patches
		var filePatches []TsonPatch
		if err := json.Unmarshal(fileContent, &filePatches); err != nil {
			return fmt.Errorf("failed to parse patch file %s: %v", path, err)
		}

		// Add to our collection
		patches = append(patches, filePatches...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return patches, nil
}

// Helper function to generate sample timestamps for testing
func generateTimestamps(patches []TsonPatch, count int) []int64 {
	// Find min and max timestamps
	var minTime, maxTime int64

	if len(patches) > 0 {
		minTime = patches[0].Timestamp
		maxTime = patches[0].Timestamp

		for _, patch := range patches {
			if patch.Timestamp < minTime {
				minTime = patch.Timestamp
			}
			if patch.Timestamp > maxTime {
				maxTime = patch.Timestamp
			}
		}
	} else {
		// Default to current time if no patches
		now := time.Now().Unix()
		minTime = now - 3600 // 1 hour ago
		maxTime = now
	}

	// Generate timestamps spread across the range
	timestamps := make([]int64, count)
	timeRange := maxTime - minTime

	for i := 0; i < count; i++ {
		position := float64(i) / float64(count-1)
		timestamps[i] = minTime + int64(position*float64(timeRange))
	}

	return timestamps
}

// Helper function to get the full time range from patches
func getTimeRange(patches []TsonPatch) (int64, int64) {
	if len(patches) == 0 {
		now := time.Now().Unix()
		return now - 3600, now
	}

	minTime := patches[0].Timestamp
	maxTime := patches[0].Timestamp

	for _, patch := range patches {
		if patch.Timestamp < minTime {
			minTime = patch.Timestamp
		}
		if patch.Timestamp > maxTime {
			maxTime = patch.Timestamp
		}
	}

	return minTime, maxTime
}

// findLatestTimestampDir finds the most recent experiment timestamp directory
func findLatestTimestampDir(scenarioDir string) (string, error) {
	// List all timestamp directories
	entries, err := os.ReadDir(scenarioDir)
	if err != nil {
		return "", fmt.Errorf("error reading directory: %v", err)
	}

	// Find directories that match the timestamp pattern
	var timestampDirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it's a timestamp directory (format: YYYYMMDD_HHMMSS)
			if matched, _ := filepath.Match("[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]_[0-9][0-9][0-9][0-9][0-9][0-9]", entry.Name()); matched {
				timestampDirs = append(timestampDirs, entry.Name())
			}
		}
	}

	if len(timestampDirs) == 0 {
		return "", fmt.Errorf("no timestamp directories found in %s", scenarioDir)
	}

	// Sort to find the latest
	sort.Strings(timestampDirs)
	latestDir := filepath.Join(scenarioDir, timestampDirs[len(timestampDirs)-1])

	return latestDir, nil
}
