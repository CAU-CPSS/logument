package exp

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"
)

// DataStorage handles the task of saving experimental data to files
type DataStorage struct {
    OutputDir string
    Scenario  string
    Timestamp string
}

func NewDataStorage(baseDir, scenario string) *DataStorage {
    // Generate a timestamp for a unique directory
    timestamp := time.Now().Format("20060102_150405")

    // Create the full path including scenario and timestamp
    fullPath := filepath.Join(baseDir, scenario, timestamp)

    // Create the directory if it doesn't exist
    os.MkdirAll(fullPath, os.ModePerm)

    return &DataStorage{
        OutputDir: fullPath,
        Scenario:  scenario,
        Timestamp: timestamp,
    }
}

// SaveInitialJSON saves the initial JSON document
func (ds *DataStorage) SaveInitialJSON(data any) error {
    return ds.saveJSONToFile(data, "initial_json.json")
}

// SaveInitialTSON saves the initial TSON document
func (ds *DataStorage) SaveInitialTSON(data any) error {
    return ds.saveJSONToFile(data, "initial_tson.json")
}

// SaveJSONPatch saves a JSON patch
func (ds *DataStorage) SaveJSONPatch(patchIndex int, patch any) error {
    filename := fmt.Sprintf("json_patch_%04d.json", patchIndex)
    return ds.saveJSONToFile(patch, filename)
}

// SaveTSONPatch saves a TSON patch
func (ds *DataStorage) SaveTSONPatch(patchIndex int, patch any) error {
    filename := fmt.Sprintf("tson_patch_%04d.json", patchIndex)
    return ds.saveJSONToFile(patch, filename)
}

// SaveAllJSONPatches saves all JSON patches into a single file
func (ds *DataStorage) SaveAllJSONPatches(patches any) error {
    return ds.saveJSONToFile(patches, "all_json_patches.json")
}

// SaveAllTSONPatches saves all TSON patches into a single file
func (ds *DataStorage) SaveAllTSONPatches(patches any) error {
    return ds.saveJSONToFile(patches, "all_tson_patches.json")
}

// SaveFinalJSON saves the final JSON document
func (ds *DataStorage) SaveFinalJSON(data any) error {
    return ds.saveJSONToFile(data, "final_json.json")
}

// SaveFinalTSON saves the final TSON document
func (ds *DataStorage) SaveFinalTSON(data any) error {
    return ds.saveJSONToFile(data, "final_tson.json")
}

// SaveJSONSnapshot saves a JSON snapshot at a specific time
func (ds *DataStorage) SaveJSONSnapshot(timeMs int64, data any) error {
    filename := fmt.Sprintf("json_snapshot_%09d.json", timeMs)
    return ds.saveJSONToFile(data, filename)
}

// SaveTSONSnapshot saves a TSON snapshot at a specific time
func (ds *DataStorage) SaveTSONSnapshot(timeMs int64, data any) error {
    filename := fmt.Sprintf("tson_snapshot_%09d.json", timeMs)
    return ds.saveJSONToFile(data, filename)
}

// saveJSONToFile is a helper method to save JSON data to a file
func (ds *DataStorage) saveJSONToFile(data any, filename string) error {
    filePath := filepath.Join(ds.OutputDir, filename)

    // Marshal the data into indented JSON for readability
    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return fmt.Errorf("JSON marshaling error: %v", err)
    }

    // Write to the file
    if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
        return fmt.Errorf("File write error %s: %v", filePath, err)
    }

    // fmt.Printf("File saved successfully: %s\n", filePath)
    return nil
}
