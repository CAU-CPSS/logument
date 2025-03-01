# **_LOGUMENT_**

---

## What is **_LOGUMENT_**?

### Abstract

Modern vehicles generate vast amounts of data from a growing number of sensors and actuators, and continuous reflection of evolving vehicle states is highly demanded in the automotive and mobility industries.
Although the digital twin is a key enabling technology, few studies examine how to handle their evolving states.
To address the issue, we first adopt JSON, a widely used document format, and explore its support for logging, synchronization, storing, and temporal queries in digital twins.
While a naive approach of transmitting entire snapshots of JSONs can achieve the synchronization of digital twins, it incurs redundancy and inefficiency.
Instead, transmitting partial changes using JSON Patch improves the efficiency in synchronizing and storing vehicle state.
Meanwhile, most applications in digital twins demand temporal queries about "_WHEN, what, and how has an event occurred in vehicles?_".
To deal with these challenges, we propose **_LOGUMENT_**, a novel data type that integrates TSON and TSON Patch, which _Time-Stamped_ JSON and JSON Patch, respectively.
By defining their operations, we materialize not only efficient logging, synchronizing, storing ever-changing data, but also practical temporal queries.
Through discussions on implementation and concurrent synchronization, we demonstrate that _LOGUMENT_ serves as a foundation to manage real-time evolving data from a wider range of IoT devices.

---

### Structure

- **Version** []uint64: An array of versions that **_LOGUMENT_** manages; should be continuos
- **Snapshot** map[uint64]tson.Tson: A map which contains an initial Snapshot (by `Create`) and Snapshots from `Snapshot` Function (version, Snapshot)
- **Patches** map[uint64]tsonpatch.Patch: A map of Patches managed internally in _LOGUMENT_ (version, Patches)
- **PatchPool** tsonpatch.Patch: Patches to be managed in **_LOGUMENT_**

---

## Interface for **_LOGUMENT_**

### Primitive operations

- **Create(snapshot tson.Tson, patches jsonpatch.Patch)**: Make a new _LOGUMENT_ using an initial snapshot (**_Note_**: The function name in the implementation is `NewLogument`)

- **Store(patches jsonpatch.Patch)**: Store new patches in the PatchPool temporarily; These patches are queued and will be later integrated into the Logument state via the `Append` operation

- **Append()**: Incorporate all pending patches from the PatchPool into the Patches, update(increase) the version, and clear the PatchPool

> 💡 Implementation detail
>
> Append, as defined in the _LOGUMENT_ paper, is implemented by sequentially executing `Store` and `Append`.

- **Track(vi, vj uint64)**: Extract patches that have changed _Values_ between the vi and the vj, preparing them for transmission

- **Snapshot(vk uint64)**: Generate a snapshot representing the _LOGUMENT_ state at the specified version by applying the corresponding patches to the nearest previous snapshot

- **Slice(vi, vj uint64)**: Extract a subset of the _LOGUMENT_ that includes all snapshots and patches between the version vi and vj (inclusive)

### Supporting operations

- **Set**:

- **Unset**:

- **TestSet**:

- **TestUnset**:

- **TemporalTrack(tsi, tsj int64)**: Extract patches between the tsi and the tsj, enabling to query the evolution and history of data over time

- **TemporalSnapshot(tsk int64)**: Create a snapshot based on a target timestamp

- **TemporalSlice(tsi, tsj int64)**: Extract a subset of the _LOGUMENT_ document based on the specified start and end timestamps

### Additional supporting operation

- **Compact(targetPath string)**: For the specified targetPath, remove patches where only the timestamp has changed (i.e., retain only those patches where the _Value_ has actually been modified)

- **History(targetPath string)**: Retrieve the history of changes at the specified target path; This includes all patches that have modified the value at the target path

---

## About TSON

For more about **_TSON_**, please refer to [internal/tson/README.md].

---

## Contribute

- **_LOGUMENT_** interface: Sunghwan Park
- **_TSON & TSON Patch_** implementation: Sunwoo Na ([Karu](https://github.com/karu-rress))
- Data Synchronize Framework: Sunwoo Na ([Karu](https://github.com/karu-rress))

- `vss.json` from [Vehicle Signal Specification](https://github.com/COVESA/vehicle_signal_specification)`