package domain

import "time"

type Token struct {
	Username  string
	Value     string
	CreatedAt time.Time
	Valid     bool
}
