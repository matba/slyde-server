package db

import "time"

type SecurityInformation struct {
	Password string
}

// User keeps the information about user
type User struct {
	ID           string
	Email        string
	SecurityInfo SecurityInformation
	Name         string
	CreationDate time.Time
	ImageQuota   int
	QuotaUsed    int
}
