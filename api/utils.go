package api

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/matba/slyde-server/internals/cacher"
	"github.com/matba/slyde-server/internals/db"
	"go.mongodb.org/mongo-driver/bson"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func SetJsonContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func WriteErrorOnResponse(error string, w *http.ResponseWriter, status int) {
	(*w).WriteHeader(http.StatusInternalServerError)
	errorResp := errorResponse{Description: error}
	js, _ := json.Marshal(errorResp)
	(*w).Write(js)
}

// GenerateVerificationKey creates a random string of capital letters with specific size.
func GenerateVerificationKey(length int) string {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// GetUser if the session token in request is valid it will get the user email
// otherwise it will writes appropriate stuff in response and return empty string
func GetUser(w http.ResponseWriter, r *http.Request) string {
	// We can obtain the session token from the requests cookies, which come with every request
	c, err := r.Cookie(sessionTokenKey)
	if err != nil {
		log.Printf(errLogTemplate, "Cookie issues", "AUTH", "", err.Error())
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			WriteErrorOnResponse(errUnAuthorized, &w, http.StatusUnauthorized)
			return ""
		}
		// For any other type of error, return a bad request status
		WriteErrorOnResponse(errBadRequest, &w, http.StatusBadRequest)
		return ""
	}
	sessionToken := c.Value

	// We then get the name of the user from our cache, where we set the session token
	response, err := cacher.GetCache().GetKeyValue(signInSessionCacheKey + sessionToken)
	if err != nil && err != cacher.NotFound {
		log.Printf(errLogTemplate, errLogCacheFailure, "AUTH", sessionToken, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return ""
	}

	if err == cacher.NotFound || response == "" {
		log.Printf(errLogTemplate, errLogIvalidSessionToken, "AUTH", sessionToken, err.Error())
		WriteErrorOnResponse(errUnAuthorized, &w, http.StatusUnauthorized)
		return ""
	}

	return response
}

// GetUserByEmail Gets the user information for user if there is an error appropriate response is return to output
func GetUserByEmail(w http.ResponseWriter, email string, serviceName string) (*db.User, error) {
	client, err := db.CreateMongoClient()
	defer db.CloseClient(client)
	if err != nil {
		log.Printf(errLogTemplate, errLogCannotConnectToDb, serviceName, email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return nil, err
	}

	collection := (*client).Database(db.MainDbName).Collection(db.UsersCollection)

	filter := bson.D{{"email", email}}
	var matchUser db.User
	err = collection.FindOne(context.TODO(), filter).Decode(&matchUser)
	if err != nil {
		log.Printf(errLogTemplate, errLogCannotRetrieveFromDb, serviceName, email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return nil, err
	}

	return &matchUser, nil
}
