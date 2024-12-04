package jpatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// PatchOperation 구조체는 단일 JSON Patch 작업을 나타냅니다.
type PatchOperation struct {
	Op    string          `json:"op"`
	Path  string          `json:"path"`
	From  string          `json:"from,omitempty"`
	Value json.RawMessage `json:"value,omitempty"`
}

// ApplyPatch 함수는 PatchOperation의 배열을 문서에 적용합니다.
func ApplyPatch(document *interface{}, patch []PatchOperation) error {
	for _, op := range patch {
		switch op.Op {
		case "add":
			var value interface{}
			if err := json.Unmarshal(op.Value, &value); err != nil {
				return err
			}
			if err := addValue(document, op.Path, value); err != nil {
				return err
			}
		case "remove":
			if err := removeValue(document, op.Path); err != nil {
				return err
			}
		case "replace":
			var value interface{}
			if err := json.Unmarshal(op.Value, &value); err != nil {
				return err
			}
			if err := replaceValue(document, op.Path, value); err != nil {
				return err
			}
		case "move":
			if err := moveValue(document, op.From, op.Path); err != nil {
				return err
			}
		case "copy":
			if err := copyValue(document, op.From, op.Path); err != nil {
				return err
			}
		case "test":
			var value interface{}
			if err := json.Unmarshal(op.Value, &value); err != nil {
				return err
			}
			if err := testValue(document, op.Path, value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("잘못된 연산: %s", op.Op)
		}
	}
	return nil
}

// ParseJSONPointer 함수는 JSON Pointer를 파싱하여 참조 토큰의 슬라이스를 반환합니다.
func ParseJSONPointer(path string) ([]string, error) {
	if path == "" {
		return []string{}, nil
	}
	if path[0] != '/' {
		return nil, fmt.Errorf("잘못된 JSON Pointer: %s", path)
	}
	tokens := strings.Split(path[1:], "/")
	for i, token := range tokens {
		token = strings.ReplaceAll(token, "~1", "/")
		token = strings.ReplaceAll(token, "~0", "~")
		tokens[i] = token
	}
	return tokens, nil
}

// getValue 함수는 문서에서 주어진 경로의 값을 반환합니다.
func getValue(document interface{}, path string) (interface{}, error) {
	tokens, err := ParseJSONPointer(path)
	if err != nil {
		return nil, err
	}
	return getValueByTokens(document, tokens)
}

func getValueByTokens(document interface{}, tokens []string) (interface{}, error) {
	current := document
	for _, token := range tokens {
		switch curr := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = curr[token]
			if !ok {
				return nil, fmt.Errorf("경로를 찾을 수 없음: %s", token)
			}
		case []interface{}:
			index, err := strconv.Atoi(token)
			if err != nil {
				return nil, fmt.Errorf("잘못된 배열 인덱스: %s", token)
			}
			if index < 0 || index >= len(curr) {
				return nil, fmt.Errorf("배열 인덱스 범위 초과: %d", index)
			}
			current = curr[index]
		default:
			return nil, fmt.Errorf("예상치 못한 타입: %s", token)
		}
	}
	return current, nil
}

// setValue 함수는 문서에서 주어진 경로에 값을 설정합니다.
func setValue(document *interface{}, path string, value interface{}) error {
	tokens, err := ParseJSONPointer(path)
	if err != nil {
		return err
	}
	return setValueByTokens(document, tokens, value)
}

func setValueByTokens(document *interface{}, tokens []string, value interface{}) error {
	if len(tokens) == 0 {
		*document = value
		return nil
	}
	token := tokens[0]
	switch curr := (*document).(type) {
	case map[string]interface{}:
		if len(tokens) == 1 {
			curr[token] = value
			return nil
		} else {
			next, ok := curr[token]
			if !ok {
				next = make(map[string]interface{})
				curr[token] = next
			}
			return setValueByTokens(&next, tokens[1:], value)
		}
	case []interface{}:
		index := 0
		if token == "-" {
			index = len(curr)
		} else {
			index, err := strconv.Atoi(token)
			if err != nil {
				return fmt.Errorf("잘못된 배열 인덱스: %s", token)
			}
			if index < 0 || index > len(curr) {
				return fmt.Errorf("배열 인덱스 범위 초과: %d", index)
			}
		}
		if len(tokens) == 1 {
			if token == "-" {
				curr = append(curr, value)
				*document = curr
			} else if index == len(curr) {
				curr = append(curr, value)
				*document = curr
			} else {
				curr[index] = value
			}
			return nil
		} else {
			if index == len(curr) {
				return errors.New("존재하지 않는 배열 요소의 자식을 설정할 수 없습니다.")
			}
			next := curr[index]
			err := setValueByTokens(&next, tokens[1:], value)
			if err != nil {
				return err
			}
			curr[index] = next
			return nil
		}
	default:
		return fmt.Errorf("예상치 못한 타입: %s", token)
	}
}

