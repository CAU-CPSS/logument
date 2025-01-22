/**
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
)

// LeafValue is a type that can be used as a value of JSON-R leaf.
type LeafValue interface {
	~string | ~float64 | ~bool
}

// Leaf stores the value and timestamp of a JSON-R leaf.
type Leaf[T LeafValue] struct {
	Value     T     `json:"Value"`     // number, string, boolean
	Timestamp int64 `json:"Timestamp"` // Unix timestamp
}

// Value stores the value and timestamp of a JSON-R value.
type Value interface {
	isValue()
}

func (Leaf[T]) isValue() {
}

// Object stores an object of JSON-R.
type Object map[string]Value

func (Object) isValue() {
}

// Array stores an array of JSON-R.
type Array []Value

func (Array) isValue() {
}

// JSON-R type
type JsonR Value

//////////////////////////////////
///////// PARSER
//////////////////////////////////

// Parse parses the JSON-R data and returns the result.
func Parse(data []byte) (JsonR, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	return parseValue(raw)
}

func parseValue(raw any) (JsonR, error) {
	switch v := raw.(type) {
	case map[string]any: // type is Object
		if isLeaf(v) {
			return parseLeaf(v)
		} else {
			return parseObject(v)
		}
	case []any: // type is Array
		return parseArray(v)
	default: // type is Leaf
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
