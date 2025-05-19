package vssgen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GetData 메서드는 VssJson의 내부 데이터를 반환합니다
func (vss *VssJson) GetData() interface{} {
	return vss.data
}

// ToMapStringInterface 메서드는 VSS 데이터를 map[string]interface{} 형식으로 변환합니다
func (vss *VssJson) ToMapStringInterface() (map[string]interface{}, error) {
	// 임시 파일에 저장하고 다시 읽어오기
	tempDir, err := os.MkdirTemp("", "vssgen")
	if err != nil {
		return nil, fmt.Errorf("임시 디렉토리 생성 실패: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "temp_data.json")
	vss.Save(tempFile)

	// 파일에서 데이터 읽기
	jsonBytes, err := os.ReadFile(tempFile)
	if err != nil {
		return nil, fmt.Errorf("임시 파일 읽기 실패: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON 파싱 실패: %v", err)
	}

	return result, nil
}

// BuildFromLeafNodes 메서드는 LeafNodes()로 얻은 데이터를 map[string]interface{}로 변환합니다
func (vss *VssJson) BuildFromLeafNodes() (map[string]interface{}, error) {
	leafNodes := vss.LeafNodes()
	result := make(map[string]interface{})

	for _, leaf := range leafNodes {
		for path, value := range leaf {
			// 경로를 기준으로 중첩 맵 생성
			buildNestedStructure(result, path, value)
		}
	}

	return result, nil
}

// buildNestedStructure는 "a.b.c" 형식의 경로를 중첩된 맵 구조로 변환합니다
func buildNestedStructure(result map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	current := result

	for i, part := range parts {
		// 배열 인덱스인지 확인
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			// 배열 인덱스 추출
			idxStr := part[1 : len(part)-1]
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				continue // 인덱스가 아니면 건너뜀
			}

			// 마지막 부분인 경우
			if i == len(parts)-1 {
				// 배열 크기 확장
				if _, ok := current["array"]; !ok {
					current["array"] = make([]interface{}, idx+1)
				}
				arr := current["array"].([]interface{})
				if len(arr) <= idx {
					newArr := make([]interface{}, idx+1)
					copy(newArr, arr)
					arr = newArr
					current["array"] = arr
				}
				arr[idx] = value
			} else {
				// 배열 크기 확장
				if _, ok := current["array"]; !ok {
					current["array"] = make([]interface{}, idx+1)
				}
				arr := current["array"].([]interface{})
				if len(arr) <= idx {
					newArr := make([]interface{}, idx+1)
					copy(newArr, arr)
					arr = newArr
					current["array"] = arr
				}
				if arr[idx] == nil {
					arr[idx] = make(map[string]interface{})
				}
				current = arr[idx].(map[string]interface{})
			}
		} else {
			// 일반 객체 속성
			if i == len(parts)-1 {
				// 마지막 부분인 경우 값 저장
				current[part] = value
			} else {
				// 중간 경로에 맵 생성
				if _, ok := current[part]; !ok {
					current[part] = make(map[string]interface{})
				}
				// 다음으로 진행
				if nextMap, ok := current[part].(map[string]interface{}); ok {
					current = nextMap
				} else {
					// 맵이 아닌 경우 진행 불가
					return
				}
			}
		}
	}
}
