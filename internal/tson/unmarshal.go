package tson

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func FromJsonBytes(data []byte) (Tson, error) {
	var t Tson
	err := Unmarshal(data, &t)
	return t, err
}

func JsonToTson(o any) (Tson, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return FromJsonBytes(data)
}

// Parser struct maintains the parsing state.
type Parser struct {
	input []byte
	pos   int
}

func Unmarshal(data []byte, j *Tson) (err error) {
	p := &Parser{input: data, pos: 0}
	*j, err = p.parseTson()
	return err
}

// NewTson creates a new TSON from the given TSON string.
func NewTson(tjson string) (Tson, error) {
	var t Tson
	err := Unmarshal([]byte(tjson), &t)
	return t, err
}

// wrapIfPrimitive wraps a raw primitive with a timestamp.
func wrapIfPrimitive(v interface{}, timestamp int64) Value {
	switch val := v.(type) {
	case string:
		return Leaf[string]{Value: val, Timestamp: timestamp}
	case float64:
		return Leaf[float64]{Value: val, Timestamp: timestamp}
	case bool:
		return Leaf[bool]{Value: val, Timestamp: timestamp}
	case Leaf[string]:
		if timestamp != 0 {
			val.Timestamp = timestamp
		}
		return val
	case Leaf[float64]:
		if timestamp != 0 {
			val.Timestamp = timestamp
		}
		return val
	case Leaf[bool]:
		if timestamp != 0 {
			val.Timestamp = timestamp
		}
		return val
	default:
		return v.(Value)
	}
}

func (p *Parser) skipWhitespace() {
	for p.pos < len(p.input) &&
		(p.input[p.pos] == ' ' || p.input[p.pos] == '\t' ||
			p.input[p.pos] == '\n' || p.input[p.pos] == '\r') {
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
		return wrapIfPrimitive(prim, 0), nil
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
			return wrapIfPrimitive(true, 0), nil
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
			return wrapIfPrimitive(false, 0), nil
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
func (p *Parser) parsePrimitive() (interface{}, error) {
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
func (p *Parser) parseTimestamp() (int64, error) {
	if err := p.expect('<'); err != nil {
		return 0, err
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