// addValue 함수는 문서에 값을 추가합니다.
func addValue(document *interface{}, path string, value interface{}) error {
	return setValue(document, path, value)
}

// removeValue 함수는 문서에서 주어진 경로의 값을 제거합니다.
func removeValue(document *interface{}, path string) error {
	tokens, err := ParseJSONPointer(path)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return errors.New("루트 문서를 제거할 수 없습니다.")
	}
	return removeValueByTokens(document, tokens)
}

func removeValueByTokens(document *interface{}, tokens []string) error {
	token := tokens[0]
	switch curr := (*document).(type) {
	case map[string]interface{}:
		if len(tokens) == 1 {
			if _, ok := curr[token]; ok {
				delete(curr, token)
				return nil
			} else {
				return fmt.Errorf("경로를 찾을 수 없음: %s", token)
			}
		} else {
			next, ok := curr[token]
			if !ok {
				return fmt.Errorf("경로를 찾을 수 없음: %s", token)
			}
			return removeValueByTokens(&next, tokens[1:])
		}
	case []interface{}:
		index, err := strconv.Atoi(token)
		if err != nil {
			return fmt.Errorf("잘못된 배열 인덱스: %s", token)
		}
		if index < 0 || index >= len(curr) {
			return fmt.Errorf("배열 인덱스 범위 초과: %d", index)
		}
		if len(tokens) == 1 {
			curr = append(curr[:index], curr[index+1:]...)
			*document = curr
			return nil
		} else {
			next := curr[index]
			err := removeValueByTokens(&next, tokens[1:])
			if err != nil {
				return err
			}
			curr[index] = next
			return nil
		}
	default:
		return fmt.Errorf("예상치 못한 타입: %s", token)
	}
}

// replaceValue 함수는 문서에서 주어진 경로의 값을 대체합니다.
func replaceValue(document *interface{}, path string, value interface{}) error {
	tokens, err := ParseJSONPointer(path)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		*document = value
		return nil
	}
	return replaceValueByTokens(document, tokens, value)
}

func replaceValueByTokens(document *interface{}, tokens []string, value interface{}) error {
	token := tokens[0]
	switch curr := (*document).(type) {
	case map[string]interface{}:
		if len(tokens) == 1 {
			if _, ok := curr[token]; ok {
				curr[token] = value
				return nil
			} else {
				return fmt.Errorf("경로를 찾을 수 없음: %s", token)
			}
		} else {
			next, ok := curr[token]
			if !ok {
				return fmt.Errorf("경로를 찾을 수 없음: %s", token)
			}
			return replaceValueByTokens(&next, tokens[1:], value)
		}
	case []interface{}:
		index, err := strconv.Atoi(token)
		if err != nil {
			return fmt.Errorf("잘못된 배열 인덱스: %s", token)
		}
		if index < 0 || index >= len(curr) {
			return fmt.Errorf("배열 인덱스 범위 초과: %d", index)
		}
		if len(tokens) == 1 {
			curr[index] = value
			return nil
		} else {
			next := curr[index]
			err := replaceValueByTokens(&next, tokens[1:], value)
			if err != nil {
				return err
			}
			curr[index] = next
			return nil
		}
	default:
		return fmt.Errorf("예상치 못한 타입: %s", token)
	}
}

// copyValue 함수는 문서 내에서 값을 복사합니다.
func copyValue(document *interface{}, fromPath, path string) error {
	value, err := getValue(*document, fromPath)
	if err != nil {
		return err
	}
	return addValue(document, path, value)
}

// moveValue 함수는 문서 내에서 값을 이동합니다.
func moveValue(document *interface{}, fromPath, path string) error {
	value, err := getValue(*document, fromPath)
	if err != nil {
		return err
	}
	if err := removeValue(document, fromPath); err != nil {
		return err
	}
	return addValue(document, path, value)
}

