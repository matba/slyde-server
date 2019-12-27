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

type image struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

type images struct {
	ImageList []image `json:"images"`
}
