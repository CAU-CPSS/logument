//
// vss.go
//
// A JSON / JSON patch generator
// for VSS(Vehicle Signal Specification) JSON files.
//
//

// Package vssgen provides a JSON manager for VSS JSON files.
package vssgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"math/rand"

	"github.com/CAU-CPSS/logument/internal/jsonpatch"
	"github.com/CAU-CPSS/logument/internal/tson"
)

// If true, metadata will be saved in each JSON file
const SAVE_METADATA = false

const (
	tsonValue     = "value"
	tsonTimestamp = "timestamp"
)

// VSS JSON manager struct
type VssJson struct {
	initialized bool
	data        any
}

// NewVssJson creates a new VssJson
func NewVssJson(file string) *VssJson {
	vss := &VssJson{initialized: false}

	rawdata, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}
	json.Unmarshal(rawdata, &vss.data)
	vss.removeKeys("description", "uuid", "type", "comment", "deprecated", "deprecation")

	return vss
}

// Remove keys from JSON data recursively
func (vss *VssJson) removeKeys(keys ...string) {
	// If initialized, no need to remove keys
	if vss.initialized {
		return
	}

	// Recursive (DFS) function to remove keys from JSON data
	var remove func(data map[string]any, keysToRemove ...string)
	remove = func(data map[string]any, keysToRemove ...string) {
		for _, value := range data {
			if nested, ok := value.(map[string]any); ok {
				remove(nested, keysToRemove...)
			} else {
				for _, key := range keysToRemove {
					delete(data, key)
				}
			}
		}
	}

	remove(vss.data.(map[string]any), keys...)
}

// Print JSON data
func (vss VssJson) Print() {
	// TODO: format JSON patch differently...
	d, _ := tson.MarshalIndent(vss.data, "", "    ")
	fmt.Println(string(d))
}

// Print JSON data with line numbers
func (vss VssJson) PrintWithIndex() {
	d, _ := json.MarshalIndent(vss.data, "", "    ")

	for i, line := range strings.Split(string(d), "\n") {
		fmt.Printf("%05d: %s\n", i+1, line)
	}
}

// Get leaf nodes of the JSON data.
// If JSON patch is passed, nil is returned
func (vss VssJson) LeafNodes() []map[string]any {
	var leaf func(d any, parent string) []map[string]any
	leaf = func(d any, parent string) (r []map[string]any) {
		if d == nil {
			d = vss.data
		}

		_, isParent := d.(map[string]any)
		isTsonLeaf := false

		if isParent {
			_, isTsonLeaf = d.(map[string]any)[tsonTimestamp]
		}

		if !isTsonLeaf {
			for key, value := range d.(map[string]any) {
				fullKey := key
				if parent != "" && key != "children" {
					fullKey = parent + "." + key
				} else if parent != "" {
					fullKey = parent
				}

				if _, ok := value.(map[string]any); !ok {
					r = append(r, map[string]any{fullKey: value})
				} else {
					r = append(r, leaf(value, fullKey)...)
				}
			}
		} else {
			r = append(r, map[string]any{parent: d})
		}

		return
	}

	if _, ok := vss.data.(map[string]any); ok {
		return leaf(nil, "")
	}
	return nil
}

