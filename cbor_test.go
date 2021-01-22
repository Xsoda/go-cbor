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

	v = New(m)
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

	t.Log(JSONEncode(v).String())
}
