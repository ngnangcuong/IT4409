package models

import "time"

type User struct {
	ID          string
	Name        string
	Email       string
	Provider    string
	TimeCreated time.Time
	LastUpdated time.Time
}
