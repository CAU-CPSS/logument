package main

import (
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

func main() {
	if ONLY_GENERATE_VSS {
		metadata := map[string]any{
			"dataset":    "internal/vssgen/vss_rel_4.2.json",
			"cars":        100,
			"files":       500,
			"change_rate": 0.2,
			"size":        1.0,
		}
		outputDir := "./dataset"

		vssgen.PrepareOutputDir(outputDir)
		vssgen.SaveMetadata(metadata, outputDir)

		vssgen.Generate(metadata, outputDir)

		fmt.Printf("Saved to %s! Exiting...\n", outputDir)
	}

	// if err := run_jpatch(); err != nil {
	// 	fmt.Println("Application error: %v", err)
	// 	os.Exit(1)
	// }
	if err := run_logument(); err != nil {
		fmt.Printf("Application error: %v\n", err)
		os.Exit(1)
	}
}

func run_logument() error {
	// Logument 초기화
	data := map[string]interface{}{"speed": 100, "location": "Seoul"}
	lm := logument.NewLogument(data)

	// 초기 스냅샷 출력
	fmt.Println("Initial Snapshot:", lm)

	// Patch 추가
	// log.AddPatch("speed", 120)
	// log.AddPatch("location", "Busan")
	// fmt.Println("Patches after updates:", log.Patches)

	// 병합 및 결과 출력
	// log.MergePatches()
	// fmt.Println("Snapshot after merging patches:", log.CreateSnapshot())

	return nil
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
