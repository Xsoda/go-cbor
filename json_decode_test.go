package cbor

import "fmt"
import "testing"

func TestJSONDecode(t *testing.T) {
	json := []byte(`{      "string": "foo\t\u4F60\u4f60\u000d\n",
            // line comment
            "max_integer": 9223372036854775807,
            "min_integer": -9223372036854775808,
            /* block comment
            // line comment
            -- end comment */
            "array": [true, false, null],
            "utf-8_2": "\u0123 two-byte UTF-8",
            "utf-8_3": "\u0821 three-byte UTF-8",
            "utf-8_4": "\uD834\uDD1E surrogate, four-byte UTF-8",
            "e": 2.718281828459045235360287471352662498,
            "float64": 1.1,
            "float32": 3.4028234663852886e+38,
            "float16": 1.5
        }`)
	val, err := JSONDecode(json)
	if val == nil || err != nil {
		t.Log("decode json fail")
		t.Fail()
	}

	s := JSONEncode(val)
	if s.Len() > 0 {
		v, err := JSONDecode(s.Bytes())
		if v == nil || err != nil {
			t.Errorf("decode encoded json fail: %s", s.String())
			t.Fail()
		}
		fmt.Println(s.String())
	}

	if val.PointerGet("/float16").Integer() != 1 {
		t.Log("convert /float16 to integer fail")
		t.Fail()
	}

	if val.PointerGet("/max_integer").Float() != 9223372036854775807.0 {
		t.Log("convert /max_integer to float fail")
		t.Fail()
	}
}
