//
// tson.go
//
// Defines the TSON type and provides functions
// for TSON manipulation.
//
// TSON (Time-Stamped jsON) is a data format that
// extends JSON by adding a timestamp to each leaf value.
//
// TSON is a recursive data structure that can be
// represented as an object, array, or leaf.
//
// Author: Karu (@karu-rress)
//

package tson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Tson stores a TSON document.
type Tson = Value

// Value represents a TSON value.
type Value interface{ isValue() }

// Object stores an object of TSON.
type Object map[string]Value

func (Object) isValue() {} // Marks Object as a Value

// Array stores an array of TSON.
type Array []Value

func (Array) isValue() {} // Marks Array as a Value

// LeafValue is a type that can be used as a value of TSON leaf.
type LeafValue interface{ ~string | ~float64 | ~bool }

// Leaf[T] stores the value and timestamp of a TSON leaf.
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

// GetValue returns the value of the given key from the TSON object.
func GetValue(t Tson, path string) (v Value, err error) {
	var getValue func(t Tson, parts []string) (Value, error)
	getValue = func(t Tson, parts []string) (Value, error) {
		if len(parts) == 1 {
			// If j is an Object, return the value of the key
			if obj, ok := t.(Object); ok {
				if value, ok := obj[parts[0]]; ok {
					return value, nil
				} else {
					return nil, fmt.Errorf("key not found: %s", parts[0])
				}
			} else if arr, ok := t.(Array); ok {
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
				return nil, fmt.Errorf("invalid type: %T", t)
			}
		}

		// If j is an Object, get the value of the key and call getValue recursively
		if obj, ok := t.(Object); ok {
			if value, ok := obj[parts[0]]; ok {
				return getValue(value, parts[1:])
			} else {
				return nil, fmt.Errorf("key not found: %s", parts[0])
			}
		} else if arr, ok := t.(Array); ok {
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
			return nil, fmt.Errorf("Cannot recursively get value from %T", t)
		}
	}

	rfc6901Decoder := strings.NewReplacer("~1", "/", "~0", "~")
	path = rfc6901Decoder.Replace(path)
	parts := strings.Split(path, "/")[1:]

	return getValue(t, parts)
}

// Equal checks if two TSONs are equal, including timestamps.
func Equal(j1, j2 Tson) (ret bool, err error) {
	var (
		o1, o2 any
		b1, _  = ToCompatibleTsonBytes(j1)
		b2, _  = ToCompatibleTsonBytes(j2)
	)

	// Check if the given TSONs are valid
	if err = json.Unmarshal(b1, &o1); err != nil {
		return false, err
	}
	if err = json.Unmarshal(b2, &o2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(o1, o2), nil
}

// EqualWithoutTimestamp checks if two TSONs are equal, excluding timestamps.
func EqualWithoutTimestamp(j1, j2 Tson) (ret bool, err error) {
	var (
		o1, o2       any
		json1, json2 []byte
	)

	// Convert TSONs to JSON
	if json1, err = ToJsonBytes(j1); err != nil {
		return false, err
	}
	if json2, err = ToJsonBytes(j2); err != nil {
		return false, err
	}

	// Check if the given TSONs are equal
	if err = json.Unmarshal(json1, &o1); err != nil {
		return false, err
	}
	if err = json.Unmarshal(json2, &o2); err != nil {
		return false, err
	}

	return reflect.DeepEqual(o1, o2), nil
}

// GetLatestTimestamp returns the latest timestamp of the given TSON.
func GetLatestTimestamp(t Tson) int64 {
	updateMax := func(max *int64, ts int64) {
		if ts > *max {
			*max = ts
		}
	}

	switch v := t.(type) {
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
		panic(fmt.Sprintf("GetTimestamp: invalid type %T for TSON", t))
	}
}

//////////////////////////////////
///////// CONVERSIONS
//////////////////////////////////

// All conversion functions has 'To' prefix

// ToArray converts the given TSON array to a Go slice.
func ToArray(a Array) (arr []any, err error) {
	arr = make([]any, len(a))
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

// ToAny returns the given TSON as any type.
func ToAny(t Tson) any {
	return t
}

// ToCompatibleTsonBytes converts the given TSON to a JSON byte slice.
// NOTE: The timestamp field is kept
func ToCompatibleTsonBytes(t Tson) (j []byte, err error) {
	return json.Marshal(t)
}

// ToCompatibleTson converts the given TSON to a JSON object.
// NOTE: The timestamp field is kept
func ToCompatibleTson(t Tson, o *any) error {
	// Convert TSON to Compatible TSON (JSON with timestamp)
	barr, err := ToCompatibleTsonBytes(t)
	if err != nil {
		return err
	}
	// Convert JSON byte slice to JSON object
	if err = json.Unmarshal(barr, o); err != nil {
		return err
	}

	return nil
}

// ToJson converts the given TSON to a JSON byte slice.
// NOTE: The timestamp field is removed
func ToJson(t Tson, j *any) (err error) {
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

	ToCompatibleTson(t, j)
	removeTimestamp(*j)

	return nil
}

// ToJsonBytes converts the given TSON to a JSON byte slice.
// NOTE: The timestamp field is removed
func ToJsonBytes(t Tson) (j []byte, err error) {
	var o any
	if err = ToJson(t, &o); err != nil {
		return nil, err
	}

	return json.Marshal(o)
}

// ToTson converts the given JSON object to a TSON.
// NOTE: The timestamp field is added with the default(< 0) value.
func ToTson(o any, t *Tson) (err error) {
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

	// Step 1: Convert TSON(any) to CompatibleTSON(JSON with timestamp)
	str, err := json.Marshal(o)

	// Step 2: Convert CompatibleTSON to TSON
	*t, err = FromJsonBytes(str)

	return err
}

//////////////////////////////////
///////// PRIVATE
//////////////////////////////////

// parseValue parses the given value and returns a TSON value.
func parseValue(raw any) (Value, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case map[string]any:
		if isLeaf(v) { // type is a TSON leaf
			return parseLeaf(v)
		} else { // type is a TSON object
			return parseObject(v)
		}
	case []any: // type is a TSON array
		return parseArray(v)
	default:
		return nil, fmt.Errorf("invalid value type: %v (type: %T)", v, v)
	}
}

// parseObject parses the given map as a TSON object.
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

// parseArray parses the given array as a TSON array.
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

// parseLeaf parses the given map as a TSON leaf.
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

// isLeaf checks if the given map is a TSON leaf.
func isLeaf(m map[string]any) bool {
	// A TSON Object is a Leaf when
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
