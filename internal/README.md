# Internal Modules of **_LOGUMENT_**

## Modules

- **_TSON_**: The new data structure that replaces JSON, containing a timestamp
- **_TSON Patch_**: Replaces JSON Patch by using **_TSON_** instead of JSON
- **_VSS Generator_**: Generates VSS dataset with the form of **_TSON_** and **_TSON Patch_**

## Interactions among **_TSON_** Implementations

The brief explanation about TSON can be found in the [Main README.md file](/README.md).

![TSON Relationship](/Assets/tson.png)

- `JSON ([]byte)`: The `*.json` file itself or string / byte array of JSON data. It can be converted to JSON object (type `any`) via Go standard library.
    - Convertible to: `JSON (any)`, `TSON (struct)`
- `JSON (any)`: Unmarshaled JSON data by `json.Unmarshal`. Its type is `any` (formerly `interface{}`)
    - Convertible to: `JSON ([]byte)`, `TSON (struct)`
- `TSON ([]byte<>)`: The `<>` stands for **_TSON_**'s unique symbol, in which the timestamps are stored. This represents a `*.tson` file itself or string / byte array of **_TSON_** with `<>`.
    - Convertible to: `TSON (struct)`
- `TSON (struct)`: The most universal type in **_TSON_** system. Contains the **_TSON_** data using **_TSON_** struct. Can be converted into any type of **_TSON_** system.
    - Convertible to: `JSON ([]byte)`, `JSON (any)`, `TSON ([]byte<>)`, `Compatible TSON ([]byte)`, `Compatible TSON (any)`
- `Compatible TSON ([]byte)`: Compatible TSON is virtually JSON, which has `{ "value", "timestamp" }` object as a leaf node. It can be read by JSON parsers.
    - Convertible to: `TSON (struct)`, `Compatible TSON (any)`
- `Compatible TSON (any)`: Same as `JSON (any)`, but contains `{ "value", "timestamp" }` as a leaf node.
    - Convertible to: `Compatible TSON ([]byte)`
