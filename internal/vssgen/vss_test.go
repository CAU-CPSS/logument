//
// vss_test.go
//
// Test cases for vss.go
//
//

package vssgen

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"encoding/json"

	"github.com/appscode/jsonpatch"
)

const (
	file = "./vss_rel_4.2.json"
)

var v = &VssJson{
	initialized: false,
	data: map[string]interface{}{
		"a": "A",
		"b": "B",
		"c": map[string]interface{}{
			"c1": "C1",
			"c2": "C2",
			"c3": map[string]interface{}{
				"c3_1": "C3_1",
				"c3_2": "C3_2",
			},
			"c4": []string{"C4_1", "C4_2"},
			"d":  "D",
		},
	},
}

var v_next = &VssJson{
	initialized: false,
	data: map[string]interface{}{
		"a": "A`",
		"b": "B`",
		"c": map[string]interface{}{
			"c1": "C1",
			"c2": "C2",
			"c3": map[string]interface{}{
				"c3_1": "C3_1",
				"c3_2": "C3_2",
			},
			"c4": []string{"C4_1", "C4_2"},
			"d":  "D`",
		},
	},
}

func TestNewVssJson(t *testing.T) {
	v := NewVssJson(file)

	if v == nil {
		t.Errorf("NewVssJson(%s) returned nil", file)
	}
}

func TestRemoveKeys(t *testing.T) {
	v.removeKeys("a", "c1", "c3_1")

	if _, ok := v.data.(map[string]interface{})["a"]; ok {
		t.Errorf("removeKeys failed to remove key 'a'")
	}
	if _, ok := v.data.(map[string]interface{})["c"].(map[string]interface{})["c1"]; ok {
		t.Errorf("removeKeys failed to remove key 'c1'")
	}
	if _, ok := v.data.(map[string]interface{})["c"].(map[string]interface{})["c3"].(map[string]interface{})["c3_1"]; ok {
		t.Errorf("removeKeys failed to remove key 'c3_1'")
	}
}

func TestPrint(t *testing.T) {
	stdout := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w

	v.Print()

	w.Close()
	os.Stdout = stdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	captured := buf.String()

	expected := `{
    "a": "A",
    "b": "B",
    "c": {
        "c1": "C1",
        "c2": "C2",
        "c3": {
            "c3_1": "C3_1",
            "c3_2": "C3_2"
        },
        "c4": [
            "C4_1",
            "C4_2"
        ],
        "d": "D"
    }
}
`

	if captured != expected {
		t.Errorf("Expected: \n%s, got \n%s", expected, captured)
	}
}

func TestPrintWithIndex(t *testing.T) {
	stdout := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w

	v.PrintWithIndex()

	w.Close()
	os.Stdout = stdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	captured := buf.String()

	expected := `00001: {
00002:     "a": "A",
00003:     "b": "B",
00004:     "c": {
00005:         "c1": "C1",
00006:         "c2": "C2",
00007:         "c3": {
00008:             "c3_1": "C3_1",
00009:             "c3_2": "C3_2"
00010:         },
00011:         "c4": [
00012:             "C4_1",
00013:             "C4_2"
00014:         ],
00015:         "d": "D"
00016:     }
00017: }
`

	if captured != expected {
		t.Errorf("Expected: \n%s, got \n%s", expected, captured)
	}
}

func TestLeafNodes(t *testing.T) {
	result := v.LeafNodes()

	if len(result) != 8 {
		t.Errorf("Expected 8 leaf nodes, got %d", len(result))
	}
}

func TestGenerate(t *testing.T) {
	_v := NewVssJson(file)
	v := _v.Generate(1.0, 1)
	v.Print()
}

func TestGenerateNext(t *testing.T) {
	_v := NewVssJson(file)
	v := _v.Generate(1.0, 1)

	_, patch := v.GenerateNext(0.5, 1, 2)
	patch.Print()
}

func TestSave(t *testing.T) {
	_v := NewVssJson(file)
	v1 := _v.Generate(1.0, 1)

	result, _ := v1.GenerateNext(0.5, 1, 2)
	result.Save("./test_patch.json")
}

func TestJsonPatch(t *testing.T) {
	_v, _ := json.Marshal(v.data)
	_v_next, _ := json.Marshal(v_next.data)

	result, _ := jsonpatch.CreatePatch(_v, _v_next)
	t.Logf("Patch: %v\n", result)
}

func TestGenerateVss(t *testing.T) {
	metadata := map[string]any{
		"dataset":     "internal/vssgen/vss_rel_4.2.json",
		"cars":        100,
		"files":       300,
		"change_rate": 0.2,
		"size":        1.0,
	}
	outputDir := "../dataset"

	PrepareOutputDir(outputDir)
	SaveMetadata(metadata, outputDir)
	Generate(metadata, outputDir)

	fmt.Printf("Saved to %s! Exiting...\n", outputDir)
}
