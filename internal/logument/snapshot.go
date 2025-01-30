// package logument

// import (
// 	"encoding/json"
// 	"time"
// )

// // Snapshot TODO: Snapshot의 data를 JSON 형태로 바꿔주기
// // Snapshot 구조체
// type Snapshot struct {
// 	Version   uint64         // 스냅샷 버전, TODO: grow-only counter로 변경
// 	Timestamp time.Time      // 스냅샷 생성 시간
// 	Data      map[string]any // JSON 데이터
// }

// // ToJSON Snapshot을 JSON 형태로 변환하는 함수
// func (s *Snapshot) ToJSON() (string, error) {
// 	jsonData, err := json.Marshal(s)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(jsonData), nil
// }

// // FromJSON JSON 데이터를 Snapshot으로 변환하는 함수
// func (s *Snapshot) FromJSON(jsonString string) error {
// 	err := json.Unmarshal([]byte(jsonString), s)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // Cache
// // getTime leaf node이냐 아니냐에 따라 timestamp를 return
