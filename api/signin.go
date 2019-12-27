package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/matba/slyde-server/internals/cacher"
	"github.com/matba/slyde-server/internals/utils"
)

const signinSerivce = "SIGN_IN"
const signInTriesCacheKey = "SIGNIN_TRIES_"
const signInSessionCacheKey = "SIGNIN_KEY_"

// Signin handles API calls for signing in
func Signin(w http.ResponseWriter, r *http.Request) {
	var request credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf(errLogTemplate, errLogCannotDecode, signinSerivce, "", err.Error())
		WriteErrorOnResponse(errCannotDecode, &w, http.StatusBadRequest)
		return
	}

	log.Printf("Incoming login request for : %q", request.Email)

	triesNo := 0
	tries, err := cacher.GetCache().GetKeyValue(signInTriesCacheKey + strings.ToLower(request.Email))
	if err == nil {
		triesNo, _ = strconv.Atoi(tries)
		if triesNo > 2 {
			log.Printf(errLogTemplate, errLogTooManyTries, signinSerivce, request.Email, tries)
			WriteErrorOnResponse(errTooMayTries, &w, http.StatusBadRequest)
			return
		}
	}
	if err != nil && err != cacher.NotFound {
		log.Printf(errLogTemplate, errLogCacheFailure, signinSerivce, request.Email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	matchUser, err := GetUserByEmail(w, request.Email, signinSerivce)
	if err != nil {
		return
	}

	log.Printf("User %q found verifying passwords", request.Email)

	ok := utils.ComparePasswords(matchUser.SecurityInfo.Password, request.Password)

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	if !ok {
		cacher.GetCache().AddKeyValue(signInTriesCacheKey+strings.ToLower(request.Email),
			strconv.Itoa(triesNo+1), twelveHours)
		log.Printf(errLogTemplate, errLogWrongCredentials, signinSerivce, request.Email, "")
		WriteErrorOnResponse(errFailedLogin, &w, http.StatusUnauthorized)
		return
	}

	// Create a new random session token
	sessionToken1, _ := uuid.NewUUID()
	sessionToken := sessionToken1.String()
	// Set the token in the cache, along with the user whom it represents
	// The token has an expiry time of 24 hours
	err = cacher.GetCache().AddKeyValue(signInSessionCacheKey+sessionToken, request.Email, oneEightyDays)
	if err != nil {
		log.Printf(errLogTemplate, errLogCacheFailure, signinSerivce, request.Email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	// Finally, we set the client cookie for "session_token" as the session token we just generated
	// we also set an expiry time the same as the cache
	http.SetCookie(w, &http.Cookie{
		Name:    sessionTokenKey,
		Value:   sessionToken,
		Expires: time.Now().Add(time.Duration(oneEightyDays) * time.Second),
	})
}
