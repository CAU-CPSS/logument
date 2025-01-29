//
// jsonpatch.go
//
// A JSON patch libary for JSON-R documents.
//
// Author: Karu (karu-rress)
//

package jsonpatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
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

// Patch represents a JSON patch document.
type Patch []Operation

// String converts the Patch to a JSON string, with formatting.
func (p *Patch) String() (ret string) {
	lines := make([]string, len(*p))
	for i, op := range *p {
		lines[i] = "    " + op.String()
	}
	ret = fmt.Sprintf("[\n%s\n]", strings.Join(lines, ",\n"))
	return ret
}

func ParsePatch(b []byte) (Patch, error) {
	var patch Patch
	if err := json.Unmarshal(b, &patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// Operation represents a single JSON patch operation.
type Operation struct {
	Op        OpType `json:"op"`
	Path      string `json:"path"`
	Value     any    `json:"value,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// String converts the Operation to a JSON string.
func (p *Operation) String() (s string) {
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
	return s
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
	b = buf.Bytes()
	return b, nil
}

// NewOperation creates a new Operation instance.
func NewOperation(op OpType, path string, value any, timestamp int64) Operation {
	return Operation{op, path, value, timestamp}
}

// GeneratePatch generates a JSON patch from two JSON-R documents
func GeneratePatch(origin, modified jsonr.JsonR) (Patch, error) {
	// If the two JSON-R documents are equal, return an empty patch
	if eq, err := jsonr.Equal(origin, modified); err != nil && eq {
		return Patch{}, nil
	}
	return handleValues(origin, modified, "", Patch{})
}

// ApplyPatch applies a JSON patch to a JSON-R document
func ApplyPatch(doc jsonr.JsonR, patch Patch) (j jsonr.JsonR, err error) {
	for _, op := range patch {
		if j, err = applyOperation(doc, op); err != nil {
			return nil, err
		}
	}
	return j, nil
}

func applyOperation(doc jsonr.JsonR, op Operation) (j jsonr.JsonR, err error) {
	path := rfc6901Decoder.Replace(op.Path)

	// Split the path into parts, ignoring the first empty string
	parts := strings.Split(path, "/")[1:]

	// Traverse the JSON-R document
	if j, err = applyTraverse(doc, parts, op); err != nil {
		return nil, err
	}
	return j, nil
}

func applyTraverse(doc jsonr.JsonR, parts []string, op Operation) (j jsonr.JsonR, err error) {
	if len(parts) == 0 { // If the path is empty, return
		return doc, nil
	}

	switch doc.(type) { // If doc is a leaf node, return
	case jsonr.Leaf[string], jsonr.Leaf[float64], jsonr.Leaf[bool]:
		return doc, nil
	}

	switch part := parts[0]; j := doc.(type) {
	case jsonr.Object:
		if len(parts) == 1 { // Only a single part of path left
			switch op.Op { // switch by operation type
			case OpAdd, OpReplace:
				// switch by operation value's datatype
				switch leaf := op.Value.(type) {
				case jsonr.Leaf[string]:
					j[part] = jsonr.Leaf[string]{Value: leaf.Value, Timestamp: leaf.Timestamp}
				case jsonr.Leaf[float64]:
					j[part] = jsonr.Leaf[float64]{Value: leaf.Value, Timestamp: leaf.Timestamp}
				case jsonr.Leaf[bool]:
					j[part] = jsonr.Leaf[bool]{Value: leaf.Value, Timestamp: leaf.Timestamp}
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

		switch value := j[part].(type) { // Recursively traverse the JSON-R document
		case jsonr.Leaf[string], jsonr.Leaf[float64], jsonr.Leaf[bool]:
			panic(fmt.Sprintf("applyTraverse(): Leaf[T] should not be here"))
		case jsonr.Object:
			if j[part], err = applyTraverse(value, parts[1:], op); err != nil {
				return nil, err
			}
		case jsonr.Array:
			idx, err := getIndex(parts[1])
			if err != nil {
				return nil, err
			}

			switch op.Op {
			case OpAdd, OpReplace:
				switch elem := value[idx].(type) {
				// If leaf
				case jsonr.Leaf[string]:
					value[idx] = jsonr.Leaf[string]{Value: op.Value.(string), Timestamp: op.Timestamp}
				case jsonr.Leaf[float64]:
					value[idx] = jsonr.Leaf[float64]{Value: op.Value.(float64), Timestamp: op.Timestamp}
				case jsonr.Leaf[bool]:
					value[idx] = jsonr.Leaf[bool]{Value: op.Value.(bool), Timestamp: op.Timestamp}
				case jsonr.Object, jsonr.Array:
					if j[part], err = applyTraverse(value[idx], parts[1:], op); err != nil {
						return nil, err
					}
				default:
					return nil, fmt.Errorf("applyTraverse(): Unknown type %T for elem", elem)
				}
			case OpRemove:
				value = append(value[:idx], value[idx+1:]...)
			default:
				return nil, fmt.Errorf("applyTraverse(): Unknown operation %s", op.Op)
			}

		default:
			return nil, fmt.Errorf("traverse(): Unknown type %T for value", value)
		}

		return j, nil
	case jsonr.Array:
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
func diff(origin, modified jsonr.Object, path string, patch Patch) (Patch, error) {
	for key, modValue := range modified {
		p := makePath(path, key)
		origValue, ok := origin[key]
		// "add": Only exists in 'modified'
		if !ok {
			switch modLeaf := modValue.(type) {
			case jsonr.Leaf[string]:
				patch = append(patch, NewOperation(OpAdd, p, modLeaf.Value, modLeaf.Timestamp))
			case jsonr.Leaf[float64]:
				patch = append(patch, NewOperation(OpAdd, p, modLeaf.Value, modLeaf.Timestamp))
			case jsonr.Leaf[bool]:
				patch = append(patch, NewOperation(OpAdd, p, modLeaf.Value, modLeaf.Timestamp))
			default:
				return nil, fmt.Errorf("diff(): Unknown type %T for modValue", modValue)
			}
			continue
		}
		// "replace": Type has changed
		if reflect.TypeOf(origValue) != reflect.TypeOf(modValue) {
			switch modLeaf := modValue.(type) {
			case jsonr.Leaf[string]:
				patch = append(patch, NewOperation(OpReplace, p, modLeaf.Value, modLeaf.Timestamp))
			case jsonr.Leaf[float64]:
				patch = append(patch, NewOperation(OpReplace, p, modLeaf.Value, modLeaf.Timestamp))
			case jsonr.Leaf[bool]:
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
			case jsonr.Leaf[string]:
				patch = append(patch, NewOperation(OpRemove, p, nil, origLeaf.Timestamp))
			case jsonr.Leaf[float64]:
				patch = append(patch, NewOperation(OpRemove, p, nil, origLeaf.Timestamp))
			case jsonr.Leaf[bool]:
				patch = append(patch, NewOperation(OpRemove, p, nil, origLeaf.Timestamp))
			case jsonr.Object:
				// TODO: is there a way to calculate the timestamp?
				patch = append(patch, NewOperation(OpRemove, p, nil, -1))
			case jsonr.Array:
				patch = append(patch, NewOperation(OpRemove, p, nil, -1))
			default:
				panic(fmt.Sprintf("diff(): Unknown type %T for origValue", origValue))
			}
		}
	}
	return patch, nil
}

func handleValues(origValue, modValue jsonr.Value, path string, patch Patch) (Patch, error) {
	var err error
	switch origin := origValue.(type) {
	case jsonr.Leaf[string]:
		if !matchesValue(origValue, modValue) {
			modified := modValue.(jsonr.Leaf[string])
			patch = append(patch, NewOperation(OpReplace, path, modified.Value, modified.Timestamp))
		}
	case jsonr.Leaf[float64]:
		if !matchesValue(origValue, modValue) {
			modified := modValue.(jsonr.Leaf[float64])
			patch = append(patch, NewOperation(OpReplace, path, modified.Value, modified.Timestamp))
		}
	case jsonr.Leaf[bool]:
		if !matchesValue(origValue, modValue) {
			modified := modValue.(jsonr.Leaf[bool])
			patch = append(patch, NewOperation(OpReplace, path, modified.Value, modified.Timestamp))
		}
	case jsonr.Object:
		modified := modValue.(jsonr.Object)
		if patch, err = diff(origin, modified, path, patch); err != nil {
			return nil, err
		}
	case jsonr.Array:
		modified, ok := modValue.(jsonr.Array)
		if !ok { // jsonr.Array replaced by non-Array
			var mod any = modValue
			switch modLeaf := mod.(type) {
			case jsonr.Leaf[string]:
				patch = append(patch, NewOperation(OpReplace, path, modLeaf.Value, modLeaf.Timestamp))
			case jsonr.Leaf[float64]:
				patch = append(patch, NewOperation(OpReplace, path, modLeaf.Value, modLeaf.Timestamp))
			case jsonr.Leaf[bool]:
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
			case jsonr.Leaf[string]:
				patch = append(patch, NewOperation(OpAdd, path, modLeaf.Value, modLeaf.Timestamp))
			case jsonr.Leaf[float64]:
				patch = append(patch, NewOperation(OpAdd, path, modLeaf.Value, modLeaf.Timestamp))
			case jsonr.Leaf[bool]:
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

// Compares two JSON-R values and returns true if they are equal.
func matchesValue(origin, modified jsonr.Value) bool {
	if reflect.TypeOf(origin) != reflect.TypeOf(modified) {
		return false
	}

	switch org := origin.(type) {
	case jsonr.Leaf[string]:
		if modified.(jsonr.Leaf[string]).Value == org.Value {
			return true
		}
	case jsonr.Leaf[float64]:
		if modified.(jsonr.Leaf[float64]).Value == org.Value {
			return true
		}
	case jsonr.Leaf[bool]:
		if modified.(jsonr.Leaf[bool]).Value == org.Value {
			return true
		}
	case jsonr.Object:
		modObj := modified.(jsonr.Object)
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
	case jsonr.Array:
		modArray := modified.(jsonr.Array)
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
func compareArray(origArr, modArr jsonr.Array, p string) (patch Patch) {
	// Find elements that need to be removed
	processArray(origArr, modArr, func(i int, _ any) {
		// TODO: is there a way to calculate the timestamp?
		patch = append(patch, NewOperation(OpRemove, makePath(p, i), nil, -1))
	})

	reversed := make(Patch, len(patch))
	for i := 0; i < len(patch); i++ {
		reversed[len(patch)-1-i] = patch[i]
	}
	patch = reversed

	// Find elements that need to be added.
	processArray(modArr, origArr, func(i int, value any) {
		switch leaf := value.(type) {
		case jsonr.Leaf[string]:
			patch = append(patch, NewOperation(OpAdd, makePath(p, i), leaf.Value, leaf.Timestamp))
		case jsonr.Leaf[float64]:
			patch = append(patch, NewOperation(OpAdd, makePath(p, i), leaf.Value, leaf.Timestamp))
		case jsonr.Leaf[bool]:
			patch = append(patch, NewOperation(OpAdd, makePath(p, i), leaf.Value, leaf.Timestamp))
		default:
			panic(fmt.Sprintf("compareArray(): Unknown type %T for value", value))
		}
	})
	return patch
}

// processArray processes two arrays calling applyOp whenever a value is absent.
func processArray(origArr, modArr jsonr.Array, applyOp func(i int, value any)) {
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
