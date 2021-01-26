package cbor

import "fmt"
import "bytes"
import "strings"
import "strconv"

type CborValue struct {
	ctype int
	blob bytes.Buffer
	integer uint64
	real float64
	ctrl int

	// tag
	tag_item uint64
	tag_content *CborValue

	key, value *CborValue	// pair
	first, last *CborValue	// container
	next, prev *CborValue	// entry
	parent *CborValue
}

const (
	CBOR_TYPE_UINT int = iota
	CBOR_TYPE_NEGINT
	CBOR_TYPE_BYTESTRING
	CBOR_TYPE_STRING
	CBOR_TYPE_ARRAY
	CBOR_TYPE_MAP
	CBOR_TYPE_TAG
	CBOR_TYPE_SIMPLE
	CBOR__TYPE_PAIR
)

const (
	CBOR_SIMPLE_FALSE int = 20
	CBOR_SIMPLE_TRUE int = 21
	CBOR_SIMPLE_NULL int = 22
	CBOR_SIMPLE_UNDEF int = 23
	CBOR_SIMPLE_EXTENSION int = 24
	CBOR_SIMPLE_REAL int = 25
)

func (self *CborValue) Compare(T interface{}) bool {
	switch T.(type) {
	case string:
		v := T.(string)
		if self.ctype == CBOR_TYPE_BYTESTRING || self.ctype == CBOR_TYPE_STRING {
			if len(v) == self.blob.Len() {
				if bytes.Compare(self.blob.Bytes(), []byte(v)) == 0 {
					return true
				} else {
					return false
				}
			}
		}
	case float32:
		if self.ctype == CBOR_TYPE_SIMPLE && self.ctrl == CBOR_SIMPLE_REAL {
			return float32(self.Float()) == T.(float32)
		}
	case float64:
		if self.ctype == CBOR_TYPE_SIMPLE && self.ctrl == CBOR_SIMPLE_REAL {
			return self.Float() == T.(float64)
		}
	case nil:
		if self.IsNull() {
			return true
		}
	}
	return false
}

func (val *CborValue) IsString() bool {
	return val != nil && (val.ctype == CBOR_TYPE_STRING || val.ctype == CBOR_TYPE_BYTESTRING)
}
func (val *CborValue) IsMap() bool {
	return val != nil && val.ctype == CBOR_TYPE_MAP
}
func (val *CborValue) IsArray() bool {
	return val != nil && val.ctype == CBOR_TYPE_ARRAY
}
func (val *CborValue) IsInteger() bool {
	return val != nil && (val.ctype == CBOR_TYPE_UINT || val.ctype == CBOR_TYPE_NEGINT)
}
func (val *CborValue) IsFloat() bool {
	return val != nil && val.ctype == CBOR_TYPE_SIMPLE && val.ctrl == CBOR_SIMPLE_REAL
}
func (val *CborValue) IsBoolean() bool {
	return val != nil && val.ctype == CBOR_TYPE_SIMPLE && (val.ctrl == CBOR_SIMPLE_FALSE || val.ctrl == CBOR_SIMPLE_TRUE)
}
func (val *CborValue) IsNull() bool {
	return val != nil && val.ctype == CBOR_TYPE_SIMPLE && val.ctrl == CBOR_SIMPLE_NULL
}
func (val *CborValue) IsContainer() bool {
	return val != nil && (val.ctype == CBOR_TYPE_MAP || val.ctype == CBOR_TYPE_ARRAY)
}
func (val *CborValue) ContainerEmpty() bool {
	if val.IsContainer() {
		return val.first == nil && val.last == nil
	}
	return false
}

func New(value interface{}) *CborValue {
	switch v := value.(type) {
	case uint:
		return NewInteger(int64(v))
	case uint8:
		return NewInteger(int64(v))
	case uint16:
		return NewInteger(int64(v))
	case uint32:
		return NewInteger(int64(v))
	case uint64:
		return NewInteger(int64(v))
	case int:
		return NewInteger(int64(v))
	case int8:
		return NewInteger(int64(v))
	case int16:
		return NewInteger(int64(v))
	case int32:
		return NewInteger(int64(v))
	case int64:
		return NewInteger(int64(v))
	case bool:
		return NewBoolean(bool(v))
	case string:
		return NewString(string(v))
	case []byte:
		return NewBytestring([]byte(v))
	case nil:
		return NewNull()
	case float32:
		return NewFloat(float64(v))
	case float64:
		return NewFloat(float64(v))
	case map[string]interface{}:
		val := NewMap()
		for k, v := range map[string]interface{}(v) {
			key := NewString(k)
			ele := New(v)
			if ele != nil {
				pair := NewPair(key, ele)
				val.ContainerInsertTail(pair)
			}
		}
		return val
	case []interface{}:
		val := NewArray()
		for _, item := range []interface{}(v) {
			ele := New(item)
			val.ContainerInsertTail(ele)
		}
		return val
	case *CborValue:
		return value.(*CborValue).Duplicate()
	}
	return nil
}

