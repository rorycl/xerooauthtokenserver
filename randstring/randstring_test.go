package randstring

import (
	"testing"
)

func TestRandString(t *testing.T) {
	s := RandString(12)
	if len(s) != 12 {
		t.Errorf("Length of generated string %s is incorrect", s)
	}
}

func TestRandStringTwice(t *testing.T) {
	s := RandString(12)
	r := RandString(12)
	if s == r {
		t.Errorf("Calling RandString twice gives the same answer (%s==%s)",
			s, r)
	}
}
