package jsonpatch

import (
	"testing"

	"github.com/CAU-CPSS/logument/internal/jsonr"
	"github.com/stretchr/testify/assert"
)

const testJsonR1 = `
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
}`

const testJsonR2 = `
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
			"timestamp": 1888886400
		},
		{
			"value": "sleeping",
			"timestamp": 1888886400
		}
	]
}`

const patch = `[
	{ "op":"replace", "path":"/hobbies/1", "value":"sleeping", "timestamp":1888886400 }
]`

func TestGeneratePatch(t *testing.T) {
	var parsedJsonR1, parsedJsonR2 jsonr.JsonR
	jsonr.Unmarshal([]byte(testJsonR1), &parsedJsonR1)
	jsonr.Unmarshal([]byte(testJsonR2), &parsedJsonR2)

	patch, err := GeneratePatch(parsedJsonR1, parsedJsonR2)
	assert.Equal(t, nil, err)

	var p Patch = patch

	t.Log(p.String())
}

func TestApplyPatch(t *testing.T) {
	var parsedJsonR1, parsedJsonR2 jsonr.JsonR
	jsonr.Unmarshal([]byte(testJsonR1), &parsedJsonR1)
	jsonr.Unmarshal([]byte(testJsonR2), &parsedJsonR2)

	patch, err := GeneratePatch(parsedJsonR1, parsedJsonR2)
	assert.Equal(t, nil, err)

	var p Patch = patch

	// Apply patch
	_, err = ApplyPatch(parsedJsonR1, p)
	assert.Equal(t, nil, err)
}

func TestApplyPatchWithJson(t *testing.T) {
	var doc jsonr.JsonR
	var p any

	jsonr.Unmarshal([]byte(testJsonR1), &doc)
	p, err := ParsePatch([]byte(patch))
	assert.Equal(t, nil, err)

	newDoc, err := ApplyPatch(doc, p.(Patch))
	assert.Equal(t, nil, err)

	t.Log(jsonr.String(newDoc))
}
