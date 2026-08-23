package option

import (
	"database/sql"
	"database/sql/driver"
)

var _ driver.Valuer = Option[struct{}]{}

// Value encodes the value for a database driver, writing a None as SQL NULL.
// Conversion is the one database/sql already applies, so an int arrives as an
// int64, and a T that database/sql cannot store reports an error here.
func (o Option[T]) Value() (driver.Value, error) {
	return sql.Null[T]{V: o.value, Valid: o.present}.Value()
}

var _ sql.Scanner = (*Option[struct{}])(nil)

// Scan reads a value from a database driver, reading SQL NULL as None.
// It accepts anything database/sql can convert into T, including a T whose
// pointer implements sql.Scanner.
func (o *Option[T]) Scan(src any) error {
	var n sql.Null[T]
	if err := n.Scan(src); err != nil {
		return err
	}

	*o = FromTuple(n.V, n.Valid)
	return nil
}
