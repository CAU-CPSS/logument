//
// tsonpatch.go
//
// Defines the TSON patch type and provides functions
// for TSON patching and generation.
//
// The TSON patch is a document that describes
// changes to be made to a TSON document.
// It is represented as an array of operations.
// Each operation describes a single change,
// and only recorded when the value has changed,
// not when the timestamp has only changed.
//
// Author: Karu (@karu-rress)
//

package tsonpatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/CAU-CPSS/logument/internal/tson"
)

//////////////////////////////////
///////// JSON PATCH
//////////////////////////////////

// Represents the kind of TSON patch operations.
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

// Patch represents a TSON patch document.
type Patch []Operation

// String converts the Patch to a JSON string, with formatting.
func (p *Patch) String() string {
	lines := make([]string, len(*p))
	for i, op := range *p {
		lines[i] = "    " + op.String()
	}
	return fmt.Sprintf("[\n%s\n]", strings.Join(lines, ",\n"))
}

func FromJson(j any) (Patch, error) {
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return Unmarshal(b)
}

// Unmarshal converts a JSON byte array to a Patch.
func Unmarshal(b []byte) (Patch, error) {
	var patch Patch
	if err := json.Unmarshal(b, &patch); err != nil {
		return nil, err
	}
	return patch, nil
}

//////////////////////////////////
///////// TSON PATCH OPERATIONS
//////////////////////////////////

