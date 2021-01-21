package cbor

import "math"
import "bytes"
import "encoding/binary"

func write_word(buf *bytes.Buffer, w uint16) {
	flat := make([]byte, 2)
	binary.BigEndian.PutUint16(flat, w)
	buf.Write(flat)
}

func write_dword(buf *bytes.Buffer, dw uint32) {
	flat := make([]byte, 4)
	binary.BigEndian.PutUint32(flat, dw)
	buf.Write(flat)
}

func write_qword(buf *bytes.Buffer, qw uint64) {
	flat := make([]byte, 8)
	binary.BigEndian.PutUint64(flat, qw)
	buf.Write(flat)
}

func cbor_dump(val *CborValue) *bytes.Buffer {
	var dst = new(bytes.Buffer)
	var ctype uint8 = uint8(val.ctype)
	ctype <<= 5
	if val.ctype == CBOR_TYPE_UINT || val.ctype == CBOR_TYPE_NEGINT {
		if val.integer < 24 {
			ctype |= uint8(val.integer)
			dst.WriteByte(ctype)
		} else if val.integer <= 0xFF {
			ctype |= 24
			dst.WriteByte(ctype)
			dst.WriteByte(uint8(val.integer))
		} else if val.integer <= 0xFFFF {
			ctype |= 25
			dst.WriteByte(ctype)
			write_word(dst, uint16(val.integer))
		} else if val.integer <= 0xFFFFFFFF {
			ctype |= 26
			dst.WriteByte(ctype)
			write_dword(dst, uint32(val.integer))
		} else {
			ctype |= 27
			dst.WriteByte(ctype)
			write_qword(dst, val.integer)
		}
	} else if val.ctype == CBOR_TYPE_BYTESTRING || val.ctype == CBOR_TYPE_STRING {
		len := val.StringSize()
		if len < 24 {
			ctype |= uint8(len)
			dst.WriteByte(ctype)
			dst.Write(val.blob.Bytes())
		} else if len <= 0xFF {
			ctype |= 24
			dst.WriteByte(ctype)
			dst.WriteByte(uint8(len))
			dst.Write(val.blob.Bytes())
		} else if len <= 0xFFFF {
			ctype |= 25
			dst.WriteByte(ctype)
			write_word(dst, uint16(len))
			dst.Write(val.blob.Bytes())
		} else if len <= 0xFFFFFFFF {
			ctype |= 26
			dst.WriteByte(ctype)
			write_dword(dst, uint32(len))
			dst.Write(val.blob.Bytes())
		} else {
			ctype |= 27
			dst.WriteByte(ctype)
			write_qword(dst, uint64(len))
			dst.Write(val.blob.Bytes())
		}
	} else if val.ctype == CBOR__TYPE_PAIR {
		cbor_dump(val.key)
		cbor_dump(val.value)
	} else if val.ctype == CBOR_TYPE_ARRAY || val.ctype == CBOR_TYPE_MAP {
		ctype |= 31
		dst.WriteByte(ctype)
		for ele := val.ContainerFirst(); ele != nil; ele = val.ContainerNext(ele) {
			cbor_dump(ele)
		}
		dst.WriteByte(0xFF)
	} else if val.ctype == CBOR_TYPE_SIMPLE {
		if val.ctrl == CBOR_SIMPLE_FALSE {
			ctype |= 20
			dst.WriteByte(ctype)
		} else if val.ctrl == CBOR_SIMPLE_TRUE {
			ctype |= 21
			dst.WriteByte(ctype)
		} else if val.ctrl == CBOR_SIMPLE_NULL {
			ctype |= 22
			dst.WriteByte(ctype)
		} else if val.ctrl == CBOR_SIMPLE_REAL {
			u64 := math.Float64bits(val.real)
			exp := int(u64 >> 32) & 0x7FF
			sign := int(u64 >> 63)
			frac := u64 & 0xFFFFFFFFFFFFF
			var frac_bitcnt int
			if frac > 0 {
				for frac_bitcnt = 0; frac_bitcnt < 52; frac_bitcnt++ {
					if frac & 1 > 0 {
						break
					}
					frac >>= 1
				}
			} else {
				frac_bitcnt = 52
			}

			frac_bitcnt = 52 - frac_bitcnt
			frac = u64 & 0xFFFFFFFFFFFFF
			if exp == 0 || exp == 0x7FF {
				if frac_bitcnt <= 10 {
					ctype |= 25
					dst.WriteByte(ctype)

					u16 := uint16(frac >> (52 - 10))
					if sign == 1 {
						u16 |= 1 << 15
					}
					if exp > 0 {
						u16 |= 0x1F << 10
					}
					write_word(dst, u16)
				} else if frac_bitcnt <= 23 {
					ctype |= 26
					dst.WriteByte(ctype)

					u32 := uint32(frac >> (52 - 23))
					if sign == 1 {
						u32 |= 1 << 31
					}
					if exp > 0 {
						u32 |= 0x3F << 23
					}
					write_dword(dst, u32)
				} else {
					ctype |= 27
					dst.WriteByte(ctype)

					u64 = frac
					if sign == 1 {
						u64 |= 1 << 63
					}
					if exp > 0 {
						u64 |= 0x7FF << 52
					}
					write_qword(dst, u64)
				}
			} else {
				exp -= 1023
				if exp >= -14 && exp <= 15 && frac_bitcnt <= 10 {
					ctype |= 25
					dst.WriteByte(ctype)

					u16 := uint16(frac >> (52 - 10))
					if sign == 1 {
						u16 |= 1 << 15
					}
					u16 |= uint16(exp + 15) << 10
					write_word(dst, u16)
				} else if exp >= -126 && exp <= 127 && frac_bitcnt <= 23 {
					ctype |= 26
					dst.WriteByte(ctype)

					u32 := uint32(frac >> (52 - 23))
					if sign == 1 {
						u32 |= 1 << 31
					}
					u32 |= uint32(exp + 127) << 23
					write_dword(dst, u32)
				} else {
					ctype |= 27
					dst.WriteByte(ctype)

					u64 = frac
					if sign == 1 {
						u64 |= 1 << 63
					}
					u64 |= uint64(exp + 1023) << 52
					write_qword(dst, u64)
				}
			}

		}
	}
	return dst
}

func CBOREncode(val *CborValue) *bytes.Buffer {
	return cbor_dump(val)
}
