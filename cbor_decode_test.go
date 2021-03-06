package cbor

import "testing"

var content = []string{
    "\x00",
    "\x01",
    "\x0a",
    "\x17",
    "\x18\x18",
    "\x18\x19",
    "\x18\x64",
    "\x19\x03\xe8",
    "\x1a\x00\x0f\x42\x40",
    "\x1b\x00\x00\x00\xe8\xd4\xa5\x10\x00",
    "\x1b\xff\xff\xff\xff\xff\xff\xff\xff",
    "\xc2\x49\x01\x00\x00\x00\x00\x00\x00\x00\x00",
    "\x3b\xff\xff\xff\xff\xff\xff\xff\xff",
    "\xc3\x49\x01\x00\x00\x00\x00\x00\x00\x00\x00",
    "\x20",
    "\x29",
    "\x38\x63",
    "\x39\x03\xe7",
    "\xf9\x00\x00",
    "\xf9\x80\x00",
    "\xf9\x3c\x00",
    "\xfb\x3f\xf1\x99\x99\x99\x99\x99\x9a",
    "\xf9\x3e\x00",
    "\xf9\x7b\xff",
    "\xfa\x47\xc3\x50\x00",
    "\xfa\x7f\x7f\xff\xff",
    "\xfb\x7e\x37\xe4\x3c\x88\x00\x75\x9c",
    "\xf9\x00\x01",
    "\xf9\x04\x00",
    "\xf9\xc4\x00",
    "\xfb\xc0\x10\x66\x66\x66\x66\x66\x66",
    "\xf9\x7c\x00",
    "\xf9\x7e\x00",
    "\xf9\xfc\x00",
    "\xfa\x7f\x80\x00\x00",
    "\xfa\x7f\xc0\x00\x00",
    "\xfa\xff\x80\x00\x00",
    "\xfb\x7f\xf0\x00\x00\x00\x00\x00\x00",
    "\xfb\x7f\xf8\x00\x00\x00\x00\x00\x00",
    "\xfb\xff\xf0\x00\x00\x00\x00\x00\x00",
    "\xf4",
    "\xf5",
    "\xf6",
    "\xf7",
    "\xf0",
    "\xf8\x18",
    "\xf8\xff",
    "\xc0\x74\x32\x30\x31\x33\x2d\x30\x33\x2d\x32\x31\x54\x32\x30\x3a\x30\x34\x3a\x30\x30\x5a",
    "\xc1\x1a\x51\x4b\x67\xb0",
    "\xc1\xfb\x41\xd4\x52\xd9\xec\x20\x00\x00",
    "\xd7\x44\x01\x02\x03\x04",
    "\xd8\x18\x45\x64\x49\x45\x54\x46",
    "\xd8\x20\x76\x68\x74\x74\x70\x3a\x2f\x2f\x77\x77\x77\x2e\x65\x78\x61\x6d\x70\x6c\x65\x2e\x63\x6f\x6d",
    "\x40",
    "\x44\x01\x02\x03\x04",
    "\x60",
    "\x61\x61",
    "\x64\x49\x45\x54\x46",
    "\x62\x22\x5c",
    "\x62\xc3\xbc",
    "\x63\xe6\xb0\xb4",
    "\x64\xf0\x90\x85\x91",
    "\x80",
    "\x83\x01\x02\x03",
    "\x83\x01\x82\x02\x03\x82\x04\x05",
    "\x98\x19\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x18\x18\x19",
    "\xa0",
    "\xa2\x01\x02\x03\x04",
    "\xa2\x61\x61\x01\x61\x62\x82\x02\x03",
    "\x82\x61\x61\xa1\x61\x62\x61\x63",
    "\xa5\x61\x61\x61\x41\x61\x62\x61\x42\x61\x63\x61\x43\x61\x64\x61\x44\x61\x65\x61\x45",
    "\x5f\x42\x01\x02\x43\x03\x04\x05\xff",
    "\x7f\x65\x73\x74\x72\x65\x61\x64\x6d\x69\x6e\x67\xff",
    "\x9f\xff",
    "\x9f\x01\x82\x02\x03\x9f\x04\x05\xff\xff",
    "\x9f\x01\x82\x02\x03\x82\x04\x05\xff",
    "\x83\x01\x82\x02\x03\x9f\x04\x05\xff",
    "\x83\x01\x9f\x02\x03\xff\x82\x04\x05",
    "\x9f\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x18\x18\x19\xff",
    "\xbf\x61\x61\x01\x61\x62\x9f\x02\x03\xff\xff",
    "\x82\x61\x61\xbf\x61\x62\x61\x63\xff",
    "\xbf\x63\x46\x75\x6e\xf5\x63\x41\x6d\x74\x21\xff",
}

func TestCBORDecode(t *testing.T) {
	for idx, item := range content {
		val, err := CBORDecode([]byte(item))
		if err != nil {
			t.Errorf("decode fail: %d %v", idx, item)
			t.Fail()
		}
		buf := CBOREncode(val)
		if buf.Len() == 0 {
			t.Errorf("%d. not equal: %#v, %#v", idx, []byte(item), buf.Bytes())
		}
	}
}

func TestNetworkEndian(t *testing.T) {
	b := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}

	if read_network_endian(b, 0, 1) != 0x11 {
		t.Log("read network endian byte fail")
		t.Fail()
	}
	if read_network_endian(b, 0, 2) != 0x1122 {
		t.Log("read network endian word fail")
		t.Fail()
	}
	if read_network_endian(b, 0, 4) != 0x11223344 {
		t.Log("read network endian dword fail")
		t.Fail()
	}
	if read_network_endian(b, 0, 8) != 0x1122334455667788 {
		t.Log("read network endian qword fail")
		t.Fail()
	}
}
