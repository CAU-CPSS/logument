package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/CAU-CPSS/logument/internal/logument"
	"github.com/CAU-CPSS/logument/internal/tson"
	"github.com/CAU-CPSS/logument/internal/tsonpatch"
	"github.com/CAU-CPSS/logument/internal/vssgen"
)

const (
	// defaultDataset = "internal/vssgen/vss.json"
	carCount     = 1
	stateChanges = 600 // 10분간 600번의 상태 변경
	changeRate   = 0.2 // 실제 값이 변경되는 필드의 비율
	maintainRate = 0.8 // 값은 변경되지 않지만 타임스탬프가 업데이트되는 필드의 비율
)

func LoggingOverhead() {
	fmt.Println("Logument 성능 측정 실험 시작...")

	// 출력 디렉토리 설정
	err := os.MkdirAll("results", os.ModePerm)
	if err != nil {
		fmt.Printf("결과 디렉토리 생성 오류: %v\n", err)
		return
	}

	// 시뮬레이션 데이터 생성
	fmt.Println("자율주행 시나리오 시뮬레이션 데이터 생성...")
	generateDrivingScenario()

	// 실험 실행 및 성능 측정
	fmt.Println("성능 실험 실행...")
	runExperiments()

	fmt.Println("실험 완료. 결과는 'results' 디렉토리에 저장되었습니다.")
}

func generateDrivingScenario() {
	// VSS 초기 상태 생성
	outputDir := "./dataset"
	vssgen.PrepareOutputDir(outputDir)

	// 메타데이터 저장
	metadata := map[string]any{
		"dataset":       defaultDataset,
		"cars":          carCount,
		"files":         stateChanges + 1, // 초기 상태 + 변경사항
		"change_rate":   changeRate,
		"maintain_rate": maintainRate,
		"size":          0.5, // VSS 스키마의 50% 사용
	}
	vssgen.SaveMetadata(metadata, outputDir)

	// 초기 상태 생성 및 저장
	fmt.Println("초기 상태 생성...")
	vss := vssgen.NewVssJson(defaultDataset)
	initialState := vss.Generate(0.5, 1)

	carDir := filepath.Join(outputDir, fmt.Sprintf("car_%d/tson", 1))
	patchDir := filepath.Join(outputDir, fmt.Sprintf("car_%d/patches", 1))

	// 디렉토리 생성
	os.MkdirAll(carDir, os.ModePerm)
	os.MkdirAll(patchDir, os.ModePerm)

	// 초기 상태 저장
	initialState.Save(filepath.Join(carDir, "1_1.tson"))

	// 상태 변경 생성 및 저장
	fmt.Println("상태 변경 시뮬레이션 생성...")
	currentState := initialState

	for i := 2; i <= stateChanges+1; i++ {
		// 각 반복마다 새로운 상태와 패치 생성
		// 수정된 GenerateNext 함수 사용 (changeRate, maintainRate)
		nextState, patch := currentState.GenerateNext(changeRate, maintainRate, 1, i)

		// 새로운 상태와 패치 저장
		nextState.Save(filepath.Join(carDir, fmt.Sprintf("1_%d.tson", i)))
		patch.Save(filepath.Join(patchDir, fmt.Sprintf("1_%d.json", i)))

		// 다음 반복을 위해 현재 상태 업데이트
		currentState = nextState
	}

	fmt.Printf("시뮬레이션 데이터가 %s에 생성되었습니다\n", outputDir)
}

