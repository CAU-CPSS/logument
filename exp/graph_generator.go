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

// 그래프 생성 함수
func GenerateTimestampPatchGraph() {
	fmt.Println("Figure 1 실험 그래프 생성 중...")

	// 결과 데이터가 있는지 확인
	graphDataPath := "./dataset_timestamp_experiment/graph_data.csv"
	if _, err := os.Stat(graphDataPath); os.IsNotExist(err) {
		fmt.Printf("그래프 데이터 파일이 없습니다: %s\n먼저 실험을 실행해주세요.\n", graphDataPath)
		return
	}

	// CSV 데이터 읽기
	data, err := readGraphData(graphDataPath)
	if err != nil {
		fmt.Printf("데이터 읽기 실패: %v\n", err)
		return
	}

	// 그래프 생성
	if err := createPatchCountGraph(data); err != nil {
		fmt.Printf("그래프 생성 실패: %v\n", err)
		return
	}

	fmt.Println("그래프가 생성되었습니다.")
}

// CSV 데이터 읽기 함수
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

	// 첫 번째 행은 헤더이므로 건너뜀
	if len(records) <= 1 {
		return nil, fmt.Errorf("데이터가 충분하지 않습니다")
	}

	// 데이터 추출 [update_index, json_patch_count, tson_patch_count]
	data := make([][]float64, len(records)-1)
	for i, record := range records[1:] {
		if len(record) < 3 {
			return nil, fmt.Errorf("행 %d의 데이터가 불완전합니다", i+1)
		}

		data[i] = make([]float64, 3)
		for j := 0; j < 3; j++ {
			data[i][j], err = strconv.ParseFloat(record[j], 64)
			if err != nil {
				return nil, fmt.Errorf("행 %d, 열 %d의 데이터를 파싱할 수 없습니다: %v", i+1, j, err)
			}
		}
	}

	return data, nil
}

// 패치 개수 그래프 생성 함수
func createPatchCountGraph(data [][]float64) error {
	p := plot.New()

	p.Title.Text = "JSON vs TSON+TestSet: 누적 패치 개수 비교"
	p.X.Label.Text = "업데이트 횟수"
	p.Y.Label.Text = "누적 패치 개수"

	// JSON 패치 데이터 포인트
	jsonPoints := make(plotter.XYs, len(data))
	for i, d := range data {
		jsonPoints[i].X = d[0] // 업데이트 인덱스
		jsonPoints[i].Y = d[1] // JSON 패치 개수
	}

	// TSON 패치 데이터 포인트
	tsonPoints := make(plotter.XYs, len(data))
	for i, d := range data {
		tsonPoints[i].X = d[0] // 업데이트 인덱스
		tsonPoints[i].Y = d[2] // TSON 패치 개수
	}

	// 그래프에 선 추가
	if err := plotutil.AddLinePoints(p,
		"JSON 방식", jsonPoints,
		"TSON+TestSet 방식", tsonPoints); err != nil {
		return err
	}

	// 그래프 저장
	outputPath := "./dataset_timestamp_experiment/patch_count_comparison.png"
	if err := p.Save(8*vg.Inch, 6*vg.Inch, outputPath); err != nil {
		return err
	}

	// 감소율 그래프 생성
	createReductionRateGraph(data)

	return nil
}

// 감소율 그래프 생성 함수
func createReductionRateGraph(data [][]float64) error {
	p := plot.New()

	p.Title.Text = "TSON+TestSet의 패치 감소율"
	p.X.Label.Text = "업데이트 횟수"
	p.Y.Label.Text = "패치 감소율 (%)"

	// 감소율 계산
	reductionPoints := make(plotter.XYs, len(data))
	for i, d := range data {
		reductionPoints[i].X = d[0] // 업데이트 인덱스

		// 감소율 = (1 - TSON패치수/JSON패치수) * 100
		if d[1] > 0 {
			reductionPoints[i].Y = (1 - d[2]/d[1]) * 100
		} else {
			reductionPoints[i].Y = 0
		}
	}

	// 그래프에 선 추가
	line, points, err := plotter.NewLinePoints(reductionPoints)
	if err != nil {
		return err
	}
	line.Color = plotutil.Color(2)
	points.Color = plotutil.Color(2)

	p.Add(line, points)
	p.Legend.Add("감소율", line, points)

	// Y축 범위 설정 (0-100%)
	p.Y.Min = 0
	p.Y.Max = 100

	// 그래프 저장
	outputPath := "./dataset_timestamp_experiment/reduction_rate.png"
	if err := p.Save(8*vg.Inch, 6*vg.Inch, outputPath); err != nil {
		return err
	}

	return nil
}

// 시뮬레이션 데이터 생성 함수 (실제 실험 없이 그래프만 보기 위한 기능)
func GenerateSimulationData() {
	fmt.Println("시뮬레이션 데이터 생성 중...")

	// 출력 디렉토리 생성
	os.MkdirAll("./dataset_timestamp_experiment", os.ModePerm)

	// 파라미터
	numUpdates := 100
	changeRate := 0.2
	avgPathsPerUpdate := 50

	// 결과 저장을 위한 CSV 파일 생성
	graphCsv, err := os.Create("./dataset_timestamp_experiment/graph_data.csv")
	if err != nil {
		fmt.Printf("그래프 데이터 CSV 파일 생성 실패: %v\n", err)
		return
	}
	defer graphCsv.Close()

	graphWriter := csv.NewWriter(graphCsv)
	defer graphWriter.Flush()

	// 헤더 작성
	graphWriter.Write([]string{"Update_Index", "JSON_Patch_Count", "TSON_Patch_Count"})

	// 데이터 행 작성 (누적 패치 수)
	jsonTotal := 0
	tsonTotal := 0

	rand.New(rand.NewSource(42))

	for i := 1; i <= numUpdates; i++ {
		// 각 업데이트에서 처리되는 경로의 수 랜덤화 (평균 주변으로 분포)
		pathsInThisUpdate := int(math.Max(1, float64(avgPathsPerUpdate)*(0.8+0.4*rand.Float64())))

		// 이 업데이트에서 값이 변경되는 횟수 계산
		valueChanges := 0
		for j := 0; j < pathsInThisUpdate; j++ {
			if rand.Float64() < changeRate {
				valueChanges++
			}
		}

		// 타임스탬프만 변경되는 횟수
		// timestampOnlyChanges := pathsInThisUpdate - valueChanges

		// JSON 방식은 모든 경로에 대해 타임스탬프 패치를 생성
		jsonPatchCount := pathsInThisUpdate + valueChanges // 타임스탬프 + 값 변경
		jsonTotal += jsonPatchCount

		// TSON 방식은 값이 변경된 경로에 대해서만 패치 생성
		tsonPatchCount := valueChanges
		tsonTotal += tsonPatchCount

		// 데이터 작성
		row := []string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("%d", jsonTotal),
			fmt.Sprintf("%d", tsonTotal),
		}
		graphWriter.Write(row)
	}

	fmt.Println("시뮬레이션 데이터가 생성되었습니다.")
}
