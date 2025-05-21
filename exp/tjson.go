package exp

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
)

// TJSONDocument represents a Temporal JSON document based on the concepts from Goyal and Dyreson's paper
type TJSONDocument struct {
	Items map[string]*TJSONItem `json:"items"`
}

// TJSONItem represents an item in a temporal JSON document that keeps its identity over time
type TJSONItem struct {
	Timestamp string                `json:"timestamp"`          // Period when this item was alive (e.g., "2015-2018")
	Versions  []*TJSONVersion       `json:"versions"`           // Versions of this item
	Children  map[string]*TJSONItem `json:"children,omitempty"` // Child items if this is a composite item
}

// TJSONVersion represents a specific version of an item
type TJSONVersion struct {
	Timestamp string      `json:"timestamp"` // Period when this version was active
	Data      interface{} `json:"data"`      // Actual data for this version
}

// NewTJSONDocument creates a new empty Temporal JSON document
func NewTJSONDocument() *TJSONDocument {
	return &TJSONDocument{
		Items: make(map[string]*TJSONItem),
	}
}

// AddItem adds a new item to the TJSON document
func (doc *TJSONDocument) AddItem(path string, startTime, endTime int64, value interface{}) error {
	parts := strings.Split(path, ".")
	
	if len(parts) == 0 {
		return fmt.Errorf("invalid path: %s", path)
	}

	// Create timestamp string
	timestamp := fmt.Sprintf("%d-%d", startTime, endTime)

	// Create or get the top-level item
	rootKey := parts[0]
	rootItem, exists := doc.Items[rootKey]
	if !exists {
		rootItem = &TJSONItem{
			Timestamp: timestamp,
			Versions:  make([]*TJSONVersion, 0),
			Children:  make(map[string]*TJSONItem),
		}
		doc.Items[rootKey] = rootItem
	}

	currentItem := rootItem
	// Navigate through the path
	for i := 1; i < len(parts); i++ {
		key := parts[i]

		// 모든 경로 부분에 대해 Children 항목 생성
		childItem, exists := currentItem.Children[key]
		if !exists {
			childItem = &TJSONItem{
				Timestamp: timestamp,
				Versions:  make([]*TJSONVersion, 0),
				Children:  make(map[string]*TJSONItem),
			}
			currentItem.Children[key] = childItem
		}
		currentItem = childItem

		// 마지막 부분일 경우 값도 설정
		if i == len(parts)-1 {
			version := &TJSONVersion{
				Timestamp: timestamp,
				Data:      value,
			}
			currentItem.Versions = append(currentItem.Versions, version)
		}
	}

	return nil
}

// TemporalSnapshot returns a snapshot of the document at a specific time
func (doc *TJSONDocument) TemporalSnapshot(timestamp int64) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 각 최상위 아이템에 대해 재귀적으로 처리
	for key, item := range doc.Items {
		snapshotValue := getSnapshotForItem(item, timestamp)
		if snapshotValue != nil {
			result[key] = snapshotValue
		}
	}
	return result, nil
}

