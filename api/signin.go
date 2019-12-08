package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/matba/slyde-server/internals/cacher"
	"github.com/matba/slyde-server/internals/db"
	"github.com/matba/slyde-server/internals/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func Signin(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Printf("Cannot decode the body of request %q", r.Body)
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Incoming login request for : %q", creds.Email)

	client, err := db.CreateMongoClient()
	defer db.CloseClient(client)
	if err != nil {
		log.Printf("Cannot connect to db %q", err)
		WriteErrorOnResponse("Cannot connect to db.", &w, http.StatusInternalServerError)
		return
	}

	collection := (*client).Database("slyde").Collection("users")

	filter := bson.D{{"email", creds.Email}}
	var matchUser db.User
	err = collection.FindOne(context.TODO(), filter).Decode(&matchUser)
	if err != nil {
		log.Printf("Cannot get the info for user %q with error %q", creds.Email, err)
		WriteErrorOnResponse("Cannot retrieve user info.", &w, http.StatusInternalServerError)
		return
	}

	log.Printf("User %q found verifying passwords", creds.Email)

	ok := utils.ComparePasswords(matchUser.SecurityInfo.Password, creds.Password)

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	if !ok {
		log.Printf("Unsuccessful authorization for user %q", creds.Email)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Create a new random session token
	sessionToken1, _ := uuid.NewUUID()
	sessionToken := sessionToken1.String()
	// Set the token in the cache, along with the user whom it represents
	// The token has an expiry time of 24 hours
	var timeoutSeconds int
	timeoutSeconds = 60 * 60 * 24
	err = cacher.GetCache().AddKeyValue(sessionToken, creds.Email, timeoutSeconds)
	if err != nil {
		// If there is an error in setting the cache, return an internal server error
		log.Printf("exception occured while saving the user session cookie in cache: %q", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Finally, we set the client cookie for "session_token" as the session token we just generated
	// we also set an expiry time the same as the cache
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: time.Now().Add(time.Duration(timeoutSeconds) * time.Second),
	})
}
