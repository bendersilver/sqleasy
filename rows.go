package sqleasy

import (
	"encoding/json"
	"fmt"
	"io"

	sqlite "modernc.org/sqlite/lib"
)

func newRows(c *Conn, pstmt uintptr, allocs []uintptr, empty bool) (r *Rows, err error) {
	r = &Rows{c: c, pstmt: pstmt, allocs: allocs, empty: empty}

	defer func() {
		if err != nil {
			r.Close()
			r = nil
		}
	}()

	r.columnCount, err = c.columnCount(pstmt)
	if err != nil {
		return nil, err
	}
	r.result = make([]any, r.columnCount)
	return r, nil
}

// Close -
func (r *Rows) Close() (err error) {
	for _, v := range r.allocs {
		r.c.free(v)
	}
	r.allocs = nil
	return r.c.finalize(r.pstmt)
}

// ColumnName -
func (r *Rows) ColumnName() (c []string, err error) {
	c = make([]string, r.columnCount)
	for i := range c {
		if c[i], err = r.c.columnName(r.pstmt, i); err != nil {
			return nil, err
		}
	}
	return
}

// ColumnDeclType -
func (r *Rows) ColumnDeclType() (c []string) {
	c = make([]string, r.columnCount)
	for i := range c {
		c[i] = r.c.columnDeclType(r.pstmt, i)
	}
	return
}

// Values -
func (r *Rows) Values() ([]any, error) {
	return r.result, r.err
}

// Err -
func (r *Rows) Err() error {
	return r.err
}

// Next -
func (r *Rows) Next() bool {
	err := r.next()
	if err != nil {
		r.result = nil
		if err != io.EOF {
			r.err = err
		}
		return false
	}
	return true
}

func (r *Rows) next() (err error) {
	if r.empty {
		return io.EOF
	}

	rc := sqlite.SQLITE_ROW
	if r.doStep {
		if rc, err = r.c.step(r.pstmt); err != nil {
			return err
		}
	}

	r.doStep = true
	switch rc {
	case sqlite.SQLITE_ROW:
		for i := range r.result {
			ct, err := r.c.columnType(r.pstmt, i)
			if err != nil {
				return err
			}

			switch ct {
			case sqlite.SQLITE_INTEGER:
				v, err := r.c.columnInt64(r.pstmt, i)
				if err != nil {
					return err
				}
				if r.c.columnDeclType(r.pstmt, i) == "BOOLEAN" {
					r.result[i] = v > 0
				} else {
					r.result[i] = v
				}

			case sqlite.SQLITE_FLOAT:
				r.result[i], err = r.c.columnDouble(r.pstmt, i)
				if err != nil {
					return err
				}

			case sqlite.SQLITE_TEXT:
				r.result[i], err = r.c.columnText(r.pstmt, i)
				if err != nil {
					return err
				}

			case sqlite.SQLITE_BLOB:
				b, err := r.c.columnBlob(r.pstmt, i)
				if err != nil {
					return err
				}
				r.result[i] = json.RawMessage(b)

			case sqlite.SQLITE_NULL:
				r.result[i] = nil
			default:
				return fmt.Errorf("internal error: rc %d", rc)
			}
		}
		return nil
	case sqlite.SQLITE_DONE:
		return io.EOF
	default:
		return r.c.errstr(int32(rc))
	}
}
