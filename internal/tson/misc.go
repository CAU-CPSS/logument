package tson

import "encoding/json"

// FromCompatibleTsonBytes converts a JSON byte array (with leaf nodes as { "value": ..., "timestamp": ... })
// into a Tson object.
func FromCompatibleTsonBytes(data []byte, t *Tson) error {
	var intermediate interface{}
	if err := json.Unmarshal(data, &intermediate); err != nil {
		return err
	}
	*t = convertJSONToTson(intermediate)
	return nil
}

func convertJSONToTson(v interface{}) Value {
	switch t := v.(type) {
	case map[string]interface{}:
		// leaf 노드 판별: 딱 "value"와 "timestamp" 키만 있다면
		if isLeaf, leafVal, ts := checkLeaf(t); isLeaf {
			switch leafVal.(type) {
			case string:
				return Leaf[string]{Value: leafVal.(string), Timestamp: ts}
			case float64:
				return Leaf[float64]{Value: leafVal.(float64), Timestamp: ts}
			case bool:
				return Leaf[bool]{Value: leafVal.(bool), Timestamp: ts}
			default:
				// 지원하지 않는 타입은 무시하거나 에러 처리
				return nil
			}
		}
		// 일반 객체 처리
		obj := Object{}
		for key, val := range t {
			obj[key] = convertJSONToTson(val)
		}
		return obj
	case []interface{}:
		arr := Array{}
		for _, elem := range t {
			arr = append(arr, convertJSONToTson(elem))
		}
		return arr
	default:
		// 원시값의 경우 Leaf로 감싸 timestamp를 0으로 설정
		switch t := v.(type) {
		case string:
			return Leaf[string]{Value: t, Timestamp: 0}
		case float64:
			return Leaf[float64]{Value: t, Timestamp: 0}
		case bool:
			return Leaf[bool]{Value: t, Timestamp: 0}
		default:
			return nil
		}
	}
}

func JsonToTson(o any, t *Tson) error {
	data, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return FromCompatibleTsonBytes(data, t)
}
