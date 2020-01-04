package api

type errorResponse struct {
	Description string `json:"description"`
}

// Create a struct that models the structure of a user, both in the request body, and in the DB
type credentials struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type signupRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

type signupResponse struct {
	AlreadyRequested bool `json:"alreadyRequested"`
}

type verifyRequest struct {
	Email            string `json:"email"`
	VerificationCode string `json:"code"`
}

type userInformation struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UserImage struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
}

type UserImages struct {
	ImageList []UserImage `json:"images"`
}

type ImageDeleteRequest struct {
	ImageIds []string `json:"images"`
}

type ImageDeleteResponse struct {
	NumberDeleted int `json:"deleted"`
}
