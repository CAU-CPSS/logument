//
// jsonr.go
//
// Defines the JSON-R type and provides functions
// for JSON-R manipulation.
//
// JSON-R (JSON by Rolling Ress, or JSON with Revision)
// is a data format that extends JSON by adding
// a timestamp to each leaf value.
//
// JSON-R is a recursive data structure that can be
// represented as an object, array, or leaf.
//
// Author: Karu (@karu-rress)

package jsonr

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// JsonR stores a JSON-R document.
type JsonR = Value

// Value represents a JSON-R value.
type Value interface{ isValue() }

// Object stores an object of JSON-R.
type Object map[string]Value

func (Object) isValue() {} // Marks Object as a Value

// Array stores an array of JSON-R.
type Array []Value

func (Array) isValue() {} // Marks Array as a Value

// LeafValue is a type that can be used as a value of JSON-R leaf.
type LeafValue interface{ ~string | ~float64 | ~bool }

// Leaf[T] stores the value and timestamp of a JSON-R leaf.
type Leaf[T LeafValue] struct {
	Value     T     `json:"value"`     // number, string, boolean
	Timestamp int64 `json:"timestamp"` // Unix timestamp
}

func (Leaf[T]) isValue() {} // Marks Leaf as a Value

//////////////////////////////////
///////// OPERATIONS
//////////////////////////////////

// GetValueFromKey returns the value of the given key from the JSON-R object.
func GetValueFromKey(j JsonR, key string) (v Value, err error) {
	if obj, ok := j.(Object); !ok {
		return nil, fmt.Errorf("%v is not an jsonr.Object", j)
	} else {
		return obj[key], nil
	}
}

// Marshal converts the JSON-R to a byte slice.
func Marshal(j JsonR) (data []byte, err error) {
	return json.Marshal(j)
}

// MarshalIndent converts the JSON-R to a byte slice with indent.
func MarshalIndent(j JsonR, prefix, indent string) (data []byte, err error) {
	return json.MarshalIndent(j, prefix, indent)
}

// Equal checks if two JSON-Rs are equal, including timestamps.
func Equal(j1, j2 JsonR) (ret bool, err error) {
	var (
		o1, o2 any
		b1, _  = Marshal(j1)
		b2, _  = Marshal(j2)
	)

	// Check if the given JSON-Rs are valid
	if err = json.Unmarshal(b1, &o1); err != nil {
		return false, err
	}
	if err = json.Unmarshal(b2, &o2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(o1, o2), nil
}

// EqualWithoutTimestamp checks if two JSON-Rs are equal, excluding timestamps.
func EqualWithoutTimestamp(j1, j2 JsonR) (ret bool, err error) {
	var (
		o1, o2 any
		b1, _  = Marshal(j1)
		b2, _  = Marshal(j2)
	)

	// Check if the given JSON-Rs are valid
	if err = json.Unmarshal(b1, &o1); err != nil {
		return false, err
	}
	if err = json.Unmarshal(b2, &o2); err != nil {
		return false, err
	}

	// Remove timestamps
	var removeTimestamp func(o any)
	removeTimestamp = func(o any) {
		switch v := o.(type) {
		case map[string]any:
			delete(v, "timestamp")
			for _, value := range v {
				removeTimestamp(value)
			}
		case []any:
			for _, value := range v {
				removeTimestamp(value)
			}
		}
	}
	removeTimestamp(o1)
	removeTimestamp(o2)

	return reflect.DeepEqual(o1, o2), nil
}

// GetLatestTimestamp returns the latest timestamp of the given JSON-R.
func GetLatestTimestamp(j JsonR) int64 {
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
		panic(fmt.Sprintf("GetTimestamp: invalid type %T for JSON-R", j))
	}
}

//////////////////////////////////
///////// PARSER
//////////////////////////////////

// NewJsonR creates a new JSON-R from the given JSON-R string.
func NewJsonR(jsonr string) (JsonR, error) {
	var j JsonR
	err := Unmarshal([]byte(jsonr), &j)
	return j, err
}

// Unmarshal parses the JSON-R data and stores the result.
func Unmarshal(data []byte, j *JsonR) (err error) {
	var raw any
	if err = json.Unmarshal(data, &raw); err != nil {
		return
	}

	*j, err = parseValue(raw)
	return err
}

// String converts JSON-R to a string.
func String(j JsonR) string {
	data, err := json.MarshalIndent(j, "", "    ")
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON-R: %v", err))
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

// Any returns the given JSON-R as any(interface{}) type.
func Any(j JsonR) any {
	return j
}

//////////////////////////////////
///////// PRIVATE
//////////////////////////////////

// parseValue parses the given value and returns a JSON-R value.
func parseValue(raw any) (Value, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case map[string]any:
		if isLeaf(v) { // type is a JSON-R leaf
			return parseLeaf(v)
		} else { // type is a JSON-R object
			return parseObject(v)
		}
	case []any: // type is a JSON-R array
		return parseArray(v)
	default:
		return nil, fmt.Errorf("invalid value type: %v (type: %T)", v, v)
	}
}

// parseObject parses the given map as a JSON-R object.
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

// parseArray parses the given array as a JSON-R array.
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

// parseLeaf parses the given map as a JSON-R leaf.
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

// isLeaf checks if the given map is a JSON-R leaf.
func isLeaf(m map[string]any) bool {
	// A JSON-R Object is a Leaf when
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
