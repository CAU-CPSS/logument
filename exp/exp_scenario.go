package exp

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// 실제 자율주행차량 시나리오를 모방한 데이터 모델
type VehicleData struct {
	// 센서 그룹별 업데이트 빈도와 변경 패턴이 다름
	SensorsHighFreq  map[string]*SensorData // 고빈도 업데이트 센서 (GPS, 속도 등)
	SensorsMedFreq   map[string]*SensorData // 중간 빈도 업데이트 센서 (온도, 배터리 등)
	SensorsLowFreq   map[string]*SensorData // 저빈도 업데이트 센서 (연료, 타이어 압력 등)
	ActuatorsHighVar map[string]*SensorData // 높은 변동성 액추에이터 (브레이크, 가속 등)
	ActuatorsLowVar  map[string]*SensorData // 낮은 변동성 액추에이터 (창문, 도어 등)
	Attributes       map[string]*SensorData // 거의 변경되지 않는 속성값 (차량 ID, 모델 등)
}

// 개별 센서/액추에이터 데이터
type SensorData struct {
	Path           string             // 센서 경로
	Value          interface{}        // 현재 값
	Type           string             // 값 타입 (number, string, boolean)
	Timestamp      int64              // 마지막 업데이트 타임스탬프
	UpdateInterval int                // 업데이트 간격 (밀리초)
	ChangePattern  string             // 변경 패턴 (constant, linear, sinusoidal, random, triggered)
	ChangeParams   map[string]float64 // 변경 패턴 파라미터
	ChangeCounter  int                // 변경 카운터 (시간 경과 추적용)
}

// JSON 문서 (value-timestamp 구조)
type JsonDoc map[string]interface{}

// TSON 문서 (타임스탬프가 메타데이터)
type TsonDoc map[string]interface{}

// JSON 패치
type JsonPatch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// TSON 패치
type TsonPatch struct {
	Op        string      `json:"op"`
	Path      string      `json:"path"`
	Value     interface{} `json:"value"`
	Timestamp int64       `json:"timestamp"`
}

// 실험 결과
type ExperimentResult struct {
	Timestamp            int64   // 실험 타임스탬프
	Scenario             string  // 시나리오 이름
	SimulationTimeMs     int64   // 시뮬레이션 시간 (밀리초)
	TotalUpdates         int     // 총 업데이트 횟수
	ValueChanges         int     // 값 변경 횟수
	TimestampOnlyChanges int     // 타임스탬프만 변경된 횟수
	JsonPatchCount       int     // JSON 패치 개수
	TsonPatchCount       int     // TSON 패치 개수
	JsonPatchSize        int     // JSON 패치 크기 (바이트)
	TsonPatchSize        int     // TSON 패치 크기 (바이트)
	JsonProcessingTimeNs int64   // JSON 처리 시간 (나노초)
	TsonProcessingTimeNs int64   // TSON 처리 시간 (나노초)
	JsonBandwidthUsage   float64 // JSON 대역폭 사용량 (KB/s)
	TsonBandwidthUsage   float64 // TSON 대역폭 사용량 (KB/s)
	CumulativeTimeVec    []int64 // 누적 시간 벡터
	CumulativeJsonVec    []int   // 누적 JSON 패치 개수 벡터
	CumulativeTsonVec    []int   // 누적 TSON 패치 개수 벡터
}

// 처리 시간 기록
type TimingRecord struct {
	UpdateIndex          int    // 업데이트 인덱스
	TimeMs               int64  // 시뮬레이션 시간 (밀리초)
	SensorPath           string // 센서 경로
	JsonProcessingTimeNs int64  // JSON 처리 시간 (나노초)
	TsonProcessingTimeNs int64  // TSON 처리 시간 (나노초)
	ValueChanged         bool   // 값 변경 여부
}

// 시뮬레이션 파라미터
const (
	outputDir             = "./results/patch_experiment"
	simulationDurationSec = 300 // 5분 시뮬레이션
	simulationTimeStepMs  = 100 // 100ms 단위로 업데이트
	numHighFreqSensors    = 10  // 고빈도 센서 수
	numMedFreqSensors     = 15  // 중간 빈도 센서 수
	numLowFreqSensors     = 20  // 저빈도 센서 수
	numHighVarActuators   = 5   // 높은 변동성 액추에이터 수
	numLowVarActuators    = 10  // 낮은 변동성 액추에이터 수
	numAttributes         = 8   // 속성 수
)

// 자율주행 시나리오
var drivingScenarios = []string{
	"urban_traffic",    // 도심 교통
	"highway_cruising", // 고속도로 주행
	// "parking_maneuver", // 주차 조작
	// "traffic_jam",      // 교통 체증
	"battery_charging", // 배터리 충전
}

