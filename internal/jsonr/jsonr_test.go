/**
 * jsonr_test.go
 *
 * Testing codes for jsonr package
*/

package jsonr

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name    string
		jsonR   string
		want    Value
		wantErr bool
	}{
		{
			name: "Valid JSON-R",
			jsonR: `
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
			}`,
			want: Object{
				"name":       Leaf[string]{Value: "John Doe", Timestamp: 1678886400},
				"age":        Leaf[float64]{Value: 30, Timestamp: 1678886400},
				"is-married": Leaf[bool]{Value: true, Timestamp: 1678886400},
				"address": Object{
					"street": Leaf[string]{Value: "123 Main St", Timestamp: 1678886400},
					"city":   Leaf[string]{Value: "Anytown", Timestamp: 1678886400},
				},
				"hobbies": Array{
					Leaf[string]{Value: "reading", Timestamp: 1678886400},
					Leaf[string]{Value: "hiking", Timestamp: 1678886400},
				},
			},
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			jsonR:   `{"name": "John Doe",}`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Missing Value Key",
			jsonR:   `{"name": {"Timestamp": 1678886400}}`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Missing Timestamp Key",
			jsonR:   `{"name": {"Value": "John Doe"}}`,
			want:    nil,
			wantErr: true,
		},
		// 추가적인 테스트 케이스 추가 가능
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse([]byte(tc.jsonR))
			if (err != nil) != tc.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Parse() = %v, want %v", got, tc.want)
			}
		})
	}
}
