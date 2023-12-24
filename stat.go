// stat.go -- extended stat(2) support
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
	"os"
	"syscall"
	"time"
)

type Xstat struct {
	syscall.Stat_t

	// extended attrs
	Xattr Xattr

	Name string
}

func NewFromPath(p string, xattr bool) (*Xstat, error) {
	var x Xstat

	err := syscall.Lstat(p, &x.Stat_t)
	if err != nil {
		return nil, fmt.Errorf("%s: stat %w", p, err)
	}

	x.Name = p
	if xattr {
		xa, err := getxattr(p)
		if err != nil {
			return nil, err
		}
		x.Xattr = xa
	}
	return &x, nil
}

func New(p string, fi os.FileInfo, x Xattr) (*Xstat, error) {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("%s: can't get stat_t", p)
	}

	xt := &Xstat{
		Stat_t: *st,
		Xattr:  x,
		Name:   p,
	}
	return xt, nil
}

func (x *Xstat) String() string {
	return fmt.Sprintf("%s: size=%d, uid=%d, gid=%d, mode=%#x, mtime=%s",
		x.Name, x.Size, x.Uid, x.Gid, x.Mode, x.Mtime())
}

// helpful methods to get important attributes

func (x *Xstat) IsDir() bool {
	return (x.Mode & syscall.S_IFMT) == syscall.S_IFDIR
}

func (x *Xstat) IsRegular() bool {
	return (x.Mode & syscall.S_IFMT) == syscall.S_IFREG
}

func (x *Xstat) IsSymlink() bool {
	return (x.Mode & syscall.S_IFMT) == syscall.S_IFLNK
}

func (x *Xstat) Perm() uint {
	return uint(x.Mode & 0777)
}

func (x *Xstat) Mtime() time.Time {
	return time.Unix(x.Mtim.Sec, x.Mtim.Nsec)
}

func (x *Xstat) Atime() time.Time {
	return time.Unix(x.Atim.Sec, x.Atim.Nsec)
}
func (x *Xstat) Ctime() time.Time {
	return time.Unix(x.Ctim.Sec, x.Ctim.Nsec)
}

// return true if every important field of x matches corresponding fields of b
func (x *Xstat) Equal(b *Xstat) bool {
	if x.Name != b.Name {
		return false
	}
	if x.Size != b.Size {
		return false
	}
	if x.Uid != b.Uid {
		return false
	}
	if x.Gid != b.Gid {
		return false
	}
	if x.Mode != b.Mode {
		return false
	}
	if x.Nlink != b.Nlink {
		return false
	}
	if x.Ino != b.Ino {
		return false
	}
	if !equalTime(x.Mtim, b.Mtim) {
		return false
	}
	if !equalTime(x.Atim, b.Atim) {
		return false
	}

	if len(x.Xattr) != len(b.Xattr) {
		return false
	}
	for k, v := range x.Xattr {
		c, ok := b.Xattr[k]
		if !ok {
			return false
		}
		if c != v {
			return false
		}
	}

	return true
}

func equalTime(a, b syscall.Timespec) bool {
	if a.Sec != b.Sec {
		return false
	}
	if a.Nsec != b.Nsec {
		return false
	}
	return true
}
