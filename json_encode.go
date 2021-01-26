package cbor

import "fmt"
import "bytes"
import "unicode"
import "strconv"
import "io/ioutil"

func json_dump_indent(buf *bytes.Buffer) {
	buf.WriteByte(' ')
}

func json_dumps(buf *bytes.Buffer, val *CborValue) {
	if val.ctype == CBOR_TYPE_MAP {
		buf.WriteByte('{')
		for ele := val.ContainerFirst(); ele != nil; ele = val.ContainerNext(ele) {
			json_dumps(buf, ele.key)
			buf.WriteByte(':')
			json_dump_indent(buf)
			json_dumps(buf, ele.value)
			if val.ContainerNext(ele) != nil {
				buf.WriteByte(',')
				json_dump_indent(buf)
			}
		}
		buf.WriteByte('}')
	} else if (val.ctype == CBOR_TYPE_ARRAY) {
		buf.WriteByte('[')
		for ele := val.ContainerFirst(); ele != nil; ele = val.ContainerNext(ele) {
			json_dumps(buf, ele)
			if val.ContainerNext(ele) != nil {
				buf.WriteByte(',')
				json_dump_indent(buf)
			}
		}
		buf.WriteByte(']')
	} else if (val.ctype == CBOR_TYPE_SIMPLE) {
		if val.ctrl == CBOR_SIMPLE_TRUE {
			buf.WriteString("true")
		} else if val.ctrl == CBOR_SIMPLE_FALSE {
			buf.WriteString("false")
		} else if val.ctrl == CBOR_SIMPLE_NULL {
			buf.WriteString("null")
		} else if val.ctrl == CBOR_SIMPLE_REAL {
			buf.WriteString(strconv.FormatFloat(val.Float(), 'f', 6, 64))
		}
	} else if (val.ctype == CBOR_TYPE_UINT || val.ctype == CBOR_TYPE_NEGINT) {
		buf.WriteString(strconv.FormatInt(val.Integer(), 10))
	} else if (val.ctype == CBOR_TYPE_STRING) {
		buf.WriteByte('"')
		b := val.StringBytes()
		l := val.StringSize()
		off := 0
		for off < l {
			if b[off] == '\n' {
				buf.WriteByte('\\')
				buf.WriteByte('n')
				off++
			} else if b[off] == '\t' {
				buf.WriteByte('\\')
				buf.WriteByte('t')
				off++
			} else if b[off] == '\\' {
				buf.WriteByte('\\')
				buf.WriteByte('\\')
				off++
			} else if b[off] == '\r' {
				buf.WriteByte('\\')
				buf.WriteByte('r')
				off++
			} else if b[off] == '\f' {
				buf.WriteByte('\\')
				buf.WriteByte('f')
				off++
			} else {
				var codepoint uint32 = 0xFFFFFFFF
				if b[off] <= 0x7f {
					// 0xxxxxxx
					codepoint = uint32(b[off])
					off++
				} else if b[off] >= 0xC0 && b[off] <= 0xDF  && off + 1 < l {
					// 110xxxxx 10xxxxxx
					codepoint = uint32(b[off]) & 0x1f
					codepoint <<= 6
					codepoint |= uint32(b[off + 1]) & 0x3f
					off += 2
				} else if b[off] >= 0xE0 && b[off] <= 0xEF && off + 2 < l {
					// 1110xxxx 10xxxxxx 10xxxxxx
					codepoint = uint32(b[off]) & 0xf
					codepoint <<= 12
					codepoint |= (uint32(b[off + 1]) & 0x3f) << 6
					codepoint |= (uint32(b[off + 2]) & 0x3f)
					off += 3
				} else if b[off] >= 0xF0 && b[off] <= 0xF7 && off + 3 < l {
					// 1110xxxx 10xxxxxx 10xxxxxx 10xxxxxx
					codepoint = uint32(b[off]) & 0xf
					codepoint <<= 18
					codepoint |= (uint32(b[off + 1]) & 0x3f) << 12
					codepoint |= (uint32(b[off + 2]) & 0x3f) << 6
					codepoint |= (uint32(b[off + 3]) & 0x3f)
					off += 4
				} else {
					break
				}
				if codepoint <= 0x7F {
					if unicode.IsPrint(rune(codepoint)) {
						buf.WriteByte(byte(codepoint))
					} else {
						buf.WriteString(fmt.Sprintf("\\u%04x", codepoint))
					}
				} else if codepoint <= 0xD7FF || (codepoint >= 0xE000 && codepoint <= 0xFFFF) {
					buf.WriteString(fmt.Sprintf("\\u%04x", codepoint))
				} else if codepoint <= 0x10FFFF {
					codepoint -= 0x10000
					buf.WriteString(fmt.Sprintf("\\u%04x", ((codepoint >> 10) & 0x3FF) | 0xD800))
					buf.WriteString(fmt.Sprintf("\\u%04x", (codepoint & 0x3FF) | 0xDC00))
				}
			}
		}
		buf.WriteByte('"')
	}
}



func JSONEncode(val *CborValue) *bytes.Buffer {
	ret := new(bytes.Buffer)
	if val != nil {
		json_dumps(ret, val)
	}
	return ret
}



func JSONDumpf(val *CborValue, path string) {
	if val != nil {
		buf := JSONEncode(val)
		ioutil.WriteFile(path, buf.Bytes(), 0644)
	}
}
