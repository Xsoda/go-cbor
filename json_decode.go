package cbor

import "fmt"
import "bytes"
import "strconv"
import "unicode"
import "io/ioutil"

type json_lexer struct {
	source []byte
	eof int
	offset int
	lineno int
	linest int
	lineoff int
	error *bytes.Buffer
}

func (lexer *json_lexer) skip_comment() {
	if lexer.offset + 2 < lexer.eof && lexer.source[lexer.offset] == '/' && lexer.source[lexer.offset + 1] == '/' {
		lexer.offset += 2
		lexer.lineoff += 2
		for lexer.offset < lexer.eof {
			if lexer.offset + 2 < lexer.eof && lexer.source[lexer.offset] == '\r' && lexer.source[lexer.offset + 1] == '\n' {
				lexer.offset += 2
				lexer.lineno += 2
				lexer.linest = lexer.offset
				lexer.lineoff = 0
				break
			} else if lexer.source[lexer.offset] == '\r' || lexer.source[lexer.offset] == '\n' {
				lexer.offset += 1
				lexer.lineno += 1
				lexer.linest = lexer.offset
				lexer.lineoff = 0
				break
			} else {
				lexer.offset += 1
				lexer.lineoff += 1
			}
		}
	} else if lexer.offset + 2 < lexer.eof && lexer.source[lexer.offset] == '/' && lexer.source[lexer.offset + 1] == '*' {
		lexer.offset += 2
		lexer.lineoff += 2
		for lexer.offset < lexer.eof {
			if lexer.offset + 2 < lexer.eof && lexer.source[lexer.offset] == '/' && lexer.source[lexer.offset + 1] == '*' {
				lexer.skip_comment()
			} else if lexer.offset + 2 < lexer.eof && lexer.source[lexer.offset] == '*' && lexer.source[lexer.offset + 1] == '/' {
				lexer.offset += 2
				lexer.lineoff += 2
				break
			} else if lexer.offset + 2 < lexer.eof && lexer.source[lexer.offset] == '\r' && lexer.source[lexer.offset + 1] == '\n' {
				lexer.lineno += 1
				lexer.offset += 2
				lexer.linest = lexer.offset
				lexer.lineoff = 0
			} else if lexer.source[lexer.offset] == '\r' || lexer.source[lexer.offset] == '\n' {
				lexer.lineno += 1
				lexer.offset += 1
				lexer.linest = lexer.offset
				lexer.lineoff = 0
			} else {
				lexer.offset += 1
				lexer.lineoff += 1
			}
		}
	}
}

func (lexer *json_lexer) skip_whitespace() {
	for lexer.offset < lexer.eof {
		if lexer.offset + 2 < lexer.eof && lexer.source[lexer.offset] == '/' && lexer.source[lexer.offset + 1] == '/' {
			lexer.skip_comment()
		} else if lexer.offset + 2 < lexer.eof && lexer.source[lexer.offset] == '/' && lexer.source[lexer.offset + 1] == '*' {
			lexer.skip_comment()
		} else if lexer.offset + 2 < lexer.eof && lexer.source[lexer.offset] == '\r' && lexer.source[lexer.offset + 1] == '\n' {
			lexer.offset += 2
			lexer.lineno += 1
			lexer.linest = lexer.offset
			lexer.lineoff = 0
		} else if lexer.source[lexer.offset] == '\r' || lexer.source[lexer.offset] == '\n' {
			lexer.offset += 1
			lexer.lineno += 1
			lexer.linest = lexer.offset
			lexer.lineoff = 0
		} else if unicode.IsSpace(rune(lexer.source[lexer.offset])) {
			lexer.offset += 1
			lexer.lineoff += 1
		} else {
			break
		}
	}
}

