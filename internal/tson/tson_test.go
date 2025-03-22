//
// tson_test.go
//
// Unit tests for the tson package.
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

//////////////////////////////////
///////// Test cases
//////////////////////////////////

var testCases = []struct {
	name    string
	tson    string
	want    map[string]any
	wantErr bool
}{
	{
		name: "Valid TSON",
		tson: `
		{
			"name" <1678886400>: "John Doe",
			"age" <1678886400>: 30,
			"is-married" <1678886400>: true,
			"address": {
				"street" <1678886400>: "123 Main St",
				"city" <1678886400>: "Anytown"
			},
			"hobbies": [
				<1678886400> "reading",
				<1678886400> "hiking"
			]
		}`,
		want: map[string]any{
			"name":       Leaf[string]{Value: "John Doe", Timestamp: 1678886400},
			"age":        Leaf[float64]{Value: 30, Timestamp: 1678886400},
			"is-married": Leaf[bool]{Value: true, Timestamp: 1678886400},
			"address": map[string]any{
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
		tson:    `{ "name": "John Doe", }`,
		want:    nil,
		wantErr: true,
	},
	{
		name:    "Missing Value",
		tson:    `{"name" <1678886400>: }`,
		want:    nil,
		wantErr: true,
	},
	{
		name:    "Ommitted Timestamp",
		tson:    `{"name" <>: "John Doe"}`,
		want:    map[string]any{"name": Leaf[string]{Value: "John Doe", Timestamp: -1}},
		wantErr: false,
	},
}

const (
	tson1 = "../../examples/example1.tson"
	tson2 = "../../examples/example2.tson"
	json1 = "../../examples/example1.json"
)

//////////////////////////////////
///////// Test functions
//////////////////////////////////

// Testing TSON unmarshalling
// **NOTE: Not working after changing the Object to TreeMap
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
		stringTson, _ = os.ReadFile(tson1)
		err           = Unmarshal(stringTson, &parsedTson)
	)

	if err != nil { // If error occurs
		t.Errorf("Parse() error = %v", err)
		return
	}

	// Check the type of the parsed TSON data ('tson.Object')
	tsonType := reflect.TypeOf(parsedTson).String()
	assert.Equal(t, tsonType, "tson.Object")

	// Checking the leaf ("vehicleId": "ABC1234")
	id, _ := GetValue(parsedTson, "/vehicleId")
	assert.Equal(t, reflect.TypeOf(id).String(), "tson.Leaf[string]")
	assert.Equal(t, id.(Leaf[string]).Value, "ABC1234")

	// Checking the nested object (37.7749, -122.4194)
	lat, _ := GetValue(parsedTson, "/location/latitude")
	lon, _ := GetValue(parsedTson, "/location/longitude")
	assert.Equal(t, lat.(Leaf[float64]).Value, 37.7749)
	assert.Equal(t, lon.(Leaf[float64]).Value, -122.4194)

	// Checking the nested array ([32.1 31.8 32 31.9])
	tires, _ := GetValue(parsedTson, "/tirePressure")
	assert.Equal(t, reflect.TypeOf(tires).String(), "tson.Array")
	tarr, _ := ToArray(tires.(Array))
	assert.Equal(t, tarr, []any{32.1, 31.8, 32.0, 31.9})
}

func TestGetTimestamp(t *testing.T) {
	var (
		parsedTson    Tson
		stringTson, _ = os.ReadFile(tson2)
		err           = Unmarshal(stringTson, &parsedTson)
	)
	assert.Nil(t, err)

	// Check the timestamp of the leaf node
	timestamp := GetLatestTimestamp(parsedTson)
	assert.Equal(t, timestamp, int64(2000000000))
}

func TestGetValue(t *testing.T) {
	var (
		parsedTson    Tson
		stringTson, _ = os.ReadFile(tson1)
		err           = Unmarshal(stringTson, &parsedTson)
	)
	assert.Nil(t, err)

	// Use path to retrieve the value ({32.1 -1})
	value, _ := GetValue(parsedTson, "/tirePressure/0")
	assert.Equal(t, value.(Leaf[float64]).Value, 32.1)
}

func TestToTson(t *testing.T) {
	var (
		j          any
		tson       Tson
		strJson, _ = os.ReadFile(json1)
		err        = json.Unmarshal(strJson, &j)
	)
	assert.Nil(t, err)

	err = FromJson(j, &tson)
	assert.Nil(t, err)
	t.Log(ToCompatibleTsonString(tson))
}

func TestMarshalIndent(t *testing.T) {
	var (
		parsedTson    Tson
		stringTson, _ = os.ReadFile(tson1)
		err           = Unmarshal(stringTson, &parsedTson)
	)
	assert.Nil(t, err)

	b, _ := MarshalIndent(parsedTson, "", "  ")
	t.Log(string(b))
}
