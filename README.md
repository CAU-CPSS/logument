# This is temporary **_Logument_** repository

---

## What is **_Logument_**?

- Structure
  - Version []uint64: Logument가 관리하는 버전의 배열; should be continuos
  - Snapshot map[uint64]tson.Tson: 최초 Create 시와 Snapshot 함수에 의해 생성된 Snapshot의 map (version, Snapshot)
  - PatchMap map[uint64]jsonpatch.Patch: Logument 내에서 관리하는 Patch의 map (version, Patches)
  - PatchPool jsonpatch.Patch: Logument에서 관리 예정인 Patches

---

## Interface for **_Logument_**

- **Create(snapshot tson.Tson, patches jsonpatch.Patch)**: make a new Logument using an initial snapshot (**_Note_**: The function name in the implementation is `NewLogument`)

- **Store(patches jsonpatch.Patch)**: Append new patches to the PatchPool; These patches are queued and will be later integrated into the Logument state via the `Apply` operation

- **Apply()**: Incorporate all pending patches from the PatchPool into the PatchMap, update(increase) the version, and clear the PatchPool

- **Snapshot(targetVersion uint64)**: Generate a snapshot representing the Logument state at the specified version by applying the corresponding patches to the nearest previous snapshot

- **Compact(targetPath string)**: For the specified targetPath, remove patches where only the timestamp has changed (i.e., retain only those patches where the _Value_ has actually been modified)

- **Slice(startVersion uint64, endVersion uint64)**: Extract a subset of the Logument that includes all snapshots and patches between the start and end versions (inclusive)

- **Pack(targetVersion uint64)**: Extract patches that have changed _Values_ between the target version and the latest version, preparing them for transmission

- **History(targetPath string)**: Retrieve the history of changes at the specified target path; This includes all patches that have modified the value at the target path

- **TimeSnapshot(targetTime int64)**: Create a snapshot based on a target timestamp

- **TimeSlice(startTime, endTime int64)**: Extract a subset of the Logument document based on the specified start and end timestamps

---

## Contribute

- **_Logument_** interface: Sunghwan Park
- TSON & JSON Patch implementation: Sunwoo Na ([Karu](https://github.com/karu-rress))
- Data Synchronize Framework: Sunwoo Na ([Karu](https://github.com/karu-rress))
