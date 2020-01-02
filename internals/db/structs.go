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
	Images       []ImageInfo
}

// ImageInfo keeps information about an uploaded image
type ImageInfo struct {
	ID         string
	Width      int
	Height     int
	Size       int
	UploadDate time.Time
	Name       string
}
