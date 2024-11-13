package jsonpatch

// TODO:
// 1. change interface{} to map[string]interface{}
//   vanila JSON - can stores atomic value, such as "name": "karu"
//   but, in our case, we need to store timestamp as well
//   so, we need to store it as map[string]interface{}{"Value": "karu", "Timestamp": 1234567890}
//   HOWEVER, the "value" in patch should be atomic value, not map[string]interface{}
//
// 2. add timestamp to the patch
//   if the value is same, but the timestamp is different, it should be ignored
//   if the value is different, it should be updated with the right after timestamp
//   ex)
// 	   {"op": "replace", "path": "/name", "value": "karu", "timestamp": 1234567890}
//
// ???: Do wee need a separate leaf structure for JSON logument?
//   ex)
//   type Leaf struct {
//     Value     interface{} `json:"Value"`
//     Timestamp int64       `json:"Timestamp"`
//   }

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

const (
	TEMP_TIMESTAMP = 999999

	VALUE_KEY     = "Value"
	TIMESTAMP_KEY = "Timestamp"
)

var errBadJSONDoc = fmt.Errorf("invalid JSON Document")

type JsonPatchOperation = Operation

// Operation represents a single operation in a JSON Patch document.
type Operation struct {
	Operation string      `json:"op"`
	Path      string      `json:"path"`
	Value     interface{} `json:"value,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
}

func (j *Operation) Json() string {
	b, _ := json.Marshal(j)
	return string(b)
}

func (j *Operation) MarshalJSON() ([]byte, error) {
	// Ensure for add and replace we emit `value: null`
	if j.Value == nil && (j.Operation == "replace" || j.Operation == "add") {
		return json.Marshal(struct {
			Operation string      `json:"op"`
			Path      string      `json:"path"`
			Value     interface{} `json:"value"`
			Timestamp int64       `json:"timestamp"`
		}{
			Operation: j.Operation,
			Path:      j.Path,
		})
	}
	// otherwise just marshal normally. We cannot literally do json.Marshal(j) as it would be recursively
	// calling this function.
	return json.Marshal(Operation{
		Operation: j.Operation,
		Path:      j.Path,
		Value:     j.Value,
		Timestamp: j.Timestamp,
	})
}

type ByPath []Operation

func (a ByPath) Len() int           { return len(a) }
func (a ByPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPath) Less(i, j int) bool { return a[i].Path < a[j].Path }

func NewOperation(op, path string, value interface{}, timestamp int64) Operation {
	return Operation{Operation: op, Path: path, Value: value, Timestamp: timestamp}
}

// CreatePatch creates a patch as specified in http://jsonpatch.com/
//
// 'a' is original, 'b' is the modified document. Both are to be given as json encoded content.
// The function will return an array of JsonPatchOperations
//
// An error will be returned if any of the two documents are invalid.
func CreatePatch(a, b []byte) ([]Operation, error) {
	if bytes.Equal(a, b) {
		return []Operation{}, nil
	}
	var aI, bI interface{}
	if err := json.Unmarshal(a, &aI); err != nil {
		return nil, errBadJSONDoc
	}
	if err := json.Unmarshal(b, &bI); err != nil {
		return nil, errBadJSONDoc
	}
	return handleValues(aI, bI, "", []Operation{})
}

// Returns true if the values matches (must be json types)
// The types of the values must match, otherwise it will always return false
// If two map[string]interface{} are given, all elements must match.
func OLDmatchesValue(av, bv interface{}) bool {
	if reflect.TypeOf(av) != reflect.TypeOf(bv) {
		return false
	}
	switch at := av.(type) {
	case string:
		if bt, ok := bv.(string); ok && bt == at {
			return true
		}
	case float64: // JSON only has one number type
		if bt, ok := bv.(float64); ok && bt == at {
			return true
		}
	case bool:
		if bt, ok := bv.(bool); ok && bt == at {
			return true
		}
	case map[string]interface{}:
		bt, ok := bv.(map[string]interface{})
		if !ok {
			return false
		}
		for key := range at {
			if !OLDmatchesValue(at[key], bt[key]) {
				return false
			}
		}
		for key := range bt {
			if !OLDmatchesValue(at[key], bt[key]) {
				return false
			}
		}
		return true
	case []interface{}:
		bt, ok := bv.([]interface{})
		if !ok {
			return false
		}
		if len(bt) != len(at) {
			return false
		}
		for key := range at {
			if !OLDmatchesValue(at[key], bt[key]) {
				return false
			}
		}
		for key := range bt {
			if !OLDmatchesValue(at[key], bt[key]) {
				return false
			}
		}
		return true
	}
	return false
}

func matchesValue(av, bv interface{}) bool {
	if reflect.TypeOf(av) != reflect.TypeOf(bv) {
		return false
	}
	isAvLogument := false
	isAvValueString := false
	isAvValueNumber := false
	isAvValueBoolean := false 
	var avValue interface{} = nil
	avmap, isAvMap := av.(map[string]interface{})
	avlist, isAvList := av.([]interface{})
	{
		hasAvValue, hasAvTimestamp := false, false

		if isAvMap {
			avValue, hasAvValue = avmap[VALUE_KEY]
			_, hasAvTimestamp = avmap[TIMESTAMP_KEY]
			if hasAvValue {
				switch avValue.(type) {
				case string:
					isAvValueString = true
				case float64:
					isAvValueNumber = true
				case bool:
					isAvValueBoolean = true
				}
			}
		}

		isAvLogument = isAvMap && hasAvValue && hasAvTimestamp
	}
	switch { // at := av.(type)
	case isAvLogument && isAvValueString:
		if bm := bv.(map[string]interface{}); bm != nil {
			if bstr, ok := bm[VALUE_KEY].(string); ok && bstr == avValue {
				return true
			}
		}
	case isAvLogument && isAvValueNumber:
		if bm := bv.(map[string]interface{}); bm != nil {
			if bnum, ok := bm[VALUE_KEY].(float64); ok && bnum == avValue {
				return true
			}
		}
	case isAvLogument && isAvValueBoolean:
		if bm := bv.(map[string]interface{}); bm != nil {
			if bbool, ok := bm[VALUE_KEY].(bool); ok && bbool == avValue {
				return true
			}
		}
	case isAvMap:
		bt, ok := bv.(map[string]interface{})
		if !ok {
			return false
		}
		for key := range avmap {
			if !matchesValue(avmap[key], bt[key]) {
				return false
			}
		}
		for key := range bt {
			if !matchesValue(avmap[key], bt[key]) {
				return false
			}
		}
		return true
	case isAvList:
		bt, ok := bv.([]interface{})
		if !ok {
			return false
		}
		if len(bt) != len(avlist) {
			return false
		}
		for key := range avlist {
			if !matchesValue(avlist[key], bt[key]) {
				return false
			}
		}
		for key := range bt {
			if !matchesValue(avlist[key], bt[key]) {
				return false
			}
		}
		return true
	}
	return false
}

// From http://tools.ietf.org/html/rfc6901#section-4 :
//
// Evaluation of each reference token begins by decoding any escaped
// character sequence.  This is performed by first transforming any
// occurrence of the sequence '~1' to '/', and then transforming any
// occurrence of the sequence '~0' to '~'.
//   TODO decode support:
//   var rfc6901Decoder = strings.NewReplacer("~1", "/", "~0", "~")

var rfc6901Encoder = strings.NewReplacer("~", "~0", "/", "~1")

func makePath(path string, newPart interface{}) string {
	key := rfc6901Encoder.Replace(fmt.Sprintf("%v", newPart))
	if path == "" {
		return "/" + key
	}
	return path + "/" + key
}

// diff returns the (recursive) difference between a and b as an array of JsonPatchOperations.
// if leaf value is same and only the timestamp is different, it will be ignored
func diff(a, b map[string]interface{}, path string, patch []Operation) ([]Operation, error) {
	for key, bv := range b {
		p := makePath(path, key)
		av, ok := a[key]
		// value was added
		if !ok {
			patch = append(patch, NewOperation("add", p, bv[VALUE_KEY], bv[TIMESTAMP_KEY])) // TODO: check needed
			continue
		}
		// Types are the same, compare values
		var err error
		patch, err = handleValues(av, bv, p, patch)
		if err != nil {
			return nil, err
		}
	}
	// Now add all deleted values as nil
	for key := range a {
		_, found := b[key]
		if !found {
			p := makePath(path, key)

			patch = append(patch, NewOperation("remove", p, nil, TEMP_TIMESTAMP))
		}
	}
	return patch, nil
}

func handleValues(av, bv interface{}, p string, patch []Operation) ([]Operation, error) {
	{
		at := reflect.TypeOf(av)
		bt := reflect.TypeOf(bv)
		if at == nil && bt == nil {
			// do nothing
			return patch, nil
		} else if at != bt {
			// If types have changed, replace completely (preserves null in destination)
			return append(patch, NewOperation("replace", p, bv, TEMP_TIMESTAMP)), nil
		}
	}

	var err error
	switch at := av.(type) {
	case map[string]interface{}:
		bt := bv.(map[string]interface{})
		patch, err = diff(at, bt, p, patch)
		if err != nil {
			return nil, err
		}
	case string, float64, bool:
		if !matchesValue(av, bv) {
			patch = append(patch, NewOperation("replace", p, bv, TEMP_TIMESTAMP))
		}
	case []interface{}:
		bt := bv.([]interface{})
		n := min(len(at), len(bt))
		for i := len(at) - 1; i >= n; i-- {
			patch = append(patch, NewOperation("remove", makePath(p, i), nil, TEMP_TIMESTAMP))
		}
		for i := n; i < len(bt); i++ {
			patch = append(patch, NewOperation("add", makePath(p, i), bt[i], TEMP_TIMESTAMP))
		}
		for i := 0; i < n; i++ {
			var err error
			patch, err = handleValues(at[i], bt[i], makePath(p, i), patch)
			if err != nil {
				return nil, err
			}
		}
	default:
		panic(fmt.Sprintf("Unknown type:%T ", av))
	}
	return patch, nil
}

func min(x int, y int) int {
	if y < x {
		return y
	}
	return x
}
