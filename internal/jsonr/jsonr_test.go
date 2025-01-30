//
// jsonr_test.go
//
// Tests for the jsonr package.
//
// Author: Karu (@karu-rress)
//

package jsonr

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

const exp = "../../examples/example.jsonr"
const exp2 = "../../examples/example2.jsonr"

// Testing JSON-R unmarshalling
func TestUnmarshal(t *testing.T) {
	// Test cases are in 'testcases.go' file.
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got JsonR
			err := Unmarshal([]byte(tc.jsonR), &got)
			if ret := err != nil; ret != tc.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Parse() = %v, want %v", got, tc.want)
			}
		})
	}
}

// Checking parsed JSON-R data
func TestParsedData(t *testing.T) {
	var (
		parsedJsonR    JsonR
		stringJsonR, _ = os.ReadFile(exp)
		err            = Unmarshal(stringJsonR, &parsedJsonR)
	)

	if err != nil { // If error occurs
		t.Errorf("Parse() error = %v", err)
		return
	}

	// Check the type of the parsed JSON-R data (should be 'jsonr.Object')
	jsonRType := reflect.TypeOf(parsedJsonR).String()
	assert.Equal(t, jsonRType, "jsonr.Object")
	t.Log("Parsed JSON-R type:", jsonRType)

	// Checking the leaf
	id, _ := GetValueFromKey(parsedJsonR, "vehicleId")
	assert.Equal(t, reflect.TypeOf(id).String(), "jsonr.Leaf[string]")
	t.Log("Vehicle ID:", id.(Leaf[string]).Value)

	// Checking the nested object
	location, _ := GetValueFromKey(parsedJsonR, "location")
	assert.Equal(t, reflect.TypeOf(location).String(), "jsonr.Object")
	lat, _ := GetValueFromKey(location, "latitude")
	lon, _ := GetValueFromKey(location, "longitude")
	t.Log("Location:", lat, lon)

	// Checking the nested array
	tires, _ := GetValueFromKey(parsedJsonR, "tirePressure")
	assert.Equal(t, reflect.TypeOf(tires).String(), "jsonr.Array")
	tarr, _ := ToArray(tires.(Array))
	t.Log("Tires:", tarr)
}

func TestGetTimestamp(t *testing.T) {
	var (
		parsedJsonR    JsonR
		stringJsonR, _ = os.ReadFile(exp2)
		err            = Unmarshal(stringJsonR, &parsedJsonR)
	)

	if err != nil {
		t.Errorf("Parse() error = %v", err)
		return
	}

	timestamp := GetLatestTimestamp(parsedJsonR)
	assert.Equal(t, timestamp, int64(2000000000))
	t.Logf("Max timestamp: %d", timestamp)
}
