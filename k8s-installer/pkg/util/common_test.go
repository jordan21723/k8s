package util

import (
	"strings"
	"testing"
)

const (
	Base64EncodedPre      = "123456"
	Base64EncodedExpected = "MTIzNDU2"
	StrSplitPre           = "foo.bar.baz"
	StrSplitSep           = "."
)

var StrArr = []string{"foo", "bar", "baz"}

func TestBase64Encode(t *testing.T) {
	encoded := Base64Encode(Base64EncodedPre)
	if encoded != Base64EncodedExpected {
		t.Fatalf("%s BASE64 encoded to %s, expected %s", Base64EncodedPre, encoded, Base64EncodedExpected)
	}
	t.Logf("%s BASE64 encoded successfully", Base64EncodedPre)
}

func TestStringSplit(t *testing.T) {
	arr, err := StringSplit(StrSplitSep, StrSplitPre)
	if err != nil {
		t.Fatalf("string split error: %v", err)
	}
	t.Logf("splited string list: %v", arr)
	if strings.Join(arr, StrSplitSep) == StrSplitPre {
		t.Logf("string %s splited successfully", StrSplitPre)
	}
}

func TestStringArrayLocate(t *testing.T) {
	for i, str := range StrArr {
		e, err := StringArrayLocate(i, StrArr)
		if err != nil {
			t.Fatalf("string array element locate error: %v", err)
		}
		if e != str {
			t.Fatalf("arr[%d] supposed to be %s, now: %s", i, str, e)
		}
	}
	t.Logf("string array %v located by index successfully", StrArr)
}
