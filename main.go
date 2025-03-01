//
// main.go
//
// Web-based visualizer for the Logument project.
//
// Authors:
//   Karu (@karu-rress)
//   Sunghwan Park (@tjdghks994)
//

package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	_ "html/template"
	"log"
	"net/http"
	_ "net/http"
	"os"
	"strconv"

	"github.com/CAU-CPSS/logument/internal/logument"
	"github.com/CAU-CPSS/logument/internal/tson"
	"github.com/CAU-CPSS/logument/internal/tsonpatch"
	"github.com/CAU-CPSS/logument/internal/vssgen"
)

//go:embed examples/example1.tson
var expSnapshot []byte

const expPatch = `[
	{ "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 2000000000 },
	{ "op": "replace", "path": "/tirePressure/0", "value": 35.1, "timestamp": 2000000000 },
	{ "op": "replace", "path": "/engineOn", "value": false, "timestamp": 2000000000 }
]`

// homeHandler handles the main page
func homeHandler(w http.ResponseWriter, r *http.Request) {
	html, _ := os.ReadFile("index.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(html)
	// tmpl := template.Must(template.ParseFiles("index.html"))
	// tmpl.Execute(w, nil)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	car := r.URL.Query().Get("car")
	if patch := r.URL.Query().Get("patch"); patch == "" {
		b, _ := os.ReadFile(fmt.Sprintf("dataset/car_%s/tson/%s_1.tson", car, car))
		w.Write(b)
	} else {
		b, _ := os.ReadFile(fmt.Sprintf("dataset/car_%s/patches/%s_%s.json", car, car, patch))
		w.Write(b)
	}
}

func patchHandler(w http.ResponseWriter, r *http.Request) {
	var (
		car, _          = strconv.Atoi(r.URL.Query().Get("car"))
		maxpatch, _     = strconv.Atoi(r.URL.Query().Get("patch"))
		originalTson, _ = os.ReadFile(fmt.Sprintf("dataset/car_%d/tson/%d_1.tson", car, car))
		patches         = make([]tsonpatch.Patch, 0, maxpatch)
	)

	for i := 1; i <= maxpatch; i++ {
		// Read patch file
		patch, _ := os.ReadFile(fmt.Sprintf("dataset/car_%d/patches/%d_%d.json", car, car, i))

		// patch를 변환해서 patches 뒤에 append
		_ = patch
	}

	// TODO: 여기서 orignalTson에 patches를 적용함.

	// TODO: REMOVE
	_, _ = originalTson, patches

	w.Write( /*[]byte("") 대신 생성된 패치로 교체*/ []byte(""))
}

func main() {
	// If the flag is set, generate VSS and exit
	if arguments := os.Args[1:]; len(arguments) > 0 {
		generateVss(true)
		return
	}

	// If not set, run web server

	// 1. Create dataset (cars: 5, files: 10)
	generateVss(false)

	// 2. Run web server
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/patch", patchHandler)
	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)

	var (
		initSnapshot tson.Tson
		j            = expSnapshot
	)

	tson.Unmarshal(j, &initSnapshot)

	// Make a new Logument document
	lgm := logument.NewLogument(initSnapshot, nil)
	// Store the patch in the PatchPool
	lgm.Store(expPatch)
	// Apply the patch to the PatchMap to manage the patch history
	lgm.Append()

	// Make a snapshot at the target version
	targetVesion := uint64(1)
	lgm.Snapshot(targetVesion)

	// Print the Logument document
	lgm.Print()
}

func generateVss(userDefined bool) {
	const defaultDataset = "internal/vssgen/vss.json"
	var option map[string]any
	if userDefined {
		// User-defined VSS generation
		option = vssgen.ParseArgs(defaultDataset)
	} else {
		option = map[string]any{
			"dataset":     defaultDataset,
			"cars":        5,
			"files":       10,
			"change_rate": 0.2,
			"size":        0.1,
		}
	}
	outputDir := "./dataset"

	vssgen.PrepareOutputDir(outputDir)
	vssgen.SaveMetadata(option, outputDir)
	vssgen.Generate(option, outputDir)

	fmt.Printf("Saved to %s! Exiting...\n", outputDir)
}

func runJpatch() error {
	var original, modified tson.Tson

	// Original JSON Document
	tson.Unmarshal([]byte(`{
		"name": "Alice",
		"age": 25,
		"address": {
			"city": "Seoul",
			"zipcode": "12345"
		}
	}`), &original)

	// Modified JSON
	tson.Unmarshal([]byte(`{
		"name": "Alice",
		"age": 26,
		"address": {
			"city": "Busan",
			"zipcode": "67890"
		},
		"phone": "010-1234-5678"
	}`), &modified)

	// Generating JSON Patch
	patch, err := tsonpatch.GeneratePatch(original, modified)
	if err != nil {
		log.Fatalf("Error creating patch: %v", err)
	}
	fmt.Println(patch)

	// Printing the Patch result
	patchBytes, err := json.MarshalIndent(patch, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling patch: %v", err)
	}

	fmt.Println("Generated JSON Patch:")
	fmt.Println(string(patchBytes))

	return err
}
