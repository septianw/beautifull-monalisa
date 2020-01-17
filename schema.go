package main

import (
	"time"
)

// Database table schema.
type User struct {
	MobileNumber string    `db:"mobile_number"`
	Email        string    `db:"email"`
	Firstname    string    `db:"firstname"`
	Lastname     string    `db:"lastname"`
	DateOfBirth  time.Time `db:"date_of_birth"`
	Gender       string    `db:"gender"`
}
