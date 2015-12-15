package proxy

import "testing"

func TestTrim(t *testing.T) {
	val := " 26721						Emden "
	exp := "26721 Emden"
	res, err := trim(val)
	if nil != err {
		t.Fail()
	}
	if res != exp {
		t.Errorf("Result was '%s' but expected was '%s'", res, exp)
	}
}