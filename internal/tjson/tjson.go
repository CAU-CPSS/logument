//
// tjson.go
//
// Defines the T-JSON type and provides functions
// for T-JSON manipulation.
//
// T-JSON (Temporal JSON) is a data format that
// extends JSON by adding a timestamp to each leaf value.
//
// T-JSON is a recursive data structure that can be
// represented as an object, array, or leaf.
//
// Author: Karu (@karu-rress)

package tjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// TJson stores a T-JSON document.
type TJson = Value

// Value represents a T-JSON value.
type Value interface{ isValue() }

// Object stores an object of T-JSON.
type Object map[string]Value

func (Object) isValue() {} // Marks Object as a Value

// Array stores an array of T-JSON.
type Array []Value

func (Array) isValue() {} // Marks Array as a Value

// LeafValue is a type that can be used as a value of T-JSON leaf.
type LeafValue interface{ ~string | ~float64 | ~bool }


// Leaf[T] stores the value and timestamp of a T-JSON leaf.
type Leaf[T LeafValue] struct {
	Value     T     `json:"value"`     // number, string, boolean
	Timestamp int64 `json:"timestamp"` // Unix timestamp
}

func (Leaf[T]) isValue() {} // Marks Leaf as a Value

// DefaultTimestamp is the default timestamp value.
const DefaultTimestamp int64 = -1

//////////////////////////////////
///////// OPERATIONS
//////////////////////////////////

// GetValue returns the value of the given key from the T-JSON object.
func GetValue(j TJson, path string) (v Value, err error) {
	var getValue func(j TJson, parts []string) (Value, error)
	getValue = func(j TJson, parts []string) (Value, error) {
		if len(parts) == 1 {
			// If j is an Object, return the value of the key
			if obj, ok := j.(Object); ok {
				if value, ok := obj[parts[0]]; ok {
					return value, nil
				} else {
					return nil, fmt.Errorf("key not found: %s", parts[0])
				}
			} else if arr, ok := j.(Array); ok {
				// If j is an Array, return the value of the index
				index, err := strconv.Atoi(parts[0])
				if err != nil {
					return nil, fmt.Errorf("invalid index: %s", parts[0])
				}
				if index < 0 || index >= len(arr) {
					return nil, fmt.Errorf("index out of range: %d", index)
				}
				return arr[index], nil
			} else {
				return nil, fmt.Errorf("invalid type: %T", j)
			}
		}

		// If j is an Object, get the value of the key and call getValue recursively
		if obj, ok := j.(Object); ok {
			if value, ok := obj[parts[0]]; ok {
				return getValue(value, parts[1:])
			} else {
				return nil, fmt.Errorf("key not found: %s", parts[0])
			}
		} else if arr, ok := j.(Array); ok {
			// If j is an Array, get the value of the index and call getValue recursively
			index, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid index: %s", parts[0])
			}
			if index < 0 || index >= len(arr) {
				return nil, fmt.Errorf("index out of range: %d", index)
			}
			return getValue(arr[index], parts[1:])
		} else {
			return nil, fmt.Errorf("Cannot recursively get value from %T", j)
		}
	}

	rfc6901Decoder := strings.NewReplacer("~1", "/", "~0", "~")
	path = rfc6901Decoder.Replace(path)
	parts := strings.Split(path, "/")[1:]

	return getValue(j, parts)
}

// Marshal converts the T-JSON to a byte slice.
func Marshal(j TJson) (data []byte, err error) {
	return json.Marshal(j)
}

// MarshalIndent converts the T-JSON to a byte slice with indent.
func MarshalIndent(j TJson, prefix, indent string) (data []byte, err error) {
	return json.MarshalIndent(j, prefix, indent)
}

