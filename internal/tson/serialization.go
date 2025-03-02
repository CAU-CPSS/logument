//
// serialization.go
//
// Marshalling & Unmarshalling functions
// for TSON data with '<>' timestamp.
//
// Author: Karu (@karu-rress)
//

package tson

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ToCompatibleTsonString converts TSON to a string.
func ToCompatibleTsonString(t Tson) string {
	data, err := json.MarshalIndent(t, "", "    ")
	if err != nil {
		panic(fmt.Sprintf("failed to marshal TSON: %v", err))
	}
	return string(data)
}

// >>>>>>>>> Marshalling

const (
	top = iota
	object
	array
)

// Marshal serializes a Tson data into <>-formatted bytes.
func Marshal(t Tson) ([]byte, error) {
	s, err := marshalTson(t, top)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// marshalTson marshals a Tson value given its context:
// top: 	top-level value,
// object: 	value inside an object (its timestamp is printed with the key),
// array: 	element inside an array (timestamp is printed before the primitive).
func marshalTson(v Value, ctx int) (string, error) {
	switch val := v.(type) {
	case Object:
		return marshalObject(val)
	case Array:
		return marshalArray(val)
	case Leaf[string]:
		if ctx == array {
			// In array, prefix timestamp before primitive.
			if val.Timestamp >= 0 { // Timestamp is positive (exists)
				return fmt.Sprintf("<%d> %q", val.Timestamp, val.Value), nil
			}
			return fmt.Sprintf("<> %q", val.Value), nil // No timestamp
		} else if ctx == object {
			// In object, the timestamp is printed with the key.
			return fmt.Sprintf("%q", val.Value), nil
		} else { // top-level
			if val.Timestamp >= 0 {
				return fmt.Sprintf("%q <%d>", val.Value, val.Timestamp), nil
			}
			return fmt.Sprintf("%q <>", val.Value), nil
		}
	case Leaf[float64]:
		if ctx == array {
			if val.Timestamp >= 0 {
				return fmt.Sprintf("<%d> %v", val.Timestamp, val.Value), nil
			}
			return fmt.Sprintf("<> %v", val.Value), nil
		} else if ctx == object {
			return fmt.Sprintf("%v", val.Value), nil
		} else {
			if val.Timestamp >= 0 {
				return fmt.Sprintf("%v <%d>", val.Value, val.Timestamp), nil
			}
			return fmt.Sprintf("%v <>", val.Value), nil
		}
	case Leaf[bool]:
		if ctx == array {
			if val.Timestamp >= 0 {
				return fmt.Sprintf("<%d> %v", val.Timestamp, val.Value), nil
			}
			return fmt.Sprintf("<> %v", val.Value), nil
		} else if ctx == object {
			return fmt.Sprintf("%v", val.Value), nil
		} else {
			if val.Timestamp >= 0 {
				return fmt.Sprintf("%v <%d>", val.Value, val.Timestamp), nil
			}
			return fmt.Sprintf("%v <>", val.Value), nil
		}
	default:
		return "", fmt.Errorf("marshalTson: unknown type")
	}
}

// marshalObject marshals an Object into a TSON object string.
func marshalObject(obj Object) (string, error) {
	var (
		s     strings.Builder
		first = true
	)
	s.WriteString("{")
	for key, value := range obj {
		if !first {
			s.WriteString(", ")
		} else {
			first = false
		}
		s.WriteString(fmt.Sprintf("%q", key))
		switch leaf := value.(type) {
		case Leaf[string]:
			if leaf.Timestamp >= 0 {
				s.WriteString(fmt.Sprintf(" <%d>", leaf.Timestamp))
			} else {
				s.WriteString(" <>")
			}
		case Leaf[float64]:
			if leaf.Timestamp >= 0 {
				s.WriteString(fmt.Sprintf(" <%d>", leaf.Timestamp))
			} else {
				s.WriteString(" <>")
			}
		case Leaf[bool]:
			if leaf.Timestamp >= 0 {
				s.WriteString(fmt.Sprintf(" <%d>", leaf.Timestamp))
			} else {
				s.WriteString(" <>")
			}
		}
		s.WriteString(": ")
		marshaledValue, err := marshalTson(value, object)
		if err != nil {
			return "", err
		}
		s.WriteString(marshaledValue)
	}
	s.WriteString("}")
	return s.String(), nil
}

// marshalArray marshals an Array into a TSON array string.
func marshalArray(arr Array) (string, error) {
	var (
		s     strings.Builder
		first = true
	)
	s.WriteString("[")
	for _, elem := range arr {
		if !first {
			s.WriteString(", ")
		} else {
			first = false
		}
		marshaledElem, err := marshalTson(elem, array)
		if err != nil {
			return "", err
		}
		s.WriteString(marshaledElem)
	}
	s.WriteString("]")
	return s.String(), nil
}

// MarshalIndent serializes a JSON-like or Tson document into a TSON-formatted byte slice,
// using the given prefix and indent (similar to encoding/json.MarshalIndent).
func MarshalIndent(j any, prefix, indent string) ([]byte, error) {
	s, err := marshalIndentValue(j, prefix, indent, top)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// marshalIndentValue recursively serializes the value with pretty-printing.
// ctx indicates the context ("top", "object", or "array") which affects where the timestamp is printed.
func marshalIndentValue(v any, currentIndent, indent string, ctx int) (string, error) {
	switch t := v.(type) {
	case nil:
		return "null", nil
	case bool:
		return fmt.Sprintf("%v", t), nil
	case float64:
		return fmt.Sprintf("%.f", t), nil
	case string:
		return strconv.Quote(t), nil

	case map[string]any:
		if isLeaf, leafVal, ts := checkLeaf(t); isLeaf {
			primStr, err := formatPrimitive(leafVal)
			if err != nil {
				return "", err
			}
			switch ctx {
			case object:
				return primStr, nil
			case array:
				if ts >= 0 {
					return fmt.Sprintf("<%d> %s", ts, primStr), nil
				}
				return fmt.Sprintf("<> %s", primStr), nil
			case top:
				if ts >= 0 {
					return fmt.Sprintf("%s <%d>", primStr, ts), nil
				}
				return fmt.Sprintf("%s <>", primStr), nil
			default:
				if ts >= 0 {
					return fmt.Sprintf("%s <%d>", primStr, ts), nil
				}
				return fmt.Sprintf("%s <>", primStr), nil
			}
		}
		var s strings.Builder
		s.WriteString("{\n")
		first := true
		for key, val := range t {
			if !first {
				s.WriteString(",\n")
			}
			first = false
			keyStr := strconv.Quote(key)
			if m, ok := val.(map[string]any); ok {
				if isLeaf, leafVal, ts := checkLeaf(m); isLeaf {
					primStr, err := formatPrimitive(leafVal)
					if err != nil {
						return "", err
					}
					s.WriteString(currentIndent + indent + keyStr + " <")
					if ts >= 0 {
						s.WriteString(strconv.FormatInt(ts, 10))
					}
					s.WriteString(">: " + primStr)
					continue
				}
			}
			valStr, err := marshalIndentValue(val, currentIndent+indent, indent, object)
			if err != nil {
				return "", err
			}
			s.WriteString(currentIndent + indent + keyStr + ": " + valStr)
		}
		s.WriteString("\n" + currentIndent + "}")
		return s.String(), nil

	case []any:
		s := "[\n"
		first := true
		for _, elem := range t {
			if !first {
				s += ",\n"
			}
			first = false
			if m, ok := elem.(map[string]any); ok {
				if isLeaf, leafVal, ts := checkLeaf(m); isLeaf {
					primStr, err := formatPrimitive(leafVal)
					if err != nil {
						return "", err
					}
					s += currentIndent + indent
					if ts >= 0 {
						s += fmt.Sprintf("<%d> %s", ts, primStr)
					} else {
						s += fmt.Sprintf("<> %s", primStr)
					}
					continue
				}
			}
			elemStr, err := marshalIndentValue(elem, currentIndent+indent, indent, array)
			if err != nil {
				return "", err
			}
			s += currentIndent + indent + elemStr
		}
		s += "\n" + currentIndent + "]"
		return s, nil

	case Object:
		var s strings.Builder
		s.WriteString("{\n")
		first := true
		for key, val := range t {
			if !first {
				s.WriteString(",\n")
			}
			first = false
			keyStr := strconv.Quote(key)
			// Tson의 값이 Leaf라면, key 뒤에 timestamp를 붙여 출력
			switch leaf := val.(type) {
			case Leaf[string]:
				s.WriteString(currentIndent + indent + keyStr + " <")
				if leaf.Timestamp >= 0 {
					s.WriteString(strconv.FormatInt(leaf.Timestamp, 10))
				}
				s.WriteString(">: " + strconv.Quote(leaf.Value))
			case Leaf[float64]:
				s.WriteString(currentIndent + indent + keyStr + " <")
				if leaf.Timestamp >= 0 {
					s.WriteString(strconv.FormatInt(leaf.Timestamp, 10))
				}
				s.WriteString(">: " + fmt.Sprintf("%v", leaf.Value))
			case Leaf[bool]:
				s.WriteString(currentIndent + indent + keyStr + " <")
				if leaf.Timestamp >= 0 {
					s.WriteString(strconv.FormatInt(leaf.Timestamp, 10))
				}
				s.WriteString(">: " + fmt.Sprintf("%v", leaf.Value))
			default:
				valStr, err := marshalIndentValue(val, currentIndent+indent, indent, object)
				if err != nil {
					return "", err
				}
				s.WriteString(currentIndent + indent + keyStr + ": " + valStr)
			}
		}
		s.WriteString("\n" + currentIndent + "}")
		return s.String(), nil

	case Array:
		var s strings.Builder
		s.WriteString("[\n")
		first := true
		for _, elem := range t {
			if !first {
				s.WriteString(",\n")
			}
			first = false
			switch leaf := elem.(type) {
			case Leaf[string]:
				s.WriteString(currentIndent + indent)
				if leaf.Timestamp >= 0 {
					s.WriteString(fmt.Sprintf("<%d> %s", leaf.Timestamp, strconv.Quote(leaf.Value)))
				} else {
					s.WriteString(fmt.Sprintf("<> %s", strconv.Quote(leaf.Value)))
				}
			case Leaf[float64]:
				s.WriteString(currentIndent + indent)
				if leaf.Timestamp >= 0 {
					s.WriteString(fmt.Sprintf("<%d> %v", leaf.Timestamp, leaf.Value))
				} else {
					s.WriteString(fmt.Sprintf("<> %v", leaf.Value))
				}
			case Leaf[bool]:
				s.WriteString(currentIndent + indent)
				if leaf.Timestamp >= 0 {
					s.WriteString(fmt.Sprintf("<%d> %v", leaf.Timestamp, leaf.Value))
				} else {
					s.WriteString(fmt.Sprintf("<> %v", leaf.Value))
				}
			default:
				elemStr, err := marshalIndentValue(elem, currentIndent+indent, indent, array)
				if err != nil {
					return "", err
				}
				s.WriteString(currentIndent + indent + elemStr)
			}
		}
		s.WriteString("\n" + currentIndent + "]")
		return s.String(), nil

	case Leaf[string]:
		primStr := strconv.Quote(t.Value)
		if ctx == array {
			if t.Timestamp >= 0 {
				return fmt.Sprintf("<%d> %s", t.Timestamp, primStr), nil
			}
			return fmt.Sprintf("<> %s", primStr), nil
		} else if ctx == top {
			if t.Timestamp >= 0 {
				return fmt.Sprintf("%s <%d>", primStr, t.Timestamp), nil
			}
			return fmt.Sprintf("%s <>", primStr), nil
		}
		return primStr, nil
	case Leaf[float64]:
		primStr := fmt.Sprintf("%v", t.Value)
		if ctx == array {
			if t.Timestamp >= 0 {
				return fmt.Sprintf("<%d> %s", t.Timestamp, primStr), nil
			}
			return fmt.Sprintf("<> %s", primStr), nil
		} else if ctx == top {
			if t.Timestamp >= 0 {
				return fmt.Sprintf("%s <%d>", primStr, t.Timestamp), nil
			}
			return fmt.Sprintf("%s <>", primStr), nil
		}
		return primStr, nil
	case Leaf[bool]:
		primStr := fmt.Sprintf("%v", t.Value)
		if ctx == array {
			if t.Timestamp >= 0 {
				return fmt.Sprintf("<%d> %s", t.Timestamp, primStr), nil
			}
			return fmt.Sprintf("<> %s", primStr), nil
		} else if ctx == top {
			if t.Timestamp >= 0 {
				return fmt.Sprintf("%s <%d>", primStr, t.Timestamp), nil
			}
			return fmt.Sprintf("%s <>", primStr), nil
		}
		return primStr, nil

	default:
		return "", fmt.Errorf("unsupported type: %T", v)
	}
}

// formatPrimitive converts a primitive value (string, float64, bool) into its string representation.
func formatPrimitive(v any) (string, error) {
	switch t := v.(type) {
	case string:
		return strconv.Quote(t), nil
	case float64, int, int64, uint, uint64:
		return fmt.Sprintf("%v", t), nil
	case bool:
		return fmt.Sprintf("%v", t), nil
	default:
		return "", fmt.Errorf("unsupported primitive type: %T", v)
	}
}

// <<<<<<<<< Unmarshalling

// Parser struct maintains the parsing state.
type Parser struct {
	input []byte
	pos   int
}

// Unmarshal parses the TSON bytes with '<>' timestamp
// and stores the result in the value pointed to by t.
func Unmarshal(data []byte, t *Tson) (err error) {
	p := &Parser{input: data, pos: 0}
	*t, err = p.parseTson()
	return err
}

// wrapIfPrimitive wraps a raw primitive with a timestamp.
func wrapIfPrimitive(v any, timestamp int64) Value {
	switch val := v.(type) {
	case string:
		return Leaf[string]{Value: val, Timestamp: timestamp}
	case float64:
		return Leaf[float64]{Value: val, Timestamp: timestamp}
	case bool:
		return Leaf[bool]{Value: val, Timestamp: timestamp}
	case Leaf[string]:
		if timestamp >= 0 { // TODO: Remove duplicate code
			val.Timestamp = timestamp
		}
		return val
	case Leaf[float64]:
		if timestamp >= 0 {
			val.Timestamp = timestamp
		}
		return val
	case Leaf[bool]:
		if timestamp >= 0 {
			val.Timestamp = timestamp
		}
		return val
	default:
		return v.(Value)
	}
}

func (p *Parser) skipWhitespace() {
	for p.pos < len(p.input) &&
		(p.input[p.pos] == ' ' ||
			p.input[p.pos] == '\t' ||
			p.input[p.pos] == '\n' ||
			p.input[p.pos] == '\r') {
		p.pos++
	}
}

func (p *Parser) peek() byte {
	if p.pos < len(p.input) {
		return p.input[p.pos]
	}
	return 0
}

func (p *Parser) next() byte {
	ch := p.input[p.pos]
	p.pos++
	return ch
}

func (p *Parser) expect(ch byte) error {
	p.skipWhitespace()
	if p.peek() != ch {
		return fmt.Errorf("expected '%c', got '%c' at pos %d", ch, p.peek(), p.pos)
	}
	p.pos++
	return nil
}

// parseTson is the entry point.
func (p *Parser) parseTson() (Value, error) {
	p.skipWhitespace()
	return p.parseVal()
}

// parseVal parses a <value> as defined in the BNF.
func (p *Parser) parseVal() (Value, error) {
	p.skipWhitespace()
	// timestamp can appear at the beginning of a value
	// when it is an array element.
	if p.peek() == '<' {
		ts, err := p.parseTimestamp()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		prim, err := p.parsePrimitive()
		if err != nil {
			return nil, err
		}
		return wrapIfPrimitive(prim, ts), nil
	}

	switch p.peek() {
	case '{':
		return p.parseObject()
	case '[':
		return p.parseArray()
	case '"', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		prim, err := p.parsePrimitive()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if p.pos < len(p.input) && p.peek() == '<' {
			ts, err := p.parseTimestamp()
			if err != nil {
				return nil, err
			}
			return wrapIfPrimitive(prim, ts), nil
		}
		return wrapIfPrimitive(prim, DefaultTimestamp), nil
	default:
		// true, false, null
		if p.startsWith("true") {
			p.pos += 4
			p.skipWhitespace()
			if p.pos < len(p.input) && p.peek() == '<' {
				ts, err := p.parseTimestamp()
				if err != nil {
					return nil, err
				}
				return wrapIfPrimitive(true, ts), nil
			}
			return wrapIfPrimitive(true, DefaultTimestamp), nil
		} else if p.startsWith("false") {
			p.pos += 5
			p.skipWhitespace()
			if p.pos < len(p.input) && p.peek() == '<' {
				ts, err := p.parseTimestamp()
				if err != nil {
					return nil, err
				}
				return wrapIfPrimitive(false, ts), nil
			}
			return wrapIfPrimitive(false, DefaultTimestamp), nil
		} else if p.startsWith("null") {
			p.pos += 4
			return nil, nil
		}
	}

	return nil, fmt.Errorf("unexpected character '%c' at pos %d", p.peek(), p.pos)
}

func (p *Parser) startsWith(s string) bool {
	bytes := []byte(s)
	if p.pos+len(bytes) > len(p.input) {
		return false
	}
	for i, r := range bytes {
		if p.input[p.pos+i] != r {
			return false
		}
	}
	return true
}

// parsePrimitive parses a primitive: string, number.
func (p *Parser) parsePrimitive() (any, error) {
	p.skipWhitespace()
	switch p.peek() {
	case '"':
		return p.parseString()
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return p.parseNumber()
	default:
		return nil, fmt.Errorf("invalid primitive starting with '%c' at pos %d", p.peek(), p.pos)
	}
}

// parseString parses a JSON string.
func (p *Parser) parseString() (string, error) {
	if err := p.expect('"'); err != nil {
		return "", err
	}
	var result []byte
	for p.pos < len(p.input) {
		ch := p.next()
		if ch == '\\' { // escape
			if p.pos >= len(p.input) {
				return "", fmt.Errorf("unterminated escape sequence")
			}
			esc := p.next()
			switch esc {
			case '"', '\\', '/', '\'':
				result = append(result, esc)
			case 'b':
				result = append(result, '\b')
			case 'f':
				result = append(result, '\f')
			case 'n':
				result = append(result, '\n')
			case 'r':
				result = append(result, '\r')
			case 't':
				result = append(result, '\t')
			default:
				return "", fmt.Errorf("invalid escape character: %c", esc)
			}
		} else if ch == '"' {
			return string(result), nil
		} else {
			result = append(result, ch)
		}
	}
	return "", fmt.Errorf("unterminated string")
}

// parseNumber parses a number.
func (p *Parser) parseNumber() (float64, error) {
	start := p.pos
	if p.peek() == '-' {
		p.pos++
	}
	for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
		p.pos++
	}
	if p.pos < len(p.input) && p.peek() == '.' {
		p.pos++
		for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
			p.pos++
		}
	}
	numStr := string(p.input[start:p.pos])
	return strconv.ParseFloat(numStr, 64)
}

