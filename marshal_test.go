// marshall_test.go - test harness for marshal/unmarshal code
//
// (c) 2023- Sudhi Herle <sudhi@herle.net>
//
// Licensing Terms: GPLv2
//
// If you need a commercial license for this work, please contact
// the author.
//
// This software does not come with any express or implied
// warranty; it is provided "as is". No claim  is made to its
// suitability for any purpose.

package xstat

import (
	"fmt"
	"runtime"
	"testing"
)

func newAsserter(t *testing.T) func(cond bool, msg string, args ...interface{}) {
	return func(cond bool, msg string, args ...interface{}) {
		if cond {
			return
		}

		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "???"
			line = 0
		}

		s := fmt.Sprintf(msg, args...)
		t.Fatalf("%s: %d: Assertion failed: %s\n", file, line, s)
	}
}

// add some synthetic xattr before marshaling
func addXattr(x *Xstat) {
	x.Xattr["golang.test.xstat0"] = "foobar"
	x.Xattr["golang.test.xstat1"] = "barfoo"
}

var testPaths = [...]string{
	"/etc/profile",
	"/etc/resolv.conf",
	"/etc/hosts",
}

func TestRoundtrip(t *testing.T) {
	assert := newAsserter(t)
	p := testPaths[0]

	x, err := NewFromPath(p, true)
	assert(err == nil, "can't xstat: %s", err)

	addXattr(x)

	sz := x.MarshalSize(p)
	buf := make([]byte, sz/2)
	n, err := x.MarshalBinary(p, buf)
	assert(err != nil, "marshal accepted smol buffer?!")

	buf = make([]byte, sz)
	n, err = x.MarshalBinary(p, buf)
	assert(err == nil, "marshal: %s", err)
	assert(n == sz, "marshal: size %d != ret %d", sz, n)

	var y Xstat
	m, err := y.UnmarshalBinary(buf[:n])
	assert(err == nil, "unmarshal: %s", err)
	assert(n == m, "unmarshal: size %d != ret %d", sz, n)
}

// encode multiple xstat in a stream and decode all of them
func TestStream(t *testing.T) {
	assert := newAsserter(t)

	// suitably large space
	buf := make([]byte, 65536)

	xa := make([]*Xstat, len(testPaths))

	b := buf[:]
	tot := 0
	for i := range testPaths {
		p := testPaths[i]
		x, err := NewFromPath(p, true)
		assert(err == nil, "%d: can't xstat: %s", i, err)

		addXattr(x)
		n, err := x.MarshalBinary(p, b)
		assert(err == nil, "%d: marshall: %s", i, err)
		//t.Logf("%d: %s => %d bytes\n", i, p, n)
		b = b[n:]
		tot += n
		xa[i] = x
	}

	//t.Logf("%d entries, %d bytes total\n", len(testPaths), tot)
	b = buf[:tot]
	j := 0
	for tot > 0 {
		var z Xstat

		n, err := z.UnmarshalBinary(b)
		assert(err == nil, "unmarshal %d: tot %d, error %s", j, tot, err)
		assert(z.Equal(xa[j]), "unmarshal %d: equality fail", j)
		b = b[n:]
		tot -= n
		j++
	}

}
