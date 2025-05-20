package exp

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"
)

// DataStorage는 실험 데이터를 파일로 저장하는 작업을 처리합니다
type DataStorage struct {
    OutputDir string // 출력 디렉토리 경로
    Scenario  string // 시나리오 이름
    Timestamp string // 실행 타임스탬프
}

// NewDataStorage는 새로운 DataStorage 인스턴스를 생성합니다
func NewDataStorage(baseDir, scenario string) *DataStorage {
    // 고유 디렉토리를 위한 타임스탬프 생성
    timestamp := time.Now().Format("20060102_150405")
    
    // 시나리오와 타임스탬프를 포함한 전체 경로 생성
    fullPath := filepath.Join(baseDir, scenario, timestamp)
    
    // 디렉토리가 없으면 생성
    os.MkdirAll(fullPath, os.ModePerm)
    
    return &DataStorage{
        OutputDir: fullPath,
        Scenario:  scenario,
        Timestamp: timestamp,
    }
}

// SaveInitialJSON은 초기 JSON 문서를 저장합니다
func (ds *DataStorage) SaveInitialJSON(data interface{}) error {
    return ds.saveJSONToFile(data, "initial_json.json")
}

// SaveInitialTSON은 초기 TSON 문서를 저장합니다
func (ds *DataStorage) SaveInitialTSON(data interface{}) error {
    return ds.saveJSONToFile(data, "initial_tson.json")
}

// SaveJSONPatch는 JSON 패치를 저장합니다
func (ds *DataStorage) SaveJSONPatch(patchIndex int, patch interface{}) error {
    filename := fmt.Sprintf("json_patch_%04d.json", patchIndex)
    return ds.saveJSONToFile(patch, filename)
}

// SaveTSONPatch는 TSON 패치를 저장합니다
func (ds *DataStorage) SaveTSONPatch(patchIndex int, patch interface{}) error {
    filename := fmt.Sprintf("tson_patch_%04d.json", patchIndex)
    return ds.saveJSONToFile(patch, filename)
}

// SaveAllJSONPatches는 모든 JSON 패치를 단일 파일로 저장합니다
func (ds *DataStorage) SaveAllJSONPatches(patches interface{}) error {
    return ds.saveJSONToFile(patches, "all_json_patches.json")
}

// SaveAllTSONPatches는 모든 TSON 패치를 단일 파일로 저장합니다
func (ds *DataStorage) SaveAllTSONPatches(patches interface{}) error {
    return ds.saveJSONToFile(patches, "all_tson_patches.json")
}

// SaveFinalJSON은 최종 JSON 문서를 저장합니다
func (ds *DataStorage) SaveFinalJSON(data interface{}) error {
    return ds.saveJSONToFile(data, "final_json.json")
}

// SaveFinalTSON은 최종 TSON 문서를 저장합니다
func (ds *DataStorage) SaveFinalTSON(data interface{}) error {
    return ds.saveJSONToFile(data, "final_tson.json")
}

// SaveJSONSnapshot은 특정 시간의 JSON 스냅샷을 저장합니다
func (ds *DataStorage) SaveJSONSnapshot(timeMs int64, data interface{}) error {
    filename := fmt.Sprintf("json_snapshot_%09d.json", timeMs)
    return ds.saveJSONToFile(data, filename)
}

// SaveTSONSnapshot은 특정 시간의 TSON 스냅샷을 저장합니다
func (ds *DataStorage) SaveTSONSnapshot(timeMs int64, data interface{}) error {
    filename := fmt.Sprintf("tson_snapshot_%09d.json", timeMs)
    return ds.saveJSONToFile(data, filename)
}

// saveJSONToFile은 JSON 데이터를 파일로 저장하는 헬퍼 메소드입니다
func (ds *DataStorage) saveJSONToFile(data interface{}, filename string) error {
    filePath := filepath.Join(ds.OutputDir, filename)
    
    // 가독성을 위해 들여쓰기가 있는 JSON으로 마샬링
    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return fmt.Errorf("JSON 마샬링 오류: %v", err)
    }
    
    // 파일에 쓰기
    if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
        return fmt.Errorf("파일 쓰기 오류 %s: %v", filePath, err)
    }
    
    // fmt.Printf("파일 저장 완료: %s\n", filePath)
    return nil
}