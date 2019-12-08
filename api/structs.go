package api

type errorResponse struct {
	Description string `json:"description"`
}

// Create a struct that models the structure of a user, both in the request body, and in the DB
type credentials struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type images struct {
	ImageList []string `json:"imageUrls"`
}
