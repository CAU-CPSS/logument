package mainrr

import (
	"encoding/json"
	"fmt"
	"log"

	

	"jsonpatch"
)

func main() {
	// 원본 JSON
	original := []byte(`{
		"name": "Alice",
		"age": 25,
		"address": {
			"city": "Seoul",
			"zipcode": "12345"
		}
	}`)

	// 변경된 JSON
	modified := []byte(`{
		"name": "Alice",
		"age": 26,
		"address": {
			"city": "Busan",
			"zipcode": "67890"
		},
		"phone": "010-1234-5678"
	}`)

	// JSON Patch 생성
	patch, err := jsonpatch.CreatePatch(original, modified)
	if err != nil {
		log.Fatalf("Error creating patch: %v", err)
	}

	// Patch 결과를 출력
	patchBytes, err := json.MarshalIndent(patch, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling patch: %v", err)
	}

	fmt.Println("Generated JSON Patch:")
	fmt.Println(string(patchBytes))
}
