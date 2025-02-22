package logument

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/CAU-CPSS/logument/internal/jsonpatch"
	"github.com/CAU-CPSS/logument/internal/tson"
	"github.com/davecgh/go-spew/spew"
)

// Note: http://tools.ietf.org/html/rfc6901#section-4 :
var (
	// rfc6901Encoder = strings.NewReplacer("~", "~0", "/", "~1")
	rfc6901Decoder = strings.NewReplacer("~1", "/", "~0", "~")
)

type Snapshot = tson.Tson
type Patches = jsonpatch.Patch // []jsonpatch.Operation

// Logument 구조체
type Logument struct {
	Version   []uint64            // Version 정보 관리
	Snapshots map[uint64]Snapshot // 만들었던 Shanpshot 들의 map (version, snapshot)
	PatchMap  map[uint64]Patches  // Patch 들의 map (version, patches)
	PatchPool Patches             // Apply 되지 않은 patch
}

func NewLogument(initialSnapshot any, initialPatches any) *Logument {
	// Create a new Logument document with the given initial data
	var snapshot Snapshot

	switch initialSnapshot := initialSnapshot.(type) {
	case string:
		err := tson.Unmarshal([]byte(initialSnapshot), &snapshot)
		if err != nil {
			panic(err)
		}
	case Snapshot:
		snapshot = initialSnapshot
	default:
		panic("Invalid type for initialSnapshot. Must be string or tson.Tson.")
	}

	lgm := &Logument{
		Version:   []uint64{0},
		Snapshots: map[uint64]Snapshot{0: snapshot},
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

	sort.Slice(lgm.Version, func(i, j int) bool {
		return lgm.Version[i] < lgm.Version[j]
	})

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

func (lgm *Logument) getSortedVersionsFromSnapshot() []uint64 {
	versions := make([]uint64, 0, len(lgm.Snapshots))
	for version := range lgm.Snapshots {
		versions = append(versions, version)
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })
	return versions
}

func (lgm *Logument) getSortedVersionsFromPatch() []uint64 {
	versions := make([]uint64, 0, len(lgm.PatchMap))
	for version := range lgm.PatchMap {
		versions = append(versions, version)
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })
	return versions
}

