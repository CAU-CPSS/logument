package exp

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/CAU-CPSS/logument/internal/logument"
	"github.com/CAU-CPSS/logument/internal/tson"
	"github.com/CAU-CPSS/logument/internal/tsonpatch"
)

// TemporalQueryExperiment holds experiment data
type TemporalQueryExperiment struct {
	OutputDir         string             // Output directory
	Scenario          string             // Scenario name
	TsonPatches       []TsonPatch        // TSON patches
	TJSONDocument     *TJSONDocument     // TJSON document
	LogumentDoc       *logument.Logument // Logument document
	LogumentResultDir string             // Directory for Logument results
	TJSONResultDir    string             // Directory for TJSON results
}

// TemporalQueryResult holds experiment results
type TemporalQueryResult struct {
	Scenario            string // Scenario name
	QueryType           string // Type of query (TemporalSnapshot, TemporalTrack, EventSearch)
	Parameter           string // Query parameter (timestamp, path, etc.)
	LogumentTimeNs      int64  // Execution time in nanoseconds for Logument
	TJSONTimeNs         int64  // Execution time in nanoseconds for TJSON
	LogumentResultSize  int    // Size of result in bytes for Logument
	TJSONResultSize     int    // Size of result in bytes for TJSON
	LogumentMemoryUsage int64  // Memory used by Logument (if measurable)
	TJSONMemoryUsage    int64  // Memory used by TJSON (if measurable)
	LogumentResultFile  string // Path to the file containing Logument result
	TJSONResultFile     string // Path to the file containing TJSON result
}

// NewTemporalQueryExperiment creates a new experiment
func NewTemporalQueryExperiment(inputDir, baseOutDir, scenario string, tsonPatches []TsonPatch) (*TemporalQueryExperiment, error) {
	// Create output directory
	outputDir := filepath.Join(baseOutDir, scenario, "temporal_query")
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create Logument result directory
	logumentResultDir := filepath.Join(outputDir, "logument")
	if err := os.MkdirAll(logumentResultDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create Logument result directory: %v", err)
	}
	// Create TJSON result directory
	tjsonResultDir := filepath.Join(outputDir, "tjson")
	if err := os.MkdirAll(tjsonResultDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create TJSON result directory: %v", err)
	}

	// Path to initial JSON file
	initialJsonPath := filepath.Join(inputDir, "initial_json.json")

	// Create initial patches from initial JSON
	initialPatches, err := CreateInitialPatchesFromJSON(initialJsonPath)
	if err != nil {
		panic(fmt.Sprintf("Warning: Failed to create initial patches: %v\n", err))
	}
	// Combine initial patches with existing patches
	allPatches := make([]TsonPatch, 0, len(initialPatches)+len(tsonPatches))
	allPatches = append(allPatches, initialPatches...)
	allPatches = append(allPatches, tsonPatches...)

	// Sort by timestamp
	sort.Slice(allPatches, func(i, j int) bool {
		return allPatches[i].Timestamp < allPatches[j].Timestamp
	})

	// Convert TSON patches to TJSON document
	tjsonDoc, err := FromTSONPatches(allPatches)
	if err != nil {
		return nil, fmt.Errorf("failed to create TJSON document: %v", err)
	}

	// Create Logument document
	lgm := createLogumentFromPatches(inputDir, tsonPatches)

	return &TemporalQueryExperiment{
		OutputDir:         outputDir,
		Scenario:          scenario,
		TsonPatches:       tsonPatches,
		TJSONDocument:     tjsonDoc,
		LogumentDoc:       lgm,
		LogumentResultDir: logumentResultDir,
		TJSONResultDir:    tjsonResultDir,
	}, nil
}