// Equal checks if two T-JSONs are equal, including timestamps.
func Equal(j1, j2 TJson) (ret bool, err error) {
	var (
		o1, o2 any
		b1, _  = Marshal(j1)
		b2, _  = Marshal(j2)
	)

	// Check if the given T-JSONs are valid
	if err = json.Unmarshal(b1, &o1); err != nil {
		return false, err
	}
	if err = json.Unmarshal(b2, &o2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(o1, o2), nil
}

// EqualWithoutTimestamp checks if two T-JSONs are equal, excluding timestamps.
func EqualWithoutTimestamp(j1, j2 TJson) (ret bool, err error) {
	var (
		o1, o2       any
		json1, json2 []byte
	)

	// Convert T-JSONs to JSON
	if json1, err = ToJson(j1); err != nil {
		return false, err
	}
	if json2, err = ToJson(j2); err != nil {
		return false, err
	}

	// Check if the given T-JSONs are equal
	if err = json.Unmarshal(json1, &o1); err != nil {
		return false, err
	}
	if err = json.Unmarshal(json2, &o2); err != nil {
		return false, err
	}

	return reflect.DeepEqual(o1, o2), nil
}

// GetLatestTimestamp returns the latest timestamp of the given T-JSON.
func GetLatestTimestamp(j TJson) int64 {
	updateMax := func(max *int64, ts int64) {
		if ts > *max {
			*max = ts
		}
	}

	switch v := j.(type) {
	case Leaf[string]:
		return v.Timestamp
	case Leaf[float64]:
		return v.Timestamp
	case Leaf[bool]:
		return v.Timestamp
	case Object, Array:
		max := int64(0)
		if obj, ok := v.(Object); ok { // v is Object
			for _, value := range obj {
				updateMax(&max, GetLatestTimestamp(value))
			}
		} else { // v is Array
			for _, value := range v.(Array) {
				updateMax(&max, GetLatestTimestamp(value))
			}
		}
		return max
	default:
		panic(fmt.Sprintf("GetTimestamp: invalid type %T for T-JSON", j))
	}
}

//////////////////////////////////
///////// PARSER
//////////////////////////////////

// NewTJson creates a new T-JSON from the given T-JSON string.
func NewTJson(tjson string) (TJson, error) {
	var j TJson
	err := Unmarshal([]byte(tjson), &j)
	return j, err
}

// Unmarshal parses the T-JSON data and stores the result.
func Unmarshal(data []byte, j *TJson) (err error) {
	var raw any
	if err = json.Unmarshal(data, &raw); err != nil {
		return
	}

	*j, err = parseValue(raw)
	return err
}

// String converts T-JSON to a string.
func String(j TJson) string {
	data, err := json.MarshalIndent(j, "", "    ")
	if err != nil {
		panic(fmt.Sprintf("failed to marshal T-JSON: %v", err))
	}
	return string(data)
}

//////////////////////////////////
///////// CONVERSIONS
//////////////////////////////////

// TODO
// 1. Array -> []any
// 2. Object -> map[string]any
// ...

func ToArray(a Array) ([]any, error) {
	length := len(a)
	arr := make([]any, length)

	for i, value := range a {
		switch v := value.(type) {
		case Leaf[string]:
			arr[i] = v.Value
		case Leaf[float64]:
			arr[i] = v.Value
		case Leaf[bool]:
			arr[i] = v.Value
		default:
			return nil, fmt.Errorf("All element in a should be leaf")
		}
	}
	return arr, nil
}

// Any returns the given T-JSON as any(interface{}) type.
func Any(j TJson) any {
	return j
}

// ToJson converts the given T-JSON to a JSON byte slice.
// NOTE: The timestamp field is removed
func ToJson(j TJson) (b []byte, err error) {
	var removeTimestamp func(o any)
	removeTimestamp = func(o any) {
		switch value := o.(type) {
		case map[string]any:
			for k, v := range value {
				switch leaf := v.(type) {
				case Leaf[string]:
					value[k] = leaf.Value
				case Leaf[float64]:
					value[k] = leaf.Value
				case Leaf[bool]:
					value[k] = leaf.Value
				case map[string]any:
					if isLeaf(leaf) {
						value[k] = leaf["value"]
					} else {
						removeTimestamp(v)
					}
				default:
					removeTimestamp(v)
				}
			}
		case []any:
			for i, v := range value {
				switch leaf := v.(type) {
				case Leaf[string]:
					value[i] = leaf.Value
				case Leaf[float64]:
					value[i] = leaf.Value
				case Leaf[bool]:
					value[i] = leaf.Value
				case map[string]any:
					if isLeaf(leaf) {
						value[i] = leaf["value"]
					} else {
						removeTimestamp(leaf)
					}
				default:
					removeTimestamp(v)
				}
			}
		}
	}

	var (
		o    any
		barr []byte
	)
	if barr, err = Marshal(j); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(barr, &o); err != nil {
		return nil, err
	}

	removeTimestamp(o)

	return json.Marshal(o)
}

// ToTJson converts the given JSON object to a T-JSON.
func ToTJson(o any) (j TJson, err error) {
	var addTimestamp func(o any)
	addTimestamp = func(o any) {
		switch value := o.(type) {
		case map[string]any:
			for k, v := range value {
				switch leaf := v.(type) {
				case string:
					value[k] = Leaf[string]{Value: leaf, Timestamp: DefaultTimestamp}
				case float64:
					value[k] = Leaf[float64]{Value: leaf, Timestamp: DefaultTimestamp}
				case bool:
					value[k] = Leaf[bool]{Value: leaf, Timestamp: DefaultTimestamp}
				default:
					addTimestamp(v)
				}
			}
		case []any:
			for i, v := range value {
				switch leaf := v.(type) {
				case string:
					value[i] = Leaf[string]{Value: leaf, Timestamp: DefaultTimestamp}
				case float64:
					value[i] = Leaf[float64]{Value: leaf, Timestamp: DefaultTimestamp}
				case bool:
					value[i] = Leaf[bool]{Value: leaf, Timestamp: DefaultTimestamp}
				default:
					addTimestamp(v)
				}
			}
		}
	}

	addTimestamp(o)

	str, err := json.Marshal(o)
	err = Unmarshal(str, &j)

	return j, err
}

//////////////////////////////////
///////// PRIVATE
//////////////////////////////////

// parseValue parses the given value and returns a T-JSON value.
func parseValue(raw any) (Value, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case map[string]any:
		if isLeaf(v) { // type is a T-JSON leaf
			return parseLeaf(v)
		} else { // type is a T-JSON object
			return parseObject(v)
		}
	case []any: // type is a T-JSON array
		return parseArray(v)
	default:
		return nil, fmt.Errorf("invalid value type: %v (type: %T)", v, v)
	}
}

