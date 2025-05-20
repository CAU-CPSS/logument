package exp

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

//=============================================================================
// 상수 및 전역 변수
//=============================================================================

// 시뮬레이션 파라미터
const (
	outputDir             = "./results/patch_experiment"
	simulationDurationSec = 300 // 5분 시뮬레이션
	simulationTimeStepMs  = 100 // 100ms 단위로 업데이트
	numHighFreqSensors    = 10  // 고빈도 센서 수
	numMedFreqSensors     = 15  // 중빈도 센서 수
	numLowFreqSensors     = 20  // 저빈도 센서 수
	numHighVarActuators   = 5   // 높은 변동성 액추에이터 수
	numLowVarActuators    = 10  // 낮은 변동성 액추에이터 수
	numAttributes         = 8   // 속성 수
)

// 시나리오 종류
var scenarios = []string{
	"urban_traffic",    // 도심 교통
	"highway_cruising", // 고속도로 주행
	"battery_charging", // 배터리 충전
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

//=============================================================================
// 메인 실험 함수
//=============================================================================

// RealworldScenario는 실제 자동차 시나리오 기반 실험을 수행합니다
func RealworldScenario() {
	// 출력 디렉토리 생성
	os.MkdirAll(outputDir, os.ModePerm)

	// 모든 시나리오에 대한 실험 수행
	for _, scenario := range scenarios {
		fmt.Printf("\n=== 시나리오: %s 실험 시작 ===\n", scenario)

		// 시나리오별 결과 디렉토리 생성
		scenarioDir := filepath.Join(outputDir, scenario)
		os.MkdirAll(scenarioDir, os.ModePerm)

		// 결과 파일 생성
		resultsFile, err := os.Create(filepath.Join(scenarioDir, "results.csv"))
		if err != nil {
			fmt.Printf("%s 결과 파일 생성 오류: %v\n", scenario, err)
			continue
		}
		defer resultsFile.Close()

		csvWriter := csv.NewWriter(resultsFile)
		defer csvWriter.Flush()

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
		cumulativeFile, err := os.Create(filepath.Join(scenarioDir, "cumulative.csv"))
		if err != nil {
			fmt.Printf("%s 누적 데이터 파일 생성 오류: %v\n", scenario, err)
			continue
		}
		defer cumulativeFile.Close()

		cumWriter := csv.NewWriter(cumulativeFile)
		defer cumWriter.Flush()

		// 누적 데이터 헤더
		cumWriter.Write([]string{
			"TimeMs",
			"JsonPatchCount",
			"TsonPatchCount",
		})

		// 처리 시간 측정용 파일
		timingFile, err := os.Create(filepath.Join(scenarioDir, "timing.csv"))
		if err != nil {
			fmt.Printf("%s 처리 시간 파일 생성 오류: %v\n", scenario, err)
			continue
		}
		defer timingFile.Close()

		timingWriter := csv.NewWriter(timingFile)
		defer timingWriter.Flush()

		// 처리 시간 헤더
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

		// 누적 데이터 저장
		timeVector, jsonVector, tsonVector := results.CumulativeTimeVec, results.CumulativeJsonVec, results.CumulativeTsonVec
		for i := 0; i < len(timeVector); i++ {
			cumWriter.Write([]string{
				fmt.Sprintf("%d", timeVector[i]),
				fmt.Sprintf("%d", jsonVector[i]),
				fmt.Sprintf("%d", tsonVector[i]),
			})
		}

		fmt.Printf("\n%s 시나리오 실험 완료. 결과 파일 저장됨:\n", scenario)
		fmt.Printf("- 요약 결과: %s\n", filepath.Join(scenarioDir, "results.csv"))
		fmt.Printf("- 누적 패치: %s\n", filepath.Join(scenarioDir, "cumulative.csv"))
		fmt.Printf("- 처리 시간: %s\n", filepath.Join(scenarioDir, "timing.csv"))
	}

	fmt.Println("\n모든 실험이 완료되었습니다.")
}

// runRealisticExperiment는 지정된 시나리오에 대한 실험을 수행합니다
func runRealisticExperiment(scenario string, timingWriter *csv.Writer) (ExperimentResult, []TimingRecord) {
	fmt.Printf("시나리오 '%s' 시뮬레이션 시작...\n", scenario)

	// 이 시나리오에 대한 데이터 저장소 인스턴스 생성
	dataStorage := NewDataStorage(outputDir, scenario)

	// 센서 데이터 기록기 생성
	sensorRecorder, err := NewSensorDataRecorder(outputDir, scenario)
	if err != nil {
		fmt.Printf("센서 데이터 기록기 생성 실패: %v\n", err)
	} else {
		defer sensorRecorder.Close()
	}

	// 1. VSS 기반 초기 데이터 생성
	vehicle := InitializeVehicleData(scenario)

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
	timeVec := make([]int64, 0, simulationDurationSec*10) // 10초 간격으로 저장
	jsonPatchesVec := make([]int, 0, simulationDurationSec*10)
	tsonPatchesVec := make([]int, 0, simulationDurationSec*10)

	// 처리 시간 기록 배열
	timingRecords := make([]TimingRecord, 0)

	// 대략적인 총 업데이트 횟수 예측
	totalExpectedUpdates := calculateExpectedUpdateCount(vehicle, simulationDurationSec)
	fmt.Printf("예상 업데이트 수: 약 %d\n", totalExpectedUpdates)

	// 4. 시뮬레이션 실행
	simulationStart := time.Now()
	jsonTotalProcessingTime := int64(0)
	tsonTotalProcessingTime := int64(0)

	// 진행 상황 표시용 카운터
	updateCounter := 0
	valueChangeCounter := 0
	timestampChangeCounter := 0

	// 시뮬레이션 시간 스텝 (ms)
	for currentTimeMs := int64(0); currentTimeMs < simulationDurationSec*1000; currentTimeMs += simulationTimeStepMs {
		// 타임스탬프 벡터 업데이트 (10초 간격)
		if currentTimeMs%(10*1000) == 0 {
			timeVec = append(timeVec, currentTimeMs)
			jsonPatchesVec = append(jsonPatchesVec, len(jsonPatches))
			tsonPatchesVec = append(tsonPatchesVec, len(tsonPatches))
		}

		// 진행 상황 표시 (10% 간격)
		if currentTimeMs%(simulationDurationSec*100) == 0 {
			progressPct := float64(currentTimeMs) / float64(simulationDurationSec*1000) * 100
			fmt.Printf("\r시뮬레이션 진행: %.1f%% (시간: %d ms, 업데이트: %d, 변경: %d)",
				progressPct, currentTimeMs, updateCounter, valueChangeCounter)
		}

		// 주요 센서 데이터 기록
		if sensorRecorder != nil {
			sensorRecorder.RecordSensorData(currentTimeMs, vehicle)
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
					updateJsonDoc(jsonDoc, path, newValue, newTimestamp)

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
						updateTsonDoc(tsonDoc, path, newValue, newTimestamp)
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

	// 줄바꿈 (진행률 표시 후)
	fmt.Println()

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

	// 시뮬레이션 종료 시:
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

	fmt.Printf("시뮬레이션 완료. 총 업데이트 수: %d, 값 변경: %d (%.1f%%)\n",
		updateCounter, valueChangeCounter,
		float64(valueChangeCounter)*100/float64(updateCounter))

	return result, timingRecords
}