// RunTemporalSnapshotExperiment runs the temporal snapshot experiment
func (e *TemporalQueryExperiment) RunTemporalSnapshotExperiment(timestamps []int64) ([]TemporalQueryResult, error) {
	results := make([]TemporalQueryResult, 0, len(timestamps))

	for i, ts := range timestamps {
		result := TemporalQueryResult{
			Scenario:  e.Scenario,
			QueryType: "TemporalSnapshot",
			Parameter: fmt.Sprintf("%d", ts),
		}

		// Run Logument temporal snapshot
		lgmStart := time.Now()
		lgmSnapshot := e.LogumentDoc.TemporalSnapshot(ts)
		result.LogumentTimeNs = time.Since(lgmStart).Nanoseconds()

		// Convert to bytes to measure size
		lgmBytes, _ := json.Marshal(lgmSnapshot)
		result.LogumentResultSize = len(lgmBytes)

		// Save Logument result to file
		lgmFilename := fmt.Sprintf("snapshot_logument_%d_%d.json", i, ts)
		lgmFilePath, err := saveResultToFile(e.LogumentResultDir, lgmFilename, lgmSnapshot)
		if err != nil {
			fmt.Printf("Warning: Failed to save Logument snapshot result: %v\n", err)
		} else {
			result.LogumentResultFile = lgmFilePath
		}

		// Run TJSON temporal snapshot
		tjsonStart := time.Now()
		tjsonSnapshot, err := e.TJSONDocument.TemporalSnapshot(ts)
		if err != nil {
			return nil, fmt.Errorf("TJSON temporal snapshot failed: %v", err)
		}
		result.TJSONTimeNs = time.Since(tjsonStart).Nanoseconds()

		// Convert to bytes to measure size
		tjsonBytes, _ := json.Marshal(tjsonSnapshot)
		result.TJSONResultSize = len(tjsonBytes)

		// Save TJSON result to file
		tjsonFilename := fmt.Sprintf("snapshot_tjson_%d_%d.json", i, ts)
		tjsonFilePath, err := saveResultToFile(e.TJSONResultDir, tjsonFilename, tjsonSnapshot)
		if err != nil {
			fmt.Printf("Warning: Failed to save TJSON snapshot result: %v\n", err)
		} else {
			result.TJSONResultFile = tjsonFilePath
		}

		results = append(results, result)
	}

	return results, nil
}

// RunTemporalTrackExperiment runs the temporal track experiment
func (e *TemporalQueryExperiment) RunTemporalTrackExperiment(paths []string, startTime, endTime int64) ([]TemporalQueryResult, error) {
	results := make([]TemporalQueryResult, 0, len(paths))

	for i, path := range paths {
		pathForFilename := strings.ReplaceAll(strings.ReplaceAll(path, ".", "_"), "/", "_")

		result := TemporalQueryResult{
			Scenario:  e.Scenario,
			QueryType: "TemporalTrack",
			Parameter: fmt.Sprintf("%s (%d-%d)", path, startTime, endTime),
		}

		// Run Logument temporal track
		lgmStart := time.Now()
		lgmTrack := e.LogumentDoc.TemporalTrack(startTime, endTime)
		result.LogumentTimeNs = time.Since(lgmStart).Nanoseconds()

		// Filter to the desired path and convert to bytes to measure size
		lgmPathTrack := filterTrackByPath(lgmTrack, path)
		lgmBytes, _ := json.Marshal(lgmPathTrack)
		result.LogumentResultSize = len(lgmBytes)

		// Save Logument result to file
		lgmFilename := fmt.Sprintf("track_logument_%d_%s_%d_%d.json", i, pathForFilename, startTime, endTime)
		lgmFilePath, err := saveResultToFile(e.LogumentResultDir, lgmFilename, lgmPathTrack)
		if err != nil {
			fmt.Printf("Warning: Failed to save Logument track result: %v\n", err)
		} else {
			result.LogumentResultFile = lgmFilePath
		}

		// Run TJSON temporal track
		tjsonStart := time.Now()
		tjsonTrack, err := e.TJSONDocument.TemporalTrack(path, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("TJSON temporal track failed: %v", err)
		}
		result.TJSONTimeNs = time.Since(tjsonStart).Nanoseconds()

		// Convert to bytes to measure size
		tjsonBytes, _ := json.Marshal(tjsonTrack)
		result.TJSONResultSize = len(tjsonBytes)

		// Save TJSON result to file
		tjsonFilename := fmt.Sprintf("track_tjson_%d_%s_%d_%d.json", i, pathForFilename, startTime, endTime)
		tjsonFilePath, err := saveResultToFile(e.TJSONResultDir, tjsonFilename, tjsonTrack)
		if err != nil {
			fmt.Printf("Warning: Failed to save TJSON track result: %v\n", err)
		} else {
			result.TJSONResultFile = tjsonFilePath
		}

		results = append(results, result)
	}

	return results, nil
}

