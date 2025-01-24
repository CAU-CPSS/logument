package vssgen

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
)

func ParseArgs(defaultDataset string) Json {
	dataset := flag.String("dataset", defaultDataset, "Path to VSS JSON datset")
	cars := flag.Int("cars", -1, "[required] Number of cars to generate")
	files := flag.Int("files", -1, "[required] Number of JSON files per car")
	changeRate := flag.Float64("change_rate", 0.2, "Change rate for each car")
	size := flag.Float64("size", 1.0, "Dataset size ratio 0.0-1.0")

	flag.Parse()

	if *cars < 0 || *files < 0 {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Running with arguments:")
	fmt.Println("--dataset:", *dataset)
	fmt.Println("--cars:", *cars)
	fmt.Println("--files:", *files)
	fmt.Println("--change_rate:", *changeRate)
	fmt.Println("--size:", *size)

	return Json{
		"dataset":     *dataset,
		"cars":        *cars,
		"files":       *files,
		"change_rate": *changeRate,
		"size":        *size,
	}
}

func PrepareOutputDir(outputDir string) {
	fmt.Print("\nChecking for existing output directory...")
	if stat, err := os.Stat(outputDir); err == nil && stat.IsDir() {
		fmt.Print("\nDirectory found. Removing...")
		if err := os.RemoveAll(outputDir); err != nil {
			fmt.Printf("failed to remove directory.\n\n")
			return
		}
		fmt.Printf("removed successfully.\n\n")
	} else {
		fmt.Printf("not found.\n\n")
	}
	os.MkdirAll(outputDir, os.ModePerm)
}

func SaveMetadata(metadata Json, dataFolder string) {
	metadataFile := filepath.Join(dataFolder, "metadata.json")
	metadataData, _ := json.Marshal(metadata)
	os.WriteFile(metadataFile, metadataData, 0644)
}

func Generate(metadata Json, dataFolder string) {
	dataset, cars, files, changeRate, size := metadata["dataset"].(string), metadata["cars"].(int), metadata["files"].(int), metadata["change_rate"].(float64), metadata["size"].(float64)

	fmt.Printf("Generating %d cars with %d files each...\n", cars, files)
	vss := NewVssJson(dataset)
	bar := progressbar.Default(int64(cars*files), "Generating JSON & JSON patch...")

	for i := 1; i <= cars; i++ {
		// Create path
		carDir := filepath.Join(dataFolder, fmt.Sprintf("car_%d/json", i))
		patchDir := filepath.Join(dataFolder, fmt.Sprintf("car_%d/patches", i))

		if _, err := os.Stat(carDir); os.IsNotExist(err) {
			os.MkdirAll(carDir, os.ModePerm)
		}
		if _, err := os.Stat(patchDir); os.IsNotExist(err) {
			os.MkdirAll(patchDir, os.ModePerm)
		}

		// Generate first JSON files
		data := vss.Generate(size, i)
		data.Save(filepath.Join(carDir, fmt.Sprintf("%d_1.json", i)))
		data.Save(filepath.Join(patchDir, fmt.Sprintf("%d_1.json", i)))
		bar.Add(1)

		// Generate the rest of the JSON files
		for j := 2; j <= files; j++ {
			data, patch := data.GenerateNext(changeRate, i, j)
			data.Save(filepath.Join(carDir, fmt.Sprintf("%d_%d.json", i, j)))
			patch.Save(filepath.Join(patchDir, fmt.Sprintf("%d_%d.json", i, j)))
			bar.Add(1)
		}
	}
}
