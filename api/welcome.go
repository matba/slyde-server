package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/matba/slyde-server/internals/cacher"
)

func Welcome(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Finally, return the welcome message to the user
	w.Write([]byte(fmt.Sprintf("Welcome %s!", response)))
}
