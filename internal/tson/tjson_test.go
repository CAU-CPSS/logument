//
// tson_test.go
//
// Tests for the tson package.
//
// Author: Karu (@karu-rress)
//

package tson

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	ex1 = "../../examples/example1.tson"
	ex2 = "../../examples/example2.tson"
	js1 = "../../examples/example1.json"
)

// Testing TSON unmarshalling
func TestUnmarshal(t *testing.T) {
	// Test cases are in 'testcases.go' file.
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got Tson
			err := Unmarshal([]byte(tc.tson), &got)
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

// Checking parsed TSON data
func TestParsedData(t *testing.T) {
	var (
		parsedTson    Tson
		stringTson, _ = os.ReadFile(ex1)
		err            = Unmarshal(stringTson, &parsedTson)
	)

	if err != nil { // If error occurs
		t.Errorf("Parse() error = %v", err)
		return
	}

	// Check the type of the parsed TSON data (should be 'tson.Object')
	tsonType := reflect.TypeOf(parsedTson).String()
	assert.Equal(t, tsonType, "tson.Object")
	t.Log("Parsed TSON type:", tsonType)

	// Checking the leaf
	id, _ := GetValue(parsedTson, "/vehicleId")
	assert.Equal(t, reflect.TypeOf(id).String(), "tson.Leaf[string]")
	t.Log("Vehicle ID:", id.(Leaf[string]).Value)

	// Checking the nested object
	lat, _ := GetValue(parsedTson, "/location/latitude")
	lon, _ := GetValue(parsedTson, "/location/longitude")
	t.Log("Location:", lat, lon)

	// Checking the nested array
	tires, _ := GetValue(parsedTson, "/tirePressure")
	assert.Equal(t, reflect.TypeOf(tires).String(), "tson.Array")
	tarr, _ := ToArray(tires.(Array))
	t.Log("Tires:", tarr)
}

func TestGetTimestamp(t *testing.T) {
	var (
		parsedTson    Tson
		stringTson, _ = os.ReadFile(ex2)
		err            = Unmarshal(stringTson, &parsedTson)
	)

	if err != nil {
		t.Errorf("Parse() error = %v", err)
		return
	}

	timestamp := GetLatestTimestamp(parsedTson)
	assert.Equal(t, timestamp, int64(2000000000))
	t.Logf("Max timestamp: %d", timestamp)
}

func TestGetValue(t *testing.T) {
	var (
		parsedTson    Tson
		stringTson, _ = os.ReadFile(ex1)
		err            = Unmarshal(stringTson, &parsedTson)
	)

	assert.Nil(t, err)

	// Use path to retrieve the value
	value, err := GetValue(parsedTson, "/tirePressure/0")
	t.Log("Value:", value)
	assert.Nil(t, err)
}

func TestToTson(t *testing.T) {
	var j any

	strJson, _ := os.ReadFile(js1)
	err := json.Unmarshal(strJson, &j)
	assert.Nil(t, err)

	tson, err := ToTson(j)
	assert.Nil(t, err)
	t.Log(String(tson))
}
