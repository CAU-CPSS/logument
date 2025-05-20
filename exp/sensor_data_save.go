package exp

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// KeySensorPaths는 그래프 시각화를 위해 추적할 주요 센서 경로 목록입니다
var KeySensorPaths = []string{
	"Vehicle.CurrentLocation.Latitude",
	"Vehicle.CurrentLocation.Longitude",
	"Vehicle.Speed",
	"Vehicle.Powertrain.TractionBattery.StateOfCharge.Current",
}

// SensorDataRecorder는 주요 센서 데이터를 CSV로 기록합니다
type SensorDataRecorder struct {
	OutputDir  string                 // 출력 디렉토리
	Scenario   string                 // 시나리오 이름
	CSVFiles   map[string]*os.File    // 센서별 파일
	CSVWriters map[string]*csv.Writer // 센서별 CSV 작성자
	StartTime  time.Time              // 실험 시작 시간
}

// NewSensorDataRecorder는 새 SensorDataRecorder 인스턴스를 생성합니다
func NewSensorDataRecorder(baseDir, scenario string) (*SensorDataRecorder, error) {
	// 시나리오별 디렉토리 생성
	sensorDataDir := filepath.Join(baseDir, scenario, "sensor_data")
	if err := os.MkdirAll(sensorDataDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("센서 데이터 디렉토리 생성 실패: %v", err)
	}

	recorder := &SensorDataRecorder{
		OutputDir:  sensorDataDir,
		Scenario:   scenario,
		CSVFiles:   make(map[string]*os.File),
		CSVWriters: make(map[string]*csv.Writer),
		StartTime:  time.Now(),
	}

	// 주요 센서마다 CSV 파일 생성
	for _, path := range KeySensorPaths {
		// 파일명에 사용할 경로의 마지막 부분 추출
		shortName := filepath.Base(path)
		fileName := filepath.Join(sensorDataDir, shortName+".csv")

		file, err := os.Create(fileName)
		if err != nil {
			// 에러 발생 시 열린 모든 파일 닫기
			recorder.Close()
			return nil, fmt.Errorf("CSV 파일 생성 실패 %s: %v", fileName, err)
		}

		recorder.CSVFiles[path] = file
		writer := csv.NewWriter(file)
		recorder.CSVWriters[path] = writer

		// 헤더 쓰기
		writer.Write([]string{"TimestampMs", "SimulationTimeMs", "Value"})
		writer.Flush()
	}

	return recorder, nil
}

// RecordSensorData는 센서 데이터 포인트를 기록합니다
func (r *SensorDataRecorder) RecordSensorData(simulationTimeMs int64, vehicle *VehicleData) error {
	// 각 주요 센서에 대해
	for _, path := range KeySensorPaths {
		// 센서 값 추출
		var sensorValue interface{}
		var found bool

		// 센서가 속한 카테고리 확인 및 값 추출
		if sensor, ok := vehicle.SensorsHighFreq[path]; ok {
			sensorValue = sensor.Value
			found = true
		} else if sensor, ok := vehicle.SensorsMedFreq[path]; ok {
			sensorValue = sensor.Value
			found = true
		} else if sensor, ok := vehicle.SensorsLowFreq[path]; ok {
			sensorValue = sensor.Value
			found = true
		} else if sensor, ok := vehicle.ActuatorsHighVar[path]; ok {
			sensorValue = sensor.Value
			found = true
		} else if sensor, ok := vehicle.ActuatorsLowVar[path]; ok {
			sensorValue = sensor.Value
			found = true
		} else if sensor, ok := vehicle.Attributes[path]; ok {
			sensorValue = sensor.Value
			found = true
		}

		if !found {
			continue // 센서를 찾지 못함
		}

		// CSV 기록
		if writer, ok := r.CSVWriters[path]; ok {
			// 타임스탬프, 시뮬레이션 시간, 값
			now := time.Now()
			elapsedMs := now.Sub(r.StartTime).Milliseconds()

			writer.Write([]string{
				fmt.Sprintf("%d", elapsedMs),
				fmt.Sprintf("%d", simulationTimeMs),
				fmt.Sprintf("%v", sensorValue),
			})

			// 주기적으로 버퍼 비우기
			if simulationTimeMs%1000 == 0 {
				writer.Flush()
			}
		}
	}

	return nil
}

// Close는 모든 CSV 파일을 닫습니다
func (r *SensorDataRecorder) Close() {
	// 모든 CSV 버퍼 비우기
	for _, writer := range r.CSVWriters {
		writer.Flush()
	}

	// 모든 파일 닫기
	for _, file := range r.CSVFiles {
		file.Close()
	}

	fmt.Printf("센서 데이터 기록 완료: %s\n", r.OutputDir)
}
