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