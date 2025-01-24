package jsonpatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/CAU-CPSS/logument/internal/jsonr"
)

// Represents the kind of JSON patch operations.
type OpType string

// Enums for OperationType
const (
	OpAdd     OpType = "add"
	OpRemove  OpType = "remove"
	OpReplace OpType = "replace"
	OpMove    OpType = "move"
	OpCopy    OpType = "copy"
	OpTest    OpType = "test"
)

// Represents a single JSON patch operation.
type PatchOperation struct {
	Op        OpType `json:"op"`
	Path      string `json:"path"`
	Value     any    `json:"value,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// ToString converts the PatchOperation to a JSON string.
func (p *PatchOperation) ToString() string {
	// NOTE: Using vanila JSON here
	b, _ := json.Marshal(p)
	return string(b)
}

// Marshal converts the PatchOperation to a JSON byte Array.
func (p *PatchOperation) Marshal() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf(`{ "op":"%s", "path":"%s"`, p.Op, p.Path))
	if p.Value != nil || p.Op == OpReplace || p.Op == OpAdd || p.Op == OpTest {
		v, err := json.Marshal(p.Value)
		if err != nil {
			return nil, err
		}
		b.WriteString(`,"value":`)
		b.Write(v)
	}
	b.WriteString(fmt.Sprintf(`,"timestamp":%d }`, p.Timestamp))
	return b.Bytes(), nil
}

// NewPatchOperation creates a new PatchOperation instance.
func NewPatchOperation(operation OpType, path string, value any, timestamp int64) PatchOperation {
	return PatchOperation{operation, path, value, timestamp}
}

// GeneratePatch generates a JSON patch from two JSON-R documents
func GeneratePatch(origin, modified jsonr.JsonR) ([]PatchOperation, error) {
	// If the two JSON-R documents are equal, return an empty patch
	if eq, err := jsonr.Equal(origin, modified); err != nil && eq {
		return []PatchOperation{}, nil
	}

	return handleValues(origin, modified, "", []PatchOperation{})
}

// Returns true if the values matches (must be json types)
// The types of the values must match, otherwise it will always return false
// If two map[string]any are given, all elements must match.
func matchesValue(origin, modified jsonr.Value) bool {
	if reflect.TypeOf(origin) != reflect.TypeOf(modified) {
		return false
	}

	switch at := origin.(type) {
	case jsonr.Leaf[string]:
		if modified.(jsonr.Leaf[string]).Value == at.Value {
			return true
		}
	case jsonr.Leaf[float64]:
		if modified.(jsonr.Leaf[float64]).Value == at.Value {
			return true
		}
	case jsonr.Leaf[bool]:
		if modified.(jsonr.Leaf[bool]).Value == at.Value {
			return true
		}
	case jsonr.Object:
		bt := modified.(jsonr.Object)
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
	case jsonr.Array:
		bt := modified.(jsonr.Array)
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

// diff returns the (recursive) difference between a and b as an jsonr.Array of JsonPatchOperations.
func diff(a, b jsonr.Object, path string, patch []PatchOperation) ([]PatchOperation, error) {
	for key, bv := range b {
		p := makePath(path, key)
		av, ok := a[key]
		// value was added
		if !ok {
			switch bm := bv.(type) {
			case jsonr.Leaf[string]:
				patch = append(patch, NewPatchOperation(OpAdd, p, bm.Value, bm.Timestamp))
			case jsonr.Leaf[float64]:
				patch = append(patch, NewPatchOperation(OpAdd, p, bm.Value, bm.Timestamp))
			case jsonr.Leaf[bool]:
				patch = append(patch, NewPatchOperation(OpAdd, p, bm.Value, bm.Timestamp))
			default:
				panic(fmt.Sprintf("diff(): Unknown type %T for bv", bv))
			}
			continue
		}
		// If types have changed, replace completely
		if reflect.TypeOf(av) != reflect.TypeOf(bv) {
			switch bm := bv.(type) {
			case jsonr.Leaf[string]:
				patch = append(patch, NewPatchOperation(OpReplace, p, bm.Value, bm.Timestamp))
			case jsonr.Leaf[float64]:
				patch = append(patch, NewPatchOperation(OpReplace, p, bm.Value, bm.Timestamp))
			case jsonr.Leaf[bool]:
				patch = append(patch, NewPatchOperation(OpReplace, p, bm.Value, bm.Timestamp))
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
			switch leaf := am.(type) {
			case jsonr.Leaf[string]:
				patch = append(patch, NewPatchOperation(OpRemove, p, nil, leaf.Timestamp))
			case jsonr.Leaf[float64]:
				patch = append(patch, NewPatchOperation(OpRemove, p, nil, leaf.Timestamp))
			case jsonr.Leaf[bool]:
				patch = append(patch, NewPatchOperation(OpRemove, p, nil, leaf.Timestamp))
			case jsonr.Object:
				patch = append(patch, NewPatchOperation(OpRemove, p, nil, -1))
			case jsonr.Array:
				patch = append(patch, NewPatchOperation(OpRemove, p, nil, -1))
			default:
				panic(fmt.Sprintf("diff(): Unknown type %T for am", am))
			}
		}
	}
	return patch, nil
}

func handleValues(av, bv jsonr.JsonR, p string, patch []PatchOperation) ([]PatchOperation, error) {
	var err error
	switch at := av.(type) {
	case jsonr.Leaf[string]:
		if !matchesValue(av, bv) {
			patch = append(patch, NewPatchOperation(OpReplace, p, bv.(jsonr.Leaf[string]).Value, bv.(jsonr.Leaf[string]).Timestamp))
		}
	case jsonr.Leaf[float64]:
		if !matchesValue(av, bv) {
			patch = append(patch, NewPatchOperation(OpReplace, p, bv.(jsonr.Leaf[float64]).Value, bv.(jsonr.Leaf[float64]).Timestamp))
		}
	case jsonr.Leaf[bool]:
		if !matchesValue(av, bv) {
			patch = append(patch, NewPatchOperation(OpReplace, p, bv.(jsonr.Leaf[bool]).Value, bv.(jsonr.Leaf[bool]).Timestamp))
		}
	case jsonr.Object:
		bt := bv.(jsonr.Object)
		patch, err = diff(at, bt, p, patch)
		if err != nil {
			return nil, err
		}
	case jsonr.Array:
		bt, ok := bv.(jsonr.Array)
		if !ok {
			var bk any = bv
			// jsonr.Array replaced by non-jsonr.Array
			patch = append(patch, NewPatchOperation(OpReplace, p, bk.(map[string]any)["Value"], bk.(map[string]any)["Timestamp"].(int64)))
		} else if len(at) != len(bt) {
			// jsonr.Arrays are not the same length
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
			var bk any = bv
			patch = append(patch, NewPatchOperation(OpAdd, p, bk.(map[string]any)["Value"], bk.(map[string]any)["Timestamp"].(int64)))
		}
	default:
		panic(fmt.Sprintf("handleValues(): Unknown type %T for av", av))
	}
	return patch, nil
}

// compareArray generates remove and add operations for `av` and `bv`.
func compareArray(av, bv jsonr.Array, p string) []PatchOperation {
	retval := []PatchOperation{}

	// Find elements that need to be removed
	processArray(av, bv, func(i int, value any) {
		retval = append(retval, NewPatchOperation(OpRemove, makePath(p, i), nil, -1))
	})
	reversed := make([]PatchOperation, len(retval))
	for i := 0; i < len(retval); i++ {
		reversed[len(retval)-1-i] = retval[i]
	}
	retval = reversed
	// Find elements that need to be added.
	// NOTE we pass in `bv` then `av` so that processArray can find the missing elements.
	processArray(bv, av, func(i int, value any) {
		switch leaf := value.(type) {
		case jsonr.Leaf[string]:
			retval = append(retval, NewPatchOperation(OpAdd, makePath(p, i), leaf.Value, leaf.Timestamp))

		case jsonr.Leaf[float64]:
			retval = append(retval, NewPatchOperation(OpAdd, makePath(p, i), leaf.Value, leaf.Timestamp))

		case jsonr.Leaf[bool]:
			retval = append(retval, NewPatchOperation(OpAdd, makePath(p, i), leaf.Value, leaf.Timestamp))

		default:
			panic(fmt.Sprintf("compareArray(): Unknown type %T for value", value))
		}
	})

	return retval
}

// processArray processes `av` and `bv` calling `applyOp` whenever a value is absent.
// It keeps track of which indexes have already had `applyOp` called for and automatically skips them so you can process duplicate jsonr.Objects correctly.
func processArray(av, bv jsonr.Array, applyOp func(i int, value any)) {
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
