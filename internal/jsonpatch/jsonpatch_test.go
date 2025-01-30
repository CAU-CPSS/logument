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

const exp = "../../examples/example.jsonr"
const exp2 = "../../examples/example2.jsonr"

func TestGeneratePatch(t *testing.T) {
	var (
		parsed1, parsed2 jsonr.JsonR
		jsonr1, _        = os.ReadFile(exp)
		jsonr2, _        = os.ReadFile(exp2)
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
		jsonr1, _        = os.ReadFile(exp)
		jsonr2, _        = os.ReadFile(exp2)
	)

	jsonr.Unmarshal(jsonr1, &parsed1)
	jsonr.Unmarshal(jsonr2, &parsed2)

	patch, err := GeneratePatch(parsed1, parsed2)
	assert.Equal(t, nil, err)

	// Apply patch
	j, err := ApplyPatch(parsed1, patch)
	assert.Equal(t, nil, err)

	// Check if the result is equal to the second JSON-R
	ret, _ := jsonr.EqualWithoutTimestamp(parsed2, j)
	assert.Equal(t, true, ret)
}

func TestApplyPatchWithJson(t *testing.T) {
	var (
		doc           jsonr.JsonR
		jsonr1, _ = os.ReadFile(exp)
		p             Patch
	)

	// Unmarshal the first JSON-R
	jsonr.Unmarshal(jsonr1, &doc)
	p, err := Unmarshal([]byte(patch))
	assert.Equal(t, nil, err)

	// Apply patch
	newDoc, err := ApplyPatch(doc, p)
	assert.Equal(t, nil, err)

	// Check if the result is equal to the second JSON-R
	jsonr2, _ := os.ReadFile(exp2)
	jsonr.Unmarshal(jsonr2, &doc)
	ret, _ := jsonr.EqualWithoutTimestamp(doc, newDoc)
	assert.Equal(t, true, ret)
}
