//
// tjson_test.go
//
// Tests for the tjson package.
//
// Author: Karu (@karu-rress)
//

package tjson

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	ex1 = "../../examples/example1.tjson"
	ex2 = "../../examples/example2.tjson"
	js1 = "../../examples/example1.json"
)

// Testing T-JSON unmarshalling
func TestUnmarshal(t *testing.T) {
	// Test cases are in 'testcases.go' file.
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got TJson
			err := Unmarshal([]byte(tc.tJson), &got)
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

// Checking parsed T-JSON data
func TestParsedData(t *testing.T) {
	var (
		parsedTJson    TJson
		stringTJson, _ = os.ReadFile(ex1)
		err            = Unmarshal(stringTJson, &parsedTJson)
	)

	if err != nil { // If error occurs
		t.Errorf("Parse() error = %v", err)
		return
	}

	// Check the type of the parsed T-JSON data (should be 'tjson.Object')
	tJsonType := reflect.TypeOf(parsedTJson).String()
	assert.Equal(t, tJsonType, "tjson.Object")
	t.Log("Parsed T-JSON type:", tJsonType)

	// Checking the leaf
	id, _ := GetValue(parsedTJson, "/vehicleId")
	assert.Equal(t, reflect.TypeOf(id).String(), "tjson.Leaf[string]")
	t.Log("Vehicle ID:", id.(Leaf[string]).Value)

	// Checking the nested object
	lat, _ := GetValue(parsedTJson, "/location/latitude")
	lon, _ := GetValue(parsedTJson, "/location/longitude")
	t.Log("Location:", lat, lon)

	// Checking the nested array
	tires, _ := GetValue(parsedTJson, "/tirePressure")
	assert.Equal(t, reflect.TypeOf(tires).String(), "tjson.Array")
	tarr, _ := ToArray(tires.(Array))
	t.Log("Tires:", tarr)
}

func TestGetTimestamp(t *testing.T) {
	var (
		parsedTJson    TJson
		stringTJson, _ = os.ReadFile(ex2)
		err            = Unmarshal(stringTJson, &parsedTJson)
	)

	if err != nil {
		t.Errorf("Parse() error = %v", err)
		return
	}

	timestamp := GetLatestTimestamp(parsedTJson)
	assert.Equal(t, timestamp, int64(2000000000))
	t.Logf("Max timestamp: %d", timestamp)
}

func TestGetValue(t *testing.T) {
	var (
		parsedTJson    TJson
		stringTJson, _ = os.ReadFile(ex1)
		err            = Unmarshal(stringTJson, &parsedTJson)
	)

	assert.Nil(t, err)

	// Use path to retrieve the value
	value, err := GetValue(parsedTJson, "/tirePressure/0")
	t.Log("Value:", value)
	assert.Nil(t, err)
}

func TestToTJson(t *testing.T) {
	var j any

	strJson, _ := os.ReadFile(js1)
	err := json.Unmarshal(strJson, &j)
	assert.Nil(t, err)

	tJson, err := ToTJson(j)
	assert.Nil(t, err)
	t.Log(String(tJson))
}
