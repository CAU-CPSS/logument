package logument

import (
	"sync/atomic"
	"time"

	"github.com/CAU-CPSS/logument/internal/jsonpatch"
	"github.com/CAU-CPSS/logument/internal/jsonr"
)

type Snapshot = jsonr.JsonR
type Patch = jsonpatch.Patch

// Logument 구조체
type Logument struct {
	Version   []uint64            // Version 정보 관리
	Snapshots map[uint64]Snapshot // 만들었던 Shanpshot 들의 배열
	Patches   map[uint64]Patch    // Patch 들의 배열
}

// NewLogument TODO: initial data 를 json 으로 변경
// Create a new Logument document with the given initial data.
func NewLogument(initialSnapshot any, initialPatches any) *Logument {
	var ss Snapshot
	var pp Patch

	switch initialSnapshot.(type) {
	case string:
		err := jsonr.Unmarshal([]byte(initialSnapshot.(string)), &ss)
		if err != nil {
			panic(err)
		}
	case Snapshot:
		ss = initialSnapshot.(Snapshot)
	default:
		panic("Invalid type for initialSnapshot. Must be string or jsonr.Snapshot.")
	}

	switch initialPatches.(type) {
	case Patch:
		pp = initialPatches.(Patch)
	default:
		panic("Invalid type for initialPatches. Must be Patch or []Patch.")
	}

	return &Logument{
		Version:   []uint64{0},
		Snapshots: map[uint64]Snapshot{0: ss},
		Patches:   map[uint64]Patch{0: pp},
	}
}

// AppendPatch data를 받아 logument에 patches에 patch를 추가
// TODO: Psuedo 구현임 실제 구현 필요
func (lgm *Logument) AppendPatch(newPatch any) {
	switch p := newPatch.(type) {
	case Patch:
		lgm.Patches = append(lgm.Patches, p)
	case []Patch:
		lgm.Patches = append(lgm.Patches, p...)
	default:
		panic("Invalid type for AddPatch. Must be Patch or []Patch.")
	}
}

// Snapshot Snapshot 생성
func (lgm *Logument) Snapshot(targetVersion uint64) Snapshot {
	// Find the latest version before the target version
	var latestVersion uint64
	var latestSnapshot Snapshot
	for version, snapshot := range lgm.Snapshots {
		if version <= targetVersion {
			latestVersion = version
			latestSnapshot = snapshot
		}
	}

	if latestVersion == targetVersion {
		return latestSnapshot
	}

	latestSnapshot

	// ss := lgm.CurrentSnapshot

	// lgm.Snapshots = append(lgm.Snapshots, updatedSnapshot)
	// return updatedSnapshot
}

// Apply Snapshot과 Patch 병합
func (l *Logument) Apply() {
	for _, patch := range l.Patches {
		switch patch.Op {
		case "replace":
			l.CurrentSnapshot.Data[patch.Path] = patch.Value
		case "add":
			// 예시로 간단히 처리
			if _, exists := l.CurrentSnapshot.Data[patch.Path]; !exists {
				l.CurrentSnapshot.Data[patch.Path] = patch.Value
			}
		case "remove":
			delete(l.CurrentSnapshot.Data, patch.Path)
		}
	}
	l.Patches = []Patch{}
	atomic.AddUint64(&l.CurrentSnapshot.Version, 1)
	l.CurrentSnapshot.Timestamp = time.Now()
}

// timestamp 와 version을 같이 관리해야할 것 같음
// 논문 Proposed method 이후에 perspective 혹은 vision이 들어가야할 듯
// 내부적으로는 sqlite 같이 가벼운 DB를 사용해서 구현해보면 좋을 듯
