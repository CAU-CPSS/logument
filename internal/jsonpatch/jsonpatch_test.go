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

	"github.com/CAU-CPSS/logument/internal/tjson"
	"github.com/stretchr/testify/assert"
)

const patch = `[
	{ "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 2000000000 },
	{ "op": "replace", "path": "/tirePressure/0", "value": 35.1, "timestamp": 2000000000 },
	{ "op": "replace", "path": "/engineOn", "value": false, "timestamp": 2000000000 }
]`

const (
	tj1 = "../../examples/example1.tjson"
	tj2 = "../../examples/example2.tjson"
)

func TestGeneratePatch(t *testing.T) {
	var (
		parsed1, parsed2 tjson.TJson
		tjson1, _        = os.ReadFile(tj1)
		tjson2, _        = os.ReadFile(tj2)
	)

	tjson.Unmarshal(tjson1, &parsed1)
	tjson.Unmarshal(tjson2, &parsed2)

	patch, err := GeneratePatch(parsed1, parsed2)
	assert.Equal(t, nil, err)

	t.Log(patch.String())
}

func TestApplyPatch(t *testing.T) {
	var (
		parsed1, parsed2 tjson.TJson
		tjson1, _        = os.ReadFile(tj1)
		tjson2, _        = os.ReadFile(tj2)
		b1, b2           []byte
	)

	tjson.Unmarshal(tjson1, &parsed1)
	tjson.Unmarshal(tjson2, &parsed2)

	patch, err := GeneratePatch(parsed1, parsed2)
	assert.Equal(t, nil, err)

	// Apply patch
	j, err := ApplyPatch(parsed1, patch)
	assert.Equal(t, nil, err)

	// Check if the result is equal to the second T-JSON
	b1, err = tjson.ToJson(j)
	b2, err = tjson.ToJson(parsed2)

	t.Log(string(b1))
	t.Log("=====================================")
	t.Log(string(b2))

	ret, _ := tjson.EqualWithoutTimestamp(parsed2, j)
	assert.Equal(t, true, ret)
}

func TestApplyPatchWithJson(t *testing.T) {
	var (
		doc       tjson.TJson
		tjson1, _ = os.ReadFile(tj1)
		p         Patch
	)

	// Unmarshal the first T-JSON
	tjson.Unmarshal(tjson1, &doc)
	p, err := Unmarshal([]byte(patch))
	assert.Equal(t, nil, err)

	// Apply patch
	newDoc, err := ApplyPatch(doc, p)
	assert.Equal(t, nil, err)

	// Check if the result is equal to the second T-JSON
	tjson2, _ := os.ReadFile(tj2)
	tjson.Unmarshal(tjson2, &doc)
	ret, _ := tjson.EqualWithoutTimestamp(doc, newDoc)
	assert.Equal(t, true, ret)
}
