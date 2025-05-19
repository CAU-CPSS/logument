package logument_test

import (
	"testing"

	"github.com/CAU-CPSS/logument/internal/logument"
	"github.com/CAU-CPSS/logument/internal/tson"
	"github.com/CAU-CPSS/logument/internal/tsonpatch"
	"github.com/davecgh/go-spew/spew"
)

const initSnapshot = `{
    "vehicleId" <1700000000>: "ABC1234",
    "speed" <1700000000>: 72.5,
    "engineOn" <1700000000>: true,
    "location": {
        "latitude" <1700000000>: 37.7749,
        "longitude" <1700000000>: -122.4194
    },
    "tirePressure": [
        <> 32.1,
        <> 31.8,
        <1700000000> 32.0,
        <1700000000> 31.9
    ]
}`

var patches = []string{
	`[
		{ "op": "replace", "path": "/location/latitude", "value": 43.9409, "timestamp": 1800000000 },
		{ "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 1800000000 },
		{ "op": "replace", "path": "/tirePressure/0", "value": 35.1, "timestamp": 1900000000 }
	]`,
	`[
		{ "op": "replace", "path": "/engineOn", "value": false, "timestamp": 2000000000 }
	]`,
	`[
		{ "op": "replace", "path": "/location/latitude", "value": 43.9409, "timestamp": 2100000000 },
		{ "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 2100000000 }
	]`,
	`[
		{ "op": "replace", "path": "/tirePressure/2", "value": 33.7, "timestamp": 2300000000 },
		{ "op": "replace", "path": "/speed", "value": 94.9, "timestamp": 2400000000 }
	]`,
}

func TestCreate(t *testing.T) {
	t.Log("Make a new Logument document\n")

	// use only string format
	t.Log("Make a Logument with string format\n")
	lgm := logument.NewLogument(initSnapshot, patches[0])
	t.Log(spew.Sdump(lgm))

	// use Snapshot and Patches format
	t.Log("Make a Logument with Snapshot format\n")
	var ss tson.Tson
	if err := tson.Unmarshal([]byte(initSnapshot), &ss); err != nil {
		panic(err)
	}
	pp, err := tsonpatch.Unmarshal([]byte(patches[0]))
	if err != nil {
		t.Error(err)
	}
	lgmWithFormatdata := logument.NewLogument(ss, pp)
	t.Log(spew.Sdump(lgmWithFormatdata))

	// use []Patches format
	t.Log("Make a Logument with []Patches format\n")
	pp2, err := tsonpatch.Unmarshal([]byte(patches[1]))
	if err != nil {
		t.Error(err)
	}
	lgmWithPatches := logument.NewLogument(ss, []tsonpatch.Patch{pp, pp2})
	t.Log(spew.Sdump(lgmWithPatches))
}

func TestStore(t *testing.T) {
	t.Log("Store patches to the pool\n")
	lgm := logument.NewLogument(initSnapshot, nil)

	lgm.Store(patches[0])
	t.Log(spew.Sdump(lgm))

	lgm.Store(patches[1])
	t.Log(spew.Sdump(lgm))
}

