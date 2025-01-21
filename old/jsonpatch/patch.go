package mainrr

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/CAU-CPSS/logument/internal/jsonr"
	"github.com/CAU-CPSS/logument/internal/logument"
)

type Patch = logument.Patch

// CreatePatch creates a JSON patch from the given JSON-R documents.
func CreatePatch(original, modified jsonr.JsonR) (Patch, error) {
	return createPatchRecursive(original, modified, "/")
}

func createPatchRecursive(original, modified jsonr.JsonR, path string) (Patch, error) {
	var patch Patch

	switch orig := original.(type) {
	case jsonr.Object: // Object
		mod, ok := modified.(jsonr.Object)
		if !ok {
			// 특정 Object가 통째로 바뀌는 경우
			// Value에는 Timestamp가 제거된 (일반 JSON) 오브젝트가 들어감
			// Timestamp는 가장 최근에 바뀐 Timestamp로 설정됨
			timestamp := findLatestTimestamp(modified)
			patch = append(patch, Operation{
				Op:        OpReplace,
				Path:      path,
				Value:     removeTimestamp(modified),
				Timestamp: timestamp,
			})
			break
		}

		for key, origVal := range orig {
			if modVal, ok := mod[key]; ok {
				// 1. Has Key
				subPatch, err := createPatchRecursive(origVal, modVal, path+key+"/")
				if err != nil {
					return nil, err
				}
				patch = append(patch, subPatch...)
			} else {
				// 2. No Key (Deleted)
				if leaf, ok := origVal.(jsonr.Leaf[string]); ok {
					patch = append(patch, Operation{
						Op:        OpRemove,
						Path:      path + key,
						Timestamp: leaf.Timestamp,
					})
				} else if leaf, ok := origVal.(jsonr.Leaf[float64]); ok {
					patch = append(patch, Operation{
						Op:        OpRemove,
						Path:      path + key,
						Timestamp: leaf.Timestamp,
					})
				} else if leaf, ok := origVal.(jsonr.Leaf[bool]); ok {
					patch = append(patch, Operation{
						Op:        OpRemove,
						Path:      path + key,
						Timestamp: leaf.Timestamp,
					})
				} else {
					// TODO: 여기 섹션 날릴까. Error로.
					patch = append(patch, Operation{
						Op:        OpRemove,
						Path:      path + key,
						Timestamp: 0,
					})
				}
			}
		}
		for key, modVal := range mod {
			if _, ok := orig[key]; !ok {
				// Leaf의 특정 값이 바뀌거나 삭제되는 경우
				// Leaf의 Value는 Operation의 Value가 됨
				// Timestamp는 Operation의 Timestamp가 됨
				if leaf, ok := modVal.(jsonr.Leaf[string]); ok {
					patch = append(patch, Operation{
						Op:        OpAdd,
						Path:      path + key,
						Value:     leaf.Value,
						Timestamp: leaf.Timestamp,
					})
				} else if leaf, ok := modVal.(jsonr.Leaf[float64]); ok {
					patch = append(patch, Operation{
						Op:        OpAdd,
						Path:      path + key,
						Value:     leaf.Value,
						Timestamp: leaf.Timestamp,
					})
				} else if leaf, ok := modVal.(jsonr.Leaf[bool]); ok {
					patch = append(patch, Operation{
						Op:        OpAdd,
						Path:      path + key,
						Value:     leaf.Value,
						Timestamp: leaf.Timestamp,
					})
				} else {
					patch = append(patch, Operation{
						Op:        OpAdd,
						Path:      path + key,
						Value:     removeTimestamp(modVal),
						Timestamp: findLatestTimestamp(modVal),
					})
				}
			}
		}

	case jsonr.Array: // Array
		mod, ok := modified.(jsonr.Array)
		if !ok {
			timestamp := findLatestTimestamp(modified)
			patch = append(patch, Operation{
				Op:        OpReplace,
				Path:      path,
				Value:     removeTimestamp(modified),
				Timestamp: timestamp,
			})
			break
		}

		// 배열의 경우, 간단하게는 전체를 Replace하는 것으로 처리
		if !reflect.DeepEqual(orig, mod) {
			timestamp := findLatestTimestamp(modified)
			patch = append(patch, Operation{
				Op:        OpReplace,
				Path:      path,
				Value:     removeTimestamp(modified),
				Timestamp: timestamp,
			})
		}

	case jsonr.Leaf[string]: // Leaf
		mod, ok := modified.(jsonr.Leaf[string])
		if !ok || orig.Value != mod.Value {
			patch = append(patch, Operation{
				Op:        OpReplace,
				Path:      path,
				Value:     mod.Value,
				Timestamp: mod.Timestamp,
			})
		}

	case jsonr.Leaf[float64]: // Leaf
		mod, ok := modified.(jsonr.Leaf[float64])
		if !ok || orig.Value != mod.Value {
			patch = append(patch, Operation{
				Op:        OpReplace,
				Path:      path,
				Value:     mod.Value,
				Timestamp: mod.Timestamp,
			})
		}

	case jsonr.Leaf[bool]: // Leaf
		mod, ok := modified.(jsonr.Leaf[bool])
		if !ok || orig.Value != mod.Value {
			patch = append(patch, Operation{
				Op:        OpReplace,
				Path:      path,
				Value:     mod.Value,
				Timestamp: mod.Timestamp,
			})
		}

	default:
		return nil, fmt.Errorf("Unsupported type: %T", original)
	}

	return patch, nil
}

