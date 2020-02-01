package appy

// Model manages the data, logic and rules.
type Model struct {
	ID        int `db:"id"`
	CreatedAt int `db:"created_at"`
	UpdatedAt int `db:"updated_at"`
}
