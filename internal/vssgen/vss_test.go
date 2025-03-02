//
// vss_test.go
//
// Test cases for vss.go
//
// Author: Karu (@karu-rress)
//

package vssgen

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"encoding/json"

	"github.com/CAU-CPSS/logument/internal/tson"
	"github.com/appscode/jsonpatch"
)

const (
	file = "./vss.json"
)

var v = &VssJson{
	initialized: false,
	data: map[string]any{
		"a": "A",
		"b": "B",
		"c": map[string]any{
			"c1": "C1",
			"c2": "C2",
			"c3": map[string]any{
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
	data: map[string]any{
		"a": "A`",
		"b": "B`",
		"c": map[string]any{
			"c1": "C1",
			"c2": "C2",
			"c3": map[string]any{
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

	if _, ok := v.data.(map[string]any)["a"]; ok {
		t.Errorf("removeKeys failed to remove key 'a'")
	}
	if _, ok := v.data.(map[string]any)["c"].(map[string]any)["c1"]; ok {
		t.Errorf("removeKeys failed to remove key 'c1'")
	}
	if _, ok := v.data.(map[string]any)["c"].(map[string]any)["c3"].(map[string]any)["c3_1"]; ok {
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

	j, patch := v.GenerateNext(0.5, 1, 2)
	_ = j

	patch.Print()

}

func TestSave(t *testing.T) {
	_v := NewVssJson(file)
	v1 := _v.Generate(1.0, 1)

	result, _ := v1.GenerateNext(0.5, 1, 2)
	result.Save("./test.tson")
}

func TestJsonPatch(t *testing.T) {
	_v, _ := json.Marshal(v.data)
	_v_next, _ := json.Marshal(v_next.data)

	result, _ := jsonpatch.CreatePatch(_v, _v_next)
	t.Logf("Patch: %v\n", result)
}

func TestCase(t *testing.T) {
	// 예제 1: 일반 JSON 형식의 데이터 (map[string]any)
	jsonData := map[string]any{
		"vehicleId": map[string]any{
			"value":     "ABC1234",
			"timestamp": 1700000000,
		},
		"speed": map[string]any{
			"value":     72.5,
			"timestamp": 1700000000,
		},
		"engineOn": map[string]any{
			"value":     true,
			"timestamp": 1700000000,
		},
		"location": map[string]any{
			"latitude": map[string]any{
				"value":     37.7749,
				"timestamp": 1700000000,
			},
			"longitude": map[string]any{
				"value":     -122.4194,
				"timestamp": 1700000000,
			},
		},
		"tirePressure": []any{
			map[string]any{
				"value":     32.1,
				"timestamp": 1700000000,
			},
			map[string]any{
				"value":     31.8,
				"timestamp": 1700000000,
			},
			map[string]any{
				"value":     32.0,
				"timestamp": 1700000000,
			},
			map[string]any{
				"value":     31.9,
				"timestamp": 1700000000,
			},
		},
	}

	// 예제 2: Tson 타입을 사용한 데이터
	tsonData := tson.Object{
		"vehicleId": tson.Leaf[string]{Value: "ABC1234", Timestamp: 1700000000},
		"speed":     tson.Leaf[float64]{Value: 72.5, Timestamp: 1700000000},
		"engineOn":  tson.Leaf[bool]{Value: true, Timestamp: 1700000000},
		"location": tson.Object{
			"latitude":  tson.Leaf[float64]{Value: 37.7749, Timestamp: 1700000000},
			"longitude": tson.Leaf[float64]{Value: -122.4194, Timestamp: 1700000000},
		},
		"tirePressure": tson.Array{
			tson.Leaf[float64]{Value: 32.1, Timestamp: 1700000000},
			tson.Leaf[float64]{Value: 31.8, Timestamp: 1700000000},
			tson.Leaf[float64]{Value: 32.0, Timestamp: 1700000000},
			tson.Leaf[float64]{Value: 31.9, Timestamp: 1700000000},
		},
	}

	fmt.Println("=== JSON 데이터 ===")
	data, err := tson.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		fmt.Println("Error during MarshalIndent:", err)
		return
	}
	fmt.Println(string(data))

	fmt.Println("\n=== Tson 데이터 ===")
	data2, err := tson.MarshalIndent(tsonData, "", "  ")
	if err != nil {
		fmt.Println("Error during MarshalIndent:", err)
		return
	}
	fmt.Println(string(data2))
}