func NewTag() *CborValue {
	val := new(CborValue)
	val.ctype = CBOR_TYPE_TAG
	return val
}

func NewUndef() *CborValue {
	val := new(CborValue)
	val.ctype = CBOR_TYPE_SIMPLE
	val.ctrl = CBOR_SIMPLE_UNDEF
	return val
}

func NewExt() *CborValue {
	val := new(CborValue)
	val.ctype = CBOR_TYPE_SIMPLE
	val.ctrl = CBOR_SIMPLE_EXTENSION
	return val
}
func NewMap() *CborValue {
	val := new(CborValue)
	val.ctype = CBOR_TYPE_MAP
	return val
}

func NewArray() *CborValue {
	val := new(CborValue)
	val.ctype = CBOR_TYPE_ARRAY
	return val
}

func NewBoolean(b bool) *CborValue {
	val := new(CborValue)
	val.ctype = CBOR_TYPE_SIMPLE
	if b {
		val.ctrl = CBOR_SIMPLE_TRUE
	} else {
		val.ctrl = CBOR_SIMPLE_FALSE
	}
	return val
}

func NewInteger(i int64) *CborValue {
	val := new(CborValue)
	if i < 0 {
		val.ctype = CBOR_TYPE_NEGINT
		val.integer = uint64(-i -1)
	} else {
		val.ctype = CBOR_TYPE_UINT
		val.integer = uint64(i)
	}
	return val
}

func NewString(s string) *CborValue {
	val := new(CborValue)
	val.ctype = CBOR_TYPE_STRING
	val.blob = bytes.Buffer{}
	val.blob.WriteString(s)
	return val
}

func NewBytestring(b []byte) *CborValue {
	val := new(CborValue)
	val.ctype = CBOR_TYPE_BYTESTRING
	val.blob = bytes.Buffer{}
	val.blob.Write(b)
	return val
}

func NewNull() *CborValue {
	val := new(CborValue)
	val.ctype = CBOR_TYPE_SIMPLE
	val.ctrl = CBOR_SIMPLE_NULL
	return val
}

func NewFloat(real float64) *CborValue {
	val := new(CborValue)
	val.ctype = CBOR_TYPE_SIMPLE
	val.ctrl = CBOR_SIMPLE_REAL
	val.real = real
	return val
}
func NewPair(key *CborValue, val *CborValue) *CborValue {
	pair := new(CborValue)
	pair.ctype = CBOR__TYPE_PAIR
	pair.key = key
	pair.value = val
	key.parent = pair
	val.parent = pair
	return pair
}

func (val *CborValue) Integer() int64 {
	if val == nil {
		return 0
	}

	if val.ctype == CBOR_TYPE_UINT {
		return int64(val.integer)
	} else if val.ctype == CBOR_TYPE_NEGINT {
		return -1 - int64(val.integer)
	} else if val.ctype == CBOR_TYPE_SIMPLE && val.ctrl == CBOR_SIMPLE_REAL {
		return int64(val.real)
	}
	return 0
}

func (val *CborValue) Float() float64 {
	if val == nil {
		return 0.0
	}

	if val.ctype == CBOR_TYPE_UINT {
		return float64(val.integer)
	} else if val.ctype == CBOR_TYPE_NEGINT {
		return float64(-1 - int64(val.integer))
	} else if val.ctype == CBOR_TYPE_SIMPLE && val.ctrl == CBOR_SIMPLE_REAL {
		return val.real
	}
	return .0
}

func (val *CborValue) String() string {
	if val == nil {
		return ""
	}

	if val.ctype == CBOR_TYPE_STRING || val.ctype == CBOR_TYPE_BYTESTRING {
		return val.blob.String()
	}
	return ""
}