func RealworldScenario() {
	// 출력 디렉토리 생성
	os.MkdirAll(outputDir, os.ModePerm)

	// 모든 시나리오에 대한 실험 수행
	for _, scenario := range drivingScenarios {
		fmt.Printf("\n=== 시나리오: %s 실험 시작 ===\n", scenario)

		// 시나리오별 결과 파일 생성
		resultsFile, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("%s_results.csv", scenario)))
		if err != nil {
			fmt.Printf("%s 결과 파일 생성 오류: %v\n", scenario, err)
			continue
		}

		csvWriter := csv.NewWriter(resultsFile)

		// CSV 헤더 작성
		headers := []string{
			"Timestamp",
			"SimulationTimeMs",
			"TotalUpdates",
			"ValueChanges",
			"TimestampOnlyChanges",
			"JsonPatchCount",
			"TsonPatchCount",
			"JsonPatchSize",
			"TsonPatchSize",
			"JsonProcessingTimeNs",
			"TsonProcessingTimeNs",
			"JsonBandwidthUsage",
			"TsonBandwidthUsage",
			"PatchReduction",
		}
		csvWriter.Write(headers)

		// 누적 데이터 저장용 파일
		cumulativeFile, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("%s_cumulative.csv", scenario)))
		if err != nil {
			fmt.Printf("%s 누적 데이터 파일 생성 오류: %v\n", scenario, err)
			resultsFile.Close()
			continue
		}

		cumWriter := csv.NewWriter(cumulativeFile)

		// 누적 데이터 헤더 작성
		cumWriter.Write([]string{
			"TimeMs",
			"JsonPatchCount",
			"TsonPatchCount",
		})

		// 처리 시간 측정용 파일
		timingFile, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("%s_timing.csv", scenario)))
		if err != nil {
			fmt.Printf("%s 처리 시간 파일 생성 오류: %v\n", scenario, err)
			resultsFile.Close()
			cumulativeFile.Close()
			continue
		}

		timingWriter := csv.NewWriter(timingFile)

		// 처리 시간 헤더 작성
		timingWriter.Write([]string{
			"UpdateIndex",
			"TimeMs",
			"SensorPath",
			"JsonProcessingTimeNs",
			"TsonProcessingTimeNs",
			"ValueChanged",
			"ProcessingTimeDiffPct",
		})

		// 실험 실행
		results, timingData := runRealisticExperiment(scenario, timingWriter)

		fmt.Printf("Timing Data: %d\n", len(timingData))

		fmt.Printf("\n--- %s 시나리오 실험 결과 ---\n", scenario)
		fmt.Printf("총 업데이트 횟수: %d\n", results.TotalUpdates)
		fmt.Printf("값 변경 횟수: %d (%.1f%%)\n",
			results.ValueChanges,
			float64(results.ValueChanges)*100/float64(results.TotalUpdates))
		fmt.Printf("타임스탬프만 변경 횟수: %d (%.1f%%)\n",
			results.TimestampOnlyChanges,
			float64(results.TimestampOnlyChanges)*100/float64(results.TotalUpdates))
		fmt.Printf("JSON 패치 개수: %d\n", results.JsonPatchCount)
		fmt.Printf("TSON 패치 개수: %d\n", results.TsonPatchCount)
		fmt.Printf("패치 감소율: %.1f%%\n",
			(1-float64(results.TsonPatchCount)/float64(results.JsonPatchCount))*100)
		fmt.Printf("JSON 처리 시간: %.2f ms\n", float64(results.JsonProcessingTimeNs)/1e6)
		fmt.Printf("TSON 처리 시간: %.2f ms\n", float64(results.TsonProcessingTimeNs)/1e6)
		fmt.Printf("처리 시간 차이: %.1f%%\n",
			(1-float64(results.TsonProcessingTimeNs)/float64(results.JsonProcessingTimeNs))*100)

		// CSV에 결과 저장
		reductionRate := (1 - float64(results.TsonPatchCount)/float64(results.JsonPatchCount)) * 100

		row := []string{
			fmt.Sprintf("%d", results.Timestamp),
			fmt.Sprintf("%d", results.SimulationTimeMs),
			fmt.Sprintf("%d", results.TotalUpdates),
			fmt.Sprintf("%d", results.ValueChanges),
			fmt.Sprintf("%d", results.TimestampOnlyChanges),
			fmt.Sprintf("%d", results.JsonPatchCount),
			fmt.Sprintf("%d", results.TsonPatchCount),
			fmt.Sprintf("%d", results.JsonPatchSize),
			fmt.Sprintf("%d", results.TsonPatchSize),
			fmt.Sprintf("%d", results.JsonProcessingTimeNs),
			fmt.Sprintf("%d", results.TsonProcessingTimeNs),
			fmt.Sprintf("%.2f", results.JsonBandwidthUsage),
			fmt.Sprintf("%.2f", results.TsonBandwidthUsage),
			fmt.Sprintf("%.2f", reductionRate),
		}
		csvWriter.Write(row)
		csvWriter.Flush()

		// 누적 데이터 저장
		timeVector, jsonVector, tsonVector := results.CumulativeTimeVec, results.CumulativeJsonVec, results.CumulativeTsonVec
		for i := 0; i < len(timeVector); i++ {
			cumWriter.Write([]string{
				fmt.Sprintf("%d", timeVector[i]),
				fmt.Sprintf("%d", jsonVector[i]),
				fmt.Sprintf("%d", tsonVector[i]),
			})
		}
		cumWriter.Flush()

		// 파일 닫기
		resultsFile.Close()
		cumulativeFile.Close()
		timingFile.Close()

		fmt.Printf("\n%s 시나리오 실험 완료. 결과 파일 저장됨:\n", scenario)
		fmt.Printf("- 요약 결과: %s\n", filepath.Join(outputDir, fmt.Sprintf("%s_results.csv", scenario)))
		fmt.Printf("- 누적 패치: %s\n", filepath.Join(outputDir, fmt.Sprintf("%s_cumulative.csv", scenario)))
		fmt.Printf("- 처리 시간: %s\n", filepath.Join(outputDir, fmt.Sprintf("%s_timing.csv", scenario)))
	}

	fmt.Println("\n모든 실험이 완료되었습니다. 결과는 dataset_timestamp_experiment 디렉토리에 저장되었습니다.")
}

