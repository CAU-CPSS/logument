//
// main.go
//
// Web-based visualizer for the Logument project.
//
// Author: Karu (@karu-rress)
//

package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/CAU-CPSS/logument/internal/logument"
	"github.com/CAU-CPSS/logument/internal/tson"
	"github.com/CAU-CPSS/logument/internal/vssgen"
)

const (
	defaultDataset   = "internal/vssgen/vss.json"
	defaultCarCount  = 5
	defaultFileCount = 10
)

func main() {
	// If the flag is set, generate VSS and exit
	if arguments := os.Args[1:]; len(arguments) > 0 {
		generateVss(true)
		return
	}
	// Otherwise, run web server

	// 1. Create dataset with default settings
	generateVss(false)

	// 2. Run web server
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/patch", patchHandler)
	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func generateVss(userDefined bool) {
	var option map[string]any

	if userDefined {
		// User-defined VSS generation
		option = vssgen.ParseArgs(defaultDataset)
	} else {
		option = map[string]any{
			"dataset":     defaultDataset,
			"cars":        defaultCarCount,
			"files":       defaultFileCount,
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

// homeHandler handles the main page
func homeHandler(w http.ResponseWriter, r *http.Request) {
	html, _ := os.ReadFile("index.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(html)
}

// updateHandler handles the update request, sending the TSON or Patch
func updateHandler(w http.ResponseWriter, r *http.Request) {
	car := r.URL.Query().Get("car")
	if patch := r.URL.Query().Get("patch"); patch == "" {
		// Send the TSON snapshot
		b, _ := os.ReadFile(fmt.Sprintf("dataset/car_%s/tson/%s_1.tson", car, car))
		w.Write(b)
	} else { // Send the patch
		b, _ := os.ReadFile(fmt.Sprintf("dataset/car_%s/patches/%s_%s.json", car, car, patch))
		w.Write(b)
	}
}

// patchHandler handles the patch request, sending the patched TSON
func patchHandler(w http.ResponseWriter, r *http.Request) {
	var (
		car, _          = strconv.Atoi(r.URL.Query().Get("car"))
		maxpatch, _     = strconv.Atoi(r.URL.Query().Get("patch"))
		originalTson, _ = os.ReadFile(fmt.Sprintf("dataset/car_%d/tson/%d_1.tson", car, car))
		lgm             = logument.NewLogument(originalTson, nil)
	)

	for i := 2; i <= maxpatch; i++ {
		// Read patch file
		fileName := fmt.Sprintf("dataset/car_%d/patches/%d_%d.json", car, car, i)
		patch, _ := os.ReadFile(fileName)

		// Append it to the PatchPool
		lgm.Store(patch)
	}

	// Apply the pathes and make a snapshot
	lgm.Append()

	// Take the snapshot
	snapshot := lgm.Snapshot(1)
	result, _ := tson.MarshalIndent(snapshot, "", "  ")

	// Send the result
	w.Write(result)
}