func (val *CborValue) StringBytes() []byte {
	if val == nil {
		return []byte{}
	}

	if val.ctype == CBOR_TYPE_STRING || val.ctype == CBOR_TYPE_BYTESTRING {
		return val.blob.Bytes()
	}
	return []byte("")
}

func (val *CborValue) StringSize() int {
	if val == nil {
		return 0
	}

	if val.ctype == CBOR_TYPE_STRING || val.ctype == CBOR_TYPE_BYTESTRING {
		return val.blob.Len()
	}
	return 0
}

func (val *CborValue) Boolean() bool {
	if val == nil {
		return false
	}

	if val.ctype == CBOR_TYPE_SIMPLE {
		if val.ctrl == CBOR_SIMPLE_TRUE {
			return true
		} else if val.ctrl == CBOR_SIMPLE_FALSE {
			return false
		}
	}
	return false
}

func (pair *CborValue) PairKey() *CborValue {
	if pair != nil && pair.ctype == CBOR__TYPE_PAIR {
		return pair.key
	}
	return nil
}

func (pair *CborValue) PairValue() *CborValue {
	if pair != nil && pair.ctype == CBOR__TYPE_PAIR {
		return pair.value
	}
	return nil
}

func (pair *CborValue) SetValue(val *CborValue) {
	if val == nil || val.parent != nil {
		return
	}

	if pair != nil && pair.ctype == CBOR__TYPE_PAIR {
		v := pair.value
		v.parent = nil
		val.parent = pair
		pair.value = val
	}
}

func (s *CborValue) BlobAppendByte(b byte) {
	if s.IsString() {
		s.blob.WriteByte(b)
	}
}

func (s *CborValue) BlobAppendRune(r rune) {
	if s.IsString() {
		s.blob.WriteRune(r)
	}
}

func (s *CborValue) BlobAppend(str string) {
	if s.IsString() {
		s.blob.WriteString(str)
	}
}

func (s *CborValue) BlobAppendFormat(format string, va ...interface{}) {
	if s.IsString() {
		s.blob.WriteString(fmt.Sprintf(format, va...))
	}
}

func (container *CborValue) ContainerInsertTail(val *CborValue) {
	if val == nil || !container.IsContainer() {
		return
	}
	if val.parent != nil {
		val = val.Duplicate()
	}

	val.prev = container.last
	if container.last != nil {
		container.last.next = val;
	}
	container.last = val
	if container.first == nil {
		container.first = val
	}
	val.parent = container
}

func (container *CborValue) ContainerInsertHead(val *CborValue) {
	if val == nil || !container.IsContainer() {
		return
	}
	if val.parent != nil {
		val = val.Duplicate()
	}
	val.next = container.first
	if container.first != nil {
		container.first.prev = val
	}
	container.first = val
	if container.last == nil {
		container.last = val
	}
	val.parent = container
}

func (container *CborValue) ContainerRemove(val *CborValue) {
	if val != nil && container.IsContainer() && val.parent == container {
		prev := val.prev
		next := val.next

		if prev != nil {
			prev.next = next
		}
		if next != nil {
			next.prev = prev
		}
		if container.first == val {
			container.first = next
		}
		if container.last == val {
			container.last = prev
		}
		val.prev = nil
		val.next = nil
		val.parent = nil
	}
}

func (container *CborValue) PointerGet(path string) *CborValue {
	if !container.IsContainer() {
		return nil
	}

	var current *CborValue = nil
	split := strings.Split(path, "/")
	for i, ele := range split {
		ele = strings.Replace(ele, "~1", "/", -1)
		ele = strings.Replace(ele, "~0", "~", -1)
		if len(ele) == 0 && i == 0 {
			current = container
			continue
		} else {
			if current.IsMap() {
				var elm *CborValue = nil
				for elm = current.ContainerFirst(); elm != nil; elm = current.ContainerNext(elm) {
					if elm.PairKey().Compare(ele) {
						break
					}
				}
				if elm != nil {
					current = elm.PairValue()
					continue
				}
			} else if current.IsArray() {
				if ele == "-" {
					current = current.ContainerLast()
					continue
				} else {
					idx, err := strconv.ParseInt(ele, 10, 32)
					if err == nil && idx >= 0 {
						var elm *CborValue = nil
						for elm = current.ContainerFirst(); elm != nil && idx > 0; idx-- {
							elm = current.ContainerNext(elm)
						}
						if elm != nil {
							current = elm
							continue
						}
					}
				}
			}
		}
		current = nil
		break
	}
	return current
}