// 실제적인 실험 수행
func runRealisticExperiment(scenario string, timingWriter *csv.Writer) (ExperimentResult, []TimingRecord) {
	fmt.Printf("시나리오 '%s' 시뮬레이션 시작...\n", scenario)

	// 이 시나리오에 대한 데이터 저장소 인스턴스 생성
	dataStorage := NewDataStorage(outputDir, scenario)

	// 1. 초기 데이터 생성
	vehicle := createInitialVehicleData(scenario)

	// 2. JSON 및 TSON 문서 초기화
	jsonDoc := vehicleDataToJsonDoc(vehicle)
	tsonDoc := vehicleDataToTsonDoc(vehicle)

	// 초기 문서 저장
	if err := dataStorage.SaveInitialJSON(jsonDoc); err != nil {
		fmt.Printf("초기 JSON 저장 오류: %v\n", err)
	}
	if err := dataStorage.SaveInitialTSON(tsonDoc); err != nil {
		fmt.Printf("초기 TSON 저장 오류: %v\n", err)
	}

	// 3. 시뮬레이션 결과 변수 초기화
	result := ExperimentResult{
		Timestamp:        time.Now().Unix(),
		Scenario:         scenario,
		SimulationTimeMs: simulationDurationSec * 1000,
	}

	jsonPatches := make([]JsonPatch, 0)
	tsonPatches := make([]TsonPatch, 0)

	// 타임스탬프 수집을 위한 벡터 초기화
	timeVec := make([]int64, 0, simulationDurationSec*1000/simulationTimeStepMs)
	jsonPatchesVec := make([]int, 0, simulationDurationSec*1000/simulationTimeStepMs)
	tsonPatchesVec := make([]int, 0, simulationDurationSec*1000/simulationTimeStepMs)

	// 처리 시간 기록 배열
	timingRecords := make([]TimingRecord, 0)

	// 대략적인 총 업데이트 횟수 예측
	totalExpectedUpdates := 0

	// 각 센서 그룹에 대한 업데이트 횟수 예측
	for _, sensor := range vehicle.SensorsHighFreq {
		totalExpectedUpdates += simulationDurationSec * 1000 / sensor.UpdateInterval
	}
	for _, sensor := range vehicle.SensorsMedFreq {
		totalExpectedUpdates += simulationDurationSec * 1000 / sensor.UpdateInterval
	}
	for _, sensor := range vehicle.SensorsLowFreq {
		totalExpectedUpdates += simulationDurationSec * 1000 / sensor.UpdateInterval
	}
	for _, actuator := range vehicle.ActuatorsHighVar {
		totalExpectedUpdates += simulationDurationSec * 1000 / actuator.UpdateInterval
	}
	for _, actuator := range vehicle.ActuatorsLowVar {
		totalExpectedUpdates += simulationDurationSec * 1000 / actuator.UpdateInterval
	}

	// 4. 시뮬레이션 실행
	fmt.Printf("예상 업데이트 수: 약 %d\n", totalExpectedUpdates)

	simulationStart := time.Now()
	jsonTotalProcessingTime := int64(0)
	tsonTotalProcessingTime := int64(0)

	// 진행 상황 표시용 카운터
	updateCounter := 0
	valueChangeCounter := 0
	timestampChangeCounter := 0
	progressCounter := 0

	// 시뮬레이션 시간 스텝 (ms)
	for currentTimeMs := int64(0); currentTimeMs < simulationDurationSec*1000; currentTimeMs += simulationTimeStepMs {
		// 타임스탬프 벡터 업데이트
		if currentTimeMs%(simulationTimeStepMs*10) == 0 {
			timeVec = append(timeVec, currentTimeMs)
			jsonPatchesVec = append(jsonPatchesVec, len(jsonPatches))
			tsonPatchesVec = append(tsonPatchesVec, len(tsonPatches))
		}

		// 진행 상황 표시
		if currentTimeMs%(simulationDurationSec*100) == 0 {
			progressPct := float64(currentTimeMs) / float64(simulationDurationSec*1000) * 100
			fmt.Printf("\r시뮬레이션 진행: %.1f%% (시간: %d ms, 업데이트: %d, 변경: %d)",
				progressPct, currentTimeMs, updateCounter, valueChangeCounter)
			progressCounter++
		}

		// 각 센서를 확인하고 업데이트
		for _, sensorGroups := range []map[string]*SensorData{
			vehicle.SensorsHighFreq,
			vehicle.SensorsMedFreq,
			vehicle.SensorsLowFreq,
			vehicle.ActuatorsHighVar,
			vehicle.ActuatorsLowVar,
			vehicle.Attributes,
		} {
			for path, sensor := range sensorGroups {
				// 이 센서가 이번 타임스텝에서 업데이트되어야 하는지 확인
				if currentTimeMs%int64(sensor.UpdateInterval) == 0 {
					updateCounter++

					// 새 값과 타임스탬프 계산
					newValue := calculateNewValue(sensor, currentTimeMs, scenario)
					newTimestamp := simulationStart.UnixNano() + currentTimeMs*1000000 // ms -> ns

					// 값이 변경되었는지 확인
					valueChanged := !isEqual(sensor.Value, newValue)

					if valueChanged {
						valueChangeCounter++
						sensor.Value = newValue
					} else {
						timestampChangeCounter++
					}

					sensor.Timestamp = newTimestamp

					// JSON 처리 시작 시간
					jsonStartTime := time.Now()

					// JSON 패치 생성 (항상 timestamp 업데이트, 값이 변경된 경우 value도 업데이트)
					// 타임스탬프 패치
					jsonPatches = append(jsonPatches, JsonPatch{
						Op:    "replace",
						Path:  path + "/timestamp",
						Value: newTimestamp,
					})

					// 값이 변경된 경우 value 패치도 추가
					if valueChanged {
						jsonPatches = append(jsonPatches, JsonPatch{
							Op:    "replace",
							Path:  path + "/value",
							Value: newValue,
						})
					}

					// JSON 문서 업데이트
					updateJsonDocument(jsonDoc, path, newValue, newTimestamp)

					// JSON 처리 시간 측정
					jsonProcessingTime := time.Since(jsonStartTime).Nanoseconds()
					jsonTotalProcessingTime += jsonProcessingTime

					// TSON 처리 시작 시간
					tsonStartTime := time.Now()

					// TSON+TestSet 업데이트 (값이 변경된 경우만 패치 생성)
					if valueChanged {
						// TSON 패치 생성
						tsonPatches = append(tsonPatches, TsonPatch{
							Op:        "replace",
							Path:      path,
							Value:     newValue,
							Timestamp: newTimestamp,
						})

						// TSON 문서 업데이트
						updateTsonDocument(tsonDoc, path, newValue, newTimestamp)
					}

					// TSON 처리 시간 측정
					tsonProcessingTime := time.Since(tsonStartTime).Nanoseconds()
					tsonTotalProcessingTime += tsonProcessingTime

					// 처리 시간 기록 저장
					timingRecord := TimingRecord{
						UpdateIndex:          updateCounter,
						TimeMs:               currentTimeMs,
						SensorPath:           path,
						JsonProcessingTimeNs: jsonProcessingTime,
						TsonProcessingTimeNs: tsonProcessingTime,
						ValueChanged:         valueChanged,
					}
					timingRecords = append(timingRecords, timingRecord)

					// 처리 시간 CSV에 기록
					timingDiffPct := 0.0
					if jsonProcessingTime > 0 {
						timingDiffPct = (1.0 - float64(tsonProcessingTime)/float64(jsonProcessingTime)) * 100.0
					}

					timingWriter.Write([]string{
						fmt.Sprintf("%d", updateCounter),
						fmt.Sprintf("%d", currentTimeMs),
						path,
						fmt.Sprintf("%d", jsonProcessingTime),
						fmt.Sprintf("%d", tsonProcessingTime),
						fmt.Sprintf("%t", valueChanged),
						fmt.Sprintf("%.2f", timingDiffPct),
					})

					// 주기적으로 버퍼 비우기
					if updateCounter%100 == 0 {
						timingWriter.Flush()
					}
				}
			}
		}
	}

	// JSON 패치 크기 계산
	jsonPatchBytes, _ := json.Marshal(jsonPatches)
	jsonPatchSize := len(jsonPatchBytes)

	// TSON 패치 크기 계산
	tsonPatchBytes, _ := json.Marshal(tsonPatches)
	tsonPatchSize := len(tsonPatchBytes)

	// 대역폭 사용량 계산 (KB/s)
	simulationTimeSeconds := float64(simulationDurationSec)
	jsonBandwidth := float64(jsonPatchSize) / simulationTimeSeconds / 1024
	tsonBandwidth := float64(tsonPatchSize) / simulationTimeSeconds / 1024

	// 결과 저장
	result.TotalUpdates = updateCounter
	result.ValueChanges = valueChangeCounter
	result.TimestampOnlyChanges = timestampChangeCounter
	result.JsonPatchCount = len(jsonPatches)
	result.TsonPatchCount = len(tsonPatches)
	result.JsonPatchSize = jsonPatchSize
	result.TsonPatchSize = tsonPatchSize
	result.JsonProcessingTimeNs = jsonTotalProcessingTime
	result.TsonProcessingTimeNs = tsonTotalProcessingTime
	result.JsonBandwidthUsage = jsonBandwidth
	result.TsonBandwidthUsage = tsonBandwidth
	result.CumulativeTimeVec = timeVec
	result.CumulativeJsonVec = jsonPatchesVec
	result.CumulativeTsonVec = tsonPatchesVec

	// 최종 문서 저장
	if err := dataStorage.SaveFinalJSON(jsonDoc); err != nil {
		fmt.Printf("최종 JSON 저장 오류: %v\n", err)
	}
	if err := dataStorage.SaveFinalTSON(tsonDoc); err != nil {
		fmt.Printf("최종 TSON 저장 오류: %v\n", err)
	}

	// 모든 패치 저장
	if err := dataStorage.SaveAllJSONPatches(jsonPatches); err != nil {
		fmt.Printf("JSON 패치 저장 오류: %v\n", err)
	}
	if err := dataStorage.SaveAllTSONPatches(tsonPatches); err != nil {
		fmt.Printf("TSON 패치 저장 오류: %v\n", err)
	}

	fmt.Printf("\n시뮬레이션 완료. 총 업데이트 수: %d, 값 변경: %d (%.1f%%)\n",
		updateCounter, valueChangeCounter,
		float64(valueChangeCounter)*100/float64(updateCounter))

	return result, timingRecords
}