// RunEventSearchExperiment runs the event search experiment
func (e *TemporalQueryExperiment) RunEventSearchExperiment(paths []string, startTime, endTime int64,
	thresholds []float64) ([]TemporalQueryResult, error) {
	results := make([]TemporalQueryResult, 0, len(paths)*len(thresholds))

	for i, path := range paths {
		pathForFilename := strings.ReplaceAll(strings.ReplaceAll(path, ".", "_"), "/", "_")

		for j, threshold := range thresholds {
			paramDesc := fmt.Sprintf("%s > %.1f (%d-%d)", path, threshold, startTime, endTime)

			result := TemporalQueryResult{
				Scenario:  e.Scenario,
				QueryType: "EventSearch",
				Parameter: paramDesc,
			}

			// Run Logument event search
			lgmStart := time.Now()
			lgmEvents := findEventsInLogument(e.LogumentDoc, path, startTime, endTime, threshold)
			result.LogumentTimeNs = time.Since(lgmStart).Nanoseconds()

			// Convert to bytes to measure size
			lgmBytes, _ := json.Marshal(lgmEvents)
			result.LogumentResultSize = len(lgmBytes)

			// Save Logument result to file
			lgmFilename := fmt.Sprintf("event_logument_%d_%s_%.1f_%d_%d.json",
				i*len(thresholds)+j, pathForFilename, threshold, startTime, endTime)
			lgmFilePath, err := saveResultToFile(e.LogumentResultDir, lgmFilename, lgmEvents)
			if err != nil {
				fmt.Printf("Warning: Failed to save Logument event search result: %v\n", err)
			} else {
				result.LogumentResultFile = lgmFilePath
			}

			// Run TJSON event search
			tjsonStart := time.Now()

			// Create predicate function for this threshold
			predicate := func(v interface{}) bool {
				if val, ok := v.(float64); ok {
					return val > threshold
				}
				return false
			}

			tjsonEvents, err := e.TJSONDocument.FindEvents(path, startTime, endTime, predicate)
			if err != nil {
				return nil, fmt.Errorf("TJSON event search failed: %v", err)
			}
			result.TJSONTimeNs = time.Since(tjsonStart).Nanoseconds()

			// Convert to bytes to measure size
			tjsonBytes, _ := json.Marshal(tjsonEvents)
			result.TJSONResultSize = len(tjsonBytes)

			// Save TJSON result to file
			tjsonFilename := fmt.Sprintf("event_tjson_%d_%s_%.1f_%d_%d.json",
				i*len(thresholds)+j, pathForFilename, threshold, startTime, endTime)
			tjsonFilePath, err := saveResultToFile(e.TJSONResultDir, tjsonFilename, tjsonEvents)
			if err != nil {
				fmt.Printf("Warning: Failed to save TJSON event search result: %v\n", err)
			} else {
				result.TJSONResultFile = tjsonFilePath
			}

			results = append(results, result)
		}
	}

	return results, nil
}

// SaveResults saves experiment results to a CSV file
func (e *TemporalQueryExperiment) SaveResults(results []TemporalQueryResult, filename string) error {
	// Create the CSV file
	outputPath := filepath.Join(e.OutputDir, filename)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create results file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Scenario",
		"QueryType",
		"Parameter",
		"LogumentTimeNs",
		"TJSONTimeNs",
		"TimeDiffPct",
		"LogumentResultSize",
		"TJSONResultSize",
		"SizeDiffPct",
	}
	writer.Write(header)

	// Write data rows
	for _, result := range results {
		// Calculate percentage differences
		timeDiffPct := 0.0
		if result.TJSONTimeNs > 0 {
			timeDiffPct = float64(result.LogumentTimeNs-result.TJSONTimeNs) / float64(result.TJSONTimeNs) * 100
		}

		sizeDiffPct := 0.0
		if result.TJSONResultSize > 0 {
			sizeDiffPct = float64(result.LogumentResultSize-result.TJSONResultSize) / float64(result.TJSONResultSize) * 100
		}

		row := []string{
			result.Scenario,
			result.QueryType,
			result.Parameter,
			fmt.Sprintf("%d", result.LogumentTimeNs),
			fmt.Sprintf("%d", result.TJSONTimeNs),
			fmt.Sprintf("%.2f", timeDiffPct),
			fmt.Sprintf("%d", result.LogumentResultSize),
			fmt.Sprintf("%d", result.TJSONResultSize),
			fmt.Sprintf("%.2f", sizeDiffPct),
		}
		writer.Write(row)
	}

	fmt.Printf("Results saved to %s\n", outputPath)
	return nil
}

