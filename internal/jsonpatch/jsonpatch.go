package jsonpatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	. "github.com/CAU-CPSS/logument/internal/jsonr"
)

var errBadJSONDoc = fmt.Errorf("invalid JSON Document")

// Represents the kind of JSON patch operations.
type OperationType = string

const (
	OpAdd     OperationType = "add"
	OpRemove  OperationType = "remove"
	OpReplace OperationType = "replace"
	// OpMove   OperationType = "move"
	// OpCopy   OperationType = "copy"
	// OpTest   OperationType = "test"
)

// Represents a single JSON patch operation.
type JsonPatchOperation struct {
	Op        OperationType `json:"op"`
	Path      string        `json:"path"`
	Value     any           `json:"value,omitempty"`
	Timestamp int64         `json:"timestamp"`
}

func (j *JsonPatchOperation) Json() string {
	b, _ := json.Marshal(j)
	return string(b)
}

func (j *JsonPatchOperation) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("{")
	b.WriteString(fmt.Sprintf(`"op":"%s"`, j.Op))
	b.WriteString(fmt.Sprintf(`,"path":"%s"`, j.Path))
	// Consider omitting Value for non-nullable operations.
	if j.Value != nil || j.Op == "replace" || j.Op == "add" || j.Op == "test" {
		v, err := json.Marshal(j.Value)
		if err != nil {
			return nil, err
		}
		b.WriteString(`,"value":`)
		b.Write(v)
	}
	b.WriteString(fmt.Sprintf(`,"timestamp":%d`, j.Timestamp))
	b.WriteString("}")
	return b.Bytes(), nil
}

type ByPath []JsonPatchOperation

func (a ByPath) Len() int           { return len(a) }
func (a ByPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPath) Less(i, j int) bool { return a[i].Path < a[j].Path }

func NewPatch(operation, path string, value any, timestamp int64) JsonPatchOperation {
	return JsonPatchOperation{Op: operation, Path: path, Value: value, Timestamp: timestamp}
}

// CreatePatch creates a patch as specified in http://jsonpatch.com/
//
// 'a' is original, 'b' is the modified document. Both are to be given as json encoded content.
// The function will return an array of JsonPatchOperations
//
// An error will be returned if any of the two documents are invalid.
func CreatePatch(a, b []byte) ([]JsonPatchOperation, error) {
	if bytes.Equal(a, b) {
		return []JsonPatchOperation{}, nil
	}

	var aI JsonR
	var bI JsonR
	var err error

	if aI, err = Parse(a); err != nil {
		return nil, err
	}
	if bI, err = Parse(b); err != nil {
		return nil, err
	}

	return handleValues(aI, bI, "", []JsonPatchOperation{})
}

