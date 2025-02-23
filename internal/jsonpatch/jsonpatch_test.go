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

	"github.com/CAU-CPSS/logument/internal/tson"
	"github.com/stretchr/testify/assert"
)

const patch = `[
	{ "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 2000000000 },
	{ "op": "replace", "path": "/tirePressure/0", "value": 35.1, "timestamp": 2000000000 },
	{ "op": "replace", "path": "/engineOn", "value": false, "timestamp": 2000000000 }
]`

const (
	tj1 = "../../examples/example1.tson"
	tj2 = "../../examples/example2.tson"
)

func TestGeneratePatch(t *testing.T) {
	var (
		parsed1, parsed2 tson.Tson
		tson1, _         = os.ReadFile(tj1)
		tson2, _         = os.ReadFile(tj2)
	)

	tson.Unmarshal(tson1, &parsed1)
	tson.Unmarshal(tson2, &parsed2)

	patch, err := GeneratePatch(parsed1, parsed2)
	assert.Equal(t, nil, err)

	t.Log(patch.String())
}

func TestApplyPatch(t *testing.T) {
	var (
		parsed1, parsed2 tson.Tson
		tson1, _         = os.ReadFile(tj1)
		tson2, _         = os.ReadFile(tj2)
		b1, b2           []byte
	)

	tson.Unmarshal(tson1, &parsed1)
	tson.Unmarshal(tson2, &parsed2)

	patch, err := GeneratePatch(parsed1, parsed2)
	assert.Equal(t, nil, err)

	// Apply patch
	j, err := ApplyPatch(parsed1, patch)
	assert.Equal(t, nil, err)

	// Check if the result is equal to the second TSON
	b1, err = tson.ToJson(j)
	b2, err = tson.ToJson(parsed2)

	t.Log(string(b1))
	t.Log("=====================================")
	t.Log(string(b2))

	ret, _ := tson.EqualWithoutTimestamp(parsed2, j)
	assert.Equal(t, true, ret)
}

func TestApplyPatchWithJson(t *testing.T) {
	var (
		doc      tson.Tson
		tson1, _ = os.ReadFile(tj1)
		p        Patch
	)

	// Unmarshal the first TSON
	tson.Unmarshal(tson1, &doc)
	p, err := Unmarshal([]byte(patch))
	assert.Equal(t, nil, err)

	// Apply patch
	newDoc, err := ApplyPatch(doc, p)
	assert.Equal(t, nil, err)

	// Check if the result is equal to the second TSON
	tson2, _ := os.ReadFile(tj2)
	tson.Unmarshal(tson2, &doc)
	ret, _ := tson.EqualWithoutTimestamp(doc, newDoc)
	assert.Equal(t, true, ret)
}