func (lgm *Logument) Store(inputPatches any) {
	var patches Patches

	switch inputPatches := inputPatches.(type) {
	case string:
		var err error
		patches, err = jsonpatch.Unmarshal([]byte(inputPatches))
		if err != nil {
			panic(err)
		}
	case Patches:
		patches = inputPatches
	case []Patches:
		var tempPatches Patches
		for _, patch := range inputPatches {
			tempPatches = append(tempPatches, patch...)
		}
		patches = tempPatches
	default:
		panic("Invalid type for initialPatches. Must be Patch or []Patch.")
	}

	if lgm.PatchPool == nil {
		lgm.PatchPool = patches
	} else {
		lgm.PatchPool = append(lgm.PatchPool, patches...)
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

	if _, exist := lgm.PatchMap[latestVersion+1]; exist {
		panic("The patch for the next version already exists.")
	}
	lgm.PatchMap[latestVersion+1] = lgm.PatchPool

	lgm.Version = append(lgm.Version, latestVersion+1)
	lgm.PatchPool = nil
}

func (lgm *Logument) findLatest(targetVersion uint64) (latestVersion uint64, latestSnapshot Snapshot) {
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	if targetVersion > lgm.Version[len(lgm.Version)-1] {
		panic("Target version should be smaller than the latest version." +
			"\nTarget version: " + strconv.FormatUint(targetVersion, 10) +
			"\nLatest version: " + strconv.FormatUint(lgm.Version[len(lgm.Version)-1], 10))
	}

	versions := lgm.getSortedVersionsFromSnapshot()
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
	var timedSnapshot Snapshot

	if latestVersion != targetVersion {
		// Apply patches from the latest version to the target version
		for i := latestVersion + 1; i <= targetVersion; i++ {
			var err error
			timedSnapshot, err = jsonpatch.ApplyPatch(latestSnapshot, lgm.PatchMap[i])
			if err != nil {
				panic("Failed to make a snapshot with the given version. Error: " + err.Error())
			}
		}
	} else {
		timedSnapshot = latestSnapshot
	}

	jsonSnapshot, err := tson.ToJson(timedSnapshot)
	if err != nil {
		panic("Failed to convert the snapshot to JSON. Error: " + err.Error())
	}

	var unmarshaledJsonSnapshot any
	err = json.Unmarshal(jsonSnapshot, &unmarshaledJsonSnapshot)
	if err != nil {
		panic("Failed to unmarshal the snapshot. Error: " + err.Error())
	}

	snapshot, err := tson.ToTson(unmarshaledJsonSnapshot)
	if err != nil {
		panic("Failed to convert the snapshot to Tson. Error: " + err.Error())
	}

	return snapshot
}

func (lgm *Logument) TimeSnapshot(targetTime int64) Snapshot {
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	versions := lgm.getSortedVersionsFromSnapshot()

	// Find the latest timestamp before the target timestamp
	latestVersion := lgm.Version[0]
	for _, version := range versions {
		s := lgm.Snapshots[version]
		lts := tson.GetLatestTimestamp(s)
		if lts <= targetTime {
			latestVersion = version
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
		if p.Timestamp <= targetTime {
			latestPatches = append(latestPatches, p)
		}
	}

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

	versions := lgm.getSortedVersionsFromPatch()
	for _, version := range versions {
		ps := lgm.PatchMap[version]
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
		lgm.PatchMap[version] = compactPatches
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

	var SlicedVersions []uint64

	versionsFromSnapshot := lgm.getSortedVersionsFromSnapshot()
	SlicedSnapshots := make(map[uint64]Snapshot)
	for _, version := range versionsFromSnapshot {
		if version >= startVersion && version <= endVersion {
			SlicedSnapshots[version] = lgm.Snapshots[version]
		}
	}

	versionsFromPatch := lgm.getSortedVersionsFromPatch()
	SlicedPatches := make(map[uint64]Patches)
	for _, version := range versionsFromPatch {
		if version >= startVersion && version <= endVersion {
			// SlicedVersions always includes all versions between the start and end versions
			// because lgm.Version and lgm.PatchMap are continuous.
			// However, lgm.Snapshots may not be continuous,
			// so the following line works correctly only in this loop.
			SlicedVersions = append(SlicedVersions, version)
			SlicedPatches[version] = lgm.PatchMap[version]
		}
	}

	// Add the snapshot at the start version if it does not exist
	if len(SlicedSnapshots) == 0 {
		SlicedSnapshots[startVersion] = lgm.Snapshot(startVersion)
	}

	slicedLgm := &Logument{
		Version:   SlicedVersions,
		Snapshots: SlicedSnapshots,
		PatchMap:  SlicedPatches,
		PatchPool: nil,
	}

	return slicedLgm
}

func (lgm *Logument) TimeSlice(startTime, endTime int64) *Logument {
	// TimeSlice the Logument to make a subset of the Logument based on the timestamp
	// The subset should contain the snapshots and patches from the start time to the end time
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}
	if startTime > endTime {
		panic("Start time should be smaller than the end time." +
			"\nStart time: " + strconv.FormatInt(startTime, 10) +
			"\nEnd time: " + strconv.FormatInt(endTime, 10))
	}

	var SlicedVersions []uint64

	versionsFromSnapshot := lgm.getSortedVersionsFromSnapshot()
	SlicedSnapshots := make(map[uint64]Snapshot)
	for _, version := range versionsFromSnapshot {
		snapshot := lgm.Snapshots[version]
		latestTimestamp := tson.GetLatestTimestamp(snapshot)
		if latestTimestamp >= startTime && latestTimestamp <= endTime {
			SlicedSnapshots[version] = snapshot
		}
	}

	versionsFromPatch := lgm.getSortedVersionsFromPatch()
	SlicedPatches := make(map[uint64]Patches)
	for _, version := range versionsFromPatch {
		patches := lgm.PatchMap[version]
		for _, patch := range patches {
			if patch.Timestamp >= startTime && patch.Timestamp <= endTime {
				// SlicedVersions always includes all versions between the start and end times
				// because lgm.Version and lgm.PatchMap are continuous.
				// However, lgm.Snapshots may not be continuous,
				// so the following line works correctly only in this loop.
				SlicedVersions = append(SlicedVersions, version)
				SlicedPatches[version] = append(SlicedPatches[version], patch)
			}
		}
	}

	// Add the snapshot at the start version if it does not exist
	if len(SlicedSnapshots) == 0 {
		startVersion := SlicedVersions[0]
		SlicedSnapshots[startVersion] = lgm.TimeSnapshot(startTime)
	}

	slicedLgm := &Logument{
		Version:   SlicedVersions,
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

	if lgm.Version[len(lgm.Version)-1] < targetVersion {
		panic("Target version should be smaller than the latest version." +
			"\nTarget version: " + strconv.FormatUint(targetVersion, 10) +
			"\nLatest version: " + strconv.FormatUint(lgm.Version[len(lgm.Version)-1], 10))
	}

	versions := lgm.getSortedVersionsFromPatch()
	packedPatches := make(map[uint64]Patches)
	for _, version := range versions {
		if targetVersion <= version {
			packedPatches[version] = lgm.PatchMap[version]
		}
	}

	latestValues := make(map[string]any)

	paths := make([]uint64, 0, len(packedPatches))
	for path := range packedPatches {
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool { return paths[i] < paths[j] })

	for _, path := range paths {
		patches := packedPatches[path]
		compactPatches := make(Patches, 0, len(patches))
		for _, patch := range patches {
			// Compare to previous value if it exists at the same path
			if prev, exists := latestValues[patch.Path]; exists {
				// If value has changed, keep the patch and update the status
				if prev != patch.Value {
					compactPatches = append(compactPatches, patch)
					latestValues[patch.Path] = patch.Value
				}
				// Skip if value is the same
			} else {
				// Always keep the patch when it appears for the first time
				compactPatches = append(compactPatches, patch)
				latestValues[patch.Path] = patch.Value
			}
		}
		lgm.PatchMap[path] = compactPatches
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
	for _, patches := range lgm.PatchMap {
		for _, patch := range patches {
			path := rfc6901Decoder.Replace(patch.Path)
			if strings.HasPrefix(path, targetPath) {
				historyPatches[patch.Path] = append(historyPatches[patch.Path], patch)
			}
		}
	}

	// Add the initial value to the history
	for key := range historyPatches {
		val, err := tson.GetValue(lgm.Snapshots[0], key)
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
				Timestamp: 0,
			}},
				historyPatches[key]...) // Add the initial value to the beginning
		}
	}

	return historyPatches
}
