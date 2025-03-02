# **_LOGUMENT_**: "WHEN, What, and How has it happened in Vehicles?"

_Sunghwan Park (Chung-Ang Univ.), Hyun-Gul Roh (42dot Corp.), Sunwoo Na (Chung-Ang Univ.), Sangho Park (Chung-Ang Univ.), Yeseul Chang (Chung-Ang Univ.), and Jaewoo Lee (Chung-Ang Univ.)_

## Contents

1. [What is **_LOGUMENT_**?](#what-is-logument)
    - [Abstract](#abstract)
2. [Implementation of  **_LOGUMENT_**](#implementation-of-logument)
    - [Structure](#structure)
    - [Primitive operations](#primitive-operations)
    - [Supporting operations](#supporting-operations)
    - [Additional supporting operation](#additional-supporting-operation)
3. [About **_TSON_**](#about-tson)
    - [BNF of **_TSON_**](#bnf-of-tson)
    - [VSCode Extension](#vscode-extension)
4. [Usage](#usage)
5. [Contributions](#contributions)

---

## What is **_LOGUMENT_**?

### Abstract

Modern vehicles generate vast amounts of data from a growing number of sensors and actuators, and continuous reflection of evolving vehicle states is highly demanded in the automotive and mobility industries.
Although the digital twin is a key enabling technology, few studies examine how to handle their evolving states.
To address the issue, we first adopt [JSON](https://www.json.org/json-en.html), a widely used document format, and explore its support for logging, synchronization, storing, and temporal queries in digital twins.
While a naive approach of transmitting entire snapshots of JSONs can achieve the synchronization of digital twins, it incurs redundancy and inefficiency.
Instead, transmitting partial changes using JSON Patch improves the efficiency in synchronizing and storing vehicle state.
Meanwhile, most applications in digital twins demand temporal queries about "_WHEN, what, and how has an event occurred in vehicles?_".
To deal with these challenges, we propose **_LOGUMENT_**, a novel data type that integrates _**TSON**_ and _**TSON**_ Patch, which are _Time-Stamped_ JSON and JSON Patch, respectively.
By defining their operations, we materialize not only efficient logging, synchronizing, storing ever-changing data, but also practical temporal queries.
Through discussions on implementation and concurrent synchronization, we demonstrate that _LOGUMENT_ serves as a foundation to manage real-time evolving data from a wider range of IoT devices.

![Digital_twin_usage](./paper/logument.png)
The prospective usages of **_TSON_** and **_LOGUMENT_** in the vehicle digital twin.

---

## Implementation of **_LOGUMENT_**

### Structure

- **Version _[]uint64_**: An array of versions that _LOGUMENT_ manages; should be continuos
- **Snapshot _map[uint64]tson.Tson_**: A map which contains an initial Snapshot (by `Create`) and Snapshots from `Snapshot` Function {version: Snapshot}
- **Patches _map[uint64]tsonpatch.Patch_**: A map of Patches managed internally in _LOGUMENT_ {version: Patches}
- **PatchPool _tsonpatch.Patch_**: Patches to be managed in _LOGUMENT_

### Primitive operations

- **Create(_snapshot tson.Tson, patches tsonpatch.Patch_)**: Make a new _LOGUMENT_ using an initial snapshot (**_Note_**: The function name in the implementation is `NewLogument`)

- **Store(_patches tsonpatch.Patch_)**: Store new patches in the PatchPool temporarily; These patches are queued and will be later integrated into the _LOGUMENT_ state via the `Append` operation

- **Append()**: Incorporate all pending patches from the PatchPool into the Patches, update(increase) the version, and clear the PatchPool

  > 💡 Implementation detail
  >
  > Append, as defined in the _LOGUMENT_ paper, is implemented by sequentially executing `Store` and `Append`.

- **Track(_vi, vj uint64_)**: Extract patches that have changed _Values_ between the vi and the vj, preparing them for transmission

- **Snapshot(_vk uint64_)**: Generate a snapshot representing the _LOGUMENT_ state at the specified version by applying the corresponding patches to the nearest previous snapshot

- **Slice(_vi, vj uint64_)**: Extract a subset of the _LOGUMENT_ that includes all snapshots and patches between the version vi and vj (inclusive)

### Supporting operations

- **Set(_vk uint64, op tsonpatch.OpType, path string, value any_)**: Update the state with a JSON-supported _value_, and create the patch of which _op_ is either `add` or `replace`

- **Unset(_vk uint64, op tsonpatch.OpType, path string_)**: Remove the specific path, and create the patch that _op_ is `remove`

- **TestSet(_vk uint64, op tsonpatch.OpType, path string, value any_)**: `Set` only if _value_ has changed

- **TestUnset(_vk uint64, op tsonpatch.OpType, path string_)**: `Unset` only if _path_ has removed

- **TemporalTrack(_tsi, tsj int64_)**: Extract patches between the tsi and the tsj, enabling to query the evolution and history of data over time

- **TemporalSnapshot(_tsk int64_)**: Create a snapshot based on a target timestamp

- **TemporalSlice(_tsi, tsj int64_)**: Extract a subset of the _LOGUMENT_ document based on the specified start and end timestamps

### Additional supporting operation

- **Compact(_path string_)**: For the specified targetPath, remove patches where only the timestamp has changed (i.e., retain only those patches where the _Value_ has actually been modified)

- **History(_path string_)**: Retrieve the history of changes at the specified target path; This includes all patches that have modified the value at the target path

---

## About **_TSON_**

**_TSON_**, which stands for _Time-Stamped_ JSON, is used to store timestamps as well as data values.

### BNF of **_TSON_**

```ebnf
<tson> ::= <object> | <array> | <value>

<object> ::= "{" <members>? "}"
<members> ::= <pair> ("," <pair>)*
<pair> ::= <string> <timestamp>? ":" <value>

<array> ::= "[" <elements>? "]"
<elements> ::= <value> ("," <value>)*

<value> ::= <primitive> <timestamp>? | <object> | <array> | "null"
<primitive> ::= <string> | <number> | <boolean>

<timestamp> ::= "<" <timestamp_value> ">"
```

### VSCode Extension

There is a Visual Studio Code extension, which enables **_TSON_** syntax highlighting.  
Please refer to [VSCode-TSON-Extension](https://github.com/CAU-CPSS/VSCode-TSON-Extension) repository.

Please check out [**_TSON_** extension for VSCode Marketplace](https://marketplace.visualstudio.com/items?itemName=rollingress.tson)


---

## Usage

`go run main.go [ARGUMENTS...]`

**Parameters:**

- `cars <int>`: \[required\] the number of cars
- `files <int>`: \[required\] the number of files per each car
- `change_rate <float64>`: Change rate for each file
- `size <float64>`: Dataset size ratio (0.5: Use 50% attributes of VSS)
- `dataset <string>`: Path to VSS JSON dataset

---

## Contributions

- Supervising: Hyun-Gul Roh and Jaewoo Lee
- **_LOGUMENT_** interface: Sunghwan Park
- **_TSON & TSON Patch_** implementation: Sunwoo Na ([Karu](https://github.com/karu-rress))
- Data Synchronize Framework: Sunwoo Na ([Karu](https://github.com/karu-rress))

- `vss.json` from [Vehicle Signal Specification](https://github.com/COVESA/vehicle_signal_specification)