func TestApply(t *testing.T) {
	t.Log("Apply patches\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(patches[0])
	lgm.Store(patches[1])
	t.Log(spew.Sdump(lgm))

	if err := lgm.Append(); err != nil {
		t.Error(err)
	}
	t.Log(spew.Sdump(lgm))
}

func TestSnapshot(t *testing.T) {
	t.Log("Take a snapshot\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(patches[0])
	lgm.Store(patches[1])
	lgm.Append()

	lgm.Print()

	// Take a snapshot already taken
	snapshot := lgm.Snapshot(0)
	t.Log(spew.Sdump(snapshot))

	// Take a snapshot
	snapshot = lgm.Snapshot(1)
	t.Log(spew.Sdump(snapshot))

	// Requests exceeding latest version
	// snapshot = lgm.Snapshot(3)
	// t.Log(spew.Sdump(snapshot))
}

func TestTemporalSnapshot(t *testing.T) {
	t.Log("Take a timed snapshot\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(patches[0])
	lgm.Store(patches[1])
	lgm.Append()

	// Take a snapshot already taken
	// snapshot := lgm.TimedSnapshot(1700000000)
	// t.Log(spew.Sdump(snapshot))

	// Take a snapshot
	snapshot := lgm.TemporalSnapshot(1900000000)
	t.Log(spew.Sdump(snapshot))

	// Requests exceeding latest version
	// snapshot = lgm.TimedSnapshot(2100000000)
	// t.Log(spew.Sdump(snapshot))
}

func TestSlice(t *testing.T) {
	t.Log("Slice Logument\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(patches[0])
	lgm.Store(patches[1])
	if err := lgm.Append(); err != nil { // version 1
		t.Error(err)
	}

	lgm.Store(patches[2])
	if err := lgm.Append(); err != nil { // version 2
		t.Error(err)
	}

	lgm.Store(patches[3])
	if err := lgm.Append(); err != nil { // version 3
		t.Error(err)
	}

	// Slice patches
	lgmSubset := lgm.Slice(1, 3)
	lgmSubset.Print()
}

func TestTrack(t *testing.T) {
	t.Log("Track changes\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(patches[0])
	lgm.Store(patches[1])
	lgm.Append() // version 1
	lgm.Store(patches[2])
	lgm.Append() // version 2
	lgm.Store(patches[3])
	lgm.Append() // version 3

	// Track changes
	// Expected:
	// { "op": "replace", "path": "/location/latitude", "value": 43.9409, "timestamp": 2100000000 },
	// { "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 2100000000 },
	// { "op": "replace", "path": "/tirePressure/2", "value": 33.7, "timestamp": 2300000000 },
	// { "op": "replace", "path": "/speed", "value": 94.9, "timestamp": 2400000000 }
	changes := lgm.Track(2, 3)
	t.Log(spew.Sdump(changes))
}

func TestTemporalTrack(t *testing.T) {
	t.Log("Track changes\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(patches[0])
	lgm.Store(patches[1])
	lgm.Append() // version 1
	lgm.Store(patches[2])
	lgm.Append() // version 2
	lgm.Store(patches[3])
	lgm.Append() // version 3

	// Track changes
	// Expected:
	// { "op": "replace", "path": "/tirePressure/0", "value": 35.1, "timestamp": 1900000000 },
	// { "op": "replace", "path": "/engineOn", "value": false, "timestamp": 2000000000 },
	// { "op": "replace", "path": "/location/latitude", "value": 43.9409, "timestamp": 2100000000 },
	// { "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 2100000000 }
	changes := lgm.TemporalTrack(1900000000, 2100000000)
	t.Log(spew.Sdump(changes))
}

func TestSet(t *testing.T) {
	t.Log("Set a value\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(patches[0])
	lgm.Store(patches[1])
	lgm.Append()

	lgm.Store(patches[2])
	lgm.Append()

	lgm.Set(2, tsonpatch.Operation{
		Op:    "replace",
		Path:  "/location/latitude",
		Value: 42.4242,
	})
	lgm.Print()
}

func TestValidSet(t *testing.T) {
	t.Log("Set a value\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(patches[0])
	lgm.Store(patches[1])
	lgm.Append()

	lgm.Store(patches[2])
	lgm.Append()

	lgm.TestSet(2, tsonpatch.Operation{
		Op:    "replace",
		Path:  "/location/latitude",
		Value: 42.4242,
	})
	lgm.Print()
}

func TestCompact(t *testing.T) {
	t.Log("Compact patches\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(patches[0])
	lgm.Store(patches[1])
	lgm.Append()

	lgm.Store(patches[2])
	lgm.Append()

	lgm.Compact("/location")
	lgm.Print()
}

func TestHistory(t *testing.T) {
	t.Log("Show history\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(patches[0])
	lgm.Store(patches[1])
	lgm.Append()

	lgm.Store(patches[2])
	lgm.Append()

	// Show history of the "/location"
	his := lgm.History("/location")
	t.Log(spew.Sdump(his))
}

func TestSimpleCaseWithLogument(t *testing.T) {
	// 간단한 초기 상태
	initialState := tson.Object{
		"value": tson.Leaf[float64]{Value: 10.0, Timestamp: 1000},
	}

	// 두 개의 패치: 하나는 값 변경, 하나는 타임스탬프만 변경
	patches := []tsonpatch.Patch{
		{
			tsonpatch.Operation{Op: "replace", Path: "/value", Value: 20.0, Timestamp: 2000},
		},
		{
			tsonpatch.Operation{Op: "replace", Path: "/value", Value: 20.0, Timestamp: 3000},
		},
		{
			tsonpatch.Operation{Op: "replace", Path: "/value", Value: 40.0, Timestamp: 4000},
		},
	}

	// Logument로 적용
	lgm := logument.NewLogument(initialState, nil)

	// 첫 번째 패치: 값 변경 (10.0 != 20.0)
	for _, p := range patches[0] {
		lgm.TestSet(uint64(len(lgm.Version)), p)
	}
	lgm.Snapshot(uint64(len(lgm.Patches)))

	// 두 번째 패치: 타임스탬프만 변경 (20.0 == 20.0)
	for _, p := range patches[1] {
		lgm.TestSet(uint64(len(lgm.Version)), p)
	}
	lgm.Snapshot(uint64(len(lgm.Patches)))

	// 세 번째 패치:  값 변경 (20.0 != 40.0)
	for _, p := range patches[2] {
		lgm.TestSet(uint64(len(lgm.Version)), p)
	}
	lgm.Snapshot(uint64(len(lgm.Patches)))


	// 결과 확인
	lgm.Print()
	t.Logf("패치 수: %d\n", len(lgm.Patches))
	for v, p := range lgm.Patches {
		t.Logf("버전 %d: %d 작업\n", v, len(p))
	}
}

func TestTsonReference(t *testing.T) {
	// 초기 객체 생성
	initialObj := tson.Object{
		"value": tson.Leaf[float64]{Value: 10.0, Timestamp: 1000},
	}

	// 객체를 그대로 할당 (복사 없음)
	objRef := initialObj

	// 참조된 객체의 값 변경
	if leaf, ok := objRef["value"].(tson.Leaf[float64]); ok {
		leaf.Value = 20.0
		objRef["value"] = leaf
	}

	// 원본 객체 확인
	if leaf, ok := initialObj["value"].(tson.Leaf[float64]); ok {
		t.Logf("원본 값: %v", leaf.Value) // 10.0이 나와야 함
	}
}

func TestTsonMapReference(t *testing.T) {
    // 초기 객체 생성
    initialObj := tson.Object{
        "value": tson.Leaf[float64]{Value: 10.0, Timestamp: 1000},
    }
    
    // 맵을 그대로 할당 (얕은 복사)
    objRef := initialObj
    
    // 참조된 맵에 직접 새 키 추가
    objRef["newKey"] = tson.Leaf[string]{Value: "test", Timestamp: 1000}
    
    // 원본 맵 확인
    t.Logf("원본 맵 키 수: %d", len(initialObj)) // 1이 나와야 하지만 2가 나올 것
    t.Logf("원본 맵에 newKey 존재: %v", initialObj["newKey"] != nil) // false여야 하지만 true일 것
}

func TestLogumentSnapshotReference(t *testing.T) {
    // 초기 상태 생성
    initialState := tson.Object{
        "value": tson.Leaf[float64]{Value: 10.0, Timestamp: 1000},
    }
    
    // Logument 초기화
    lgm := logument.NewLogument(initialState, nil)
    
    // 초기 스냅샷에 직접 접근
    snapshot0 := lgm.Snapshots[0]
    
    // 스냅샷 내부 값 직접 변경 시도 (참조 테스트)
    if obj, ok := snapshot0.(tson.Object); ok {
        if leaf, ok := obj["value"].(tson.Leaf[float64]); ok {
            // 값 변경
            leaf.Value = 20.0
            obj["value"] = leaf
        }
    }
    
    // 변경 후 다시 확인
    updatedSnapshot0 := lgm.Snapshots[0]
    if obj, ok := updatedSnapshot0.(tson.Object); ok {
        if leaf, ok := obj["value"].(tson.Leaf[float64]); ok {
            t.Logf("변경 후 값: %v", leaf.Value) // 10.0이 나와야 하지만 20.0이 나올 것
        }
    }
}