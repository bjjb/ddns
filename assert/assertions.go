package assert

import (
	"regexp"
	"testing"
)

// Equal gets a function which errors if the string versions of all a aren't
// equal.
func Equal(t *testing.T) (f func(a ...interface{})) {
	f = func(a ...interface{}) {
		t.Helper()
		if len(a) < 2 {
			return
		}
		if l, r := repr(a[0]), repr(a[1]); l != r {
			t.Errorf("%q != %q", l, r)
		}
		f(a[1:]...)
	}
	return
}

// Match errors if each a doesn't match /r/.
func Match(t *testing.T) (f func(r string, a ...interface{})) {
	f = func(r string, a ...interface{}) {
		t.Helper()
		if len(a) < 1 {
			return
		}
		s := repr(a[0])
		if !regexp.MustCompile(r).MatchString(s) {
			t.Errorf("%q !~ /%s/", s, r)
		}
		if len(a) > 1 {
			f(r, a[1:]...)
		}
	}
	return
}
