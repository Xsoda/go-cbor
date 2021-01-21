package cbor

import "fmt"
import "math"
import "encoding/binary"

func read_network_endian(buf []byte, offset int, size int) uint64 {
	var integer uint64 = 0
	for i := 0; i < size; i++ {
		integer <<= 8
		integer |= uint64(buf[offset + i])
	}
	return integer
}

func cbor_parse(buf []byte, offset int) (*CborValue, error, int) {
	var val *CborValue = nil
	var err error = nil
	var origin int = offset
	ctype := int(uint32(buf[offset]) >> 5)
	addition := int(uint32(buf[offset]) & 0x1F)
	if ctype == CBOR_TYPE_UINT {
		offset++
		val = new(CborValue)
		val.ctype = CBOR_TYPE_UINT
		if addition < 24 {
			val.integer = uint64(addition)
		} else if addition == 24 && offset + 1 <= len(buf) {
			val.integer = read_network_endian(buf, offset, 1)
			offset++
		} else if addition == 25 && offset + 2 <= len(buf) {
			val.integer = read_network_endian(buf, offset, 2)
			offset += 2
		} else if addition == 26 && offset + 4 <= len(buf) {
			val.integer = read_network_endian(buf, offset, 4)
			offset += 4
		} else if addition == 27 && offset + 8 <= len(buf) {
			val.integer = read_network_endian(buf, offset, 8)
			offset += 8
		} else {
			val = nil
			err = fmt.Errorf("unknown addition value when decode unsigned integer")
		}
	} else if ctype == CBOR_TYPE_NEGINT {
		val = new(CborValue)
		val.ctype = CBOR_TYPE_NEGINT
		offset++
		if addition < 24 {
			val.integer = uint64(addition)
		} else if addition == 24 && offset + 1 <= len(buf) {
			val.integer = read_network_endian(buf, offset, 1)
			offset++
		} else if addition == 25 && offset + 2 <= len(buf) {
			val.integer = read_network_endian(buf, offset, 2)
			offset += 2
		} else if addition == 26 && offset + 4 <= len(buf) {
			val.integer = read_network_endian(buf, offset, 4)
			offset += 4
		} else if addition == 27 && offset + 8 <= len(buf) {
			val.integer = read_network_endian(buf, offset, 8)
			offset += 8
		} else {
			val = nil
			err = fmt.Errorf("unknown addition value when decode integer")
		}

	} else if ctype == CBOR_TYPE_BYTESTRING {
		val = new(CborValue)
		val.ctype = CBOR_TYPE_BYTESTRING
		offset++
		var size uint64 = 0
		if addition < 24 {
			size = uint64(addition)
		} else if addition == 24 && offset + 1 <= len(buf) {
			size = read_network_endian(buf, offset, 1)
			offset++
		} else if addition == 25 && offset + 2 <= len(buf) {
			size = read_network_endian(buf, offset, 2)
			offset += 2
		} else if addition == 26 && offset + 4 <= len(buf) {
			size = read_network_endian(buf, offset, 4)
			offset += 4
		} else if addition == 27 && offset + 8 <= len(buf) {
			size = read_network_endian(buf, offset, 8)
			offset += 8
		} else if addition == 31 {
			for offset < len(buf) {
				if buf[offset] == 0xFF {
					offset++
					break
				}
				subval, suberr, subconsume := cbor_parse(buf, offset)
				if subval != nil && subval.ctype == CBOR_TYPE_BYTESTRING {
					offset += subconsume
					val.blob.Write(subval.blob.Bytes())
				} else {
					val = nil
					err = suberr
					break
				}
			}
		} else {
			val = nil
			err = fmt.Errorf("unknown addition value when decode bytestring")
		}
		if val != nil && addition != 31 && offset + int(size) <= len(buf) {
			val.blob.Write(buf[offset:offset+int(size)])
			offset += int(size)
		}
	} else if ctype == CBOR_TYPE_STRING {
		val = new(CborValue)
		val.ctype = CBOR_TYPE_STRING
		offset++
		var size uint64 = 0
		if addition < 24 {
			size = uint64(addition)
		} else if addition == 24 && offset + 1 <= len(buf) {
			size = read_network_endian(buf, offset, 1)
			offset++
		} else if addition == 25 && offset + 2 <= len(buf) {
			size = read_network_endian(buf, offset, 2)
			offset += 2
		} else if addition == 26 && offset + 4 <= len(buf) {
			size = read_network_endian(buf, offset, 4)
			offset += 4
		} else if addition == 27 && offset + 8 <= len(buf) {
			size = read_network_endian(buf, offset, 8)
			offset += 8
		} else if addition == 31 {
			for offset < len(buf) {
				if buf[offset] == 0xFF {
					offset++
					break
				}
				subval, suberr, subconsume := cbor_parse(buf, offset)
				if subval != nil && subval.ctype == CBOR_TYPE_STRING {
					offset += subconsume
					val.blob.Write(subval.blob.Bytes())
				} else {
					val = nil
					err = suberr
					break
				}
			}
		} else {
			val = nil
			err = fmt.Errorf("unknown addition value when decode string")
		}
		if addition != 31 && offset + int(size) <= len(buf) {
			val.blob.Write(buf[offset:offset+int(size)])
			offset += int(size)
		}
	} else if ctype == CBOR_TYPE_ARRAY {
		val = NewArray()
		offset++
		var size uint64 = 0
		if addition < 24 {
			size = uint64(addition)
		} else if addition == 24 && offset + 1 <= len(buf) {
			size = read_network_endian(buf, offset, 1)
			offset++
		} else if addition == 25 && offset + 2 <= len(buf) {
			size = read_network_endian(buf, offset, 2)
			offset += 2
		} else if addition == 26 && offset + 4 <= len(buf) {
			size = read_network_endian(buf, offset, 4)
			offset += 4
		} else if addition == 27 && offset + 8 <= len(buf) {
			size = read_network_endian(buf, offset, 8)
			offset += 8
		} else if addition == 31 {
			for offset < len(buf) {
				if buf[offset] == 0xFF {
					offset++
					break
				}
				subval, suberr, subconsume := cbor_parse(buf, offset)
				if subval != nil {
					offset += subconsume
					val.ContainerInsertTail(subval)
				} else {
					val = nil
					err = suberr
					break
				}
			}
		} else {
			val = nil
			err = fmt.Errorf("unknown addition value when decode array")
		}
		if val != nil && addition != 31 {
			for i := 0; i < int(size) && offset < len(buf); i++ {
				subval, suberr, subconsume := cbor_parse(buf, offset)
				if subval != nil {
					offset += subconsume
					val.ContainerInsertTail(subval)
				} else {
					val = nil
					err = suberr
					break
				}
			}
		}
	} else if ctype == CBOR_TYPE_MAP {
		val = NewMap()
		offset++
		var size uint64 = 0
		if addition < 24 {
			size = uint64(addition)
		} else if addition == 24 && offset + 1 <= len(buf) {
			size = read_network_endian(buf, offset, 1)
			offset++
		} else if addition == 25 && offset + 2 <= len(buf) {
			size = read_network_endian(buf, offset, 2)
			offset += 2
		} else if addition == 26 && offset + 4 <= len(buf) {
			size = read_network_endian(buf, offset, 4)
			offset += 4
		} else if addition == 27 && offset + 8 <= len(buf) {
			size = read_network_endian(buf, offset, 8)
			offset += 8
		} else if addition == 31 {
			for offset < len(buf) {
				if buf[offset] == 0xFF {
					offset++
					break
				}
				subkey, suberr, subconsume := cbor_parse(buf, offset)
				if subkey != nil {
					offset += subconsume
					subval, suberr, subconsume := cbor_parse(buf, offset)
					if subval != nil {
						offset += subconsume
						pair := NewPair(subkey, subval)
						val.ContainerInsertTail(pair)
					} else {
						val = nil
						err = suberr
						break
					}
				} else {
					val = nil
					err = suberr
					break
				}
			}
		} else {
			val = nil
			err = fmt.Errorf("unknown addition value when decode map")
		}
		if val != nil && addition != 31 {
			for i := 0; i < int(size) && offset < len(buf); i++ {
				subkey, suberr, subconsume := cbor_parse(buf, offset)
				if subkey != nil {
					offset += subconsume
					subval, suberr, subconsume := cbor_parse(buf, offset)
					if subval != nil {
						offset += subconsume
						pair := NewPair(subkey, subval)
						val.ContainerInsertTail(pair)
					} else {
						val = nil
						err = suberr
						break
					}
				} else {
					val = nil
					err = suberr
					break
				}
			}
		}
	} else if ctype == CBOR_TYPE_TAG {
		val = NewTag()
		offset++
		if addition < 24 {
			val.tag_item = uint64(addition);
		} else if addition == 24 && offset + 1 <= len(buf) {
			val.tag_item = read_network_endian(buf, offset, 1)
			offset++
		} else if addition == 25 && offset + 2 <= len(buf) {
			val.tag_item = read_network_endian(buf, offset, 2)
			offset += 2
		} else if addition == 26 && offset + 4 <= len(buf) {
			val.tag_item = read_network_endian(buf, offset, 4)
			offset += 4
		} else if addition == 27 && offset + 8 <= len(buf) {
			val.tag_item = read_network_endian(buf, offset, 8)
			offset += 8
		}
		content, err, consume := cbor_parse(buf, offset)
		if content != nil && offset + consume <= len(buf) {
			offset += consume
			val.tag_content = content
		} else {
			err = err
			val = nil
		}
	} else if ctype == CBOR_TYPE_SIMPLE {
		offset++
		if addition < 20 {
			val = NewUndef()
			val.ctrl = addition
		} else if addition == 20 {
			val = NewBoolean(false)
		} else if addition == 21 {
			val = NewBoolean(true)
		} else if addition == 22 {
			val = NewNull()
		} else if addition == 23 {
			val = NewUndef()
		} else if addition == 24 && offset + 1 <= len(buf) {
			val = NewExt()
			val.ctrl = int(read_network_endian(buf, offset, 1))
			offset += 1
		} else if addition == 25 && offset + 2 <= len(buf) {
			// float16
			u64 := uint64(binary.BigEndian.Uint16(buf[offset:]))
			sign := (u64 & 0x8000) >> 15
			exp := (u64 >> 10) & 0x1F
			frac := u64 & 0x3FF

			u64 = frac << (52 - 10)
			if sign == 1 {
				u64 |= 1 << 63
			}
			if exp == 0 {

			} else if exp == 31 {
				u64 |= 0x7FF << 52
			} else {
				u64 |= (exp - 15 + 1023) << 52
			}
			offset += 2
			val = NewFloat(math.Float64frombits(u64))
		} else if addition == 26 && offset + 4 <= len(buf) {
			// float32
			u64 := uint64(binary.BigEndian.Uint32(buf[offset:]))
			sign := u64 >> 31
			exp := (u64 >> 23) & 0xFF
			frac := u64 & 0x7FFFFF

			u64 = frac << (52 - 23)
			if sign == 1 {
				u64 |= 1 << 63
			}
			if exp == 0 {

			} else if exp == 255 {
				u64 |= 0x7FF << 52
			} else {
				u64 |= (exp - 127 + 1023) << 52
			}
			offset += 4
			val = NewFloat(math.Float64frombits(u64))
		} else if addition == 27 && offset + 8 <= len(buf) {
			// float64
			u64 := uint64(binary.BigEndian.Uint64(buf[offset:]))
			offset += 8
			val = NewFloat(math.Float64frombits(u64))
		} else {
			val = nil
			err = fmt.Errorf("unknown addition value when decode simple value")
		}
	} else {
		val = nil
		err = fmt.Errorf("unknown decode cbor type")
	}
	return val, err, offset - origin
}

func CBORDecode(buf []byte) (val *CborValue, err error) {
	val, err, _ = cbor_parse(buf, 0)
	return
}