// getSnapshotForItem은 주어진 시점에 아이템의 상태를 재귀적으로 계산
func getSnapshotForItem(item *TJSONItem, timestamp int64) interface{} {
	// 아이템이 해당 시점에 존재하는지 확인 !isTimeInRange(timestamp, item.Timestamp)
	if len(item.Children) == 0 && len(item.Versions) == 0 {
		return nil
	}

	// 아이템의 버전 데이터와 자식 노드의 결합된 데이터를 담을 맵
	itemSnapshot := make(map[string]interface{})

	// 해당 시점에 유효한 버전을 찾아 데이터 추출
	var versionData interface{}
	var versionFound bool

	// 버전에서 데이터 찾기 (시간 범위 내의 가장 마지막 버전 사용)
	for i := 0; i < len(item.Versions); i++ {
		version := item.Versions[i]
		if isTimeInRange(timestamp, version.Timestamp) {
			versionData = version.Data
			versionFound = true
			break
		}

		versionData = version.Data
		versionFound = true
	}

	// 버전 데이터가 맵인 경우
	if versionFound {
		if versionMap, isMap := versionData.(map[string]interface{}); isMap {
			// 맵 데이터 복사
			for k, v := range versionMap {
				itemSnapshot[k] = v
			}
		} else {
			// 맵이 아닌 경우는 바로 반환 (리프 노드)
			return versionData
		}
	}

	// 자식 아이템들 처리
	if len(item.Children) > 0 {
		// 자식 아이템 결과를 저장할 맵
		childrenData := make(map[string]interface{})
		childrenFound := false

		// 모든 자식 아이템에 대해 재귀적으로 처리
		for childKey, childItem := range item.Children {
			childSnapshot := getSnapshotForItem(childItem, timestamp)
			if childSnapshot != nil {
				childrenData[childKey] = childSnapshot
				childrenFound = true
			}
		}

		// 자식 데이터가 존재하면 결과에 추가
		if childrenFound {
			// 아직 itemSnapshot이 비어있고 버전 데이터가 없었다면
			if len(itemSnapshot) == 0 && !versionFound {
				return childrenData
			}

			// 아니면 자식 데이터를 itemSnapshot에 병합
			for k, v := range childrenData {
				itemSnapshot[k] = v
			}
		}
	}

	// 버전 데이터도 없고 자식 데이터도 없으면 nil 반환
	if len(itemSnapshot) == 0 && !versionFound {
		return nil
	}

	return itemSnapshot
}

// TemporalTrack returns changes to a specific path within a time range
// This is implemented as a time slice over the document
func (doc *TJSONDocument) TemporalTrack(path string, startTime, endTime int64) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0)
	parts := strings.Split(path, ".")

	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid path: %s", path)
	}

	// 최상위 아이템 찾기
	rootItem, exists := doc.Items[parts[0]]
	if !exists {
		return nil, fmt.Errorf("path not found: %s (root item missing)", path)
	}

	// 경로를 따라 중첩된 아이템 찾기
	targetItem := rootItem

	// 첫 번째 부분(index 0)은 이미 처리했으므로 1부터 시작
	for i := 1; i < len(parts); i++ {
		childItem, exists := targetItem.Children[parts[i]]
		if !exists {
			return nil, fmt.Errorf("path not found: %s (segment %s missing)", path, parts[i])
		}
		targetItem = childItem
	}

	// 찾은 아이템의 모든 버전 중 시간 범위 내에 있는 것들 선택
	for _, version := range targetItem.Versions {
		versionStart, versionEnd := parseTimeRange(version.Timestamp)

		// 요청된 시간 범위와 중첩되는지 확인
		if versionStart <= endTime && versionEnd >= startTime {
			result := map[string]interface{}{
				"timestamp": version.Timestamp,
				"data":      version.Data,
			}
			results = append(results, result)
		}
	}

	// 타임스탬프 기준으로 결과 정렬
	sort.Slice(results, func(i, j int) bool {
		iStart, _ := parseTimeRange(results[i]["timestamp"].(string))
		jStart, _ := parseTimeRange(results[j]["timestamp"].(string))
		return iStart < jStart
	})

	return results, nil
}

