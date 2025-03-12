//
// tson.go
//
// Defines the TSON type and provides functions
// for TSON manipulation.
//
// TSON (Time-Stamped JSON) is a data format that
// extends JSON by adding a timestamp to each leaf value.
//
// Author: Karu (@karu-rress)
//

package tson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
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
		if len(parts) == 1 { // If Object, return the value of the key
			if obj, ok := t.(Object); ok {
				if value, ok := obj[parts[0]]; ok {
					return value, nil
				} else {
					return nil, fmt.Errorf("key not found: %s", parts[0])
				}
			} else if arr, ok := t.(Array); ok { // If Array, return the value of the index
				index, err := strconv.Atoi(parts[0])
				if err != nil {
					return nil, fmt.Errorf("invalid index: %s", parts[0])
				}
				if index < 0 || index >= len(arr) {
					return nil, fmt.Errorf("index out of range: %d", index)
				}
				return arr[index], nil
			} else { // If neither Object nor Array, return error
				return nil, fmt.Errorf("invalid type: %T", t)
			}
		}

		if obj, ok := t.(Object); ok { // If Object, call getValue recursively using the key
			if value, ok := obj[parts[0]]; ok {
				return getValue(value, parts[1:])
			} else {
				return nil, fmt.Errorf("key not found: %s", parts[0])
			}
		} else if arr, ok := t.(Array); ok { // If Array, call getValue recursively using the index
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
func Equal(t1, t2 Tson) (ret bool, err error) {
	var (
		o1, o2 any
		b1, _  = ToCompatibleTsonBytes(t1)
		b2, _  = ToCompatibleTsonBytes(t2)
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
func EqualWithoutTimestamp(t1, t2 Tson) (ret bool, err error) {
	var (
		o1, o2 any
		j1, _  = ToJsonBytes(t1)
		j2, _  = ToJsonBytes(t2)
	)

	// Check if the given TSONs (and converted JSONs) are valid
	if err = json.Unmarshal(j1, &o1); err != nil {
		return false, err
	}
	if err = json.Unmarshal(j2, &o2); err != nil {
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

// Every conversion functions has 'To' or 'From' prefix

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

// FromCompatibleTsonBytes converts a JSON byte array (with leaf nodes as { "value": ..., "timestamp": ... })
// into a Tson object.
func FromCompatibleTsonBytes(data []byte, t *Tson) error {
	var convert func(v any) Value
	convert = func(v any) Value {
		switch t := v.(type) {
		case map[string]any:
			if isLeaf, leafVal, ts := checkLeaf(t); isLeaf {
				switch leafVal.(type) {
				case string:
					return Leaf[string]{Value: leafVal.(string), Timestamp: ts}
				case float64:
					return Leaf[float64]{Value: leafVal.(float64), Timestamp: ts}
				case bool:
					return Leaf[bool]{Value: leafVal.(bool), Timestamp: ts}
				default:
					return nil // NOT supported
				}
			}
			obj := Object{}
			for key, val := range t {
				obj[key] = convert(val)
			}
			return obj
		case []any:
			arr := Array{}
			for _, elem := range t {
				arr = append(arr, convert(elem))
			}
			return arr
		default:
			panic("FromCompatibleTsonBytes: Not a compatible TSON data")
		}
	}

	var intermediate any
	if err := json.Unmarshal(data, &intermediate); err != nil {
		return err
	}
	*t = convert(intermediate)
	return nil
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
					if isLeaf, leafVal, _ := checkLeaf(leaf); isLeaf {
						value[k] = leafVal
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
					if isLeaf, leafVal, _ := checkLeaf(leaf); isLeaf {
						value[i] = leafVal
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

// FromJson converts the given JSON object to a TSON.
// NOTE: The timestamp field is added with the default(< 0) value.
func FromJson(o any, t *Tson) (err error) {
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
	return FromCompatibleTsonBytes(str, t)
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

// FromJsonBytes converts the given JSON byte slice to a TSON.
// NOTE: The timestamp field is added with the default(< 0) value.
func FromJsonBytes(j []byte, t *Tson) error {
	var o any
	if err := json.Unmarshal(j, &o); err != nil {
		return err
	}

	return FromJson(o, t)
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
		if isLeaf, _, _ := checkLeaf(v); isLeaf {
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

// checkLeaf determines whether m (map[string]any) is a leaf node.
func checkLeaf(m map[string]any) (bool, any, int64) {
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

// ToCompatibleTsonBytesSorted converts the TSON to a JSON byte slice
// but ensures that keys in objects are sorted (stable ordering).
func ToCompatibleTsonBytesSorted(t Tson) ([]byte, error) {
	v, err := toSortedInterface(t)
	if err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

// toSortedInterface recursively converts TSON -> map/slice with sorted keys.
func toSortedInterface(t Tson) (any, error) {
	switch val := t.(type) {
	case Leaf[string]:
		return map[string]any{"value": val.Value, "timestamp": val.Timestamp}, nil
	case Leaf[float64]:
		return map[string]any{"value": val.Value, "timestamp": val.Timestamp}, nil
	case Leaf[bool]:
		return map[string]any{"value": val.Value, "timestamp": val.Timestamp}, nil

	case Object:
		// 1. 수집된 key들을 사전순으로 정렬
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// 2. 각 key 순서대로 재귀 변환
		result := make(map[string]any, len(val))
		for _, k := range keys {
			child, err := toSortedInterface(val[k])
			if err != nil {
				return nil, err
			}
			result[k] = child
		}
		return result, nil

	case Array:
		arr := make([]any, len(val))
		for i, child := range val {
			conv, err := toSortedInterface(child)
			if err != nil {
				return nil, err
			}
			arr[i] = conv
		}
		return arr, nil

	default:
		return nil, fmt.Errorf("unexpected type in toSortedInterface: %T", t)
	}
}
