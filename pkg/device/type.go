package device

import (
	"time"
)

type Device struct {
	ID          int       `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Name        string    `json:"name" db:"name"`
	Type        string    `json:"type" db:"type"`
	IP          string    `json:"ip" db:"ip"`
	MAC         string    `json:"mac" db:"mac"`
	Description *string   `json:"description" db:"description"`
	Employee    *string   `json:"employee" db:"employee"`
}
