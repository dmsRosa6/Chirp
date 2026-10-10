package grammar

import (
	"strings"
	"testing"
)

func TestValidateSubject(t *testing.T) {
	long := strings.Repeat("a", MaxSubjectLen)
	tests := []struct {
		in      string
		wantErr bool
	}{
		{"foo", false},
		{"foo.bar", false},
		{"Foo_1.bar-2.Baz", false},
		{long, false},

		{"", true},
		{".", true},
		{"a..b", true},
		{".a", true},
		{"a.", true},
		{"a.*", true}, // wildcard is for SUB only
		{"a*", true},
		{"a b", true},
		{"a/b", true},
		{"héllo", true},
		{long + "a", true},
	}
	for _, tc := range tests {
		err := ValidateSubject(tc.in)
		if (err != nil) != tc.wantErr {
			t.Errorf("ValidateSubject(%q): err=%v, wantErr=%v", tc.in, err, tc.wantErr)
		}
	}
}

func TestValidatePattern(t *testing.T) {
	tests := []struct {
		in      string
		wantErr bool
	}{
		{"foo", false},
		{"foo.bar", false},
		{"foo.b*", false},
		{"foo.*", false},
		{"foo*", false},
		{"*", false},

		{"", true},
		{"a..b", true},
		{".a", true},
		{"a.", true},
		{"a*.b", true}, // wildcard only in the last token
		{"a**", true},  // only one trailing wildcard
		{"a*b", true},  // wildcard must be the last character
		{"*.a", true},
		{"a.b c*", true},
	}
	for _, tc := range tests {
		err := ValidatePattern(tc.in)
		if (err != nil) != tc.wantErr {
			t.Errorf("ValidatePattern(%q): err=%v, wantErr=%v", tc.in, err, tc.wantErr)
		}
	}
}

func TestValidateSubID(t *testing.T) {
	if err := ValidateSubID("sub1"); err != nil {
		t.Errorf("sub1: %v", err)
	}
	if err := ValidateSubID(""); err == nil {
		t.Error("empty id should be rejected")
	}
	if err := ValidateSubID(strings.Repeat("x", MaxSubIDLen+1)); err == nil {
		t.Error("overlong id should be rejected")
	}
}