// parseTimestamp parses a timestamp in the form <number>.
// If timestamp is ommitted, it returns negative value(-1).
func (p *Parser) parseTimestamp() (int64, error) {
	if err := p.expect('<'); err != nil { // expect '<'
		return 0, err
	}
	if p.peek() == '>' { // timestamp is ommitted
		p.pos++
		return -1, nil
	}

	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
		p.pos++
	}
	if start == p.pos {
		return 0, fmt.Errorf("expected digits for timestamp at pos %d", p.pos)
	}
	numStr := string(p.input[start:p.pos])
	ts, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, err
	}
	if err := p.expect('>'); err != nil {
		return 0, err
	}
	return ts, nil
}

// parseObject parses an object: { <members>? }.
func (p *Parser) parseObject() (Value, error) {
	if err := p.expect('{'); err != nil {
		return nil, err
	}
	obj := Object{}
	p.skipWhitespace()
	if p.peek() == '}' {
		p.pos++
		return obj, nil
	}
	for {
		p.skipWhitespace()
		// key must be a string
		key, err := p.parseString()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		// optional timestamp after key
		var ts int64 = 0
		if p.peek() == '<' {
			ts, err = p.parseTimestamp()
			if err != nil {
				return nil, err
			}
			p.skipWhitespace()
		}
		if err := p.expect(':'); err != nil {
			return nil, err
		}
		p.skipWhitespace()
		val, err := p.parseVal()
		if err != nil {
			return nil, err
		}
		obj[key] = wrapIfPrimitive(val, ts)
		p.skipWhitespace()
		if p.peek() == ',' {
			p.pos++
			continue
		} else if p.peek() == '}' {
			p.pos++
			break
		} else {
			return nil, fmt.Errorf("expected ',' or '}', got '%c' at pos %d", p.peek(), p.pos)
		}
	}
	return obj, nil
}

// parseArray parses an array: [ <elements>? ].
func (p *Parser) parseArray() (Value, error) {
	if err := p.expect('['); err != nil {
		return nil, err
	}
	arr := Array{}
	p.skipWhitespace()
	if p.peek() == ']' {
		p.pos++
		return arr, nil
	}
	for {
		p.skipWhitespace()
		var ts int64 = 0
		if p.peek() == '<' {
			var err error
			ts, err = p.parseTimestamp()
			if err != nil {
				return nil, err
			}
			p.skipWhitespace()
		}
		val, err := p.parseVal()
		if err != nil {
			return nil, err
		}
		arr = append(arr, wrapIfPrimitive(val, ts))
		p.skipWhitespace()
		if p.peek() == ',' {
			p.pos++
			continue
		} else if p.peek() == ']' {
			p.pos++
			break
		} else {
			return nil, fmt.Errorf("expected ',' or ']', got '%c' at pos %d", p.peek(), p.pos)
		}
	}
	return arr, nil
}
