package api

import (
	"encoding/json"
	"log"
	"net/http"
)

const userService = "USER"

// HandleUser handles users endpoint
func HandleUser(w http.ResponseWriter, r *http.Request) {
	log.Printf("Incoming call for getting user")
	email := GetUser(w, r)
	if email == "" {
		return
	}
	log.Printf("The user was recognized as %q", email)
	user, err := GetUserByEmail(w, email, userService)
	if err != nil {
		log.Printf(errLogTemplate, errLogDb, userService, email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	js, _ := json.Marshal(userInformation{
		Email: user.Email,
		Name:  user.Name,
	})
	w.Write(js)
}