// 성능 실험 실행
func runExperiments() {
	// 1. 시뮬레이션 데이터 로드
	initialData, err := os.ReadFile("dataset/car_1/tson/1_1.tson")
	if err != nil {
		fmt.Printf("초기 상태 로드 오류: %v\n", err)
		return
	}

	fmt.Println("초기 상태 데이터 로드 완료")

	// 초기 상태 파싱
	var initialTson tson.Tson
	err = tson.Unmarshal(initialData, &initialTson)
	if err != nil {
		fmt.Printf("초기 상태 파싱 오류: %v\n", err)
		return
	}

	// 모든 패치 로드
	fmt.Println("패치 데이터 로드 중...")
	patches := make([]tsonpatch.Patch, stateChanges)
	for i := 1; i <= stateChanges; i++ {
		patchFile := fmt.Sprintf("dataset/car_1/patches/1_%d.json", i+1)
		patchData, err := os.ReadFile(patchFile)
		if err != nil {
			fmt.Printf("패치 %d 로드 오류: %v\n", i, err)
			return
		}

		err = json.Unmarshal(patchData, &patches[i-1])
		if err != nil {
			fmt.Printf("패치 %d 파싱 오류: %v\n", i, err)
			return
		}
	}

	fmt.Println("모든 패치 데이터 로드 완료")

	// 2. 저장 오버헤드 측정
	fmt.Println("로깅 오버헤드 측정 중...")
	snapshotSize := measureSnapshotSize(initialTson, patches)
	patchOnlySize := measurePatchOnlySize(initialTson, patches)
	logumentSize := measureLogumentSize(initialTson, patches)

	fmt.Printf("저장 오버헤드 (바이트): Snapshot=%d Byte, PatchOnly=%d Byte, Logument=%d Byte\n",
		snapshotSize, patchOnlySize, logumentSize)

	// 3. 세 가지 방식의 작업 시간 측정
	fmt.Println("각 방식의 상태 변경 시간 측정 중...")
	snapshotTimes := measureSnapshotTimes(initialTson, patches)
	patchOnlyTimes := measurePatchOnlyTimes(initialTson, patches)
	logumentTimes := measureLogumentTimes(initialTson, patches)

	// 4. Logument 특정 작업 시간 측정
	fmt.Println("Logument 작업별 시간 측정 중...")
	storeTimes := measureStoreTimes(initialTson, patches)
	appendTimes := measureAppendTimes(initialTson, patches)
	trackTimes := measureTrackTimes(initialTson, patches)
	temporalTrackTimes := measureTemporalTrackTimes(initialTson, patches)

	// 5. 결과를 CSV로 저장
	fmt.Println("결과를 CSV로 저장 중...")
	writeOverheadCSV(snapshotSize, patchOnlySize, logumentSize)
	writeApproachTimesCSV(snapshotTimes, patchOnlyTimes, logumentTimes)
	writeLogumentOperationTimesCSV(storeTimes, appendTimes, trackTimes, temporalTrackTimes)
}

// 스냅샷 방식의 오버헤드 측정
func measureSnapshotSize(initialState tson.Tson, patches []tsonpatch.Patch) int64 {
	// 초기 상태 크기
	initialBytes, _ := tson.MarshalIndent(initialState, "", "  ")
	totalSize := int64(len(initialBytes))

	// 각 스냅샷의 크기 계산
	currentState := initialState
	for _, patch := range patches {
		// 현재 시점에서 변경된 모든 패치를 하나의 snapshot으로 적용
		newState, _ := tsonpatch.ApplyPatch(currentState, patch)
		currentState = newState

		// 이 스냅샷의 크기 계산
		stateBytes, _ := tson.MarshalIndent(currentState, "", "  ")
		totalSize += int64(len(stateBytes))

		// 패치를 하나씩 적용하여 다음 상태 얻기
		// for _, op := range patch {
		// 	newState, _ := tsonpatch.ApplyOperation(currentState, op)
		// 	currentState = newState

		//  이 스냅샷의 크기 계산
		// 	stateBytes, _ := tson.MarshalIndent(currentState, "", "  ")
		// 	totalSize += int64(len(stateBytes))
		// }
	}

	return totalSize
}

// 패치만 사용하는 방식의 오버헤드 측정
func measurePatchOnlySize(initialState tson.Tson, patches []tsonpatch.Patch) int64 {
	// 초기 상태 크기
	initialBytes, _ := tson.MarshalIndent(initialState, "", "  ")
	totalSize := int64(len(initialBytes))

	// // 각 스냅샷의 크기 계산
	// currentState := initialState
	// for _, patch := range patches {
	// 	// 현재 시점에서 변경된 모든 패치를 하나의 snapshot으로 적용
	// 	newState, _ := tsonpatch.ApplyPatch(currentState, patch)
	// 	currentState = newState

	// 	// 이 스냅샷의 크기 계산
	// 	stateBytes, _ := tson.MarshalIndent(currentState, "", "  ")
	// 	totalSize += int64(len(stateBytes))
	// }

	// 각 패치의 크기 추가
	for _, patch := range patches {
		patchBytes, _ := json.MarshalIndent(patch, "", "  ")
		totalSize += int64(len(patchBytes))
	}

	return totalSize
}