// Generate an initial random dataset based on the JSON schema
func (vss VssJson) Generate(datasetSize float64, id int) *VssJson {
	timestamp := time.Now().UnixNano()
	result := make(map[string]any)
	leafs := make(map[string]map[string]any)

	for _, leafNode := range vss.LeafNodes() {
		for key, val := range leafNode {
			idx := strings.LastIndex(key, ".")
			if _, ok := leafs[key[:idx]]; !ok {
				leafs[key[:idx]] = make(map[string]any)
			}
			leafs[key[:idx]][key[idx+1:]] = val
		}
	}

	for parent, metadata := range leafs {
		// !! only add the node by datasetSize
		if rand.Float64() > datasetSize {
			continue
		}
		new := result
		var node string
		idx := 0

		for _, node = range strings.Split(parent, ".") {
			idx++
			// If node does not exist, create it
			if _, ok := new[node]; !ok {
				new[node] = make(map[string]any)
				if idx <= strings.Count(parent, ".") {
					new = new[node].(map[string]any)
				}
			} else {
				new = new[node].(map[string]any)
			}
		}

		dtype := metadata["datatype"]
		_, allowed_ok := metadata["allowed"]
		switch {
		case dtype == "boolean":
			new[node] = map[string]any{
				tsonValue:     rand.Float64() < 0.5,
				tsonTimestamp: timestamp,
			}
		case dtype == "int8" || dtype == "uint8" || dtype == "float" && metadata["unit"] == "percent":
			f := rand.Float64() * 100
			if dtype == "float" {
				new[node] = map[string]any{
					tsonValue:     f,
					tsonTimestamp: timestamp,
				}
			} else {
				new[node] = map[string]any{
					tsonValue:     int(f),
					tsonTimestamp: timestamp,
				}
			}
		case allowed_ok:
			array := metadata["allowed"].([]any)
			new[node] = map[string]any{
				tsonValue:     array[rand.Intn(len(array))],
				tsonTimestamp: timestamp,
			}
		case dtype == "double" || dtype == "float":
			new[node] = map[string]any{
				tsonValue:     rand.Float64() * 100,
				tsonTimestamp: timestamp,
			}
		case dtype == "float[]":
			new[node] = make([]any, 0)
			for i := 0; i < rand.Intn(5)+1; i++ {
				new[node] = append(new[node].([]any), map[string]any{
					tsonValue:     rand.Float64() * 100,
					tsonTimestamp: timestamp,
				})
			}
		case dtype == "int8" || dtype == "int16" || dtype == "int32":
			new[node] = map[string]any{
				tsonValue:     rand.Intn(201) - 100,
				tsonTimestamp: timestamp,
			}
		case dtype == "string":
			new[node] = map[string]any{
				tsonValue:     genRandomString(15),
				tsonTimestamp: timestamp,
			}
		case dtype == "string[]":
			new[node] = make([]any, 0)
			for i := 0; i < rand.Intn(5)+1; i++ {
				new[node] = append(new[node].([]any), map[string]any{
					tsonValue:     genRandomString(15),
					tsonTimestamp: timestamp,
				})
			}
		case dtype == "uint8" || dtype == "uint16" || dtype == "uint32":
			new[node] = map[string]any{
				tsonValue:     rand.Intn(101),
				tsonTimestamp: timestamp,
			}
		case dtype == "uint8[]":
			new[node] = make([]any, 0)
			for i := 0; i < rand.Intn(5)+1; i++ {
				new[node] = append(new[node].([]any), map[string]any{
					tsonValue:     rand.Intn(101),
					tsonTimestamp: timestamp,
				})
			}
		case dtype == nil:
			new[node] = map[string]any{
				tsonValue:     nil,
				tsonTimestamp: timestamp,
			}
		}
	}

	if SAVE_METADATA {
		result["metadata"] = map[string]any{
			"CarId":  id,
			"FileNo": 1,
			"Time":   timestamp,
		}
	}
	return &VssJson{initialized: true, data: result}
}