// ApplyPatch applies the given JSON patch to the original JSON-R document.
func ApplyPatch(doc jsonr.JsonR, patch Patch) (jsonr.JsonR, error) {
	return applyPatchRecursive(doc, patch, "/")
}

func applyPatchRecursive(doc jsonr.JsonR, patch Patch, path string) (jsonr.JsonR, error) {
	for _, op := range patch {
		if !strings.HasPrefix(op.Path, path) {
			continue
		}

		switch op.Op {
		case OpAdd:
			// split op.Path with "/"
			pathSegments := strings.Split(op.Path, "/")[1:]
			if len(pathSegments) < 1 {
				return nil, fmt.Errorf("Invalid path: %s", op.Path)
			}

			// 마지막 세그먼트를 제외한 나머지로 순회
			var current any = doc
			for _, segment := range pathSegments[:len(pathSegments)-1] {
				switch val := current.(type) {
				case jsonr.Object:
					var ok bool
					if current, ok = val[segment]; !ok {
						return nil, fmt.Errorf("Path not found: %s", op.Path)
					}

				case jsonr.Array:
					index, err := parseIndex(segment)
					if err != nil {
						return nil, fmt.Errorf("Invalid array index: %s", segment)
					}

					if index < 0 || index >= len(val) {
						return nil, fmt.Errorf("Index out of range: %s", segment)
					}
					current = val[index]

				default:
					return nil, fmt.Errorf("Invalid path: %s", op.Path)
				}
			}

			// 마지막 세그먼트 Add
			switch val := current.(type) {
			case jsonr.Object:
				val[pathSegments[len(pathSegments)-1]] = jsonr.Leaf[string]{
					Value:     fmt.Sprintf("%v", op.Value),
					Timestamp: op.Timestamp,
				}

			case jsonr.Array:
				index, err := parseIndex(pathSegments[len(pathSegments)-1])
				if err != nil {
					return nil, fmt.Errorf("Invalid array index: %s", pathSegments[len(pathSegments)-1])
				}
				if index < 0 || index > len(val) {
					return nil, fmt.Errorf("Index out of range: %s", pathSegments[len(pathSegments)-1])
				}

				// 배열에 추가. op.Value 타입 변환
				switch opValue := op.Value.(type) {
				case string:
					val = append(val[:index], append(
						[]jsonr.Array{jsonr.Leaf[string]{Value: opValue, Timestamp: op.Timestamp}},
						val[index:]...)...)
				case float64:
					val = append(val[:index], append(
						[]jsonr.Array{jsonr.Leaf[float64]{Value: opValue, Timestamp: op.Timestamp}},
						val[index:]...)...)
				case bool:
					val = append(val[:index], append(
						[]jsonr.Array{jsonr.Leaf[bool]{Value: opValue, Timestamp: op.Timestamp}},
						val[index:]...)...)
				default: // Object, Array, ...
					val = append(val[:index], append(json.Array{op.Value}, val[index:]...)...)
				}
			default:
				return nil, fmt.Errorf("Invalid path for add operation: %s", op.Path)

			}

		case OpRemove:
			// split op.Path with "/"
			pathSegments := strings.Split(op.Path, "/")[1:]
			if len(pathSegments) < 1 {
				return nil, fmt.Errorf("Invalid path: %s", op.Path)
			}

			// 마지막 세그먼트를 제외한 나머지로 순회
			var current any = doc
			for _, segment := range pathSegments[:len(pathSegments)-1] {
				switch val := current.(type) {
				case jsonr.Object:
					var ok bool
					if current, ok = val[segment]; !ok {
						return nil, fmt.Errorf("Path not found: %s", op.Path)
					}

				case jsonr.Array:
					index, err := parseIndex(segment)
					if err != nil {
						return nil, fmt.Errorf("Invalid array index: %s", segment)
					}

					if index < 0 || index >= len(val) {
						return nil, fmt.Errorf("Index out of range: %s", segment)
					}
					current = val[index]

				default:
					return nil, fmt.Errorf("Invalid path: %s", op.Path)
				}
			}

			// 마지막 세그먼트 Remove
			switch val := current.(type) {
			case jsonr.Object:
				delete(val, pathSegments[len(pathSegments)-1])
			case jsonr.Array:
				index, err := parseIndex(pathSegments[len(pathSegments)-1])
				if err != nil {
					return nil, fmt.Errorf("Invalid array index: %s", pathSegments[len(pathSegments)-1])
				}

				if index < 0 || index >= len(val) {
					return nil, fmt.Errorf("Index out of range: %s", pathSegments[len(pathSegments)-1])
				}
				val = append(val[:index], val[index+1:]...)

			default:
				return nil, fmt.Errorf("Invalid path: %s", op.Path)
			}

		case OpReplace:
			// split op.Path with "/"
			pathSegments := strings.Split(op.Path, "/")[1:]
			if len(pathSegments) < 1 {
				return nil, fmt.Errorf("Invalid path: %s", op.Path)
			}

			// 마지막 세그먼트를 제외한 나머지로 순회
			var current any = doc
			for _, segment := range pathSegments[:len(pathSegments)-1] {
				switch val := current.(type) {
				case jsonr.Object:
					var ok bool
					if current, ok = val[segment]; !ok {
						return nil, fmt.Errorf("Path not found: %s", op.Path)
					}

				case jsonr.Array:
					index, err := parseIndex(segment)
					if err != nil {
						return nil, fmt.Errorf("Invalid array index: %s", segment)
					}

					if index < 0 || index >= len(val) {
						return nil, fmt.Errorf("Index out of range: %s", segment)
					}
					current = val[index]

				default:
					return nil, fmt.Errorf("Invalid path: %s", op.Path)
				}
			}

			// 마지막 세그먼트 Replace
			switch val := current.(type) {
			case jsonr.Object:
				switch opValue := op.Value.(type) {
				case map[string]any: // Object
					obj := jsonr.Object{}
					for k, v := range opValue {
						obj[k] = v
					}

					val[pathSegments[len(pathSegments)-1]] = obj
				default:
					return nil, fmt.Errorf("Invalid value type: %T", op.Value)
				}
			case jsonr.Array:
				index, err := parseIndex(pathSegments[len(pathSegments)-1])
				if err != nil {
					return nil, fmt.Errorf("Invalid array index: %s", pathSegments[len(pathSegments)-1])
				}

				if index < 0 || index >= len(val) {
					return nil, fmt.Errorf("Index out of range: %s", pathSegments[len(pathSegments)-1])
				}

				switch opValue := op.Value.(type) {
				case map[string]any: // Object

					obj := jsonr.Object{}
					for k, v := range opValue {
						obj[k] = v
					}
					val[index] = obj

				case []any: // Array
					arr := jsonr.Array{}
					for _, v := range opValue {
						arr = append(arr, v)
					}
					val[index] = arr

				case string:
					val[index] = jsonr.Leaf[string]{Value: opValue, Timestamp: op.Timestamp}

				case float64:
					val[index] = jsonr.Leaf[float64]{Value: opValue, Timestamp: op.Timestamp}

				case bool:
					val[index] = jsonr.Leaf[bool]{Value: opValue, Timestamp: op.Timestamp}

				default:
					return nil, fmt.Errorf("Invalid value type for replace: %T", op.Value)
				}

			case jsonr.Leaf[string]:
				switch opValue := op.Value.(type) {
				case string:
					val.Value = opValue
					val.Timestamp = op.Timestamp
				default:
					return nil, fmt.Errorf("Invalid value type for replace: %T", op.Value)
				}
			case jsonr.Leaf[float64]:
				switch opValue := op.Value.(type) {
				case float64:
					val.Value = opValue
					val.Timestamp = op.Timestamp
				default:
					return nil, fmt.Errorf("Invalid value type for replace: %T", op.Value)
				}
			case jsonr.Leaf[bool]:
				switch opValue := op.Value.(type) {
				case bool:
					val.Value = opValue
					val.Timestamp = op.Timestamp
				default:
					return nil, fmt.Errorf("Invalid value type for replace: %T", op.Value)
				}
			default:
				return nil, fmt.Errorf("Invalid path for replace operation: %s", op.Path)
			}
		}
	}
	return doc, nil
}
