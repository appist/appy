package support

import (
	"gopkg.in/guregu/null.v4"
)

type (
	// NBool is a nullable bool. If the value is zero, it equals to null in JSON
	// and NULL in SQL.
	NBool = null.Bool

	// NFloat64 is a nullable float64. If the value is zero, it equals to null in
	// JSON and NULL in SQL.
	NFloat64 = null.Float

	// NInt64 is a nullable int64. If the value is zero, it equals to null in
	// JSON and NULL in SQL.
	NInt64 = null.Int

	// NString is a nullable string. If the value is zero, it equals to null in
	// JSON and NULL in SQL.
	NString = null.String

	// NTime is a nullable time. If the value is zero, it equals to null in
	// JSON and NULL in SQL.
	NTime = null.Time
)

var (
	// NewNBool creates a new NBool.
	NewNBool = null.BoolFrom

	// NewNFloat64 creates a new NFloat64.
	NewNFloat64 = null.FloatFrom

	// NewNInt64 creates a new NInt64.
	NewNInt64 = null.IntFrom

	// NewNString creates a new NString.
	NewNString = null.StringFrom

	// NewNTime creates a new NTime.
	NewNTime = null.TimeFrom
)
