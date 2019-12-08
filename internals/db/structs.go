package db

type SecurityInformation struct {
	Password string
}

type User struct {
	Id           string
	Email        string
	SecurityInfo SecurityInformation
}