// 누적 패치 개수 벡터 생성
func generateCumulativeVectors(scenario string) ([]int64, []int, []int) {
	// 저장된 벡터 파일 읽기
	vectorsFilename := filepath.Join(outputDir, fmt.Sprintf("%s_vectors.csv", scenario))
	file, err := os.Open(vectorsFilename)
	if err != nil {
		fmt.Printf("벡터 파일 읽기 오류: %v\n", err)
		return []int64{}, []int{}, []int{}
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Printf("CSV 읽기 오류: %v\n", err)
		return []int64{}, []int{}, []int{}
	}

	// 헤더 건너뛰기
	records = records[1:]

	timeVec := make([]int64, len(records))
	jsonVec := make([]int, len(records))
	tsonVec := make([]int, len(records))

	for i, record := range records {
		timeVec[i], _ = strconv.ParseInt(record[0], 10, 64)
		jsonVec[i], _ = strconv.Atoi(record[1])
		tsonVec[i], _ = strconv.Atoi(record[2])
	}

	return timeVec, jsonVec, tsonVec
}

// 시뮬레이션 벡터 저장
func saveSimulationVectors(scenario string, timeVec []int64, jsonVec, tsonVec []int) {
	filename := filepath.Join(outputDir, fmt.Sprintf("%s_vectors.csv", scenario))
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("벡터 파일 생성 오류: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 헤더 작성
	writer.Write([]string{"TimeMs", "JsonPatches", "TsonPatches"})

	// 데이터 작성
	for i := 0; i < len(timeVec); i++ {
		writer.Write([]string{
			fmt.Sprintf("%d", timeVec[i]),
			fmt.Sprintf("%d", jsonVec[i]),
			fmt.Sprintf("%d", tsonVec[i]),
		})
	}
}

// JSON 문서 업데이트
func updateJsonDocument(doc JsonDoc, path string, value interface{}, timestamp int64) {
	pathParts := strings.Split(path, "/")
	current := doc

	// 마지막 부분까지 경로 탐색
	for i := 0; i < len(pathParts)-1; i++ {
		if pathParts[i] == "" {
			continue
		}

		if nextMap, ok := current[pathParts[i]].(map[string]interface{}); ok {
			current = nextMap
		} else {
			// 중간 경로가 없으면 생성
			current[pathParts[i]] = make(map[string]interface{})
			current = current[pathParts[i]].(map[string]interface{})
		}
	}

	// 마지막 부분 업데이트
	lastPart := pathParts[len(pathParts)-1]
	if lastPart != "" {
		if _, ok := current[lastPart].(map[string]interface{}); !ok {
			current[lastPart] = make(map[string]interface{})
		}

		valueMap := current[lastPart].(map[string]interface{})
		valueMap["value"] = value
		valueMap["timestamp"] = timestamp
	}
}

// TSON 문서 업데이트
func updateTsonDocument(doc TsonDoc, path string, value interface{}, timestamp int64) {
	pathParts := strings.Split(path, "/")
	current := doc

	// 마지막 부분까지 경로 탐색
	for i := 0; i < len(pathParts)-1; i++ {
		if pathParts[i] == "" {
			continue
		}

		if nextMap, ok := current[pathParts[i]].(map[string]interface{}); ok {
			current = nextMap
		} else {
			// 중간 경로가 없으면 생성
			current[pathParts[i]] = make(map[string]interface{})
			current = current[pathParts[i]].(map[string]interface{})
		}
	}

	// 마지막 부분 업데이트
	lastPart := pathParts[len(pathParts)-1]
	if lastPart != "" {
		// TSON에서는 값과 타임스탬프가 구분되어 저장됨
		current[lastPart] = map[string]interface{}{
			"value": value,
			"_ts":   timestamp, // 타임스탬프는 메타데이터로 저장
		}
	}
}

// TSON 문서에서 값 가져오기
func getTsonDocValue(doc TsonDoc, path string) interface{} {
	pathParts := strings.Split(path, "/")
	current := doc

	// 경로 탐색
	for i := 0; i < len(pathParts); i++ {
		if pathParts[i] == "" {
			continue
		}

		if i == len(pathParts)-1 {
			// 마지막 부분이면 값 반환
			if valueObj, ok := current[pathParts[i]].(map[string]interface{}); ok {
				return valueObj["value"]
			}
			return nil
		} else {
			// 중간 경로면 다음 맵으로 이동
			if nextMap, ok := current[pathParts[i]].(map[string]interface{}); ok {
				current = nextMap
			} else {
				return nil
			}
		}
	}

	return nil
}

// 값 동등 비교
func isEqual(v1, v2 interface{}) bool {
	// 타입이 다르면 다른 값
	if reflect.TypeOf(v1) != reflect.TypeOf(v2) {
		return false
	}

	// 타입별 비교
	switch val1 := v1.(type) {
	case float64:
		// 부동소수점 비교 (작은 차이는 무시)
		val2 := v2.(float64)
		return math.Abs(val1-val2) < 0.000001
	case int:
		return val1 == v2.(int)
	case bool:
		return val1 == v2.(bool)
	case string:
		return val1 == v2.(string)
	default:
		// 기타 타입은 직접 비교
		return reflect.DeepEqual(v1, v2)
	}
}

// 시나리오별 초기 차량 데이터 생성
func createInitialVehicleData(scenario string) *VehicleData {
	vehicle := &VehicleData{
		SensorsHighFreq:  make(map[string]*SensorData),
		SensorsMedFreq:   make(map[string]*SensorData),
		SensorsLowFreq:   make(map[string]*SensorData),
		ActuatorsHighVar: make(map[string]*SensorData),
		ActuatorsLowVar:  make(map[string]*SensorData),
		Attributes:       make(map[string]*SensorData),
	}

	// 시나리오별 구성
	switch scenario {
	case "urban_traffic":
		setupUrbanTrafficScenario(vehicle)
	case "highway_cruising":
		setupHighwayCruisingScenario(vehicle)
	case "parking_maneuver":
		setupParkingManeuverScenario(vehicle)
	case "traffic_jam":
		setupTrafficJamScenario(vehicle)
	case "battery_charging":
		setupBatteryChargingScenario(vehicle)
	default:
		setupDefaultScenario(vehicle)
	}

	return vehicle
}

// 기본 시나리오 설정
func setupDefaultScenario(vehicle *VehicleData) {
	setupUrbanTrafficScenario(vehicle)
}

// 차량 데이터를 JSON 문서로 변환 (value-timestamp 구조)
func vehicleDataToJsonDoc(vehicle *VehicleData) JsonDoc {
	doc := make(JsonDoc)

	// 모든 센서 그룹에 대해 처리
	for _, sensorGroup := range []map[string]*SensorData{
		vehicle.SensorsHighFreq,
		vehicle.SensorsMedFreq,
		vehicle.SensorsLowFreq,
		vehicle.ActuatorsHighVar,
		vehicle.ActuatorsLowVar,
		vehicle.Attributes,
	} {
		for path, sensor := range sensorGroup {
			updateJsonDocument(doc, path, sensor.Value, sensor.Timestamp)
		}
	}

	return doc
}

// 차량 데이터를 TSON 문서로 변환 (타임스탬프가 메타데이터)
func vehicleDataToTsonDoc(vehicle *VehicleData) TsonDoc {
	doc := make(TsonDoc)

	// 모든 센서 그룹에 대해 처리
	for _, sensorGroup := range []map[string]*SensorData{
		vehicle.SensorsHighFreq,
		vehicle.SensorsMedFreq,
		vehicle.SensorsLowFreq,
		vehicle.ActuatorsHighVar,
		vehicle.ActuatorsLowVar,
		vehicle.Attributes,
	} {
		for path, sensor := range sensorGroup {
			updateTsonDocument(doc, path, sensor.Value, sensor.Timestamp)
		}
	}

	return doc
}

// 새 값 계산 (패턴에 따라)
func calculateNewValue(sensor *SensorData, currentTimeMs int64, scenario string) interface{} {
	// 타입에 따라 다른 처리
	switch sensor.Type {
	case "number":
		// 값이 숫자가 아니면 안전하게 변환
		_, ok := sensor.Value.(float64)
		if !ok {
			// 값이 숫자가 아닌 경우 처리
			fmt.Printf("경고: 센서 %s의 타입은 'number'이지만 실제 값은 %T입니다. 0.0으로 초기화합니다.\n",
				sensor.Path, sensor.Value)
			sensor.Value = 0.0
			return 0.0
		}
		return calculateNewNumberValue(sensor, currentTimeMs, scenario)
	case "boolean":
		// 값이 불리언이 아니면 안전하게 변환
		_, ok := sensor.Value.(bool)
		if !ok {
			// 값이 불리언이 아닌 경우 처리
			fmt.Printf("경고: 센서 %s의 타입은 'boolean'이지만 실제 값은 %T입니다. false로 초기화합니다.\n",
				sensor.Path, sensor.Value)
			sensor.Value = false
			return false
		}
		return calculateNewBooleanValue(sensor, currentTimeMs, scenario)
	case "string":
		return calculateNewStringValue(sensor, currentTimeMs, scenario)
	default:
		// 알 수 없는 타입 처리
		fmt.Printf("경고: 센서 %s의 타입 '%s'는 지원되지 않습니다.\n", sensor.Path, sensor.Type)
		return sensor.Value
	}
}

// 숫자형 값 계산
func calculateNewNumberValue(sensor *SensorData, currentTimeMs int64, scenario string) interface{} {
	currValue := sensor.Value.(float64)

	// 시간 경과 추적용 카운터 증가
	sensor.ChangeCounter++

	// 패턴에 따라 값 계산
	switch sensor.ChangePattern {
	case "constant":
		return currValue

	case "constant_with_noise":
		baseline := sensor.ChangeParams["baseline"]
		noise := sensor.ChangeParams["noise"]
		return baseline + (rand.Float64()*2.0-1.0)*noise

	case "linear":
		slope := sensor.ChangeParams["slope"]
		min := sensor.ChangeParams["min"]
		max := sensor.ChangeParams["max"]

		// 선형 변화
		newValue := currValue + slope*float64(sensor.UpdateInterval)

		// 범위 제한
		if newValue < min {
			newValue = min
		} else if newValue > max {
			newValue = max
		}

		return newValue

	case "sinusoidal":
		amplitude := sensor.ChangeParams["amplitude"]
		period := sensor.ChangeParams["period"]
		baseline := sensor.ChangeParams["baseline"]
		phase := sensor.ChangeParams["phase"]

		// 사인 함수 계산
		timeInCycle := float64(currentTimeMs+int64(phase)) / period
		return baseline + amplitude*math.Sin(2.0*math.Pi*timeInCycle)

	case "random_walk":
		stepSize := sensor.ChangeParams["step_size"]
		min := sensor.ChangeParams["min"]
		max := sensor.ChangeParams["max"]

		// 랜덤 워크
		step := (rand.Float64()*2.0 - 1.0) * stepSize
		newValue := currValue + step

		// 범위 제한
		if newValue < min {
			newValue = min
		} else if newValue > max {
			newValue = max
		}

		return newValue

	case "stepped":
		stepSize := sensor.ChangeParams["step_size"]
		stepInterval := sensor.ChangeParams["step_interval"]
		min := sensor.ChangeParams["min"]
		max := sensor.ChangeParams["max"]

		// 일정 간격으로 계단식 변화
		if float64(sensor.ChangeCounter*sensor.UpdateInterval) >= stepInterval {
			// 계단식 증가 또는 감소
			direction := 1.0
			if rand.Float64() < 0.5 {
				direction = -1.0
			}

			newValue := currValue + direction*stepSize

			// 범위 제한
			if newValue < min {
				newValue = min
			} else if newValue > max {
				newValue = max
			}

			// 카운터 리셋
			sensor.ChangeCounter = 0

			return newValue
		}

		return currValue

	case "triggered":
		threshold := sensor.ChangeParams["threshold"]
		resetInterval := sensor.ChangeParams["reset_interval"]
		minValue := sensor.ChangeParams["min_value"]
		maxValue := sensor.ChangeParams["max_value"]

		// 특정 확률로 트리거되는 변화
		if rand.Float64() < threshold || float64(sensor.ChangeCounter*sensor.UpdateInterval) >= resetInterval {
			// 범위 내 랜덤 값으로 점프
			newValue := minValue + rand.Float64()*(maxValue-minValue)

			// 카운터 리셋
			sensor.ChangeCounter = 0

			return newValue
		}

		return currValue

	default:
		return currValue
	}
}

// 불리언 값 계산
func calculateNewBooleanValue(sensor *SensorData, currentTimeMs int64, scenario string) interface{} {
	currValue := sensor.Value.(bool)

	// 패턴에 따라 값 계산
	switch sensor.ChangePattern {
	case "constant":
		return currValue

	case "toggle":
		toggleProb := sensor.ChangeParams["toggle_probability"]

		// 특정 확률로 토글
		if rand.Float64() < toggleProb {
			return !currValue
		}

		return currValue

	case "triggered":
		threshold := sensor.ChangeParams["threshold"]

		// 특정 확률로 트리거
		if rand.Float64() < threshold {
			return !currValue
		}

		return currValue

	default:
		return currValue
	}
}

// 문자열 값 계산
func calculateNewStringValue(sensor *SensorData, currentTimeMs int64, scenario string) interface{} {
	// 타입 검사: 값이 이미 문자열인지 확인
	if currValue, ok := sensor.Value.(string); ok {
		// 패턴에 따라 값 계산
		switch sensor.ChangePattern {
		case "constant":
			return currValue

		case "random_selection":
			numChoices := int(sensor.ChangeParams["num_choices"])
			changeProb := sensor.ChangeParams["change_probability"]

			// 기본값 설정
			if changeProb == 0.0 {
				changeProb = 0.05
			}

			// 특정 확률로 변경
			if rand.Float64() < changeProb {
				prefix := ""

				// 경로에 따라 다른 접두사 사용
				if strings.Contains(sensor.Path, "Transmission") {
					prefix = "gear_"
				} else if strings.Contains(sensor.Path, "HVAC") {
					prefix = "setting_"
				} else if strings.Contains(sensor.Path, "ChargingStatus") {
					// 충전 상태 값
					options := []string{"charging", "full", "error", "pending"}
					return options[rand.Intn(len(options))]
				} else if strings.Contains(sensor.Path, "ChargingMode") {
					// 충전 모드 값
					options := []string{"fast", "normal", "eco", "scheduled"}
					return options[rand.Intn(len(options))]
				} else {
					prefix = "value_"
				}

				// 랜덤 값 선택 (현재 값과 다름)
				var newValue string
				for {
					newValue = fmt.Sprintf("%s%d", prefix, rand.Intn(numChoices))
					if newValue != currValue {
						break
					}
				}

				return newValue
			}

			return currValue

		default:
			return currValue
		}
	} else {
		// 값이 문자열이 아닌 경우 에러 처리 (안전한 기본값 반환)
		fmt.Printf("경고: 센서 %s의 타입은 'string'이지만 실제 값은 %T입니다. 기본 문자열 값으로 변환합니다.\n",
			sensor.Path, sensor.Value)

		// 경로에 따라 적절한 문자열 값 반환
		if strings.Contains(sensor.Path, "ChargingStatus") {
			return "charging"
		} else if strings.Contains(sensor.Path, "ChargingMode") {
			return "fast"
		} else {
			return fmt.Sprintf("value_%d", rand.Intn(10))
		}
	}
}