// Operation represents a single TSON patch operation.
type Operation struct {
	Op        OpType `json:"op"`
	Path      string `json:"path"`
	Value     any    `json:"value,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// String converts the Operation to a JSON string.
func (p *Operation) String() string {
	var ( // Using vanila JSON here
		op        = p.Op
		path      = p.Path
		value, _  = json.Marshal(p.Value)
		timestamp = p.Timestamp
		buf       bytes.Buffer
	)
	fmt.Fprintf(&buf,
		`{ "op": "%s", "path": "%s", "value": %s, "timestamp": %d }`,
		op, path, value, timestamp)
	return buf.String()
}

// Marshal converts the Operation to a JSON byte Array.
func (p *Operation) Marshal() (b []byte, err error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `{ "op": "%s", "path": "%s"`, p.Op, p.Path)
	if p.Value != nil || p.Op == OpReplace || p.Op == OpAdd || p.Op == OpTest {
		if v, err := json.Marshal(p.Value); err != nil {
			return nil, err
		} else {
			fmt.Fprintf(&buf, `, "value": %s`, v)
		}
	}
	fmt.Fprintf(&buf, `, "timestamp": %d }`, p.Timestamp)
	return buf.Bytes(), nil
}

// NewOperation creates a new Operation instance.
func NewOperation(op OpType, path string, value any, timestamp int64) Operation {
	return Operation{op, path, value, timestamp}
}

// GeneratePatch generates a JSON patch from two TSON documents
func GeneratePatch(origin, modified tson.Tson) (Patch, error) {
	// If the two TSON documents are equal, return an empty patch
	if eq, err := tson.Equal(origin, modified); err != nil && eq {
		return Patch{}, nil
	}
	return handleValues(origin, modified, "", Patch{})
}

// ApplyPatch applies a JSON patch to a TSON document
func ApplyPatch(doc tson.Tson, patch Patch) (t tson.Tson, err error) {
	for _, op := range patch {
		if t, err = applyOperation(doc, op); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func applyOperation(doc tson.Tson, op Operation) (t tson.Tson, err error) {
	path := rfc6901Decoder.Replace(op.Path)

	// Split the path into parts, ignoring the first empty string
	parts := strings.Split(path, "/")[1:]

	// Traverse the TSON document
	if t, err = applyTraverse(doc, parts, op); err != nil {
		return nil, err
	}
	return t, nil
}

func applyTraverse(doc tson.Tson, parts []string, op Operation) (t tson.Tson, err error) {
	if len(parts) == 0 { // If the path is empty, return
		return doc, nil
	}

	// fmt.Printf("doc: %v\n", doc)
	// fmt.Printf("parts: %v\n", parts)

	switch doc.(type) { // If doc is a leaf node, return
	case tson.Leaf[string], tson.Leaf[float64], tson.Leaf[bool]:
		return doc, nil
	}

	switch part := parts[0]; j := doc.(type) {
	case tson.Object:
		if len(parts) == 1 { // Only a single part of path left
			switch op.Op { // switch by operation type
			case OpAdd, OpReplace:
				switch j[part].(type) {
				case tson.Leaf[string]:
					j[part] = tson.Leaf[string]{Value: op.Value.(string), Timestamp: op.Timestamp}
				case tson.Leaf[float64]:
					j[part] = tson.Leaf[float64]{Value: op.Value.(float64), Timestamp: op.Timestamp}
				case tson.Leaf[bool]:
					j[part] = tson.Leaf[bool]{Value: op.Value.(bool), Timestamp: op.Timestamp}
				default:
					return nil, fmt.Errorf("applyTraverse(): Unknown type %T for op.Value", op.Value)
				}
			case OpRemove:
				delete(j, part)
			case OpMove, OpCopy, OpTest:
				panic(fmt.Sprintf("applyTraverse(): Operation %s not implemented", op.Op))
			default:
				return nil, fmt.Errorf("applyTraverse(): Unknown operation %s", op.Op)
			}
			return j, nil
		}

		switch value := j[part].(type) { // Recursively traverse the TSON document
		case tson.Leaf[string], tson.Leaf[float64], tson.Leaf[bool]:
			panic("applyTraverse(): Leaf[T] should not be here")
		case tson.Object:
			if j[part], err = applyTraverse(value, parts[1:], op); err != nil {
				return nil, err
			}
		case tson.Array:
			idx, err := getIndex(parts[1])
			if err != nil {
				return nil, err
			}

			switch op.Op {
			case OpAdd, OpReplace:
				switch elem := value[idx].(type) {
				// If leaf
				case tson.Leaf[string]:
					value[idx] = tson.Leaf[string]{Value: op.Value.(string), Timestamp: op.Timestamp}
				case tson.Leaf[float64]:
					value[idx] = tson.Leaf[float64]{Value: op.Value.(float64), Timestamp: op.Timestamp}
				case tson.Leaf[bool]:
					value[idx] = tson.Leaf[bool]{Value: op.Value.(bool), Timestamp: op.Timestamp}
				case tson.Object, tson.Array:
					if j[part], err = applyTraverse(value[idx], parts[1:], op); err != nil {
						return nil, err
					}
				default:
					return nil, fmt.Errorf("applyTraverse(): Unknown type %T for elem", elem)
				}
			case OpRemove:
				j[part] = append(value[:idx], value[idx+1:]...)
			default:
				return nil, fmt.Errorf("applyTraverse(): Unknown operation %s", op.Op)
			}

		default:
			return nil, fmt.Errorf("traverse(): Unknown type %T for value", value)
		}

		return j, nil
	case tson.Array:
		panic("applyTraverse(): Array is not implemented")
	default:
		return nil, fmt.Errorf("applyTraverse(): Unknown type %T for doc", doc)
	}
}

func getIndex(part string) (idx int, err error) {
	if idx, err = strconv.Atoi(part); err != nil {
		return -1, fmt.Errorf("getIndex(): Invalid index %s", part)
	}
	return idx, nil
}

// Note: http://tools.ietf.org/html/rfc6901#section-4 :
var (
	rfc6901Encoder = strings.NewReplacer("~", "~0", "/", "~1")
	rfc6901Decoder = strings.NewReplacer("~1", "/", "~0", "~")
)

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

// diff returns the (recursive) difference between a and b.
func diff(origin, modified tson.Object, path string, patch Patch) (Patch, error) {
	for key, modValue := range modified {
		p := makePath(path, key)
		origValue, ok := origin[key]
		// "add": Only exists in 'modified'
		if !ok {
			switch modLeaf := modValue.(type) {
			case tson.Leaf[string]:
				patch = append(patch, NewOperation(OpAdd, p, modLeaf.Value, modLeaf.Timestamp))
			case tson.Leaf[float64]:
				patch = append(patch, NewOperation(OpAdd, p, modLeaf.Value, modLeaf.Timestamp))
			case tson.Leaf[bool]:
				patch = append(patch, NewOperation(OpAdd, p, modLeaf.Value, modLeaf.Timestamp))
			default:
				return nil, fmt.Errorf("diff(): Unknown type %T for modValue", modValue)
			}
			continue
		}
		// "replace": Type has changed
		if reflect.TypeOf(origValue) != reflect.TypeOf(modValue) {
			switch modLeaf := modValue.(type) {
			case tson.Leaf[string]:
				patch = append(patch, NewOperation(OpReplace, p, modLeaf.Value, modLeaf.Timestamp))
			case tson.Leaf[float64]:
				patch = append(patch, NewOperation(OpReplace, p, modLeaf.Value, modLeaf.Timestamp))
			case tson.Leaf[bool]:
				patch = append(patch, NewOperation(OpReplace, p, modLeaf.Value, modLeaf.Timestamp))
			default:
				return nil, fmt.Errorf("diff(): Unknown type %T for modValue", modValue)
			}
			continue
		}
		// Types are the same, compare values
		var err error
		patch, err = handleValues(origValue, modValue, p, patch)
		if err != nil {
			return nil, err
		}
	}
	// "remove": Only exists in 'origin'
	for key := range origin {
		_, found := modified[key]
		if !found {
			p := makePath(path, key)
			origValue := origin[key]
			switch origLeaf := origValue.(type) {
			case tson.Leaf[string]:
				patch = append(patch, NewOperation(OpRemove, p, nil, origLeaf.Timestamp))
			case tson.Leaf[float64]:
				patch = append(patch, NewOperation(OpRemove, p, nil, origLeaf.Timestamp))
			case tson.Leaf[bool]:
				patch = append(patch, NewOperation(OpRemove, p, nil, origLeaf.Timestamp))
			case tson.Object:
				// TODO: is there a way to calculate the timestamp?
				patch = append(patch, NewOperation(OpRemove, p, nil, 0))
			case tson.Array:
				patch = append(patch, NewOperation(OpRemove, p, nil, 0))
			default:
				panic(fmt.Sprintf("diff(): Unknown type %T for origValue", origValue))
			}
		}
	}
	return patch, nil
}

func handleValues(origValue, modValue tson.Value, path string, patch Patch) (Patch, error) {
	var err error
	switch origin := origValue.(type) {
	case tson.Leaf[string]:
		if !matchesValue(origValue, modValue) {
			modified := modValue.(tson.Leaf[string])
			patch = append(patch, NewOperation(OpReplace, path, modified.Value, modified.Timestamp))
		}
	case tson.Leaf[float64]:
		if !matchesValue(origValue, modValue) {
			modified := modValue.(tson.Leaf[float64])
			patch = append(patch, NewOperation(OpReplace, path, modified.Value, modified.Timestamp))
		}
	case tson.Leaf[bool]:
		if !matchesValue(origValue, modValue) {
			modified := modValue.(tson.Leaf[bool])
			patch = append(patch, NewOperation(OpReplace, path, modified.Value, modified.Timestamp))
		}
	case tson.Object:
		modified := modValue.(tson.Object)
		if patch, err = diff(origin, modified, path, patch); err != nil {
			return nil, err
		}
	case tson.Array:
		modified, ok := modValue.(tson.Array)
		if !ok { // tson.Array replaced by non-Array
			var mod any = modValue
			switch modLeaf := mod.(type) {
			case tson.Leaf[string]:
				patch = append(patch, NewOperation(OpReplace, path, modLeaf.Value, modLeaf.Timestamp))
			case tson.Leaf[float64]:
				patch = append(patch, NewOperation(OpReplace, path, modLeaf.Value, modLeaf.Timestamp))
			case tson.Leaf[bool]:
				patch = append(patch, NewOperation(OpReplace, path, modLeaf.Value, modLeaf.Timestamp))
			default:
				return nil, fmt.Errorf("handleValues(): Unknown type %T for modValue", modValue)
			}
		} else if len(origin) != len(modified) { // Different array lengths
			patch = append(patch, compareArray(origin, modified, path)...)
		} else { // Same length, compare elements
			for i := range modified {
				patch, err = handleValues(origin[i], modified[i], makePath(path, i), patch)
				if err != nil {
					return nil, err
				}
			}
		}
	case nil:
		switch modValue.(type) {
		case nil:
			// Both nil, fine.
		default:
			// Replace nil with value
			var mod any = modValue
			switch modLeaf := mod.(type) {
			case tson.Leaf[string]:
				patch = append(patch, NewOperation(OpAdd, path, modLeaf.Value, modLeaf.Timestamp))
			case tson.Leaf[float64]:
				patch = append(patch, NewOperation(OpAdd, path, modLeaf.Value, modLeaf.Timestamp))
			case tson.Leaf[bool]:
				patch = append(patch, NewOperation(OpAdd, path, modLeaf.Value, modLeaf.Timestamp))
			default:
				return nil, fmt.Errorf("handleValues(): Unknown type %T for modValue", modValue)
			}
		}
	default:
		return nil, fmt.Errorf("handleValues(): Unknown type %T for origValue", origValue)
	}
	return patch, nil
}

// Compares two TSON values and returns true if they are equal.
func matchesValue(origin, modified tson.Value) bool {
	if reflect.TypeOf(origin) != reflect.TypeOf(modified) {
		return false
	}

	switch org := origin.(type) {
	case tson.Leaf[string]:
		if modified.(tson.Leaf[string]).Value == org.Value {
			return true
		}
	case tson.Leaf[float64]:
		if modified.(tson.Leaf[float64]).Value == org.Value {
			return true
		}
	case tson.Leaf[bool]:
		if modified.(tson.Leaf[bool]).Value == org.Value {
			return true
		}
	case tson.Object:
		modObj := modified.(tson.Object)
		for key := range org {
			if !matchesValue(org[key], modObj[key]) {
				return false
			}
		}
		for key := range modObj {
			if !matchesValue(org[key], modObj[key]) {
				return false
			}
		}
		return true
	case tson.Array:
		modArray := modified.(tson.Array)
		if len(modArray) != len(org) {
			return false
		}
		for key := range org {
			if !matchesValue(org[key], modArray[key]) {
				return false
			}
		}
		for key := range modArray {
			if !matchesValue(org[key], modArray[key]) {
				return false
			}
		}
		return true
	}
	return false
}

// compareArray generates remove and add operations
func compareArray(origArr, modArr tson.Array, p string) (patch Patch) {
	// Find elements that need to be removed
	processArray(origArr, modArr, func(i int, _ any) {
		// TODO: is there a way to calculate the timestamp?
		patch = append(patch, NewOperation(OpRemove, makePath(p, i), nil, 0))
	})

	reversed := make(Patch, len(patch))
	for i := 0; i < len(patch); i++ {
		reversed[len(patch)-1-i] = patch[i]
	}
	patch = reversed

	// Find elements that need to be added.
	processArray(modArr, origArr, func(i int, value any) {
		switch leaf := value.(type) {
		case tson.Leaf[string]:
			patch = append(patch, NewOperation(OpAdd, makePath(p, i), leaf.Value, leaf.Timestamp))
		case tson.Leaf[float64]:
			patch = append(patch, NewOperation(OpAdd, makePath(p, i), leaf.Value, leaf.Timestamp))
		case tson.Leaf[bool]:
			patch = append(patch, NewOperation(OpAdd, makePath(p, i), leaf.Value, leaf.Timestamp))
		default:
			panic(fmt.Sprintf("compareArray(): Unknown type %T for value", value))
		}
	})
	return patch
}

// processArray processes two arrays calling applyOp whenever a value is absent.
func processArray(origArr, modArr tson.Array, applyOp func(i int, value any)) {
	// Note: map[T]struct{} is used to simulate a set.
	foundIndexes := make(map[int]struct{}, len(origArr))
	reverseFoundIndexes := make(map[int]struct{}, len(origArr))

	for idx1, value1 := range origArr {
		for idx2, value2 := range modArr {
			if _, ok := reverseFoundIndexes[idx2]; ok {
				// This one is already found.
				continue
			}
			if reflect.DeepEqual(value1, value2) {
				// Mark this as found
				foundIndexes[idx1] = struct{}{}
				reverseFoundIndexes[idx2] = struct{}{}
				break
			}
		}
		if _, ok := foundIndexes[idx1]; !ok {
			applyOp(idx1, value1)
		}
	}
}