func (lexer *json_lexer) parse_object() (*CborValue, error) {
	lexer.offset += 1
	lexer.lineoff += 1
	container := NewMap()
	for lexer.offset < lexer.eof {
		lexer.skip_whitespace()
		if lexer.source[lexer.offset] == '}' {
			if container.ContainerEmpty() {
				return container, nil
			} else {
				return nil, fmt.Errorf("%d:%d excepted key-value pair", lexer.lineno, lexer.lineoff)
			}
		}

		key, err := lexer.parse()
		if err != nil {
			return nil, err
		}
		if key == nil || key.ctype != CBOR_TYPE_STRING {
			return nil, fmt.Errorf("%d:%d excepted string element as object key", lexer.lineno, lexer.lineoff)
		}

		lexer.skip_whitespace()

		if lexer.source[lexer.offset] == ':' {
			lexer.offset += 1
			lexer.lineoff += 1
		} else {
			return nil, fmt.Errorf("%d:%d excepted `:` as key-value sperator", lexer.lineno, lexer.lineoff)
		}

		val, err := lexer.parse()
		if err != nil {
			return nil, err
		}
		pair := NewPair(key, val)
		container.ContainerInsertTail(pair)

		lexer.skip_whitespace()
		if lexer.source[lexer.offset] == '}' {
			lexer.offset += 1
			lexer.lineoff += 1
			break
		} else if lexer.source[lexer.offset] == ',' {
			lexer.offset += 1
			lexer.lineoff += 1
			continue
		} else {
			return nil, fmt.Errorf("%d:%d excepted `,` or `}` in object, character `%c`", lexer.lineno, lexer.lineoff, lexer.source[lexer.offset])
		}
	}
	return container, nil
}

func (lexer *json_lexer) parse_array() (*CborValue, error) {
	lexer.offset += 1
	lexer.lineoff += 1
	container := NewArray()
	for lexer.offset < lexer.eof {
		lexer.skip_whitespace()
		if lexer.source[lexer.offset] == ']' {
			if container.ContainerEmpty() {
				lexer.offset += 1
				lexer.lineoff += 1
				return container, nil
			} else {
				return nil, fmt.Errorf("%d:%d excepted array element", lexer.lineno, lexer.lineoff)
			}
		}
		elm, err := lexer.parse()
		if err != nil {
			return nil, err
		}
		container.ContainerInsertTail(elm)

		lexer.skip_whitespace()
		if lexer.source[lexer.offset] == ']' {
			lexer.offset += 1
			lexer.lineoff += 1
			break;
		} else if lexer.source[lexer.offset] == ',' {
			lexer.offset += 1
			lexer.lineoff += 1
			continue
		} else {
			return nil, fmt.Errorf("%d:%d excepted `,` or `]` in array", lexer.lineno, lexer.lineoff)
		}
	}
	return container, nil
}

func (lexer *json_lexer) read_utf16() (uint32, error) {
	var surrogate uint32 = 0
	for i := 0; i < 4; i++ {
		surrogate <<= 4
		if lexer.source[lexer.offset] >= '0' && lexer.source[lexer.offset] <= '9' {
			surrogate |= uint32(lexer.source[lexer.offset] - '0')
			lexer.offset += 1
			lexer.lineoff += 1
		} else if lexer.source[lexer.offset] >= 'a' && lexer.source[lexer.offset] <= 'f' {
			surrogate |= uint32(lexer.source[lexer.offset] - 'a') + 10
			lexer.offset += 1
			lexer.lineoff += 1
		} else if lexer.source[lexer.offset] >= 'A' && lexer.source[lexer.offset] <= 'F' {
			surrogate |= uint32(lexer.source[lexer.offset] - 'A') + 10
			lexer.offset += 1
			lexer.lineoff += 1
		} else {
			return 0, fmt.Errorf("%d:%d unexcepted character `%c` when read string", lexer.lineno, lexer.lineoff, lexer.source[lexer.offset])
		}
	}
	return surrogate, nil
}

