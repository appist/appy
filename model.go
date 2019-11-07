package appy

import (
	"github.com/go-pg/pg/v9/orm"
)

type (
	AfterScanHook    = orm.AfterScanHook
	AfterSelectHook  = orm.AfterSelectHook
	BeforeInsertHook = orm.BeforeInsertHook
	AfterInsertHook  = orm.AfterInsertHook
	BeforeUpdateHook = orm.BeforeUpdateHook
	AfterUpdateHook  = orm.AfterUpdateHook
	BeforeDeleteHook = orm.BeforeDeleteHook
	AfterDeleteHook  = orm.AfterDeleteHook
)
