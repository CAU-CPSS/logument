package exp

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
)

//=============================================================================
// 유틸리티 함수
//=============================================================================

// calculateExpectedUpdateCount는 예상 업데이트 횟수를 계산합니다
func calculateExpectedUpdateCount(vehicle *VehicleData, durationSec int) int {
	totalExpectedUpdates := 0

	// 각 센서 그룹의 업데이트 횟수 예측
	for _, sensor := range vehicle.SensorsHighFreq {
		totalExpectedUpdates += durationSec * 1000 / sensor.UpdateInterval
	}
	for _, sensor := range vehicle.SensorsMedFreq {
		totalExpectedUpdates += durationSec * 1000 / sensor.UpdateInterval
	}
	for _, sensor := range vehicle.SensorsLowFreq {
		totalExpectedUpdates += durationSec * 1000 / sensor.UpdateInterval
	}
	for _, actuator := range vehicle.ActuatorsHighVar {
		totalExpectedUpdates += durationSec * 1000 / actuator.UpdateInterval
	}
	for _, actuator := range vehicle.ActuatorsLowVar {
		totalExpectedUpdates += durationSec * 1000 / actuator.UpdateInterval
	}

	return totalExpectedUpdates
}

// updateJsonDoc은 JSON 문서에 센서 값과 타임스탬프를 업데이트합니다
func updateJsonDoc(doc JsonDoc, path string, value any, timestamp int64) {
	pathParts := strings.Split(path, ".")
	current := doc

	// 마지막 부분까지 경로 탐색
	for i := range len(pathParts)-1 {
		if nextMap, ok := current[pathParts[i]].(map[string]any); ok {
			current = nextMap
		} else {
			// 중간 경로가 없으면 생성
			current[pathParts[i]] = make(map[string]any)
			current = current[pathParts[i]].(map[string]any)
		}
	}

	// 마지막 부분 업데이트
	lastPart := pathParts[len(pathParts)-1]
	if valueMap, ok := current[lastPart].(map[string]any); ok {
		valueMap["value"] = value
		valueMap["timestamp"] = timestamp
	} else {
		current[lastPart] = map[string]any{
			"value":     value,
			"timestamp": timestamp,
		}
	}
}

// updateTsonDoc은 TSON 문서에 센서 값과 타임스탬프를 업데이트합니다
func updateTsonDoc(doc TsonDoc, path string, value any, timestamp int64) {
	pathParts := strings.Split(path, ".")
	current := doc

	// 마지막 부분까지 경로 탐색
	for i := range len(pathParts)-1 {
		if nextMap, ok := current[pathParts[i]].(map[string]any); ok {
			current = nextMap
		} else {
			// 중간 경로가 없으면 생성
			current[pathParts[i]] = make(map[string]any)
			current = current[pathParts[i]].(map[string]any)
		}
	}

	// 마지막 부분 업데이트
	lastPart := pathParts[len(pathParts)-1]
	// TSON에서는 값과 타임스탬프가 구분되어 저장됨
	current[lastPart] = map[string]any{
		"value": value,
		"_ts":   timestamp, // 타임스탬프는 메타데이터로 저장
	}
}

