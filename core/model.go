package core

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9/orm"
)

// AppModel associates a type struct to a database table.
type AppModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

var _ orm.AfterScanHook = (*AppModel)(nil)
var _ orm.AfterSelectHook = (*AppModel)(nil)
var _ orm.BeforeInsertHook = (*AppModel)(nil)
var _ orm.AfterInsertHook = (*AppModel)(nil)
var _ orm.BeforeUpdateHook = (*AppModel)(nil)
var _ orm.AfterUpdateHook = (*AppModel)(nil)
var _ orm.BeforeDeleteHook = (*AppModel)(nil)
var _ orm.AfterDeleteHook = (*AppModel)(nil)

// AfterScan is the hook to trigger after a model's scan.
func (*AppModel) AfterScan(ctx context.Context) error {
	return nil
}

// AfterSelect is the hook to trigger after a model's SELECT query.
func (m *AppModel) AfterSelect(ctx context.Context) error {
	return nil
}

// BeforeInsert is the hook to trigger before a model's INSERT query.
func (m *AppModel) BeforeInsert(ctx context.Context) (context.Context, error) {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}

	return ctx, nil
}

// AfterInsert is the hook to trigger after a model's INSERT query.
func (m *AppModel) AfterInsert(ctx context.Context) error {
	return nil
}

// BeforeUpdate is the hook to trigger before a model's UPDATE query.
func (m *AppModel) BeforeUpdate(ctx context.Context) (context.Context, error) {
	m.UpdatedAt = time.Now()

	return ctx, nil
}

// AfterUpdate is the hook to trigger after a model's UPDATE query.
func (m *AppModel) AfterUpdate(ctx context.Context) error {
	return nil
}

// BeforeDelete is the hook to trigger before a model's DELETE query.
func (m *AppModel) BeforeDelete(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

// AfterDelete is the hook to trigger after a model's DELETE query.
func (m *AppModel) AfterDelete(ctx context.Context) error {
	return nil
}
