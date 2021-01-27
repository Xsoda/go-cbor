package cbor

import "testing"

func TestNew(t *testing.T) {
	v := New("abcde")
	if v.String() != "abcde" {
		t.Log("New string fail")
		t.Fail()
	}

	v = New(3.14159)
	if v.Float() != 3.14159 {
		t.Log("New float fail")
		t.Fail()
	}

	v = New(false)
	if v.Boolean() != false {
		t.Log("New boolean fail")
		t.Fail()
	}

	v = New(123456789)
	if v.Integer() != 123456789 {
		t.Log("New integer fail")
		t.Fail()
	}

	v = New(-1234567890)
	if v.Integer() != -1234567890 {
		t.Log("New integer fail", v.Integer())
		t.Fail()
	}

	v = New(nil)
	if !v.IsNull() {
		t.Log("New null fail")
		t.Fail()
	}

	m := map[string]interface{}{}
	m["float"] = 3.14159
	m["integer"] = 921313123
	m["boolean"] = false
	m["string"] = "string"

	n := New(m)
	v = n.Duplicate()
	if v == nil {
		t.Log("New map fail")
		t.Fail()
	}

	if v.PointerGet("/float").Float() != 3.14159 {
		t.Log("/float value fail")
		t.Fail()
	}

	v.PointerSet("/integer", -11111)
	if v.PointerGet("/integer").Integer() != -11111 {
		t.Log("set /integer value fail")
		t.Fail()
	}

	if v.PointerGet("/boolean").Boolean() != false {
		t.Log("get /boolean value fail")
		t.Fail()
	}

	if v.PointerGet("/string").String() != "string" {
		t.Log("get /string value fail")
		t.Fail()
	}
	v.PointerAdd("/another", New("another"))

	if v.PointerGet("/another").String() != "another" {
		t.Log("get /another fail")
		t.Fail()
	}

	if v.PointerRemove("/another").String() != "another" {
		t.Log("remove /another fail")
		t.Fail()
	}

	v.PointerSet("/dup", v)

	s := []interface{}{"foo", "bar"}
	v.PointerSet("/array", s)

	if !v.PointerGet("/array").IsArray() {
		t.Log("pointer set array fail")
		t.Fail()
	}

	v.PointerMove("/string", "/array/-")
	if v.PointerGet("/string").String() != "" {
		t.Errorf("PointerMove fail: %s\n", JSONEncode(v).String())
		t.Fail()
	}

	if v.PointerGet("/array/2").String() != "string" {
		t.Log("pointer move fail")
		t.Fail()
	}
	v.PointerMove("/array/-", "/moved")

	if v.PointerGet("/moved").String() != "string" {
		t.Log("pointer move fail")
		t.Fail()
	}

	v.PointerAdd("/array/-", New(100))
	if v.PointerGet("/array/-").Integer() != 100 {
		t.Log("pointer add fail")
		t.Fail()
	}

	v.PointerAdd("/array/0", New(1000))

	if v.PointerGet("/array/0").Integer() != 1000 {
		t.Log("pointer add fail")
		t.Fail()
	}

	v.PointerMove("/array/-", "/array/0")
	if !v.PointerGet("/array/0").IsInteger() || v.PointerGet("/array/0").Integer() != 100 {
		t.Log("pointer move fail")
		t.Fail()
	}
}

func TestPointerGet(t *testing.T) {
	v, _ := JSONDecode([]byte(` {
            "foo": ["bar", "baz"],
            "": 0,
            "a/b": 1,
            "c%d": 2,
            "e^f": 3,
            "g|h": 4,
            "i\\j": 5,
            "k\"l": 6,
            " ": 7,
            "m~n": 8
        }`))

	if v == nil {
		t.Log("parse json document fail")
		t.Fail()
	}

	if v.PointerGet("") != v {
		t.Log("get empty path fail")
		t.Fail()
	}

	if !v.PointerGet("/foo").IsArray() {
		t.Log("get /foo fail")
		t.Fail()
	}

	if v.PointerGet("/foo/0").String() != "bar" {
		t.Log("get /foo/0 fail")
		t.Fail()
	}

	if v.PointerGet("/foo").PointerGet("/-").String() != "baz" {
		t.Log("get /foo /- fail")
		t.Fail()
	}

	if v.PointerGet("/").Integer() != 0 {
		t.Log("get / fail")
		t.Fail()
	}

	if v.PointerGet("/a~1b").Integer() != 1 {
		t.Log("get /a~1b fail")
		t.Fail()
	}

	if v.PointerGet("/c%d").Integer() != 2 {
		t.Log("get /c_%_d fail")
		t.Fail()
	}

	if v.PointerGet("/e^f").Integer() != 3 {
		t.Log("get /e^f fail")
		t.Fail()
	}

	if v.PointerGet("/g|h").Integer() != 4 {
		t.Log("get /g|h fail")
		t.Fail()
	}

	if v.PointerGet("/i\\j").Integer() != 5 {
		t.Log("get /i\\j fail")
		t.Fail()
	}

	if v.PointerGet("/k\"l").Integer() != 6 {
		t.Log("get /k\"l fail")
		t.Fail()
	}

	if v.PointerGet("/ ").Integer() != 7 {
		t.Log("get `/ ` fail")
		t.Fail()
	}

	if v.PointerGet("/m~0n").Integer() != 8 {
		t.Log("get /m~0n fail")
		t.Fail()
	}

	if v.PointerGet("/Bar").Boolean() != false {
		t.Log("get /Bar fail")
		t.Fail()
	}
}

func TestContainer(t *testing.T) {
	v := NewArray()
	v.ContainerInsertHead(New(0))
	v.ContainerInsertTail(New(1))

	ele := v.PointerGet("/0")
	if ele == nil || ele.Integer() != 0 {
		t.Log("insert head fail")
		t.Fail()
	}

	ele = v.PointerGet("/-")
	if ele == nil || ele.Integer() != 1 {
		t.Log("insert tail fail")
		t.Fail()
	}

	if ele != v.PointerGet("/1") {
		t.Log("Pointerget /1 fail")
		t.Fail()
	}

	v.ContainerInsertAfter(ele, New(3))
	ele = v.PointerGet("/2")
	if ele == nil || ele.Integer() != 3 {
		t.Log("InsertAfter fail")
		t.Fail()
	}

	v.ContainerInsertBefore(ele, New(2))
	ele = v.PointerGet("/2")
	if ele == nil || ele.Integer() != 2 {
		t.Log("InsertBefore fail")
		t.Fail()
	}

	ele = v.PointerGet("/-")
	if ele == nil || ele.Integer() != 3 {
		t.Log("assume tail element fail")
		t.Fail()
	}

	ele = v.ContainerPrev(ele)
	if ele == nil || ele.Integer() != 2 {
		t.Log("container prev fail")
		t.Fail()
	}

	ele = v.ContainerNext(ele)
	if ele == nil || ele.Integer() != 3 {
		t.Log("container next fail")
		t.Fail()
	}

	ele = v.ContainerNext(ele)
	if ele != nil {
		t.Log("container tail next fail")
		t.Fail()
	}

	ele = v.ContainerFirst()
	if ele == nil || ele.Integer() != 0 {
		t.Log("container first fail")
		t.Fail()
	}

	if ele = v.ContainerNext(ele); ele == nil || ele.Integer() != 1 {
		t.Log("container first next fail")
		t.Fail()
	}

	ele = v.ContainerLast()
	if ele == nil || ele.Integer() != 3 {
		t.Log("container last fail")
		t.Fail()
	}
}
