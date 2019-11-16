package orm

import "github.com/go-pg/pg/v9/orm"

type (
	// AfterScanHook is the hook to trigger after a model's scan.
	AfterScanHook = orm.AfterScanHook

	// AfterSelectHook is the hook to trigger after a model's select.
	AfterSelectHook = orm.AfterSelectHook

	// BeforeInsertHook is the hook to trigger before a model's insert.
	BeforeInsertHook = orm.BeforeInsertHook

	// AfterInsertHook is the hook to trigger after a model's insert.
	AfterInsertHook = orm.AfterInsertHook

	// BeforeUpdateHook is the hook to trigger before a model's update.
	BeforeUpdateHook = orm.BeforeUpdateHook

	// AfterUpdateHook is the hook to trigger after a model's update.
	AfterUpdateHook = orm.AfterUpdateHook

	// BeforeDeleteHook is the hook to trigger before a model's delete.
	BeforeDeleteHook = orm.BeforeDeleteHook

	// AfterDeleteHook is the hook to trigger after a model's delete.
	AfterDeleteHook = orm.AfterDeleteHook
)
