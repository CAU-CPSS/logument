package jsonpatch

import (
	"testing"

	"github.com/CAU-CPSS/logument/internal/jsonr"
	"github.com/stretchr/testify/assert"
)

var testJsonR1 = `
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

var testJsonR2 = `
{
	"name": {
		"Value": "John Doe",
		"Timestamp": 1678886400
	},
	"age": {
		"Value": 10,
		"Timestamp": 1678886400
	},
	"is-married": {
		"Value": true,
		"Timestamp": 1678886400
	},
	"hobbies": [
		{
			"Value": "reading",
			"Timestamp": 1678886400
		},
		{
			"Value": "hiking",
			"Timestamp": 1678886400
		},
		{
			"Value": "swimming",
			"Timestamp": 2000000000
		}
	]
}`

var parsedJsonR1, _ = jsonr.Parse([]byte(testJsonR1))
var parsedJsonR2, _ = jsonr.Parse([]byte(testJsonR2))

func TestPrintJsonR(t *testing.T) {
	t.Log(jsonr.ToString(parsedJsonR1))
	t.Log(jsonr.ToString(parsedJsonR2))
}

/*
	patch, e := jsonpatch.CreatePatch([]byte(simpleA), []byte(simpleA))
	if e != nil {
		fmt.Printf("Error creating JSON patch:%v", e)
		return
	}
	for _, operation := range patch {
		fmt.Printf("%s\n", operation.Json())
	}
*/

func TestCreatePatch(t *testing.T) {
	patch, err := GeneratePatch(parsedJsonR1, parsedJsonR2)
	assert.Equal(t, nil, err)

	for _, operation := range patch {
		t.Logf("%s\n", operation.ToString())
	}
}