func (container *CborValue) PointerRemove(path string) *CborValue {
	remval := container.PointerGet(path)
	if remval != nil {
		parent := remval.parent
		if parent != nil {
			if parent.ctype == CBOR__TYPE_PAIR {
				pair := parent
				parent = pair.parent
				if parent != nil {
					parent.ContainerRemove(pair)
					pair.value = nil
					return remval
				}
			} else if parent.ctype == CBOR_TYPE_ARRAY {
				parent.ContainerRemove(remval)
				return remval
			}
		}
	}
	return nil
}

func (container *CborValue) PointerMove(path string, dest string) *CborValue {
	var current *CborValue = nil
	var root *CborValue = nil
	var value *CborValue = nil
	last := false
	if !container.IsContainer() {
		return nil
	}

	split := strings.Split(path, "/")
	for i, ele := range split {
		ele = strings.Replace(ele, "~1", "/", -1)
		ele = strings.Replace(ele, "~0", "~", -1)
		if i == len(split) - 1 {
			last = true
		}
		if len(ele) == 0 && i == 0 {
			current = container
			continue
		} else {
			if current.IsMap() {
				var elm *CborValue = nil
				for elm = current.ContainerFirst(); elm != nil; elm = current.ContainerNext(elm) {
					if elm.PairKey().Compare(ele) {
						break
					}
				}
				if elm != nil {
					if last {
						root = current
						value = elm
					} else {
						current = elm.PairValue()
					}
					continue
				}
			} else if current.IsArray() {
				if ele == "-" {
					if last {
						root = current
						value = current.ContainerLast()
					} else {
						current = current.ContainerLast()
					}
					continue
				} else {
					idx, err := strconv.ParseInt(ele, 10, 32)
					if err == nil && idx >= 0 {
						var elm *CborValue = nil
						for elm = current.ContainerFirst(); elm != nil && idx > 0; idx-- {
							elm = current.ContainerNext(elm)
						}
						if elm != nil {
							if last {
								root = current
								value = elm
							} else {
								current = elm
							}
							continue
						}
					}
				}
			}
			current = nil
			break
		}
	}

	if current == nil || root == nil || value == nil {
		return nil
	}

	last = false
	current = nil

	split = strings.Split(dest, "/")
	for i, ele := range split {
		ele = strings.Replace(ele, "~1", "/", -1)
		ele = strings.Replace(ele, "~0", "~", -1)
		if i == len(split) - 1 {
			last = true
		}
		if current == value {
			current = nil
			break
		}
		if len(ele) == 0 && i == 0 {
			current = container
			continue
		} else {
			if current.IsMap() {
				var elm *CborValue = nil
				for elm = current.ContainerFirst(); elm != nil; elm = current.ContainerNext(elm) {
					if elm.PairKey().Compare(ele) {
						break
					}
				}
				if elm != nil {
					if last {
						root.ContainerRemove(value)
						if root.IsMap() {
							tmp := value.PairValue()
							tmp.parent = nil
							value.value = nil
							value = tmp
						}
						elm.SetValue(value)
					} else {
						current = elm.PairValue()
					}
					continue
				}
			} else if current.IsArray() {
				if ele == "-" {
					if last {
						root.ContainerRemove(value)
						if root.IsMap() {
							tmp := value.PairValue()
							tmp.parent = nil
							value.value = nil
							value = tmp
						}
						current.ContainerInsertTail(value)
					} else {
						current = current.ContainerLast()
					}
					continue
				} else {
					idx, err := strconv.ParseInt(ele, 10, 32)
					if err == nil && idx >= 0 {
						var elm *CborValue = nil
						for elm = current.ContainerFirst(); elm != nil && idx > 0; idx-- {
							elm = current.ContainerNext(elm)
						}
						if elm != nil {
							if last {
								root.ContainerRemove(value)
								if root.IsMap() {
									tmp := value.PairValue()
									tmp.parent = nil
									value.value = nil
									value = tmp
								}
								current.ContainerInsertBefore(elm, value)
							} else {
								current = elm
							}
							continue
						}
					}
				}
			}
			current = nil
			break
		}
	}
	return current
}