// Helper function to create a Logument document from TSON patches
func createLogumentFromPatches(patchesDir string, patches []TsonPatch) *logument.Logument {
	// 1. initial_json.json 파일 로드
	initialJsonPath := filepath.Join(filepath.Join(patchesDir), "initial_json.json")
	initialJsonBytes, err := os.ReadFile(initialJsonPath)
	if err != nil {
		panic(err)
	}

	// 2. JSON 파싱
	var initialJsonData map[string]interface{}
	if err := json.Unmarshal(initialJsonBytes, &initialJsonData); err != nil {
		panic(err)
	}

	fmt.Printf("Loaded initial JSON with %d top-level keys\n", len(initialJsonData))

	// 3. 구조 변환: {timestamp:..., value:...} 형태를 값만 있는 형태로 변환
	initialData := transformJsonStructure(initialJsonData)

	// 4. TSON으로 변환
	var initialTson tson.Tson
	if err := tson.FromJson(initialData, &initialTson); err != nil {
		panic(err)
	}

	// // Create an initial empty TSON
	// var initialTson tson.Tson = tson.Object{}

	// Convert TsonPatch array to tsonpatch.Patch
	tsonPatches := make([]tsonpatch.Operation, 0, len(patches))
	for _, p := range patches {
		tsonPatches = append(tsonPatches, tsonpatch.Operation{
			Op:        tsonpatch.OpType(p.Op),
			Path:      p.Path,
			Value:     p.Value,
			Timestamp: p.Timestamp,
		})
	}

	// Create the Logument
	lgm := logument.NewLogument(initialTson, tsonPatches)
	return lgm
}

// transformJsonStructure 함수 개선
func transformJsonStructure(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// timestamp와 value 키가 있는지 확인
		if _, hasTs := v["timestamp"]; hasTs {
			if value, hasValue := v["value"]; hasValue {
				// 이 노드는 {timestamp:..., value:...} 형태이므로 값만 반환
				// fmt.Printf("Found leaf node with timestamp %v and value %v (type: %T)\n", timestamp, value, value)

				// 타입에 따라 적절하게 반환
				switch val := value.(type) {
				case float64:
					// JSON에서는 모든 숫자가 float64로 파싱되므로 그대로 반환
					return val
				case string:
					return val
				case bool:
					return val
				case nil:
					// nil 값은 빈 문자열로 처리
					return ""
				default:
					fmt.Printf("Warning: Unknown leaf value type: %T\n", val)
					return val
				}
			}
		}

		// 일반 객체인 경우 각 필드를 재귀적으로 처리
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = transformJsonStructure(val)
		}
		return result

	case []interface{}:
		// 배열인 경우 각 요소를 재귀적으로 처리
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = transformJsonStructure(val)
		}
		return result

	default:
		// 기본 값 타입은 그대로 반환
		return v
	}
}

// // transformJsonStructure은 JSON 구조를 변환: {timestamp:..., value:...} 형태를 값만 있는 형태로 변환
// func transformJsonStructure(data interface{}) interface{} {
// 	switch v := data.(type) {
// 	case map[string]interface{}:
// 		// timestamp와 value 키가 있는지 확인
// 		if _, hasTs := v["timestamp"]; hasTs {
// 			if value, hasValue := v["value"]; hasValue {
// 				// 이 노드는 {timestamp:..., value:...} 형태이므로 값만 반환
// 				// fmt.Printf("Found leaf node with timestamp %v and value %v\n", timestamp, value)
// 				return value
// 			}
// 		}

// 		// 일반 객체인 경우 각 필드를 재귀적으로 처리
// 		result := make(map[string]interface{})
// 		for key, val := range v {
// 			result[key] = transformJsonStructure(val)
// 		}
// 		return result

// 	case []interface{}:
// 		// 배열인 경우 각 요소를 재귀적으로 처리
// 		result := make([]interface{}, len(v))
// 		for i, val := range v {
// 			result[i] = transformJsonStructure(val)
// 		}
// 		return result

// 	default:
// 		// 기본 값 타입은 그대로 반환
// 		return v
// 	}
// }

