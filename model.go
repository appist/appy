package appy

import (
	"github.com/go-pg/pg/v9/orm"
)

type (
	// AfterScanHook is the hook to trigger after a model's scan.
	AfterScanHook = orm.AfterScanHook

	// AfterSelectHook is the hook to trigger after a model's SELECT query.
	AfterSelectHook = orm.AfterSelectHook

	// BeforeInsertHook is the hook to trigger before a model's INSERT query.
	BeforeInsertHook = orm.BeforeInsertHook

	// AfterInsertHook is the hook to trigger after a model's INSERT query.
	AfterInsertHook = orm.AfterInsertHook

	// BeforeUpdateHook is the hook to trigger before a model's UPDATE query.
	BeforeUpdateHook = orm.BeforeUpdateHook

	// AfterUpdateHook is the hook to trigger after a model's UPDATE query.
	AfterUpdateHook = orm.AfterUpdateHook

	// BeforeDeleteHook is the hook to trigger before a model's DELETE query.
	BeforeDeleteHook = orm.BeforeDeleteHook

	// AfterDeleteHook is the hook to trigger after a model's DELETE query.
	AfterDeleteHook = orm.AfterDeleteHook
)
