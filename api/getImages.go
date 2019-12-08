package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/matba/slyde-server/internals/cacher"
)

func GetImages(w http.ResponseWriter, r *http.Request) {
	// We can obtain the session token from the requests cookies, which come with every request
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			log.Printf("Unauthorized access %q", err)
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	// We then get the name of the user from our cache, where we set the session token
	response, err := cacher.GetCache().GetKeyValue(sessionToken)
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		log.Printf("Cannot find a session token %q in cache. %q", sessionToken, err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if response == "" {
		// If there is an error fetching from cache, return an internal server error status
		log.Printf("Cannot find a session token %q in cache. %q", sessionToken, err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	returnImages := images{ImageList: []string{"http://ls3.rnet.ryerson.ca/people/mahdi/images/my_picture.jpg"}}

	js, err := json.Marshal(returnImages)
	if err != nil {
		log.Printf("Cannot convert images object to json %q", err)
		WriteErrorOnResponse("Cannot cannot convert the result to JSON.", &w, http.StatusInternalServerError)
		return
	}
	w.Write(js)
}
