package main

import (
	"math/rand"

	"github.com/CAU-CPSS/logument/exp"
)

const (
	defaultDataset   = "internal/vssgen/vss.json"
	defaultCarCount  = 5
	defaultFileCount = 10
)

const (
	// 실험 파라미터
	carCount     = 1
	stateChanges = 600 // 10분간 600번의 상태 변경

	changeRate   = 0.2 // 실제 값이 변경되는 필드의 비율
	maintainRate = 0.0 // 값은 변경되지 않지만 타임스탬프가 업데이트되는 필드의 비율

	vssDatasetSize = 0.2 // VSS 데이터셋 크기 비율 (20%)
	numUpdates     = 100 // 업데이트 횟수
)

const (
	datasetPath     = "./dataset_timestamp_experiment"
	initialJsonPath = "./dataset_timestamp_experiment/initial_json.json"
	initialTsonPath = "./dataset_timestamp_experiment/initial_tson.tson"
	resultsCsvPath  = "./dataset_timestamp_experiment/results.csv"
	graphDataPath   = "./dataset_timestamp_experiment/graph_data.csv"
)

func main() {
	rand.New(rand.NewSource(42))
	// ScenarioExperiment()
	TemporalQueryExperiment()
}

func ScenarioExperiment() {
	// 실험 실행
	exp.RealworldScenario()
}

func TemporalQueryExperiment() {
	exp.RunTemporalQueryExperiments()
}
