# This is temporary **_Logument_** repository

---

## What is **_Logument_**?

- Structure
  - Version []uint64: Logument가 관리하는 버전의 배열
  - Snapshot map[uint64]jsonr.JsonR: 최초 Create 시와 Snapshot 함수에 의해 생성된 Snapshot의 map (version, Snapshot)
  - PatchMap map[uint64]jsonpatch.Patch: Logument 내에서 관리하는 Patch의 map (version, Patches)
  - PatchPool jsonpatch.Patch: Logument에서 관리 예정인 Patches

---

## Interface for **_Logument_**

- **Create(snapshot jsonr.JsonR, patches jsonpatch.Patch)**: make a new Logument (function name in implementation is 'NewLogument')
- **Store(patches jsonpatch.Patch)**: Queue new patches in the PatchPool before they are managed in Logument
- **Apply()**: Move pending patches from PatchPool to PatchMap, and then increase(and append) the version
- **Snapshot(targetVersion)**: Make a snapshot at target version
- Compact
- Slice
- Pack
- History
- Next
- TimeSnapshot
- TimeSlice

---

## Contribute

- **_Logument_** interface: Sunghwan Park
- JSON Patch implementation: Sunwoo Na (Karu)
- Data Synchronize Framework: Sunwoo Na (Karu)