// isEqual은 두 값이 동일한지 비교합니다
func isEqual(v1, v2 any) bool {
	// 타입이 다르면 다른 값
	if reflect.TypeOf(v1) != reflect.TypeOf(v2) {
		return false
	}

	// 타입별 비교
	switch val1 := v1.(type) {
	case float64:
		// 부동소수점 비교 (작은 차이는 무시)
		val2 := v2.(float64)
		return math.Abs(val1-val2) < 0.01
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

// vehicleDataToJsonDoc은 VehicleData를 JSON 문서로 변환합니다
func vehicleDataToJsonDoc(vehicle *VehicleData) map[string]any {
	doc := make(map[string]any)

	// 센서 그룹 반복
	for _, sensorMap := range []map[string]*SensorData{
		vehicle.SensorsHighFreq,
		vehicle.SensorsMedFreq,
		vehicle.SensorsLowFreq,
		vehicle.ActuatorsHighVar,
		vehicle.ActuatorsLowVar,
		vehicle.Attributes,
	} {
		// 각 센서 처리
		for path, sensor := range sensorMap {
			addPathToJsonDoc(doc, path, sensor.Value, sensor.Timestamp)
		}
	}

	return doc
}

// vehicleDataToTsonDoc은 VehicleData를 TSON 문서로 변환합니다
func vehicleDataToTsonDoc(vehicle *VehicleData) map[string]any {
	doc := make(map[string]any)

	// 센서 그룹 반복
	for _, sensorMap := range []map[string]*SensorData{
		vehicle.SensorsHighFreq,
		vehicle.SensorsMedFreq,
		vehicle.SensorsLowFreq,
		vehicle.ActuatorsHighVar,
		vehicle.ActuatorsLowVar,
		vehicle.Attributes,
	} {
		// 각 센서 처리
		for path, sensor := range sensorMap {
			addPathToTsonDoc(doc, path, sensor.Value, sensor.Timestamp)
		}
	}

	return doc
}

// addPathToJsonDoc은 JSON 문서에 경로를 추가합니다 (JSON 형식)
func addPathToJsonDoc(doc map[string]any, path string, value any, timestamp int64) {
	pathParts := strings.Split(path, ".")
	current := doc

	// 마지막 노드 전까지 경로 탐색
	for i := range len(pathParts)-1 {
		part := pathParts[i]

		// 중간 노드가 없으면 생성
		if _, exists := current[part]; !exists {
			current[part] = make(map[string]any)
		}

		// 다음 레벨로 이동
		current = current[part].(map[string]any)
	}

	// 마지막 노드 추가
	lastPart := pathParts[len(pathParts)-1]
	valueObj := map[string]any{
		"value":     value,
		"timestamp": timestamp,
	}
	current[lastPart] = valueObj
}

// addPathToTsonDoc은 TSON 문서에 경로를 추가합니다 (TSON 형식)
func addPathToTsonDoc(doc map[string]any, path string, value any, timestamp int64) {
	pathParts := strings.Split(path, ".")
	current := doc

	// 마지막 노드 전까지 경로 탐색
	for i := range len(pathParts)-1 {
		part := pathParts[i]

		// 중간 노드가 없으면 생성
		if _, exists := current[part]; !exists {
			current[part] = make(map[string]any)
		}

		// 다음 레벨로 이동
		current = current[part].(map[string]any)
	}

	// 마지막 노드 추가
	lastPart := pathParts[len(pathParts)-1]

	// TSON에서는 값과 메타데이터 타임스탬프 분리
	valueObj := map[string]any{
		"value": value,
		"_ts":   timestamp, // 메타데이터로 타임스탬프 저장
	}
	current[lastPart] = valueObj
}

//=============================================================================
// 센서 값 계산 함수
//=============================================================================

// calculateNewValue는 센서 패턴에 따라 새 값을 계산합니다
func calculateNewValue(sensor *SensorData, currentTimeMs int64, scenario string) any {
	// 센서 카운터 증가
	sensor.ChangeCounter++

	// 센서 타입에 따른 처리
	switch sensor.Type {
	case "number":
		return calculateNewNumberValue(sensor, currentTimeMs, scenario)
	case "boolean":
		return calculateNewBooleanValue(sensor, currentTimeMs, scenario)
	case "string":
		return calculateNewStringValue(sensor, currentTimeMs, scenario)
	default:
		// 알 수 없는 타입은 현재 값 유지
		return sensor.Value
	}
}

// calculateNewNumberValue는 숫자형 센서에 대한 새 값을 계산합니다
func calculateNewNumberValue(sensor *SensorData, currentTimeMs int64, scenario string) any {
	currValue, ok := sensor.Value.(float64)
	if !ok {
		// 값이 숫자가 아니면 0으로 초기화
		fmt.Printf("경고: 센서 %s의 타입은 'number'이지만 실제 값은 %T입니다. 0.0으로 초기화합니다.\n",
			sensor.Path, sensor.Value)
		return 0.0
	}

	// 패턴에 따른 값 계산
	switch sensor.ChangePattern {
	case "constant":
		// 값 변경 없음
		return currValue

	case "constant_with_noise":
		// 기준 값 주위로 노이즈 추가
		baseline := sensor.ChangeParams["baseline"]
		noise := sensor.ChangeParams["noise"]
		return baseline + (rand.Float64()*2.0-1.0)*noise

	case "linear":
		// 직선적 변화 (기울기에 따라 증가/감소)
		slope := sensor.ChangeParams["slope"]
		min := sensor.ChangeParams["min"]
		max := sensor.ChangeParams["max"]

		// 선형 변화
		// newValue := currValue + slope*float64(sensor.UpdateInterval/1000.0)
		newValue := currValue + slope*float64(sensor.UpdateInterval)/1000.0

		// 범위 제한
		if newValue < min {
			newValue = min
		} else if newValue > max {
			newValue = max
		}

		return newValue

	case "sinusoidal":
		// 사인파 변화
		amplitude := sensor.ChangeParams["amplitude"]
		period := sensor.ChangeParams["period"]
		baseline := sensor.ChangeParams["baseline"]
		phase := sensor.ChangeParams["phase"]

		// 사인파 계산
		timeInCycle := float64(currentTimeMs+int64(phase)) / period
		return baseline + amplitude*math.Sin(2.0*math.Pi*timeInCycle)

	case "random_walk":
		// 랜덤 워크 (이전 값에서 무작위로 이동)
		stepSize := sensor.ChangeParams["step_size"]
		min := sensor.ChangeParams["min"]
		max := sensor.ChangeParams["max"]

		// 랜덤 움직임
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
		// 계단식 변화 (일정 시간마다 단계적 변화)
		stepSize := sensor.ChangeParams["step_size"]
		stepInterval := sensor.ChangeParams["step_interval"]
		min := sensor.ChangeParams["min"]
		max := sensor.ChangeParams["max"]

		// 일정 간격마다 계단식 변화
		if float64(sensor.ChangeCounter*sensor.UpdateInterval) >= stepInterval {
			// 방향 결정 (증가 또는 감소)
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
		// 트리거 기반 변화 (특정 확률로 급격한 변화)
		threshold := sensor.ChangeParams["threshold"]
		resetInterval := sensor.ChangeParams["reset_interval"]
		minValue := sensor.ChangeParams["min_value"]
		maxValue := sensor.ChangeParams["max_value"]

		// 트리거 조건 확인
		if rand.Float64() < threshold || float64(sensor.ChangeCounter*sensor.UpdateInterval) >= resetInterval {
			// 범위 내 랜덤 값
			newValue := minValue + rand.Float64()*(maxValue-minValue)

			// 카운터 리셋
			sensor.ChangeCounter = 0

			return newValue
		}

		return currValue

	default:
		// 알 수 없는 패턴은 현재 값 유지
		return currValue
	}
}

// calculateNewBooleanValue는 불리언 타입 센서에 대한 새 값을 계산합니다
func calculateNewBooleanValue(sensor *SensorData, currentTimeMs int64, scenario string) any {
	currValue, ok := sensor.Value.(bool)
	if !ok {
		// 값이 불리언이 아니면 false로 초기화
		fmt.Printf("경고: 센서 %s의 타입은 'boolean'이지만 실제 값은 %T입니다. false로 초기화합니다.\n",
			sensor.Path, sensor.Value)
		return false
	}

	// 패턴에 따른 값 계산
	switch sensor.ChangePattern {
	case "constant":
		// 값 변경 없음
		return currValue

	case "toggle":
		// 특정 확률로 값 토글
		toggleProb := sensor.ChangeParams["toggle_probability"]

		if rand.Float64() < toggleProb {
			return !currValue
		}

		return currValue

	case "triggered":
		// 트리거 기반 토글
		threshold := sensor.ChangeParams["threshold"]

		if rand.Float64() < threshold {
			return !currValue
		}

		return currValue

	default:
		// 알 수 없는 패턴은 현재 값 유지
		return currValue
	}
}

// calculateNewStringValue는 문자열 타입 센서에 대한 새 값을 계산합니다
func calculateNewStringValue(sensor *SensorData, currentTimeMs int64, scenario string) any {
	currValue, ok := sensor.Value.(string)
	if !ok {
		// 값이 문자열이 아니면 기본값으로 초기화
		fmt.Printf("경고: 센서 %s의 타입은 'string'이지만 실제 값은 %T입니다. 기본 문자열로 초기화합니다.\n",
			sensor.Path, sensor.Value)
		return "unknown_value"
	}

	// 패턴에 따른 값 계산
	switch sensor.ChangePattern {
	case "constant":
		// 값 변경 없음
		return currValue

	case "random_selection":
		// 여러 값 중에서 무작위 선택
		numChoices := int(sensor.ChangeParams["num_choices"])
		changeProb := sensor.ChangeParams["change_probability"]

		// 기본값 설정
		if changeProb == 0.0 {
			changeProb = 0.05
		}

		// 변경 확률에 따라 결정
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

			// 현재 값과 다른 새 값 선택
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
		// 알 수 없는 패턴은 현재 값 유지
		return currValue
	}
}