// Generate a new dataset based on the current dataset
func (vss VssJson) GenerateNext(changeRate float64, id int, fileNo int) (*VssJson, *VssJson) {
	if !vss.initialized {
		panic("VssJson must be initialized with VssJson.Generate()")
	}

	timestamp := time.Now().UnixNano()
	result := make(map[string]any)
	leafs := vss.LeafNodes()

	for _, leafNode := range leafs {
		for path, value := range leafNode {
			new := result
			var node string
			idx := 0

			for _, node = range strings.Split(path, ".") {
				idx++
				// If node does not exist, create it
				if _, ok := new[node]; !ok {
					new[node] = make(map[string]any)
					if idx <= strings.Count(path, ".") {
						new = new[node].(map[string]any)
					}
				} else {
					new = new[node].(map[string]any)
				}
			}

			if arr, ok := value.([]any); ok {
				new[node] = make([]any, 0)
				for _, item := range arr {
					new[node] = append(new[node].([]any), map[string]any{
						tsonValue:     item.(map[string]any)[tsonValue],
						tsonTimestamp: timestamp,
					})
				}
				continue
			}

			if rand.Float64() > changeRate {
				new[node] = map[string]any{
					tsonValue:     value.(map[string]any)[tsonValue],
					tsonTimestamp: timestamp,
				}
				continue
			}

			switch value.(map[string]any)[tsonValue].(type) {
			case string:
				new[node] = map[string]any{
					tsonValue:     genRandomString(15),
					tsonTimestamp: timestamp,
				}
			case bool:
				v := value.(map[string]any)[tsonValue].(bool)
				new[node] = map[string]any{
					tsonValue:     !v,
					tsonTimestamp: timestamp,
				}
			case int:
				new[node] = map[string]any{
					tsonValue:     rand.Intn(201) - 100,
					tsonTimestamp: timestamp,
				}
			case float64:
				new[node] = map[string]any{
					tsonValue:     rand.Float64() * 100,
					tsonTimestamp: timestamp,
				}
			}
		}
	}

	if SAVE_METADATA {
		result["metadata"] = map[string]any{
			"CarId":  id,
			"FileNo": fileNo,
			"Time":   timestamp,
		}
	}

	// Step 1. Create a patch using the current dataset and the new dataset
	// var j any
	var origin, modified tson.Tson

	b := mapToJson(vss.data.(map[string]any))
	origin, err := tson.FromJsonBytes(b)
	if err != nil {
		panic(err)
	}

	b = mapToJson(result)
	modified, err = tson.FromJsonBytes(b)
	if err != nil {
		panic(err)
	}

	ops, _ := jsonpatch.GeneratePatch(origin, modified)

	// Step 2. Convert the patch object to a []byte
	bytes, _ := json.Marshal(ops)

	// Step 3. Unmarshal the []byte to a JSON object
	var patch any
	json.Unmarshal(bytes, &patch)

	// Step 4. Return the new dataset and the patch
	return &VssJson{initialized: true, data: result}, &VssJson{initialized: true, data: patch}
}

// Save JSON data to file
func (vss *VssJson) Save(file string) {
	var data []byte

	// If vanilla JSON object
	if _, ok := vss.data.(map[string]any); ok {
		var err error
		data, err = tson.MarshalIndent(vss.data, "", "    ")
		if err != nil {
			fmt.Printf("Error marshalling JSON: %v\n", err)
		}
	} else { // If JSON patch
		var lines []string
		for _, item := range vss.data.([]any) {
			var buf bytes.Buffer
			d := item.(map[string]any)

			fmt.Fprintf(&buf,
				`{ "op": "%s", "path": "%s", "value": %#v, "timestamp": %d }`,
				d["op"], d["path"], d[tsonValue], int64(d[tsonTimestamp].(float64)))
			lines = append(lines, "    "+buf.String())
		}

		// Join all lines with commas and wrap them in square brackets
		data = []byte(fmt.Sprintf("[\n%s\n]", strings.Join(lines, ",\n")))
	}

	// Write data to file
	if err := os.WriteFile(file, data, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
	}
}

// Generate a random string of a given length
func genRandomString(_ int) string {
	return testStrings[rand.Intn(len(testStrings))]
}

// Convert a map to JSON []byte
func mapToJson(data map[string]any) []byte {
	jsonData, _ := json.Marshal(data)
	return jsonData
}

var testStrings = [...]string{
	"TEST_STRING_VALUE_1",
	"TEST_STRING_VALUE_2",
	"TEST_STRING_VALUE_3",
	"TEST_STRING_VALUE_4",
	"TEST_STRING_VALUE_5",
	"TEST_STRING_VALUE_6",
	"TEST_STRING_VALUE_7",
	"TEST_STRING_VALUE_8",
	"TEST_STRING_VALUE_9",
	"TEST_STRING_VALUE_10",

	"TEST_STRING_VALUE_11",
	"TEST_STRING_VALUE_12",
	"TEST_STRING_VALUE_13",
	"TEST_STRING_VALUE_14",
	"TEST_STRING_VALUE_15",
	"TEST_STRING_VALUE_16",
	"TEST_STRING_VALUE_17",
	"TEST_STRING_VALUE_18",
	"TEST_STRING_VALUE_19",
	"TEST_STRING_VALUE_20",
}
