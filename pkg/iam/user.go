package iam

import (
	"time"
)

// represent IAM user
type User struct {
	Account      string `xml:"-" json:"account"`
	Path         string `xml:"Path" json:"path"`
	CreatedAt    time.Time `xml:"CreatedAt" json:"createdAt"`
	LastActivity time.Time `xml:"LastActivity" json:"lastActivity"`
}
