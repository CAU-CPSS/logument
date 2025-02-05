package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/CAU-CPSS/logument/internal/jsonpatch"
	"github.com/CAU-CPSS/logument/internal/jsonr"
	"github.com/CAU-CPSS/logument/internal/logument"
	"github.com/CAU-CPSS/logument/internal/vssgen"
)

const ONLY_GENERATE_VSS = true

const (
	NumCars = 5
	NumPatches = 30
	ChangeRate = 0.2
)

//go:embed examples/example1.jsonr
var expSnapshot []byte

const expPatch = `[
	{ "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 2000000000 },
	{ "op": "replace", "path": "/tirePressure/0", "value": 35.1, "timestamp": 2000000000 },
	{ "op": "replace", "path": "/engineOn", "value": false, "timestamp": 2000000000 }
]`

func main() {
	arguments := os.Args[1:]

	if len(arguments) > 0 {
		generate_vss()
		return
	}

	var (
		initSnapshot jsonr.JsonR
		j            = expSnapshot
	)

	jsonr.Unmarshal(j, &initSnapshot)

	// Make a new Logument document
	lgm := logument.NewLogument(initSnapshot, nil)
	// Store the patch in the PatchPool
	lgm.Store(expPatch)
	// Apply the patch to the PatchMap to manage the patch history
	lgm.Apply()

	// Make a snapshot at the target version
	targetVesion := uint64(1)
	lgm.Snapshot(targetVesion)

	// Print the Logument document
	lgm.Print()
}

func generate_vss() {
	option := vssgen.ParseArgs("internal/vssgen/vss.json")
	outputDir := "./dataset"

	vssgen.PrepareOutputDir(outputDir)
	vssgen.SaveMetadata(option, outputDir)
	vssgen.Generate(option, outputDir)

	fmt.Printf("Saved to %s! Exiting...\n", outputDir)
}

func run_jpatch() error {
	// 원본 JSON 문서
	original, _ := jsonr.NewJsonR(`{
		"name": "Alice",
		"age": 25,
		"address": {
			"city": "Seoul",
			"zipcode": "12345"
		}
	}`)

	// 변경된 JSON
	modified, _ := jsonr.NewJsonR(`{
		"name": "Alice",
		"age": 26,
		"address": {
			"city": "Busan",
			"zipcode": "67890"
		},
		"phone": "010-1234-5678"
	}`)

	// JSON Patch 생성
	patch, err := jsonpatch.GeneratePatch(original, modified)
	if err != nil {
		log.Fatalf("Error creating patch: %v", err)
	}
	fmt.Println(patch)

	// Patch 결과를 출력
	patchBytes, err := json.MarshalIndent(patch, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling patch: %v", err)
	}

	fmt.Println("Generated JSON Patch:")
	fmt.Println(string(patchBytes))

	return err
}
