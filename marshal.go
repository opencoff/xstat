// marshal.go - Xstat marshal/unmarshal implementation
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
	"encoding/binary"
	"errors"
	"syscall"

	"github.com/opencoff/xstat/internal/proto"
)

// maximum size of marshalled data; used to prevent malformed packets from
// causing buffer overflows
const _maxMarshalSize uint32 = 65536

// MarshalBinary encodes xstat into a binary buffer using nm as the file name
// The caller may wish to use a truncated or shorter name
func (x *Xstat) MarshalBinary(nm string, buf []byte) (int, error) {
	if len(nm) == 0 {
		nm = x.Name
	}

	m := &proto.Xstat{
		Name:  nm,
		Size:  x.Size,
		Uid:   x.Uid,
		Gid:   x.Gid,
		Mode:  x.Mode,
		Mtime: toProtoTime(x.Mtim),
		Atime: toProtoTime(x.Atim),
		Nlink: x.Nlink,
		Ino:   x.Ino,
		Xattr: x.Xattr,
	}

	sz := m.SizeVT()

	if len(buf) < (sz + 4) {
		return 0, ErrNoSpace
	}

	binary.BigEndian.PutUint32(buf[:4], uint32(sz))
	n, err := m.MarshalToVT(buf[4:])
	if err != nil {
		return 0, err
	}
	return n + 4, nil
}

func (x *Xstat) MarshalSize(nm string) int {
	m := &proto.Xstat{
		Name:  nm,
		Size:  x.Size,
		Uid:   x.Uid,
		Gid:   x.Gid,
		Mode:  x.Mode,
		Mtime: toProtoTime(x.Mtim),
		Atime: toProtoTime(x.Atim),
		Nlink: x.Nlink,
		Ino:   x.Ino,
		Xattr: x.Xattr,
	}

	return m.SizeVT() + 4
}

func (x *Xstat) UnmarshalBinary(buf []byte) (int, error) {
	if len(buf) < 4 {
		return 0, ErrTooSmall
	}

	sz := binary.BigEndian.Uint32(buf[:4])
	if sz > _maxMarshalSize {
		return 0, ErrTooBig
	}

	b := buf[4:]
	if uint32(len(b)) < sz {
		return 0, ErrTooSmall
	}

	var m proto.Xstat
	if err := m.UnmarshalVT(b[:sz]); err != nil {
		return 0, err
	}

	*x = Xstat{
		Stat_t: syscall.Stat_t{
			Size:  m.Size,
			Uid:   m.Uid,
			Gid:   m.Gid,
			Mode:  m.Mode,
			Mtim:  fromProtoTime(m.Mtime),
			Atim:  fromProtoTime(m.Atime),
			Nlink: m.Nlink,
			Ino:   m.Ino,
		},
		Xattr: m.Xattr,
		Name:  m.Name,
	}
	return len(b) + 4, nil
}

func toProtoTime(t syscall.Timespec) *proto.Time {
	p := &proto.Time{
		Sec:  t.Sec,
		Nsec: uint32(t.Nsec),
	}
	return p
}

func fromProtoTime(p *proto.Time) syscall.Timespec {
	return syscall.Timespec{
		Sec:  p.Sec,
		Nsec: int64(p.Nsec),
	}
}

var (
	ErrNoSpace  = errors.New("not enough space in buffer")
	ErrTooSmall = errors.New("input buf not big enough for unmarshalling")
	ErrTooBig   = errors.New("unsafe marshalled size")
)