func (lexer *json_lexer) parse_string() (*CborValue, error) {
	lexer.offset += 1
	lexer.linest += 1
	str := NewString("")
	for lexer.offset < lexer.eof && lexer.source[lexer.offset] != '"' {
		if lexer.source[lexer.offset] == '\\' {
			if lexer.source[lexer.offset + 1] == 'r' {
				str.BlobAppendByte('\r')
				lexer.offset += 2
				lexer.lineoff += 2
			} else if lexer.source[lexer.offset + 1] == 'n' {
				str.BlobAppendByte('\n')
				lexer.offset += 2
				lexer.lineoff += 2
			} else if lexer.source[lexer.offset + 1] == 't' {
				str.BlobAppendByte('\t')
				lexer.offset += 2
				lexer.lineoff += 2
			} else if lexer.source[lexer.offset + 1] == 'f' {
				str.BlobAppendByte('\f')
				lexer.offset += 2
				lexer.lineoff += 2
			} else if lexer.source[lexer.offset + 1] == '"' {
				str.BlobAppendByte('"')
				lexer.offset += 2
				lexer.lineoff += 2
			} else if lexer.source[lexer.offset + 1] == '\\' {
				str.BlobAppendByte('\\')
				lexer.offset += 2
				lexer.lineoff += 2
			} else if lexer.source[lexer.offset + 1] == '/' {
				str.BlobAppendByte('/')
				lexer.offset += 2
				lexer.lineoff += 2
			} else if lexer.source[lexer.offset + 1] == 'b' {
				str.BlobAppendByte('\b')
				lexer.offset += 2
				lexer.lineoff += 2
			} else if lexer.source[lexer.offset + 1] == 'u' {
				var high_surrogate uint32 = 0
				var low_surrogate uint32 = 0
				lexer.offset += 2
				lexer.lineoff += 2
				high_surrogate, err := lexer.read_utf16()
				if err != nil {
					return nil, err
				}
				if high_surrogate >= 0xD800 && high_surrogate <= 0xDBFF {
					if lexer.offset + 6 < lexer.eof && lexer.source[lexer.offset] == '\\' && lexer.source[lexer.offset + 1] == 'u' {
						lexer.offset += 2
						lexer.lineoff += 2
						low_surrogate, err = lexer.read_utf16()
						if err != nil {
							return nil, err
						}
						if low_surrogate >= 0xDC00 && low_surrogate <= 0xDFFF {
							high_surrogate &= 0x3FF
							high_surrogate <<= 10
							low_surrogate &= 0x3FF
							low_surrogate |= high_surrogate
							low_surrogate += 0x10000
							str.BlobAppendRune(rune(low_surrogate));
						} else {
							return nil, fmt.Errorf("%d:%d utf-16 surrogate error", lexer.lineno, lexer.lineoff)
						}
					} else {
						return nil, fmt.Errorf("%d:%d excepted next utf-16 surrogate", lexer.lineno, lexer.lineoff)
					}
				} else {
					str.BlobAppendRune(rune(high_surrogate));
				}
			}
		} else if lexer.source[lexer.offset] == '\r' || lexer.source[lexer.offset] == '\n' {
			return nil, fmt.Errorf("%d:%d unexcepted line-break in string", lexer.lineno, lexer.lineoff)
		} else {
			str.BlobAppendByte(lexer.source[lexer.offset])
			lexer.offset += 1
			lexer.lineoff += 1
		}
	}
	if lexer.offset < lexer.eof && lexer.source[lexer.offset] == '"' {
		lexer.offset += 1
		lexer.lineoff += 1
		return str, nil
	} else {
		return nil, fmt.Errorf("%d:%d infinity string", lexer.lineno, lexer.lineoff)
	}
}