// Helper function to filter Logument track results by path
func filterTrackByPath(track map[uint64]tsonpatch.Patch, path string) map[uint64][]tsonpatch.Operation {
	result := make(map[uint64][]tsonpatch.Operation)

	for version, patches := range track {
		for _, patch := range patches {
			if patch.Path == path {
				if _, exists := result[version]; !exists {
					result[version] = make([]tsonpatch.Operation, 0)
				}
				result[version] = append(result[version], patch)
			}
		}
	}

	return result
}

// Helper function to find events in Logument that exceed a threshold
func findEventsInLogument(lgm *logument.Logument, path string, startTime, endTime int64, threshold float64) []int64 {
	// Get tracked changes in the time range
	changes := lgm.TemporalTrack(startTime, endTime)

	// Find all timestamps where the value exceeds the threshold
	eventTimes := make([]int64, 0)

	for _, patches := range changes {
		for _, patch := range patches {
			if patch.Path == path {
				if val, ok := patch.Value.(float64); ok && val > threshold {
					eventTimes = append(eventTimes, patch.Timestamp)
				}
			}
		}
	}

	return eventTimes
}

// 결과를 JSON 파일로 저장하는 헬퍼 함수 추가
func saveResultToFile(outputDir, filename string, data interface{}) (string, error) {
	// 파일 경로 생성
	outputPath := filepath.Join(outputDir, filename)

	// JSON으로 인코딩
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON encoding failed: %v", err)
	}

	// 파일에 저장
	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}

	return outputPath, nil
}

// CreateInitialPatchesFromJSON creates initial "add" patches from initial_json.json
func CreateInitialPatchesFromJSON(initialJsonPath string) ([]TsonPatch, error) {
	// 1. Load initial_json.json file
	initialJsonBytes, err := os.ReadFile(initialJsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read initial JSON file: %v", err)
	}

	// 2. Parse JSON
	var initialJsonData map[string]interface{}
	if err := json.Unmarshal(initialJsonBytes, &initialJsonData); err != nil {
		return nil, fmt.Errorf("failed to parse initial JSON: %v", err)
	}

	fmt.Printf("Loaded initial JSON with %d top-level keys\n", len(initialJsonData))

	// 3. Transform structure: convert {timestamp:..., value:...} structure to values only
	initialData := transformJsonStructure(initialJsonData)

	// 4. Create initial patches
	patches := make([]TsonPatch, 0)

	// Generate patches recursively
	generateInitialPatches("", initialData, &patches)

	// fmt.Println(spew.Sdump(patches))

	fmt.Printf("Generated %d initial patches from initial JSON\n", len(patches))

	return patches, nil
}

// Helper function to recursively generate patches from initial data
func generateInitialPatches(path string, data interface{}, patches *[]TsonPatch) {
	switch v := data.(type) {
	case map[string]interface{}:
		// Process each field in the object
		for key, value := range v {
			newPath := key
			if path != "" {
				newPath = path + "." + key
			}

			// Create add patch for the nested object itself (if it has primitive values)
			hasPrimitives := false
			for _, fieldValue := range v {
				if isPrimitive(fieldValue) {
					hasPrimitives = true
					break
				}
			}

			if hasPrimitives {
				primitiveValues := make(map[string]interface{})
				for fieldKey, fieldValue := range v {
					if isPrimitive(fieldValue) {
						primitiveValues[fieldKey] = fieldValue
					}
				}

				if len(primitiveValues) > 0 {
					*patches = append(*patches, TsonPatch{
						Op:        "add",
						Path:      path,
						Value:     primitiveValues,
						Timestamp: int64(0),
					})
				}
			}

			// Recursively process nested structures
			generateInitialPatches(newPath, value, patches)
		}

	case []interface{}:
		// Process each item in the array
		for i, item := range v {
			newPath := fmt.Sprintf("%s.%d", path, i)

			// Add patch for array item
			if isPrimitive(item) {
				*patches = append(*patches, TsonPatch{
					Op:        "add",
					Path:      newPath,
					Value:     item,
					Timestamp: int64(0),
				})
			} else {
				generateInitialPatches(newPath, item, patches)
			}
		}

	default:
		// Add patch for primitive value
		if path != "" {
			*patches = append(*patches, TsonPatch{
				Op:        "add",
				Path:      path,
				Value:     v,
				Timestamp: int64(0),
			})
		}
	}
}

// Helper function to check if a value is primitive
func isPrimitive(value interface{}) bool {
	switch value.(type) {
	case string, float64, bool, nil:
		return true
	default:
		return false
	}
}
