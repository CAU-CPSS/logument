/**
 *
 * JSON-R (jsonr.go)
 *
 * Defines JSON-R, which consist of Value and Timestamp.
 *
 *
 * Example of JSON-R:
{
  "name": {
    "value": "John Doe",
    "timestamp": 1678886400
  },
  "age": {
    "value": 30,
    "timestamp": 1678886400
  },
  "isMarried": {
    "value": true,
    "timestamp": 1678886400
  },
  "address": {
    "street": {
      "value": "123 Main St",
      "timestamp": 1678886400
    },
    "city": {
      "value": "Anytown",
      "timestamp": 1678886400
    }
  },
  "hobbies": [
    {
      "value": "reading",
      "timestamp": 1678886400
    },
    {
      "value": "hiking",
      "timestamp": 1678886400
    }
  ]
}
*/

package jsonr

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// LeafValue is a type that can be used as a value of JSON-R leaf.
type LeafValue interface {
	~string | ~float64 | ~bool
}

// Leaf[T] stores the value and timestamp of a JSON-R leaf.
type Leaf[T LeafValue] struct {
	Value     T     `json:"value"`     // number, string, boolean
	Timestamp int64 `json:"timestamp"` // Unix timestamp
}

// isValue is a method that marks Leaf[T] as a Value.
func (Leaf[T]) isValue() {
	// This method is intentionally left empty.
}

// Value represents a JSON-R value, including Object and Array.
type Value interface {
	isValue()
}

// Object stores an object of JSON-R.
type Object map[string]Value

// isValue is a method that marks Object as a Value.
func (Object) isValue() {
	// This method is intentionally left empty.
}

// Array stores an array of JSON-R.
type Array []Value

// isValue is a method that marks Array as a Value.
func (Array) isValue() {
	// This method is intentionally left empty.
}

// The JSON-R type
type JsonR Value

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

// Equal checks if two JSON-Rs are equal.
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

// Unmarshal parses the JSON-R data and returns the result.
func Unmarshal(data []byte, j *JsonR) (err error) {
	var raw any
	if err = json.Unmarshal(data, &raw); err != nil {
		return
	}

	*j, err = parseValue(raw)
	return err
}

// ToString converts JSON-R to a string.
func ToString(j JsonR) string {
	data, err := json.Marshal(j)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON-R: %v", err))
	}
	return string(data)
}

//////////////////////////////////
///////// PRIVATE
//////////////////////////////////

func parseValue(raw any) (JsonR, error) {
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
