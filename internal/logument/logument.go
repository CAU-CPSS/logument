package logument

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/CAU-CPSS/logument/internal/jsonpatch"
	"github.com/CAU-CPSS/logument/internal/jsonr"
)

type Snapshot = jsonr.JsonR
type Patches = jsonpatch.Patch

// Logument 구조체
type Logument struct {
	Version   []uint64            // Version 정보 관리
	Snapshots map[uint64]Snapshot // 만들었던 Shanpshot 들의 map (version, snapshot)
	PatchMap  map[uint64]Patches  // Patch 들의 map (version, patches)
	PatchPool Patches             // Apply 되지 않은 patch
}

// NewLogument TODO: initial data 를 json 으로 변경
// Create a new Logument document with the given initial data.
func NewLogument(initialSnapshot any, initialPatches any) *Logument {
	var ss Snapshot

	switch initialSnapshot := initialSnapshot.(type) {
	case string:
		err := jsonr.Unmarshal([]byte(initialSnapshot), &ss)
		if err != nil {
			panic(err)
		}
	case Snapshot:
		ss = initialSnapshot
	default:
		panic("Invalid type for initialSnapshot. Must be string or jsonr.Snapshot.")
	}

	lgm := &Logument{
		Version:   []uint64{0},
		Snapshots: map[uint64]Snapshot{0: ss},
		PatchMap:  make(map[uint64]Patches),
		PatchPool: nil,
	}

	if initialPatches != nil {
		lgm.Store(initialPatches)
	}

	return lgm

}

func (lgm *Logument) isContinuous() bool {
	// Check if the versions are continuous
	// If the versions are not continuous, return false
	// Otherwise, return true
	if len(lgm.Version) == 0 && len(lgm.Version) == 1 {
		return true
	}

	for idx, v := range lgm.Version {
		if idx == 0 {
			continue
		}
		if v != lgm.Version[idx-1]+1 {
			return false
		}
	}
	return true
}

func (lgm *Logument) Store(inputPatches any) {
	var pp Patches

	switch inputPatches := inputPatches.(type) {
	case string:
		var err error
		pp, err = jsonpatch.Unmarshal([]byte(inputPatches))
		if err != nil {
			panic(err)
		}
	case Patches:
		pp = inputPatches
	case []Patches:
		var tempPatches Patches
		for _, p := range inputPatches {
			tempPatches = append(tempPatches, p...)
		}
		pp = tempPatches
	default:
		panic("Invalid type for initialPatches. Must be Patch or []Patch.")
	}

	if lgm.PatchPool == nil {
		lgm.PatchPool = pp
	} else {
		lgm.PatchPool = append(lgm.PatchPool, pp...)
	}
}

// Apply PatchPool에 있는 patch를 PatchMap에 적용하고, version을 증가시킴
func (lgm *Logument) Apply() {
	// This operation applies a new patch $P(v_n, v_{n+1})$ to an existing Logument instance at $V_n$,
	// resulting in an updated instance at $V_{n+1}$.
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	latestVersion := lgm.Version[len(lgm.Version)-1]
	lgm.PatchMap[latestVersion+1] = lgm.PatchPool
	lgm.Version = append(lgm.Version, latestVersion+1)
	lgm.PatchPool = nil
}

func (lgm *Logument) findLatest(targetVersion uint64) (latestVersion uint64, latestSnapshot Snapshot) {
	if targetVersion > lgm.Version[len(lgm.Version)-1] {
		panic("Target version should be smaller than the latest version." +
			"\nTarget version: " + strconv.FormatUint(targetVersion, 10) +
			"\nLatest version: " + strconv.FormatUint(lgm.Version[len(lgm.Version)-1], 10))
	}
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	versions := make([]uint64, 0, len(lgm.Snapshots))
	for ver := range lgm.Snapshots {
		versions = append(versions, ver)
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i] < versions[j]
	})
	idx := sort.Search(len(versions), func(i int) bool {
		return versions[i] > targetVersion
	})
	if idx == 0 {
		panic("No version found <= targetVersion")
	}

	return versions[idx-1], lgm.Snapshots[latestVersion]

}

// Snapshot Snapshot 생성
func (lgm *Logument) Snapshot(targetVersion uint64) Snapshot {
	// Find the latest version before the target version
	latestVersion, latestSnapshot := lgm.findLatest(targetVersion)

	if latestVersion == targetVersion {
		return latestSnapshot
	}

	// Apply patches from the latest version to the target version
	for i := latestVersion + 1; i <= targetVersion; i++ {
		var err error
		fmt.Println(reflect.TypeOf(latestSnapshot))
		fmt.Println(reflect.TypeOf(lgm.PatchMap[i]))
		latestSnapshot, err = jsonpatch.ApplyPatch(latestSnapshot, lgm.PatchMap[i])
		if err != nil {
			panic("Failed to make a snapshot with the given version. Error: " + err.Error())
		}
	}

	return latestSnapshot
}

// timestamp 와 version을 같이 관리해야할 것 같음
// 논문 Proposed method 이후에 perspective 혹은 vision이 들어가야할 듯
// 내부적으로는 sqlite 같이 가벼운 DB를 사용해서 구현해보면 좋을 듯
