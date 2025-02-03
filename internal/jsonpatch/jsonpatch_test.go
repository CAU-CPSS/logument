//
// jsonpatch_test.go
//
// Tests for the jsonpatch package.
//
// Author: Karu (@karu-rress)
//

package jsonpatch

import (
	"os"
	"testing"

	"github.com/CAU-CPSS/logument/internal/jsonr"
	"github.com/stretchr/testify/assert"
)

const patch = `[
	{ "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 2000000000 },
	{ "op": "replace", "path": "/tirePressure/0", "value": 35.1, "timestamp": 2000000000 },
	{ "op": "replace", "path": "/engineOn", "value": false, "timestamp": 2000000000 }
]`

const (
	ex1 = "../../examples/example1.jsonr"
	ex2 = "../../examples/example2.jsonr"
)

func TestGeneratePatch(t *testing.T) {
	var (
		parsed1, parsed2 jsonr.JsonR
		jsonr1, _        = os.ReadFile(ex1)
		jsonr2, _        = os.ReadFile(ex2)
	)

	jsonr.Unmarshal(jsonr1, &parsed1)
	jsonr.Unmarshal(jsonr2, &parsed2)

	patch, err := GeneratePatch(parsed1, parsed2)
	assert.Equal(t, nil, err)

	t.Log(patch.String())
}

func TestApplyPatch(t *testing.T) {
	var (
		parsed1, parsed2 jsonr.JsonR
		jsonr1, _        = os.ReadFile(ex1)
		jsonr2, _        = os.ReadFile(ex2)
		b1, b2           []byte
	)

	jsonr.Unmarshal(jsonr1, &parsed1)
	jsonr.Unmarshal(jsonr2, &parsed2)

	patch, err := GeneratePatch(parsed1, parsed2)
	assert.Equal(t, nil, err)

	// Apply patch
	j, err := ApplyPatch(parsed1, patch)
	assert.Equal(t, nil, err)

	// Check if the result is equal to the second JSON-R
	b1, err = jsonr.ToJson(j)
	b2, err = jsonr.ToJson(parsed2)

	t.Log(string(b1))
	t.Log("=====================================")
	t.Log(string(b2))

	ret, _ := jsonr.EqualWithoutTimestamp(parsed2, j)
	assert.Equal(t, true, ret)
}

func TestApplyPatchWithJson(t *testing.T) {
	var (
		doc       jsonr.JsonR
		jsonr1, _ = os.ReadFile(ex1)
		p         Patch
	)

	// Unmarshal the first JSON-R
	jsonr.Unmarshal(jsonr1, &doc)
	p, err := Unmarshal([]byte(patch))
	assert.Equal(t, nil, err)

	// Apply patch
	newDoc, err := ApplyPatch(doc, p)
	assert.Equal(t, nil, err)

	// Check if the result is equal to the second JSON-R
	jsonr2, _ := os.ReadFile(ex2)
	jsonr.Unmarshal(jsonr2, &doc)
	ret, _ := jsonr.EqualWithoutTimestamp(doc, newDoc)
	assert.Equal(t, true, ret)
}