func (container *CborValue) PointerAdd(path string, val *CborValue) *CborValue {
	var current *CborValue = nil
	last := false
	if !container.IsContainer() || val == nil {
		return nil
	}

	split := strings.Split(path, "/")
	for i, ele := range split {
		ele = strings.Replace(ele, "~1", "/", -1)
		ele = strings.Replace(ele, "~0", "~", -1)
		if i == len(split) - 1 {
			last = true
		}
		if len(ele) == 0 && i == 0 {
			current = container
			continue
		} else {
			if current.IsMap() {
				var elm *CborValue = nil
				for elm = current.ContainerFirst(); elm != nil; elm = current.ContainerNext(elm) {
					if elm.PairKey().Compare(ele) {
						break
					}
				}
				if elm != nil {
					if last {
						elm.SetValue(val)
					} else {
						current = elm.PairValue()
					}
					continue
				} else if last {
					key := NewString(ele)
					pair := NewPair(key, val)
					current.ContainerInsertTail(pair)
					continue
				}
			} else if current.IsArray() {
				if ele == "-" {
					if last {
						current.ContainerInsertTail(val)
					} else {
						current = current.ContainerLast()
					}
					continue
				} else {
					idx, err := strconv.ParseInt(ele, 10, 32)
					if err == nil && idx >= 0 {
						var elm *CborValue = nil
						for elm = current.ContainerFirst(); elm != nil && idx > 0; idx-- {
							elm = current.ContainerNext(elm)
						}
						if elm != nil {
							if last {
								current.ContainerInsertBefore(val, elm)
							} else {
								current = elm
							}
							continue
						}
					}
				}
			}
		}
		current = nil
		break
	}
	return current
}

func (container *CborValue) PointerSet(path string, value interface{}) *CborValue {
	val := New(value)
	if val != nil {
		return container.PointerAdd(path, val)
	}
	return nil
}

func (val *CborValue) Duplicate() *CborValue { // deep copy
	if val == nil {
		return New(nil)
	}

	if val.ctype == CBOR_TYPE_UINT || val.ctype == CBOR_TYPE_NEGINT {
		return NewInteger(val.Integer())
	} else if val.ctype == CBOR_TYPE_STRING {
		return NewString(val.String())
	} else if val.ctype == CBOR_TYPE_SIMPLE {
		if val.ctrl == CBOR_SIMPLE_FALSE {
			return NewBoolean(false)
		} else if val.ctrl == CBOR_SIMPLE_TRUE {
			return NewBoolean(true)
		} else if val.ctrl == CBOR_SIMPLE_NULL {
			return NewNull()
		} else if val.ctrl == CBOR_SIMPLE_REAL {
			return NewFloat(val.Float())
		}
	} else if val.ctype == CBOR__TYPE_PAIR {
		return NewPair(val.PairKey().Duplicate(), val.PairValue().Duplicate())
	} else if val.IsContainer() {
		dup := new(CborValue)
		dup.ctype = val.ctype
		for ele := val.ContainerFirst(); ele != nil; ele = val.ContainerNext(ele) {
			dup.ContainerInsertTail(ele.Duplicate())
		}
		return dup
	} else if val.ctype == CBOR_TYPE_BYTESTRING {
		return NewBytestring(val.blob.Bytes())
	}
	return nil
}

func (container *CborValue) ContainerInsertBefore(elm *CborValue, val *CborValue) {
	if container.IsContainer() && elm != nil && elm.parent == container {
		prev := elm.prev
		elm.prev = val
		if prev != nil {
			prev.next = val
		} else {
			container.first = val
		}
		val.prev = prev
		val.next = elm
		val.parent = container
	}
}

func (container *CborValue) ContainerInsertAfter(elm *CborValue, val *CborValue) {
	if container.IsContainer() && elm != nil && elm.parent == container {
		next := elm.next
		elm.next = val
		if next != nil {
			next.prev = val
		} else {
			container.last = val
		}
		val.prev = elm
		val.next = next
		val.parent = container
	}
}

func (container *CborValue) ContainerFirst() *CborValue {
	if container.IsContainer() {
		return container.first
	}
	return nil
}
func (container *CborValue) ContainerLast() *CborValue {
	if container.IsContainer() {
		return container.last
	}
	return nil
}

func (container *CborValue) ContainerNext(val *CborValue) *CborValue {
	if val != nil && container.IsContainer() && val.parent == container {
		return val.next
	}
	return nil
}

func (container *CborValue) ContainerPrev(val *CborValue) *CborValue {
	if val != nil && container.IsContainer() && val.parent == container {
		return val.prev
	}
	return nil
}