// Logument 방식의 오버헤드 측정
func measureLogumentSize(initialState tson.Tson, patches []tsonpatch.Patch) int64 {
	// Logument 초기화
	lgm := logument.NewLogument(initialState, nil)

	totalOps := 0    // 원본 작업 수
	filteredOps := 0 // 필터링 후 작업 수

	for i, patch := range patches {
		totalOps += len(patch)

		// 적용 전 패치 사이즈 확인
		fmt.Printf("패치 %d: 적용 전 %d 작업\n", i+1, len(patch))

		// TestSet 적용
		for _, p := range patch {
			lgm.TestSet(uint64(len(lgm.Version)), p)
		}
		lgm.Append() // 여기서 version이 증가함
		lgm.Snapshot(uint64(len(lgm.Version) - 1))

		// 적용 후 패치 사이즈 확인
		currentVersion := lgm.Version[len(lgm.Version)-1]
		if versionPatch, exists := lgm.Patches[currentVersion]; exists {
			filteredOps += len(versionPatch)
			fmt.Printf("패치 %d: 필터링 후 %d 작업 (제거율: %.1f%%)\n",
				i+1, len(versionPatch),
				100.0*(1.0-float64(len(versionPatch))/float64(len(patch))))
		}

		lgm.Append()
	}

	fmt.Printf("총 필터링: %d → %d 작업 (제거율: %.1f%%)\n",
		totalOps, filteredOps,
		100.0*(1.0-float64(filteredOps)/float64(totalOps)))

	// 저장 크기 계산
	var totalSize int64 = 0

	// 초기 상태 크기
	initialBytes, _ := tson.MarshalIndent(initialState, "", "  ")
	totalSize += int64(len(initialBytes))

	// 각 패치 버전의 크기 계산
	for version := range lgm.Patches {
		patchBytes, _ := json.MarshalIndent(lgm.Patches[version], "", "  ")
		totalSize += int64(len(patchBytes))
	}

	return totalSize
}

// 스냅샷 방식의 시간 측정
func measureSnapshotTimes(initialState tson.Tson, patches []tsonpatch.Patch) []int64 {
	times := make([]int64, len(patches))
	currentState := initialState

	for i, patch := range patches {
		runtime.GC()
		startTime := time.Now().UnixNano()

		// 현재 상태에 패치 적용
		newState, _ := tsonpatch.ApplyPatch(currentState, patch)
		currentState = newState

		// 스냅샷 방식에서는 전체 상태를 직렬화
		_, _ = tson.MarshalIndent(currentState, "", "  ")

		endTime := time.Now().UnixNano()
		times[i] = endTime - startTime
	}

	return times
}

// 패치만 사용하는 방식의 시간 측정
func measurePatchOnlyTimes(initialState tson.Tson, patches []tsonpatch.Patch) []int64 {
	times := make([]int64, len(patches))
	currentState := initialState

	for i, patch := range patches {
		runtime.GC()
		startTime := time.Now().UnixNano()

		// 현재 상태에 패치 적용
		newState, _ := tsonpatch.ApplyPatch(currentState, patch)
		currentState = newState

		// 패치만 직렬화
		_, _ = json.MarshalIndent(patch, "", "  ")

		endTime := time.Now().UnixNano()
		times[i] = endTime - startTime
	}

	return times
}

// Logument 방식의 시간 측정
func measureLogumentTimes(initialState tson.Tson, patches []tsonpatch.Patch) []int64 {
	times := make([]int64, len(patches))

	// Logument 초기화
	lgm := logument.NewLogument(initialState, nil)

	for i, patch := range patches {
		runtime.GC()
		startTime := time.Now().UnixNano()

		// 각 패치 작업에 대해 TestSet 적용
		for _, p := range patch {
			lgm.TestSet(uint64(len(lgm.Version)), p)
		}
		lgm.Append() // 여기서 version이 증가함

		endTime := time.Now().UnixNano()
		times[i] = endTime - startTime
	}

	return times
}

// Store 작업 시간 측정
func measureStoreTimes(initialState tson.Tson, patches []tsonpatch.Patch) []int64 {
	times := make([]int64, len(patches))

	// Logument 초기화
	lgm := logument.NewLogument(initialState, nil)

	for i, patch := range patches {
		startTime := time.Now().UnixNano()
		lgm.Store(patch)
		endTime := time.Now().UnixNano()

		times[i] = endTime - startTime

		// 다음 반복을 위해 상태 유지
		lgm.Append()
	}

	return times
}

// Append 작업 시간 측정
func measureAppendTimes(initialState tson.Tson, patches []tsonpatch.Patch) []int64 {
	times := make([]int64, len(patches))

	// Logument 초기화
	lgm := logument.NewLogument(initialState, nil)

	for i, patch := range patches {
		// 시간 측정 없이 패치 저장
		lgm.Store(patch)

		// Append 작업 시간만 측정
		startTime := time.Now().UnixNano()
		lgm.Append()
		endTime := time.Now().UnixNano()

		times[i] = endTime - startTime
	}

	return times
}

