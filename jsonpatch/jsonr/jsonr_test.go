package jsonr

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	jsonRString := `
	{
		"name": {
			"Value": "John Doe",
			"Timestamp": 1678886400
		},
		"age": {
			"Value": 30,
			"Timestamp": 1678886400
		},
		"is-married": {
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
	}`

	jsonR, err := Parse([]byte(jsonRString))
	if err != nil {
		t.Fatalf("Failed to parse JSON-R: %v", err)
	}

	// 파싱된 결과 확인 (간단한 예시)
	switch v := jsonR.(type) {
	case Object:
		if nameLeaf, ok := v["name"].(Leaf[string]); ok {
			fmt.Println("Name:", nameLeaf.Value)
			fmt.Println("Timestamp:", nameLeaf.Timestamp)
		} else {
			t.Errorf("Expected 'name' to be a Leaf[string]")
		}
	default:
		t.Errorf("Expected the root to be an Object")
	}
}
