package logument

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/CAU-CPSS/logument/internal/jsonpatch"
	"github.com/CAU-CPSS/logument/internal/jsonr"
	"github.com/davecgh/go-spew/spew"
)

// Note: http://tools.ietf.org/html/rfc6901#section-4 :
var (
	// rfc6901Encoder = strings.NewReplacer("~", "~0", "/", "~1")
	rfc6901Decoder = strings.NewReplacer("~1", "/", "~0", "~")
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
		panic("Invalid type for initialSnapshot. Must be string or jsonr.JsonR.")
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

func (lgm *Logument) Print() {
	fmt.Println(spew.Sdump(lgm))
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
func (lgm *Logument) Snapshot(targetVersion uint64) any {
	// Find the latest version before the target version
	latestVersion, latestSnapshot := lgm.findLatest(targetVersion)

	if latestVersion == targetVersion {
		s, err := jsonr.ToJson(latestSnapshot)
		if err != nil {
			panic("Failed to convert the snapshot to JSON. Error: " + err.Error())
		}

		return s
	}

	var timedSnapshot Snapshot

	// Apply patches from the latest version to the target version
	for i := latestVersion + 1; i <= targetVersion; i++ {
		var err error
		timedSnapshot, err = jsonpatch.ApplyPatch(latestSnapshot, lgm.PatchMap[i])
		if err != nil {
			panic("Failed to make a snapshot with the given version. Error: " + err.Error())
		}
	}

	s, err := jsonr.ToJson(timedSnapshot)
	if err != nil {
		panic("Failed to convert the snapshot to JSON. Error: " + err.Error())
	}

	var jsonSnapshot any
	err = json.Unmarshal(s, &jsonSnapshot)
	if err != nil {
		panic("Failed to unmarshal the snapshot. Error: " + err.Error())
	}

	return jsonSnapshot
}

func (lgm *Logument) TimedSnapshot(targetTimestamp int64) Snapshot {
	// Find the latest timestamp before the target timestamp
	latestVersion := lgm.Version[0]
	for v, s := range lgm.Snapshots {
		lts := jsonr.GetLatestTimestamp(s)
		if lts <= targetTimestamp {
			latestVersion = v
		} else {
			break
		}
	}

	// Get the latest snapshot
	latestSnapshot := lgm.Snapshots[latestVersion]

	if latestVersion > lgm.Version[len(lgm.Version)-1] {
		if lgm.PatchPool != nil {
			lgm.Apply()
		} else {
			return latestSnapshot
		}
	}

	var latestPatches Patches
	for _, p := range lgm.PatchMap[latestVersion+1] {
		if p.Timestamp <= targetTimestamp {
			latestPatches = append(latestPatches, p)
		}
	}

	fmt.Print("latestPatches: ", latestPatches)

	// Apply patches
	timedSnapshot, err := jsonpatch.ApplyPatch(latestSnapshot, latestPatches)
	if err != nil {
		panic("Failed to make a snapshot with the given version. Error: " + err.Error())
	}

	return timedSnapshot
}

func (lgm *Logument) Compact(targetPath string) {
	// Compact the Logument document
	// Remove the patches that have changed only the value without changing the TIMESTAMP ts
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	latestValues := make(map[string]any)

	keys := make([]uint64, 0, len(lgm.PatchMap))
	for key := range lgm.PatchMap {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, key := range keys {
		ps := lgm.PatchMap[key]
		compactPatches := make(Patches, 0, len(ps))
		for _, p := range ps {
			pt := rfc6901Decoder.Replace(p.Path)
			if strings.HasPrefix(pt, targetPath) {
				// Compare to previous value if it exists at the same path
				if prev, exists := latestValues[p.Path]; exists {
					// If value has changed, keep the patch and update the status
					if prev != p.Value {
						compactPatches = append(compactPatches, p)
						latestValues[p.Path] = p.Value
					}
					// Skip if value is the same
				} else {
					// Always keep the patch when it appears for the first time
					compactPatches = append(compactPatches, p)
					latestValues[p.Path] = p.Value
				}
			} else {
				// Keep patches that do not match the targetPath
				compactPatches = append(compactPatches, p)
			}
		}
		lgm.PatchMap[key] = compactPatches
	}
}

func (lgm *Logument) Slice(startVersion, endVersion uint64) *Logument {
	// Slice the Logument to make a subset of the Logument
	// The subset should contain the snapshots and patches from the start version to the end version
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	if startVersion > endVersion {
		panic("Start version should be smaller than the end version." +
			"\nStart version: " + strconv.FormatUint(startVersion, 10) +
			"\nEnd version: " + strconv.FormatUint(endVersion, 10))
	}

	if startVersion > lgm.Version[len(lgm.Version)-1] || endVersion > lgm.Version[len(lgm.Version)-1] {
		panic("Start version and end version should be smaller than the latest version." +
			"\nStart version: " + strconv.FormatUint(startVersion, 10) +
			"\nEnd version: " + strconv.FormatUint(endVersion, 10) +
			"\nLatest version: " + strconv.FormatUint(lgm.Version[len(lgm.Version)-1], 10))
	}

	SlicedSnapshots := make(map[uint64]Snapshot)
	for ver, snap := range lgm.Snapshots {
		if ver >= startVersion && ver <= endVersion {
			SlicedSnapshots[ver] = snap
		}
	}

	SlicedPatches := make(map[uint64]Patches)
	for ver, patch := range lgm.PatchMap {
		if ver >= startVersion && ver <= endVersion {
			SlicedPatches[ver] = patch
		}
	}

	slicedLgm := &Logument{
		Version:   lgm.Version[startVersion : endVersion+1],
		Snapshots: SlicedSnapshots,
		PatchMap:  SlicedPatches,
		PatchPool: nil,
	}

	return slicedLgm
}

func (lgm *Logument) Pack(targetVersion uint64) map[uint64]Patches {
	// Pack the Logument document to make a patch that contains all the changes
	// from the target version to the latest version
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	if targetVersion > lgm.Version[len(lgm.Version)-1] {
		panic("Target version should be smaller than the latest version." +
			"\nTarget version: " + strconv.FormatUint(targetVersion, 10) +
			"\nLatest version: " + strconv.FormatUint(lgm.Version[len(lgm.Version)-1], 10))
	}

	packedPatches := make(map[uint64]Patches)
	for i := targetVersion + 1; i <= lgm.Version[len(lgm.Version)-1]; i++ {
		packedPatches[i] = lgm.PatchMap[i]
	}

	latestValues := make(map[string]any)

	keys := make([]uint64, 0, len(packedPatches))
	for key := range packedPatches {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, key := range keys {
		ps := packedPatches[key]
		compactPatches := make(Patches, 0, len(ps))
		for _, p := range ps {
			// Compare to previous value if it exists at the same path
			if prev, exists := latestValues[p.Path]; exists {
				// If value has changed, keep the patch and update the status
				if prev != p.Value {
					compactPatches = append(compactPatches, p)
					latestValues[p.Path] = p.Value
				}
				// Skip if value is the same
			} else {
				// Always keep the patch when it appears for the first time
				compactPatches = append(compactPatches, p)
				latestValues[p.Path] = p.Value
			}
		}
		lgm.PatchMap[key] = compactPatches
	}

	return packedPatches
}

func (lgm *Logument) History(targetPath string) map[string]Patches {
	// Get the history of the changes at the target path
	// The history should contain all the patches that have changed the value at the target path
	// The patches should be sorted by the timestamp in ascending order
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	lgm.Compact(targetPath)

	historyPatches := make(map[string]Patches)
	for _, ps := range lgm.PatchMap {
		for _, p := range ps {
			path := rfc6901Decoder.Replace(p.Path)
			if strings.HasPrefix(path, targetPath) {
				historyPatches[p.Path] = append(historyPatches[p.Path], p)
			}
		}
	}

	for key := range historyPatches {
		val, err := jsonr.GetValue(lgm.Snapshots[0], key)
		fmt.Println("key: ", key)
		fmt.Println("val: ", val)
		if err != nil {
			panic("Failed to get the value from the snapshot. Error: " + err.Error())
		}
		if val != nil {
			historyPatches[key] = append([]jsonpatch.Operation{{
				Op:        "add",
				Path:      key,
				Value:     val,
				Timestamp: 0}},
				historyPatches[key]...)
		}
	}

	return historyPatches
}
