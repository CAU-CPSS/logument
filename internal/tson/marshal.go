package tson

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func ToJsonBytes(t Tson) ([]byte, error) {
	return json.Marshal(t)
}

// Marshal serializes a Tson value into TSON-formatted bytes.
func Marshal(t Tson) ([]byte, error) {
	s, err := marshalTson(t, "top")
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// marshalTson marshals a Tson value given its context:
// "top": top-level value,
// "object": value inside an object (its timestamp is printed with the key),
// "array": element inside an array (timestamp is printed before the primitive).
func marshalTson(v Value, ctx string) (string, error) {
	switch val := v.(type) {
	case Object:
		return marshalObject(val)
	case Array:
		return marshalArray(val)
	case Leaf[string]:
		if ctx == "array" {
			// In array, prefix timestamp before primitive.
			if val.Timestamp != 0 {
				return fmt.Sprintf("<%d> %q", val.Timestamp, val.Value), nil
			}
			return fmt.Sprintf("%q", val.Value), nil
		} else if ctx == "object" {
			// In object, the timestamp is printed with the key.
			return fmt.Sprintf("%q", val.Value), nil
		} else { // top-level
			if val.Timestamp != 0 {
				return fmt.Sprintf("%q <%d>", val.Value, val.Timestamp), nil
			}
			return fmt.Sprintf("%q", val.Value), nil
		}
	case Leaf[float64]:
		if ctx == "array" {
			if val.Timestamp != 0 {
				return fmt.Sprintf("<%d> %v", val.Timestamp, val.Value), nil
			}
			return fmt.Sprintf("%v", val.Value), nil
		} else if ctx == "object" {
			return fmt.Sprintf("%v", val.Value), nil
		} else {
			if val.Timestamp != 0 {
				return fmt.Sprintf("%v <%d>", val.Value, val.Timestamp), nil
			}
			return fmt.Sprintf("%v", val.Value), nil
		}
	case Leaf[bool]:
		if ctx == "array" {
			if val.Timestamp != 0 {
				return fmt.Sprintf("<%d> %v", val.Timestamp, val.Value), nil
			}
			return fmt.Sprintf("%v", val.Value), nil
		} else if ctx == "object" {
			return fmt.Sprintf("%v", val.Value), nil
		} else {
			if val.Timestamp != 0 {
				return fmt.Sprintf("%v <%d>", val.Value, val.Timestamp), nil
			}
			return fmt.Sprintf("%v", val.Value), nil
		}
	default:
		return "", fmt.Errorf("marshalTson: unknown type")
	}
}

// marshalObject marshals an Object into a TSON object string.
func marshalObject(obj Object) (string, error) {
	s := "{"
	first := true
	for key, value := range obj {
		if !first {
			s += ", "
		} else {
			first = false
		}
		keyStr := fmt.Sprintf("%q", key)
		tsStr := ""
		switch leaf := value.(type) {
		case Leaf[string]:
			if leaf.Timestamp != 0 {
				tsStr = fmt.Sprintf(" <%d>", leaf.Timestamp)
			}
		case Leaf[float64]:
			if leaf.Timestamp != 0 {
				tsStr = fmt.Sprintf(" <%d>", leaf.Timestamp)
			}
		case Leaf[bool]:
			if leaf.Timestamp != 0 {
				tsStr = fmt.Sprintf(" <%d>", leaf.Timestamp)
			}
		}
		s += keyStr + tsStr + ": "
		marshaledValue, err := marshalTson(value, "object")
		if err != nil {
			return "", err
		}
		s += marshaledValue
	}
	s += "}"
	return s, nil
}

// marshalArray marshals an Array into a TSON array string.
func marshalArray(arr Array) (string, error) {
	s := "["
	first := true
	for _, elem := range arr {
		if !first {
			s += ", "
		} else {
			first = false
		}
		marshaledElem, err := marshalTson(elem, "array")
		if err != nil {
			return "", err
		}
		s += marshaledElem
	}
	s += "]"
	return s, nil
}

// //

// MarshalIndent serializes a JSON-like or Tson document into a TSON-formatted byte slice,
// using the given prefix and indent (similar to encoding/json.MarshalIndent).
func MarshalIndent(j interface{}, prefix, indent string) ([]byte, error) {
	s, err := marshalIndentValue(j, prefix, indent, "top")
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// marshalIndentValue recursively serializes the value with pretty-printing.
// ctx indicates the context ("top", "object", or "array") which affects where the timestamp is printed.
func marshalIndentValue(v interface{}, currentIndent, indent, ctx string) (string, error) {
	switch t := v.(type) {
	// 기본 원시 타입들
	case nil:
		return "null", nil
	case bool:
		return fmt.Sprintf("%v", t), nil
	case float64:
		return fmt.Sprintf("%v", t), nil
	case string:
		return strconv.Quote(t), nil

	// 일반 JSON 객체: map[string]interface{}
	case map[string]interface{}:
		// 만약 이 map가 leaf라면 (즉, 딱 "value"와 "timestamp"만 있다면)
		if isLeaf, leafVal, ts := checkLeaf(t); isLeaf {
			primStr, err := formatPrimitive(leafVal)
			if err != nil {
				return "", err
			}
			switch ctx {
			case "object":
				return primStr, nil
			case "array":
				return fmt.Sprintf("<%d> %s", ts, primStr), nil
			case "top":
				return fmt.Sprintf("%s <%d>", primStr, ts), nil
			default:
				return fmt.Sprintf("%s <%d>", primStr, ts), nil
			}
		}
		// 일반 객체 처리
		s := "{\n"
		first := true
		for key, val := range t {
			if !first {
				s += ",\n"
			}
			first = false
			keyStr := strconv.Quote(key)
			// 만약 val이 leaf라면 key에 timestamp를 붙여 출력한다.
			if m, ok := val.(map[string]interface{}); ok {
				if isLeaf, leafVal, ts := checkLeaf(m); isLeaf {
					primStr, err := formatPrimitive(leafVal)
					if err != nil {
						return "", err
					}
					s += currentIndent + indent + keyStr + " <" + strconv.FormatInt(ts, 10) + ">: " + primStr
					continue
				}
			}
			valStr, err := marshalIndentValue(val, currentIndent+indent, indent, "object")
			if err != nil {
				return "", err
			}
			s += currentIndent + indent + keyStr + ": " + valStr
		}
		s += "\n" + currentIndent + "}"
		return s, nil

	// 일반 JSON 배열: []interface{}
	case []interface{}:
		s := "[\n"
		first := true
		for _, elem := range t {
			if !first {
				s += ",\n"
			}
			first = false
			// 배열 요소가 leaf인 경우 처리 (timestamp를 요소 앞에 붙임)
			if m, ok := elem.(map[string]interface{}); ok {
				if isLeaf, leafVal, ts := checkLeaf(m); isLeaf {
					primStr, err := formatPrimitive(leafVal)
					if err != nil {
						return "", err
					}
					s += currentIndent + indent + fmt.Sprintf("<%d> %s", ts, primStr)
					continue
				}
			}
			elemStr, err := marshalIndentValue(elem, currentIndent+indent, indent, "array")
			if err != nil {
				return "", err
			}
			s += currentIndent + indent + elemStr
		}
		s += "\n" + currentIndent + "]"
		return s, nil

	// Tson 타입: Object
	case Object:
		s := "{\n"
		first := true
		for key, val := range t {
			if !first {
				s += ",\n"
			}
			first = false
			keyStr := strconv.Quote(key)
			// Tson의 값이 Leaf라면, key 뒤에 timestamp를 붙여 출력
			switch leaf := val.(type) {
			case Leaf[string]:
				s += currentIndent + indent + keyStr + " <" + strconv.FormatInt(leaf.Timestamp, 10) + ">: " + strconv.Quote(leaf.Value)
			case Leaf[float64]:
				s += currentIndent + indent + keyStr + " <" + strconv.FormatInt(leaf.Timestamp, 10) + ">: " + fmt.Sprintf("%v", leaf.Value)
			case Leaf[bool]:
				s += currentIndent + indent + keyStr + " <" + strconv.FormatInt(leaf.Timestamp, 10) + ">: " + fmt.Sprintf("%v", leaf.Value)
			default:
				valStr, err := marshalIndentValue(val, currentIndent+indent, indent, "object")
				if err != nil {
					return "", err
				}
				s += currentIndent + indent + keyStr + ": " + valStr
			}
		}
		s += "\n" + currentIndent + "}"
		return s, nil

	// Tson 타입: Array
	case Array:
		s := "[\n"
		first := true
		for _, elem := range t {
			if !first {
				s += ",\n"
			}
			first = false
			switch leaf := elem.(type) {
			case Leaf[string]:
				s += currentIndent + indent + fmt.Sprintf("<%d> %s", leaf.Timestamp, strconv.Quote(leaf.Value))
			case Leaf[float64]:
				s += currentIndent + indent + fmt.Sprintf("<%d> %v", leaf.Timestamp, leaf.Value)
			case Leaf[bool]:
				s += currentIndent + indent + fmt.Sprintf("<%d> %v", leaf.Timestamp, leaf.Value)
			default:
				elemStr, err := marshalIndentValue(elem, currentIndent+indent, indent, "array")
				if err != nil {
					return "", err
				}
				s += currentIndent + indent + elemStr
			}
		}
		s += "\n" + currentIndent + "]"
		return s, nil

	// Tson 타입: Leaf
	case Leaf[string]:
		primStr := strconv.Quote(t.Value)
		if ctx == "array" {
			return fmt.Sprintf("<%d> %s", t.Timestamp, primStr), nil
		} else if ctx == "top" {
			return fmt.Sprintf("%s <%d>", primStr, t.Timestamp), nil
		}
		return primStr, nil
	case Leaf[float64]:
		primStr := fmt.Sprintf("%v", t.Value)
		if ctx == "array" {
			return fmt.Sprintf("<%d> %s", t.Timestamp, primStr), nil
		} else if ctx == "top" {
			return fmt.Sprintf("%s <%d>", primStr, t.Timestamp), nil
		}
		return primStr, nil
	case Leaf[bool]:
		primStr := fmt.Sprintf("%v", t.Value)
		if ctx == "array" {
			return fmt.Sprintf("<%d> %s", t.Timestamp, primStr), nil
		} else if ctx == "top" {
			return fmt.Sprintf("%s <%d>", primStr, t.Timestamp), nil
		}
		return primStr, nil

	default:
		return "", fmt.Errorf("unsupported type: %T", v)
	}
}

// checkLeaf determines whether m (map[string]interface{}) is a leaf node.
// leaf node이면 딱 "value"와 "timestamp" 키만 가진다고 판단한다.
func checkLeaf(m map[string]interface{}) (bool, interface{}, int64) {
	if len(m) == 2 {
		val, okVal := m["value"]
		tsVal, okTs := m["timestamp"]
		if okVal && okTs {
			switch t := tsVal.(type) {
			case float64:
				return true, val, int64(t)
			case int:
				return true, val, int64(t)
			case int64:
				return true, val, t
			}
		}
	}
	return false, nil, 0
}

// formatPrimitive converts a primitive value (string, float64, bool) into its string representation.
func formatPrimitive(v interface{}) (string, error) {
	switch t := v.(type) {
	case string:
		return strconv.Quote(t), nil
	case float64, int, int64, uint, uint64:
		return fmt.Sprintf("%v", t), nil
	case bool:
		return fmt.Sprintf("%v", t), nil
	default:
		return "", fmt.Errorf("unsupported primitive type: %T", v)
	}
}