// Track 작업 시간 측정
func measureTrackTimes(initialState tson.Tson, patches []tsonpatch.Patch) []int64 {
	times := make([]int64, len(patches))

	// Logument 초기화 및 모든 패치 적용
	lgm := logument.NewLogument(initialState, nil)
	for _, patch := range patches {
		lgm.Store(patch)
		lgm.Append()
	}

	// 다양한 범위로 Track 작업 시간 측정
	for i := 0; i < len(patches); i++ {
		startVersion := uint64(0)
		endVersion := uint64(i + 1)

		startTime := time.Now().UnixNano()
		lgm.Track(startVersion, endVersion)
		endTime := time.Now().UnixNano()

		times[i] = endTime - startTime
	}

	return times
}

// TemporalTrack 작업 시간 측정
func measureTemporalTrackTimes(initialState tson.Tson, patches []tsonpatch.Patch) []int64 {
	times := make([]int64, len(patches))

	// Logument 초기화 및 모든 패치 적용
	lgm := logument.NewLogument(initialState, nil)

	// 타임스탬프 수집
	timestamps := make([]int64, len(patches))
	for i, patch := range patches {
		// 패치의 첫 작업에서 타임스탬프 추출
		if len(patch) > 0 {
			timestamps[i] = patch[0].Timestamp
		} else {
			// 패치에 작업이 없으면 현재 시간 사용
			timestamps[i] = time.Now().Unix()
		}

		lgm.Store(patch)
		lgm.Append()
	}

	// 다양한 범위로 TemporalTrack 작업 시간 측정
	for i := 0; i < len(patches); i++ {
		startTime := time.Now().UnixNano()
		lgm.TemporalTrack(timestamps[0], timestamps[i])
		endTime := time.Now().UnixNano()

		times[i] = endTime - startTime
	}

	return times
}

// 오버헤드 데이터를 CSV로 저장
func writeOverheadCSV(snapshotSize, patchOnlySize, logumentSize int64) {
	csvFile, err := os.Create("results/overhead_data.csv")
	if err != nil {
		fmt.Printf("오버헤드 CSV 생성 오류: %v\n", err)
		return
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// 헤더 작성
	csvWriter.Write([]string{"Approach", "Total_Size_Bytes"})

	// 데이터 행 작성
	csvWriter.Write([]string{"TSON_Snapshot", fmt.Sprintf("%d", snapshotSize)})
	csvWriter.Write([]string{"TSON_Patch_Only", fmt.Sprintf("%d", patchOnlySize)})
	csvWriter.Write([]string{"Logument", fmt.Sprintf("%d", logumentSize)})
}

// 각 방식의 시간을 CSV로 저장
func writeApproachTimesCSV(snapshotTimes, patchOnlyTimes, logumentTimes []int64) {
	csvFile, err := os.Create("results/approach_times.csv")
	if err != nil {
		fmt.Printf("방식별 시간 CSV 생성 오류: %v\n", err)
		return
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// 헤더 작성
	csvWriter.Write([]string{"Change", "Snapshot_Time_ns", "PatchOnly_Time_ns", "Logument_Time_ns"})

	// 데이터 행 작성
	for i := 0; i < len(snapshotTimes); i++ {
		csvWriter.Write([]string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%d", snapshotTimes[i]),
			fmt.Sprintf("%d", patchOnlyTimes[i]),
			fmt.Sprintf("%d", logumentTimes[i]),
		})
	}
}

// Logument 작업 시간을 CSV로 저장
func writeLogumentOperationTimesCSV(storeTimes, appendTimes, trackTimes, temporalTrackTimes []int64) {
	csvFile, err := os.Create("results/logument_operation_times.csv")
	if err != nil {
		fmt.Printf("작업 시간 CSV 생성 오류: %v\n", err)
		return
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// 헤더 작성
	csvWriter.Write([]string{"Change", "Store_Time_ns", "Append_Time_ns", "Track_Time_ns", "TemporalTrack_Time_ns"})

	// 데이터 행 작성
	for i := 0; i < len(storeTimes); i++ {
		csvWriter.Write([]string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%d", storeTimes[i]),
			fmt.Sprintf("%d", appendTimes[i]),
			fmt.Sprintf("%d", trackTimes[i]),
			fmt.Sprintf("%d", temporalTrackTimes[i]),
		})
	}
}
