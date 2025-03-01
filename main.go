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
	"html/template"
	"log"
	"net/http"
	_ "net/http"
	"os"

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
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

// AJAX request handler for patching
func patchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	input1 := r.FormValue("input1")
	input2 := r.FormValue("input2")

	_ = input1
	_ = input2

	// func1 호출
	// result := func1(input1, input2)

	// 결과 반환
	// w.Write([]byte(result))
}

func main() {
	// If the flag is set, generate VSS and exit
	if arguments := os.Args[1:]; len(arguments) > 0 {
		generateVss()
		return
	}

	////

	///

	///

	//

	////

	///

	http.HandleFunc("/", homeHandler)
	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)

	http.HandleFunc("/compute", patchHandler) // AJAX 요청 처리

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

func generateVss() {
	option := vssgen.ParseArgs("internal/vssgen/vss.json")
	outputDir := "./dataset"

	vssgen.PrepareOutputDir(outputDir)
	vssgen.SaveMetadata(option, outputDir)
	vssgen.Generate(option, outputDir)

	fmt.Printf("Saved to %s! Exiting...\n", outputDir)
}

func runJpatch() error {
	// Original JSON Document
	original, _ := tson.DEPRECATEDNewTson(`{
		"name": "Alice",
		"age": 25,
		"address": {
			"city": "Seoul",
			"zipcode": "12345"
		}
	}`)

	// Modified JSON
	modified, _ := tson.DEPRECATEDNewTson(`{
		"name": "Alice",
		"age": 26,
		"address": {
			"city": "Busan",
			"zipcode": "67890"
		},
		"phone": "010-1234-5678"
	}`)

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