// testValue 함수는 문서에서 주어진 경로의 값이 기대하는 값과 일치하는지 확인합니다.
func testValue(document *interface{}, path string, value interface{}) error {
	currentValue, err := getValue(*document, path)
	if err != nil {
		return err
	}
	if !deepEqual(currentValue, value) {
		return fmt.Errorf("test 연산 실패: 경로 %s", path)
	}
	return nil
}

// deepEqual 함수는 두 값을 비교합니다.
func deepEqual(a, b interface{}) bool {
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(aJSON) == string(bJSON)
}

// CreatePatch 함수는 두 개의 JSON 문서를 비교하여 변경된 내역에 대한 []PatchOperation을 생성합니다.
func CreatePatch(path string, fromDoc, toDoc interface{}) ([]PatchOperation, error) {
	var patch []PatchOperation
	switch from := fromDoc.(type) {
	case map[string]interface{}:
		to, ok := toDoc.(map[string]interface{})
		if !ok {
			// 타입이 다르면 전체를 교체
			op, err := replaceOp(path, toDoc)
			if err != nil {
				return nil, err
			}
			patch = append(patch, op)
			return patch, nil
		}
		// 키 비교
		visited := make(map[string]bool)
		for key, fromValue := range from {
			newPath := path + "/" + escapeJSONPointer(key)
			if toValue, exists := to[key]; exists {
				visited[key] = true
				subPatch, err := CreatePatch(newPath, fromValue, toValue)
				if err != nil {
					return nil, err
				}
				patch = append(patch, subPatch...)
			} else {
				// 삭제된 키
				patch = append(patch, PatchOperation{
					Op:   "remove",
					Path: newPath,
				})
			}
		}
		for key, toValue := range to {
			if visited[key] {
				continue
			}
			newPath := path + "/" + escapeJSONPointer(key)
			// 추가된 키
			op, err := addOp(newPath, toValue)
			if err != nil {
				return nil, err
			}
			patch = append(patch, op)
		}
	case []interface{}:
		to, ok := toDoc.([]interface{})
		if !ok {
			// 타입이 다르면 전체를 교체
			op, err := replaceOp(path, toDoc)
			if err != nil {
				return nil, err
			}
			patch = append(patch, op)
			return patch, nil
		}
		// 배열 비교 (간단하게 인덱스별로 비교)
		minLen := len(from)
		if len(to) < minLen {
			minLen = len(to)
		}
		for i := 0; i < minLen; i++ {
			newPath := fmt.Sprintf("%s/%d", path, i)
			subPatch, err := CreatePatch(newPath, from[i], to[i])
			if err != nil {
				return nil, err
			}
			patch = append(patch, subPatch...)
		}
		if len(from) > len(to) {
			// 요소 삭제
			for i := len(from) - 1; i >= len(to); i-- {
				newPath := fmt.Sprintf("%s/%d", path, i)
				patch = append(patch, PatchOperation{
					Op:   "remove",
					Path: newPath,
				})
			}
		} else if len(to) > len(from) {
			// 요소 추가
			for i := len(from); i < len(to); i++ {
				newPath := fmt.Sprintf("%s/-", path)
				op, err := addOp(newPath, to[i])
				if err != nil {
					return nil, err
				}
				patch = append(patch, op)
			}
		}
	default:
		if !deepEqual(fromDoc, toDoc) {
			// 값이 다르면 교체
			op, err := replaceOp(path, toDoc)
			if err != nil {
				return nil, err
			}
			patch = append(patch, op)
		}
	}
	return patch, nil
}

// addOp 함수는 add 연산을 생성합니다.
func addOp(path string, value interface{}) (PatchOperation, error) {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return PatchOperation{}, err
	}
	return PatchOperation{
		Op:    "add",
		Path:  path,
		Value: json.RawMessage(valueJSON),
	}, nil
}

// replaceOp 함수는 replace 연산을 생성합니다.
func replaceOp(path string, value interface{}) (PatchOperation, error) {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return PatchOperation{}, err
	}
	return PatchOperation{
		Op:    "replace",
		Path:  path,
		Value: json.RawMessage(valueJSON),
	}, nil
}

// escapeJSONPointer 함수는 JSON Pointer에서 사용하는 특수 문자를 이스케이프 처리합니다.
func escapeJSONPointer(token string) string {
	token = strings.ReplaceAll(token, "~", "~0")
	token = strings.ReplaceAll(token, "/", "~1")
	return token
}
