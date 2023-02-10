package sqleasy

import (
	"database/sql/driver"
	"sync"
	"time"
	"unsafe"

	"modernc.org/libc"
	"modernc.org/libc/sys/types"
	sqlite "modernc.org/sqlite/lib"
)

const (
	ptrSize                 = unsafe.Sizeof(uintptr(0))
	sqliteLockedSharedcache = sqlite.SQLITE_LOCKED | (1 << 8)
)

// Conn -
type Conn struct {
	db  uintptr
	tls *libc.TLS
	mx  sync.Mutex
}

// Rows -
type Rows struct {
	allocs []uintptr
	c      *Conn
	pstmt  uintptr

	columnCount int
	doStep      bool
	empty       bool

	result []any
	err    error
}

type mutex struct {
	sync.Mutex
}

func mutexAlloc(tls *libc.TLS) uintptr {
	return libc.Xcalloc(tls, 1, types.Size_t(unsafe.Sizeof(mutex{})))
}

func mutexFree(tls *libc.TLS, m uintptr) { libc.Xfree(tls, m) }

func toNamedValues(vals ...driver.Value) (r []driver.NamedValue) {
	r = make([]driver.NamedValue, len(vals))
	for i, val := range vals {
		switch v := val.(type) {
		case time.Time:
			val = v.UnixMicro()
		}
		r[i] = driver.NamedValue{Value: val, Ordinal: i + 1}
	}
	return r
}

// Error -
type Error struct {
	msg  string
	code int
}

// Error -
func (e *Error) Error() string { return e.msg }

// Code -
func (e *Error) Code() int { return e.code }
