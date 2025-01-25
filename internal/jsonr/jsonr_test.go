/**
 * jsonr_test.go
 *
 * Testing codes for jsonr package
 */

package jsonr

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var stringJsonR = `{
	"name": {"value": "John Doe", "timestamp": 1678886400},
	"age": {"value": 30, "timestamp": 1678886400},
	"is-married": {"value": true, "timestamp": 1678886400},
	"address": {
		"street": {"value": "123 Main St", "timestamp": 1678886400},
		"city": {"value": "Anytown", "timestamp": 1678886400}
	},
	"hobbies": [
		{"value": "reading", "timestamp": 1678886400},
		{"value": "hiking", "timestamp": 2078886400}
	]
}`

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
					"value": "John Doe",
					"timestamp": 1678886400
				},
				"age": {
					"value": 30,
					"timestamp": 1678886400
				},
				"is-married": {
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
			jsonR:   `{"name": {"timestamp": 1678886400}}`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Missing Timestamp Key",
			jsonR:   `{"name": {"value": "John Doe"}}`,
			want:    nil,
			wantErr: true,
		},
		// 추가적인 테스트 케이스 추가 가능
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got JsonR
			err := Unmarshal([]byte(tc.jsonR), &got)
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

func TestParsedData(t *testing.T) {
	var parsedJsonR JsonR
	err := Unmarshal([]byte(stringJsonR), &parsedJsonR)
	if err != nil {
		t.Errorf("Parse() error = %v", err)
		return
	}

	t.Log(reflect.TypeOf(parsedJsonR))
	t.Log(parsedJsonR)

	value, _ := GetValueFromKey(parsedJsonR, "address")
	t.Log(reflect.TypeOf(value))
	t.Log(GetValueFromKey(parsedJsonR, "address"))

	value, _ = GetValueFromKey(parsedJsonR, "address")
	value, _ = GetValueFromKey(value, "city")
	assert.Equal(t, reflect.TypeOf(value).String(), "jsonr.Leaf[string]")

	value, _ = GetValueFromKey(parsedJsonR, "address")
	t.Log(GetValueFromKey(value, "city"))
}

func TestGetTimestamp(t *testing.T) {
	var parsedJsonR JsonR
	err := Unmarshal([]byte(stringJsonR), &parsedJsonR)
	if err != nil {
		t.Errorf("Parse() error = %v", err)
		return
	}

	timestamp := GetLatestTimestamp(parsedJsonR)
	t.Logf("Max timestamp: %d", timestamp)
}