// parseObject parses the given map as a T-JSON object.
func parseObject(raw map[string]any) (obj Object, err error) {
	obj = make(Object)
	for key, value := range raw {
		val, err := parseValue(value)
		if err != nil {
			return nil, err
		}
		obj[key] = val
	}
	return obj, nil
}

// parseArray parses the given array as a T-JSON array.
func parseArray(raw []any) (arr Array, err error) {
	arr = make(Array, len(raw))
	for i, value := range raw {
		val, err := parseValue(value)
		if err != nil {
			return nil, err
		}
		arr[i] = val
	}
	return arr, nil
}

// parseLeaf parses the given map as a T-JSON leaf.
func parseLeaf(raw map[string]any) (Value, error) {
	// Get value and timestamp
	// The existence of "value" and "timestamp" is already checked in isLeaf()
	var (
		value = raw["value"]
		ts    = int64(raw["timestamp"].(float64))
	)

	// Create Leaf with the type of Value
	switch v := value.(type) {
	case string:
		return Leaf[string]{Value: v, Timestamp: ts}, nil
	case float64:
		return Leaf[float64]{Value: v, Timestamp: ts}, nil
	case bool:
		return Leaf[bool]{Value: v, Timestamp: ts}, nil
	default:
		return nil, fmt.Errorf("invalid leaf type: %v (type: %T)", v, v)
	}
}

// isLeaf checks if the given map is a T-JSON leaf.
func isLeaf(m map[string]any) bool {
	// A T-JSON Object is a Leaf when
	//  1: it has "value" and "timestamp" keys
	//  2: no other keys exist
	_, ok1 := m["value"]
	ts, ok2 := m["timestamp"]
	if !ok1 || !ok2 || len(m) != 2 {
		return false
	}

	if _, ok := ts.(float64); !ok {
		panic(fmt.Sprintf("invalid timestamp type: %T (expected float64)", ts))
	}
	return true
}