func (lexer *json_lexer) parse_number() (*CborValue, error) {
	buf := bytes.Buffer{}
	for lexer.offset < lexer.eof {
		if lexer.source[lexer.offset] == '+' || lexer.source[lexer.offset] == '-' {
			buf.WriteByte(lexer.source[lexer.offset])
			lexer.offset += 1
			lexer.lineoff += 1
		} else if lexer.source[lexer.offset] == '.' {
			buf.WriteByte(lexer.source[lexer.offset])
			lexer.offset += 1
			lexer.lineoff += 1
		} else if lexer.source[lexer.offset] == 'e' || lexer.source[lexer.offset] == 'E' {
			buf.WriteByte(lexer.source[lexer.offset])
			lexer.offset += 1
			lexer.lineoff += 1
		} else if lexer.source[lexer.offset] >= '0' && lexer.source[lexer.offset] <= '9' {
			buf.WriteByte(lexer.source[lexer.offset])
			lexer.offset += 1
			lexer.lineoff += 1
		} else {
			break
		}
	}
	integer, err := strconv.ParseInt(buf.String(), 10, 64)
	if err == nil {
		return NewInteger(integer), nil
	} else {
		number, err := strconv.ParseFloat(buf.String(), 10)
		if err == nil {
			return NewFloat(number), nil
		}
		return nil, err
	}
}

func (lexer *json_lexer) parse() (*CborValue, error) {
	lexer.skip_whitespace()
	for lexer.offset < lexer.eof {
		if lexer.source[lexer.offset] == '{' {
			return lexer.parse_object()
		} else if lexer.source[lexer.offset] == '[' {
			return lexer.parse_array()
		} else if lexer.source[lexer.offset] == '"' {
			return lexer.parse_string()
		} else if lexer.source[lexer.offset] == 'f' {
			if lexer.source[lexer.offset + 1] == 'a' && lexer.source[lexer.offset + 2] == 'l' && lexer.source[lexer.offset + 3] == 's' && lexer.source[lexer.offset + 4] == 'e' {
				lexer.offset += 5
				lexer.lineoff += 5
				return NewBoolean(false), nil
			} else {
				return nil, fmt.Errorf("%d:%d except `false`", lexer.lineno, lexer.lineoff)
			}
		} else if lexer.source[lexer.offset] == 't' {
			if lexer.source[lexer.offset + 1] == 'r' && lexer.source[lexer.offset + 2] == 'u' && lexer.source[lexer.offset + 3] == 'e' {
				lexer.offset += 4
				lexer.lineoff += 4
				return NewBoolean(true), nil
			 } else {
				return nil, fmt.Errorf("%d:%d except `true`", lexer.lineno, lexer.lineoff)
			}
		} else if lexer.source[lexer.offset] == 'n' {
			if lexer.source[lexer.offset + 1] == 'u' && lexer.source[lexer.offset + 2] == 'l' && lexer.source[lexer.offset + 3] == 'l' {
				lexer.offset += 4
				lexer.lineoff += 4
				return NewNull(), nil
			} else {
				return nil, fmt.Errorf("%d:%d except `null`", lexer.lineno, lexer.lineoff)
			}
		} else if lexer.source[lexer.offset] >= '0' && lexer.source[lexer.offset] <= '9' {
			return lexer.parse_number()
		} else if lexer.source[lexer.offset] == '-' || lexer.source[lexer.offset] == '+' {
			return lexer.parse_number()
		} else {
			return nil, fmt.Errorf("%d:%d unexcepted character `%c`", lexer.lineno, lexer.lineoff, lexer.source[lexer.offset])
		}
	}
	return nil, nil
}

func JSONDecode(buf []byte) (val *CborValue, err error) {
	lexer := &json_lexer{
		source: buf,
		eof: len(buf),
		offset: 0,
		lineno: 1,
		linest: 0,
		lineoff: 0,
	}
	val, err = lexer.parse()
	return
}

func JSONLoadf(path string) *CborValue {
	content, err := ioutil.ReadFile(path)
	if err == nil {
		val, _ := JSONDecode(content)
		return val
	}
	return nil
}
