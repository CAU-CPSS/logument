package exp

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// Graph generation function
func GenerateTimestampPatchGraph() {
	fmt.Println("Generating Figure 1 experimental graph...")

	// Check if result data exists
	graphDataPath := "./dataset_timestamp_experiment/graph_data.csv"
	if _, err := os.Stat(graphDataPath); os.IsNotExist(err) {
		fmt.Printf("Graph data file not found: %s\nPlease run the experiment first.\n", graphDataPath)
		return
	}

	// Read CSV data
	data, err := readGraphData(graphDataPath)
	if err != nil {
		fmt.Printf("Failed to read data: %v\n", err)
		return
	}

	// Generate graph
	if err := createPatchCountGraph(data); err != nil {
		fmt.Printf("Failed to generate graph: %v\n", err)
		return
	}

	fmt.Println("Graph has been generated.")
}

// Function to read CSV data
func readGraphData(filePath string) ([][]float64, error) {
	csvFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Skip the first row as it is the header
	if len(records) <= 1 {
		return nil, fmt.Errorf("Insufficient data")
	}

	// Extract data [update_index, json_patch_count, tson_patch_count]
	data := make([][]float64, len(records)-1)
	for i, record := range records[1:] {
		if len(record) < 3 {
			return nil, fmt.Errorf("Row %d has incomplete data", i+1)
		}

		data[i] = make([]float64, 3)
		for j := 0; j < 3; j++ {
			data[i][j], err = strconv.ParseFloat(record[j], 64)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse data at row %d, column %d: %v", i+1, j, err)
			}
		}
	}

	return data, nil
}

// Function to generate patch count graph
func createPatchCountGraph(data [][]float64) error {
	p := plot.New()

	p.Title.Text = "JSON vs TSON+TestSet: Cumulative Patch Count Comparison"
	p.X.Label.Text = "Number of Updates"
	p.Y.Label.Text = "Cumulative Patch Count"

	// JSON patch data points
	jsonPoints := make(plotter.XYs, len(data))
	for i, d := range data {
		jsonPoints[i].X = d[0] // Update index
		jsonPoints[i].Y = d[1] // JSON patch count
	}

	// TSON patch data points
	tsonPoints := make(plotter.XYs, len(data))
	for i, d := range data {
		tsonPoints[i].X = d[0] // Update index
		tsonPoints[i].Y = d[2] // TSON patch count
	}

	// Add lines to the graph
	if err := plotutil.AddLinePoints(p,
		"JSON Method", jsonPoints,
		"TSON+TestSet Method", tsonPoints); err != nil {
		return err
	}

	// Save the graph
	outputPath := "./dataset_timestamp_experiment/patch_count_comparison.png"
	if err := p.Save(8*vg.Inch, 6*vg.Inch, outputPath); err != nil {
		return err
	}

	// Generate reduction rate graph
	createReductionRateGraph(data)

	return nil
}

// Function to generate reduction rate graph
func createReductionRateGraph(data [][]float64) error {
	p := plot.New()

	p.Title.Text = "Patch Reduction Rate of TSON+TestSet"
	p.X.Label.Text = "Number of Updates"
	p.Y.Label.Text = "Reduction Rate (%)"

	// Calculate reduction rate
	reductionPoints := make(plotter.XYs, len(data))
	for i, d := range data {
		reductionPoints[i].X = d[0] // Update index

		// Reduction rate = (1 - TSON_Patch_Count/JSON_Patch_Count) * 100
		if d[1] > 0 {
			reductionPoints[i].Y = (1 - d[2]/d[1]) * 100
		} else {
			reductionPoints[i].Y = 0
		}
	}

	// Add lines to the graph
	line, points, err := plotter.NewLinePoints(reductionPoints)
	if err != nil {
		return err
	}
	line.Color = plotutil.Color(2)
	points.Color = plotutil.Color(2)

	p.Add(line, points)
	p.Legend.Add("Reduction Rate", line, points)

	// Set Y-axis range (0-100%)
	p.Y.Min = 0
	p.Y.Max = 100

	// Save the graph
	outputPath := "./dataset_timestamp_experiment/reduction_rate.png"
	if err := p.Save(8*vg.Inch, 6*vg.Inch, outputPath); err != nil {
		return err
	}

	return nil
}

// Function to generate simulation data (for viewing graphs without actual experiments)
func GenerateSimulationData() {
	fmt.Println("Generating simulation data...")

	// Create output directory
	os.MkdirAll("./dataset_timestamp_experiment", os.ModePerm)

	// Parameters
	numUpdates := 100
	changeRate := 0.2
	avgPathsPerUpdate := 50

	// Create CSV file to save results
	graphCsv, err := os.Create("./dataset_timestamp_experiment/graph_data.csv")
	if err != nil {
		fmt.Printf("Failed to create graph data CSV file: %v\n", err)
		return
	}
	defer graphCsv.Close()

	graphWriter := csv.NewWriter(graphCsv)
	defer graphWriter.Flush()

	// Write header
	graphWriter.Write([]string{"Update_Index", "JSON_Patch_Count", "TSON_Patch_Count"})

	// Write data rows (cumulative patch count)
	jsonTotal := 0
	tsonTotal := 0

	rand.New(rand.NewSource(42))

	for i := 1; i <= numUpdates; i++ {
		// Randomize the number of paths processed in each update (distributed around the average)
		pathsInThisUpdate := int(math.Max(1, float64(avgPathsPerUpdate)*(0.8+0.4*rand.Float64())))

		// Calculate the number of value changes in this update
		valueChanges := 0
		for j := 0; j < pathsInThisUpdate; j++ {
			if rand.Float64() < changeRate {
				valueChanges++
			}
		}

		// Number of timestamp-only changes
		// timestampOnlyChanges := pathsInThisUpdate - valueChanges

		// JSON method generates timestamp patches for all paths
		jsonPatchCount := pathsInThisUpdate + valueChanges // Timestamps + value changes
		jsonTotal += jsonPatchCount

		// TSON method generates patches only for paths with value changes
		tsonPatchCount := valueChanges
		tsonTotal += tsonPatchCount

		// Write data
		row := []string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("%d", jsonTotal),
			fmt.Sprintf("%d", tsonTotal),
		}
		graphWriter.Write(row)
	}

	fmt.Println("Simulation data has been generated.")
}
