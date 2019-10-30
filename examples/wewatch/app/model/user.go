package model

import (
	"time"

	"github.com/appist/appy"
)

// User contains the logic for users table.
type User struct {
	appy.AppModel
	ID        int
	Email     string
	DeletedAt time.Time `pg:",soft_delete"`
}