// FindEvents searches for events matching certain criteria within a time range
func (doc *TJSONDocument) FindEvents(path string, startTime, endTime int64,
	predicate func(interface{}) bool) ([]int64, error) {
	// Get the temporal track for the path
	trackResults, err := doc.TemporalTrack(path, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Find all timestamps where the predicate is true
	eventTimes := make([]int64, 0)
	for _, result := range trackResults {
		// Check if this value meets the criteria
		if predicate(result["data"]) {
			// Get the start time of this event
			start, _ := parseTimeRange(result["timestamp"].(string))
			eventTimes = append(eventTimes, start)
		}
	}

	return eventTimes, nil
}

// Helper function to check if a timestamp is within a time range
func isTimeInRange(timestamp int64, timeRange string) bool {
	start, end := parseTimeRange(timeRange)
	return timestamp >= start && timestamp <= end
}

// Helper function to parse a time range string like "2015-2018"
func parseTimeRange(timeRange string) (int64, int64) {
	parts := strings.Split(timeRange, "-")
	if len(parts) != 2 {
		return 0, 0
	}

	start, _ := strconv.ParseInt(parts[0], 10, 64)
	end, _ := strconv.ParseInt(parts[1], 10, 64)
	return start, end
}

// FromTSONPatches converts a series of TSON patches to a TJSON document
func FromTSONPatches(patches []TsonPatch) (*TJSONDocument, error) {
	doc := NewTJSONDocument()

	// Sort patches by timestamp
	sort.Slice(patches, func(i, j int) bool {
		return patches[i].Timestamp < patches[j].Timestamp
	})

	// Track the last value for each path
	lastValues := make(map[string]interface{})
	lastTimes := make(map[string]int64)

	// Process each patch
	for _, patch := range patches {
		path := patch.Path
		// Remove leading slash if present
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}
		// Replace slashes with dots for TJSON path format
		path = strings.ReplaceAll(path, "/", ".")

		// Record the value change
		if patch.Op == "add" || patch.Op == "replace" {
			// Check if value changed or this is the first occurrence
			lastValue, exists := lastValues[path]
			if !exists || !reflect.DeepEqual(lastValue, patch.Value) {
				// If we have a previous value, close its time range
				if exists {
					prevTime := lastTimes[path]
					// Add the previous item with its time range
					err := doc.AddItem(path, prevTime, patch.Timestamp-1, lastValue)
					if err != nil {
						fmt.Printf("Warning: Failed to add item for %s: %v\n", path, err)
					}
				}

				// Record the new value
				lastValues[path] = patch.Value
				lastTimes[path] = patch.Timestamp
			}
		} else if patch.Op == "remove" {
			// Handle deletion by closing the time range
			if lastValue, exists := lastValues[path]; exists {
				prevTime := lastTimes[path]
				err := doc.AddItem(path, prevTime, patch.Timestamp-1, lastValue)
				if err != nil {
					fmt.Printf("Warning: Failed to add item for %s (remove): %v\n", path, err)
				}
				delete(lastValues, path)
				delete(lastTimes, path)
			}
		}
	}

	// Close time ranges for values that still exist at the end
	currentTime := time.Now().Unix()
	for path, value := range lastValues {
		err := doc.AddItem(path, lastTimes[path], currentTime, value)
		if err != nil {
			fmt.Printf("Warning: Failed to add final item for %s: %v\n", path, err)
		}
	}

	return doc, nil
}

// Print prints the TJSON document in a readable format for debugging
func (doc *TJSONDocument) Print() {
	fmt.Println(spew.Sdump(doc))
}

// PrettyPrint prints the TJSON document in a more human-readable format
func (doc *TJSONDocument) PrettyPrint() {
	fmt.Println("TJSON Document:")
	fmt.Println("==============")

	// 각 최상위 아이템을 출력
	for key, item := range doc.Items {
		fmt.Printf("Item: %s (Timestamp: %s)\n", key, item.Timestamp)
		fmt.Println("  Versions:")

		// 각 버전 출력
		for i, version := range item.Versions {
			fmt.Printf("    [%d] Timestamp: %s\n", i, version.Timestamp)
			fmt.Printf("        Data: %v\n", version.Data)
		}

		// 자식 아이템이 있는 경우 출력
		if len(item.Children) > 0 {
			fmt.Println("  Children:")
			printChildren(item.Children, "    ")
		}

		fmt.Println()
	}
}

// Helper function to recursively print children
func printChildren(children map[string]*TJSONItem, indent string) {
	for key, child := range children {
		fmt.Printf("%sItem: %s (Timestamp: %s)\n", indent, key, child.Timestamp)

		// 각 버전 출력
		if len(child.Versions) > 0 {
			fmt.Printf("%s  Versions:\n", indent)
			for i, version := range child.Versions {
				fmt.Printf("%s    [%d] Timestamp: %s\n", indent, i, version.Timestamp)
				fmt.Printf("%s        Data: %v\n", indent, version.Data)
			}
		}

		// 재귀적으로 자식 아이템 출력
		if len(child.Children) > 0 {
			fmt.Printf("%s  Children:\n", indent)
			printChildren(child.Children, indent+"    ")
		}
	}
}
