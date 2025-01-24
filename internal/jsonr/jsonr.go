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
    "Value": "John Doe",
    "Timestamp": 1678886400
  },
  "age": {
    "Value": 30,
    "Timestamp": 1678886400
  },
  "isMarried": {
    "Value": true,
    "Timestamp": 1678886400
  },
  "address": {
    "street": {
      "Value": "123 Main St",
      "Timestamp": 1678886400
    },
    "city": {
      "Value": "Anytown",
      "Timestamp": 1678886400
    }
  },
  "hobbies": [
    {
      "Value": "reading",
      "Timestamp": 1678886400
    },
    {
      "Value": "hiking",
      "Timestamp": 1678886400
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
	Value     T     `json:"Value"`     // number, string, boolean
	Timestamp int64 `json:"Timestamp"` // Unix timestamp
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
}

// Array stores an array of JSON-R.
type Array []Value

// isValue is a method that marks Array as a Value.
func (Array) isValue() {
}

// JSON-R type
type JsonR Value

// GetValueFromKey returns the value of the given key from the JSON-R object.
func GetValueFromKey(j JsonR, key string) (Value, error) {
	obj, ok := j.(Object)
	if !ok {
		return nil, fmt.Errorf("%v is not an jsonr.Object", j)
	}
	return obj[key], nil
}

// Marshal converts the JSON-R to a byte slice.
func Marshal(j JsonR) ([]byte, error) {
	return json.Marshal(j)
}

// Equal checks if two JSON-Rs are equal.
func Equal(j1, j2 JsonR) (bool, error) {
	var o1, o2 any
	b1, _ := Marshal(j1)
	b2, _ := Marshal(j2)

	if err := json.Unmarshal(b1, &o1); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b2, &o2); err != nil {
		return false, err
	}

	return reflect.DeepEqual(o1, o2), nil
}

// GetLatestTimestamp returns the latest timestamp of the given JSON-R.
func GetLatestTimestamp(j JsonR) int64 {
	switch v := j.(type) {
	case Leaf[string]:
		return v.Timestamp
	case Leaf[float64]:
		return v.Timestamp
	case Leaf[bool]:
		return v.Timestamp
	case Object:
		var max int64
		for _, value := range v {
			timestamp := GetLatestTimestamp(value)
			if timestamp > max {
				max = timestamp
			}
		}
		return max
	case Array:
		var max int64
		for _, value := range v {
			timestamp := GetLatestTimestamp(value)
			if timestamp > max {
				max = timestamp
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
	return Parse([]byte(jsonr))
}

// Parse parses the JSON-R data and returns the result.
func Parse(data []byte) (JsonR, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	return parseValue(raw)
}

// Converts JSON-R to a string.
func ToString(j JsonR) string {
	data, err := json.Marshal(j)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON-R: %v", err))
	}
	return string(data)
}

// TODO: formatJSON-R

func parseValue(raw any) (JsonR, error) {
	switch v := raw.(type) {
	case map[string]any: // type is a JSON-R object
		if isLeaf(v) {
			return parseLeaf(v)
		} else {
			return parseObject(v)
		}
	case []any: // type is a JSON-R array
		return parseArray(v)
	// TODO: make case string, float64, bool and raise panic when default
	default: // type is a JSON-R leaf
		return parseLeaf(v)
	}
}

func parseObject(raw map[string]any) (Object, error) {
	obj := make(Object)
	for key, value := range raw {
		val, err := parseValue(value)
		if err != nil {
			return nil, err
		}
		obj[key] = val
	}
	return obj, nil
}

func parseArray(raw []any) (Array, error) {
	arr := make(Array, len(raw))
	for i, value := range raw {
		val, err := parseValue(value)
		if err != nil {
			return nil, err
		}
		arr[i] = val
	}
	return arr, nil
}

func parseLeaf(raw any) (Value, error) {
	// Check if raw is map[string]any
	rawMap, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid leaf: %v", raw)
	}

	// Check if Value and Timestamp keys exist
	valueRaw, ok := rawMap["Value"]
	if !ok {
		return nil, fmt.Errorf("missing Value key: %v", raw)
	}
	timestampRaw, ok := rawMap["Timestamp"]
	if !ok {
		return nil, fmt.Errorf("missing Timestamp key: %v", raw)
	}

	// Check if Timestamp is int64
	timestampF, ok := timestampRaw.(float64)
	timestamp := int64(timestampF)
	if !ok {
		return nil, fmt.Errorf("invalid Timestamp value: %v", timestampRaw)
	}

	// Create Leaf with the type of Value
	switch value := valueRaw.(type) {
	case string:
		return Leaf[string]{Value: value, Timestamp: timestamp}, nil
	case float64:
		return Leaf[float64]{Value: value, Timestamp: timestamp}, nil
	case int64:
		return Leaf[float64]{Value: float64(value), Timestamp: timestamp}, nil
	case bool:
		return Leaf[bool]{Value: value, Timestamp: timestamp}, nil
	default:
		return nil, fmt.Errorf("invalid Value type: %v", value)
	}
}

func isLeaf(m map[string]any) bool {
	_, ok1 := m["Value"]
	_, ok2 := m["Timestamp"]
	return ok1 && ok2
}
