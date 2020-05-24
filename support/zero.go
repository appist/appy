package support

import "gopkg.in/guregu/null.v4/zero"

type (
	// ZBool is a nullable bool. If the value is zero, it equals to false in JSON
	// and NULL in SQL.
	ZBool = zero.Bool

	// ZFloat64 is a nullable float64. If the value is zero, it equals to 0 in JSON
	// and NULL in SQL.
	ZFloat64 = zero.Float

	// ZInt64 is a nullable int64. If the value is zero, it equals to 0 in JSON
	// and NULL in SQL.
	ZInt64 = zero.Int

	// ZString is a nullable string. If the value is zero, it equals to "" in
	// JSON and NULL in SQL.
	ZString = zero.String

	// ZTime is a nullable time. If the value is zero, it equals to
	// "0001-01-01T00:00:00Z" in JSON and NULL in SQL.
	ZTime = zero.Time
)

var (
	// NewZBool creates a new ZBool.
	NewZBool = zero.BoolFrom

	// NewZFloat64 creates a new ZFloat64.
	NewZFloat64 = zero.FloatFrom

	// NewZInt64 creates a new ZInt64.
	NewZInt64 = zero.IntFrom

	// NewZString creates a new ZString.
	NewZString = zero.StringFrom

	// NewZTime creates a new ZTime.
	NewZTime = zero.TimeFrom
)
