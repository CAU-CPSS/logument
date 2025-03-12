package logument

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/CAU-CPSS/logument/internal/tson"
	"github.com/CAU-CPSS/logument/internal/tsonpatch"
	"github.com/davecgh/go-spew/spew"
)

// Note: http://tools.ietf.org/html/rfc6901#section-4 :
var (
	// rfc6901Encoder = strings.NewReplacer("~", "~0", "/", "~1")
	rfc6901Decoder = strings.NewReplacer("~1", "/", "~0", "~")
)

type tsonSnapshot = tson.Tson
type tsonPatches = tsonpatch.Patch // []jsonpatch.Operation

// Logument
type Logument struct {
	Version   []uint64                // Version list
	Snapshots map[uint64]tsonSnapshot // A map which contains an initial Snapshot (by `Create`) and Snapshots from `Snapshot` Function {version: Snapshot}
	Patches   map[uint64]tsonPatches  // A map which contains Patches from `Append` Function {version: Patches}
	PatchPool tsonPatches             // A pool of Patches from `Store` Function
}

func NewLogument(initialSnapshot any, initialPatches any) *Logument {
	// Create a new Logument document with the given initial data
	var snapshot tsonSnapshot

	switch initialSnapshot := initialSnapshot.(type) {
	case string:
		err := tson.Unmarshal([]byte(initialSnapshot), &snapshot)
		if err != nil {
			panic(err)
		}
	case []byte:
		err := tson.Unmarshal(initialSnapshot, &snapshot)
		if err != nil {
			panic(err)
		}
	case tsonSnapshot:
		snapshot = initialSnapshot
	default:
		panic("Invalid type for initialSnapshot. Must be string or tson.Tson.")
	}

	lgm := &Logument{
		Version:   []uint64{0},
		Snapshots: map[uint64]tsonSnapshot{0: snapshot},
		Patches:   make(map[uint64]tsonPatches),
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
	if len(lgm.Version) <= 1 {
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

func (lgm *Logument) getSortedVersions(source string) []uint64 {
	var versions []uint64
	switch source {
	case "snapshot":
		versions = make([]uint64, 0, len(lgm.Snapshots))
		for version := range lgm.Snapshots {
			versions = append(versions, version)
		}
	case "patch":
		versions = make([]uint64, 0, len(lgm.Patches))
		for version := range lgm.Patches {
			versions = append(versions, version)
		}
	default:
		panic("Invalid source. Must be 'snapshot' or 'patch'.")
	}

	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })

	return versions
}

func (lgm *Logument) Store(inputPatches any) {
	var patches tsonPatches

	switch inputPatches := inputPatches.(type) {
	case string:
		var err error
		patches, err = tsonpatch.Unmarshal([]byte(inputPatches))
		if err != nil {
			panic(err)
		}
	case []byte:
		var err error
		patches, err = tsonpatch.Unmarshal(inputPatches)
		if err != nil {
			panic(err)
		}
	case tsonPatches:
		patches = inputPatches
	case []tsonPatches:
		var tempPatches tsonPatches
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

func (lgm *Logument) findLatest(targetVersion uint64) (latestVersion uint64, latestSnapshot tsonSnapshot, rr error) {
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	if targetVersion > lgm.Version[len(lgm.Version)-1] {
		panic("Target version should be smaller than the latest version." +
			"\nTarget version: " + strconv.FormatUint(targetVersion, 10) +
			"\nLatest version: " + strconv.FormatUint(lgm.Version[len(lgm.Version)-1], 10))
	}

	versions := lgm.getSortedVersions("snapshot")
	idx := sort.Search(len(versions), func(i int) bool {
		return versions[i] > targetVersion
	})
	if idx == 0 {
		return 0, nil, fmt.Errorf("no version found <= targetVersion")
	}

	return versions[idx-1], lgm.Snapshots[versions[idx-1]], nil
}

// Append Append the patch from PatchPool to the Patches
func (lgm *Logument) Append() error {
	if !lgm.isContinuous() {
		return fmt.Errorf("versions are not continuous")
	}
	latestVersion := lgm.Version[len(lgm.Version)-1]
	if _, exist := lgm.Patches[latestVersion+1]; exist {
		return fmt.Errorf("the patch for the next version already exists")
	}
	lgm.Patches[latestVersion+1] = lgm.PatchPool
	lgm.Version = append(lgm.Version, latestVersion+1)
	lgm.PatchPool = nil

	return nil
}

// Snapshot Create a snapshot at the target version
func (lgm *Logument) Snapshot(vk uint64) tsonSnapshot {
	// Find the latest version before the target version
	latestVersion, latestSnapshot, err := lgm.findLatest(vk)
	if err != nil {
		panic(err)
	}

	var timedSnapshot tsonSnapshot

	if latestVersion != vk {
		// Apply patches from the latest version to the target version
		for i := latestVersion + 1; i <= vk; i++ {
			var err error
			timedSnapshot, err = tsonpatch.ApplyPatch(latestSnapshot, lgm.Patches[i])
			if err != nil {
				panic("Failed to make a snapshot with the given version. Error: " + err.Error())
			}
		}
	} else {
		timedSnapshot = latestSnapshot
	}

	if _, exists := lgm.Snapshots[vk]; !exists {
		lgm.Snapshots[vk] = timedSnapshot
	}

	return timedSnapshot
}

func (lgm *Logument) TemporalSnapshot(tsk int64) tsonSnapshot {
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	versions := lgm.getSortedVersions("snapshot")

	// Find the latest timestamp before the target timestamp
	latestVersion := lgm.Version[0]
	for _, version := range versions {
		s := lgm.Snapshots[version]
		lts := tson.GetLatestTimestamp(s)
		if lts <= tsk {
			latestVersion = version
		} else {
			break
		}
	}

	// Get the latest snapshot
	latestSnapshot := lgm.Snapshots[latestVersion]

	if latestVersion > lgm.Version[len(lgm.Version)-1] {
		if lgm.PatchPool != nil {
			lgm.Append()
		} else {
			return latestSnapshot
		}
	}

	var latestPatches tsonPatches
	for _, p := range lgm.Patches[latestVersion+1] {
		if p.Timestamp <= tsk {
			latestPatches = append(latestPatches, p)
		}
	}

	// Apply patches
	timedSnapshot, err := tsonpatch.ApplyPatch(latestSnapshot, latestPatches)
	if err != nil {
		panic("Failed to make a snapshot with the given version. Error: " + err.Error())
	}

	return timedSnapshot
}

func (lgm *Logument) Slice(vi, vj uint64) *Logument {
	// Slice the Logument to make a subset of the Logument
	// The subset should contain the snapshots and patches from vi to vj
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	if vi > vj {
		panic("Start version should be smaller than the end version." +
			"\nStart version: " + strconv.FormatUint(vi, 10) +
			"\nEnd version: " + strconv.FormatUint(vj, 10))
	}

	if vi > lgm.Version[len(lgm.Version)-1] || vj > lgm.Version[len(lgm.Version)-1] {
		panic("Start version and end version should be smaller than the latest version." +
			"\nStart version: " + strconv.FormatUint(vi, 10) +
			"\nEnd version: " + strconv.FormatUint(vj, 10) +
			"\nLatest version: " + strconv.FormatUint(lgm.Version[len(lgm.Version)-1], 10))
	}

	var SlicedVersions []uint64

	versionsFromSnapshot := lgm.getSortedVersions("snapshot")
	SlicedSnapshots := make(map[uint64]tsonSnapshot)
	for _, version := range versionsFromSnapshot {
		if version >= vi && version <= vj {
			SlicedSnapshots[version] = lgm.Snapshots[version]
		}
	}

	versionsFromPatch := lgm.getSortedVersions("patch")
	SlicedPatches := make(map[uint64]tsonPatches)
	for _, version := range versionsFromPatch {
		if version >= vi && version <= vj {
			// SlicedVersions always includes all versions between the start and end versions
			// because lgm.Version and lgm.PatchMap are continuous.
			// However, lgm.Snapshots may not be continuous,
			// so the following line works correctly only in this loop.
			SlicedVersions = append(SlicedVersions, version)
			SlicedPatches[version] = lgm.Patches[version]
		}
	}

	// Add the snapshot at the start version if it does not exist
	if len(SlicedSnapshots) == 0 {
		SlicedSnapshots[vi] = lgm.Snapshot(vi)
	}

	slicedLgm := &Logument{
		Version:   SlicedVersions,
		Snapshots: SlicedSnapshots,
		Patches:   SlicedPatches,
		PatchPool: nil,
	}

	return slicedLgm
}

func (lgm *Logument) TemporalSlice(tsi, tsj int64) *Logument {
	// TimeSlice the Logument to make a subset of the Logument based on the timestamp
	// The subset should contain the snapshots and patches from the start time to the end time
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}
	if tsi > tsj {
		panic("Start time should be smaller than the end time." +
			"\nStart time: " + strconv.FormatInt(tsi, 10) +
			"\nEnd time: " + strconv.FormatInt(tsj, 10))
	}

	var SlicedVersions []uint64

	versionsFromSnapshot := lgm.getSortedVersions("snapshot")
	SlicedSnapshots := make(map[uint64]tsonSnapshot)
	for _, version := range versionsFromSnapshot {
		snapshot := lgm.Snapshots[version]
		latestTimestamp := tson.GetLatestTimestamp(snapshot)
		if latestTimestamp >= tsi && latestTimestamp <= tsj {
			SlicedSnapshots[version] = snapshot
		}
	}

	versionsFromPatch := lgm.getSortedVersions("patch")
	SlicedPatches := make(map[uint64]tsonPatches)
	for _, version := range versionsFromPatch {
		patches := lgm.Patches[version]
		for _, patch := range patches {
			if patch.Timestamp >= tsi && patch.Timestamp <= tsj {
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
		SlicedSnapshots[startVersion] = lgm.TemporalSnapshot(tsi)
	}

	slicedLgm := &Logument{
		Version:   SlicedVersions,
		Snapshots: SlicedSnapshots,
		Patches:   SlicedPatches,
		PatchPool: nil,
	}

	return slicedLgm
}

func (lgm *Logument) Track(vi, vj uint64) map[uint64]tsonPatches {
	// Track the Logument document to make a patch that contains all the changes
	// from the version vi to the version vj
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	if lgm.Version[len(lgm.Version)-1] < vi || lgm.Version[len(lgm.Version)-1] < vj {
		panic("Target versions should be smaller than the latest version." +
			"\nTarget version vi: " + strconv.FormatUint(vi, 10) +
			"\nTarget version vj: " + strconv.FormatUint(vj, 10) +
			"\nLatest version: " + strconv.FormatUint(lgm.Version[len(lgm.Version)-1], 10))
	}

	if vi > vj {
		panic("Target version vi should be smaller than or equal to target version vj." +
			"\nTarget version vi: " + strconv.FormatUint(vi, 10) +
			"\nTarget version vj: " + strconv.FormatUint(vj, 10))
	}

	versions := lgm.getSortedVersions("patch")
	packedPatches := make(map[uint64]tsonPatches)
	for _, version := range versions {
		if vi <= version && version <= vj {
			packedPatches[version] = lgm.Patches[version]
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
		compactPatches := make(tsonPatches, 0, len(patches))
		for _, patch := range patches {
			// Compare to previous value if it exists at the same path
			if prev, exists := latestValues[patch.Path]; exists {
				// If value has changed, keep the patch and update the status bool { return paths[i] < paths[j] })
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
		lgm.Patches[path] = compactPatches
	}

	return packedPatches
}

func (lgm *Logument) TemporalTrack(tsi, tsj int64) map[uint64]tsonPatches {
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	if tsi > tsj {
		panic("Start timestamp tsi should be smaller than or equal to end timestamp tsj." +
			"\nStart timestamp: " + strconv.FormatInt(tsi, 10) +
			"\nEnd timestamp: " + strconv.FormatInt(tsj, 10))
	}

	trackedPatches := make(map[uint64]tsonPatches)
	latestValues := make(map[string]any)

	versions := lgm.getSortedVersions("patch")
	for _, version := range versions {
		ps := lgm.Patches[version]
		compactPatches := make(tsonPatches, 0, len(ps))
		for _, p := range ps {
			if p.Timestamp >= tsi && p.Timestamp <= tsj {
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
		}
		if len(compactPatches) > 0 {
			trackedPatches[version] = compactPatches
		}
	}

	return trackedPatches
}

func (lgm *Logument) Set(vk uint64, op tsonpatch.OpType, path string, value any) {
	// Set the value at the target path in the snapshot at the target version
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	if op != tsonpatch.OpReplace && op != tsonpatch.OpAdd {
		return
	}

	if _, exists := lgm.Snapshots[vk]; !exists {
		s := lgm.Snapshot(vk)
		lgm.Snapshots[vk] = s
	}

	snapshot := lgm.Snapshots[vk]

	patch := tsonpatch.Operation{
		Op:        op,
		Path:      path,
		Value:     value,
		Timestamp: time.Now().Unix(), //
	}

	lgm.Patches[vk] = append(lgm.Patches[vk], patch)

	if next, exists := lgm.Snapshots[vk+1]; exists {
		next_snapshot, err := tsonpatch.ApplyPatch(next, tsonPatches{patch})
		if err != nil {
			panic("Failed to make a snapshot with the given version. Error: " + err.Error())
		}
		lgm.Snapshots[vk+1] = next_snapshot
	} else {
		next_snapshot, err := tsonpatch.ApplyPatch(snapshot, tsonPatches{patch})
		if err != nil {
			panic("Failed to make a snapshot with the given version. Error: " + err.Error())
		}
		lgm.Snapshots[vk+1] = next_snapshot
	}
}

func (lgm *Logument) TestSet(vk uint64, op tsonpatch.OpType, path string, value any) {
	// Set the value at the target path in the snapshot at the target timestamp
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	if op != tsonpatch.OpReplace && op != tsonpatch.OpAdd {
		return
	}

	if op == tsonpatch.OpAdd {
		lgm.Set(vk, op, path, value)
	}

	if _, exists := lgm.Snapshots[vk]; !exists {
		s := lgm.Snapshot(vk)
		lgm.Snapshots[vk] = s
	}

	exist_value, err := tson.GetValue(lgm.Snapshots[vk], path)
	if err != nil {
		panic("Failed to get the value from the snapshot. Error: " + err.Error())
	}

	if exist_value == value {
		return
	}

	lgm.Set(vk, op, path, value)

}

func (lgm *Logument) Compact(targetPath string) {
	// Compact the Logument document
	// Remove the patches that have changed only the value without changing the TIMESTAMP ts
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	latestValues := make(map[string]any)

	versions := lgm.getSortedVersions("patch")
	for _, version := range versions {
		ps := lgm.Patches[version]
		compactPatches := make(tsonPatches, 0, len(ps))
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
		lgm.Patches[version] = compactPatches
	}
}

func (lgm *Logument) History(targetPath string) map[string]tsonPatches {
	// Get the history of the changes at the target path
	// The history should contain all the patches that have changed the value at the target path
	// The patches should be sorted by the timestamp in ascending order
	if !lgm.isContinuous() {
		panic("Versions are not continuous.")
	}

	lgm.Compact(targetPath)

	historyPatches := make(map[string]tsonPatches)
	for _, patches := range lgm.Patches {
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
		if err != nil {
			panic("Failed to get the value from the snapshot. Error: " + err.Error())
		}
		if val != nil {
			historyPatches[key] = append([]tsonpatch.Operation{{
				Op:        tsonpatch.OpAdd,
				Path:      key,
				Value:     val,
				Timestamp: 0,
			}},
				historyPatches[key]...) // Add the initial value to the beginning
		}
	}

	return historyPatches
}