// Returns true if the values matches (must be json types)
// The types of the values must match, otherwise it will always return false
// If two map[string]any are given, all elements must match.
func matchesValue(av, bv any) bool {
	if reflect.TypeOf(av) != reflect.TypeOf(bv) {
		return false
	}
	switch at := av.(type) {
	case Leaf[string]:
		bt := bv.(Leaf[string]).Value
		if bt == at.Value {
			return true
		}
	case Leaf[float64]:
		bt := bv.(Leaf[float64]).Value
		if bt == at.Value {
			return true
		}
	case Leaf[bool]:
		bt := bv.(Leaf[bool]).Value
		if bt == at.Value {
			return true
		}
	case Object:
		bt := bv.(Object)
		for key := range at {
			if !matchesValue(at[key], bt[key]) {
				return false
			}
		}
		for key := range bt {
			if !matchesValue(at[key], bt[key]) {
				return false
			}
		}
		return true
	case Array:
		bt := bv.(Array)
		if len(bt) != len(at) {
			return false
		}
		for key := range at {
			if !matchesValue(at[key], bt[key]) {
				return false
			}
		}
		for key := range bt {
			if !matchesValue(at[key], bt[key]) {
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

func makePath(path string, newPart any) string {
	key := rfc6901Encoder.Replace(fmt.Sprintf("%v", newPart))
	if path == "" {
		return "/" + key
	}
	if strings.HasSuffix(path, "/") {
		return path + key
	}
	return path + "/" + key
}

// diff returns the (recursive) difference between a and b as an array of JsonPatchOperations.
func diff(a, b Object, path string, patch []JsonPatchOperation) ([]JsonPatchOperation, error) {
	for key, bv := range b {
		p := makePath(path, key)
		av, ok := a[key]
		// value was added
		if !ok {
			switch bv.(type) {
			case Leaf[string]:
				bm, _ := bv.(Leaf[string])
				patch = append(patch, NewPatch("add", p, bm.Value, bm.Timestamp))
			case Leaf[float64]:
				bm, _ := bv.(Leaf[float64])
				patch = append(patch, NewPatch("add", p, bm.Value, bm.Timestamp))
			case Leaf[bool]:
				bm, _ := bv.(Leaf[bool])
				patch = append(patch, NewPatch("add", p, bm.Value, bm.Timestamp))
			default:
				panic(fmt.Sprintf("diff(): Unknown type %T for bv", bv))
			}
			continue
		}
		// If types have changed, replace completely
		if reflect.TypeOf(av) != reflect.TypeOf(bv) {
			switch bv.(type) {
			case Leaf[string]:
				bm, _ := bv.(Leaf[string])
				patch = append(patch, NewPatch("replace", p, bm.Value, bm.Timestamp))
			case Leaf[float64]:
				bm, _ := bv.(Leaf[float64])
				patch = append(patch, NewPatch("replace", p, bm.Value, bm.Timestamp))
			case Leaf[bool]:
				bm, _ := bv.(Leaf[bool])
				patch = append(patch, NewPatch("replace", p, bm.Value, bm.Timestamp))
			default:
				panic(fmt.Sprintf("diff(): Unknown type %T for bv", bv))
			}
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

			am, _ := a[key]
			switch am.(type) {
			case Leaf[string]:
				patch = append(patch, NewPatch("remove", p, nil, am.(Leaf[string]).Timestamp))
			case Leaf[float64]:
				patch = append(patch, NewPatch("remove", p, nil, am.(Leaf[float64]).Timestamp))
			case Leaf[bool]:
				patch = append(patch, NewPatch("remove", p, nil, am.(Leaf[bool]).Timestamp))
			default:
				panic(fmt.Sprintf("diff(): Unknown type %T for am", am))
			}
		}
	}
	return patch, nil
}

func handleValues(av, bv any, p string, patch []JsonPatchOperation) ([]JsonPatchOperation, error) {
	var err error
	switch at := av.(type) {
	case Leaf[string]:
		if !matchesValue(av, bv) {
			patch = append(patch, NewPatch("replace", p, bv.(Leaf[string]).Value, bv.(Leaf[string]).Timestamp))
		}
	case Leaf[float64]:
		if !matchesValue(av, bv) {
			patch = append(patch, NewPatch("replace", p, bv.(Leaf[float64]).Value, bv.(Leaf[float64]).Timestamp))
		}
	case Leaf[bool]:
		if !matchesValue(av, bv) {
			patch = append(patch, NewPatch("replace", p, bv.(Leaf[bool]).Value, bv.(Leaf[bool]).Timestamp))
		}
	case Object:
		bt := bv.(Object)
		patch, err = diff(at, bt, p, patch)
		if err != nil {
			return nil, err
		}
	case Array:
		bt, ok := bv.(Array)
		if !ok {
			// array replaced by non-array
			patch = append(patch, NewPatch("replace", p, bv.(map[string]any)["Value"], bv.(map[string]any)["Timestamp"].(int64)))
		} else if len(at) != len(bt) {
			// arrays are not the same length
			patch = append(patch, compareArray(at, bt, p)...)

		} else {
			for i := range bt {
				patch, err = handleValues(at[i], bt[i], makePath(p, i), patch)
				if err != nil {
					return nil, err
				}
			}
		}
	case nil:
		switch bv.(type) {
		case nil:
			// Both nil, fine.
		default:
			// Replace nil with value
			patch = append(patch, NewPatch("add", p, bv.(map[string]any)["Value"], bv.(map[string]any)["Timestamp"].(int64)))
		}
	default:
		panic(fmt.Sprintf("handleValues(): Unknown type %T for av", av))
	}
	return patch, nil
}

// compareArray generates remove and add operations for `av` and `bv`.
func compareArray(av, bv Array, p string) []JsonPatchOperation {
	retval := []JsonPatchOperation{}

	// Find elements that need to be removed
	processArray(av, bv, func(i int, value interface{}) {
		retval = append(retval, NewPatch("remove", makePath(p, i), nil, -1))
	})
	reversed := make([]JsonPatchOperation, len(retval))
	for i := 0; i < len(retval); i++ {
		reversed[len(retval)-1-i] = retval[i]
	}
	retval = reversed
	// Find elements that need to be added.
	// NOTE we pass in `bv` then `av` so that processArray can find the missing elements.
	processArray(bv, av, func(i int, value interface{}) {
		retval = append(retval, NewPatch("add", makePath(p, i), value.(map[string]any)["Value"], value.(map[string]any)["Timestamp"].(int64)))
	})

	return retval
}

// processArray processes `av` and `bv` calling `applyOp` whenever a value is absent.
// It keeps track of which indexes have already had `applyOp` called for and automatically skips them so you can process duplicate objects correctly.
func processArray(av, bv Array, applyOp func(i int, value interface{})) {
	foundIndexes := make(map[int]struct{}, len(av))
	reverseFoundIndexes := make(map[int]struct{}, len(av))
	for i, v := range av {
		for i2, v2 := range bv {
			if _, ok := reverseFoundIndexes[i2]; ok {
				// We already found this index.
				continue
			}
			if reflect.DeepEqual(v, v2) {
				// Mark this index as found since it matches exactly.
				foundIndexes[i] = struct{}{}
				reverseFoundIndexes[i2] = struct{}{}
				break
			}
		}
		if _, ok := foundIndexes[i]; !ok {
			applyOp(i, v)
		}
	}
}
