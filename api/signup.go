package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/matba/slyde-server/internals/db"
	"github.com/matba/slyde-server/internals/utils"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var creds credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Printf("Cannot decode the body of request %q", r.Body)
		// If the structure of the body is wrong, return an HTTP error
		WriteErrorOnResponse("Cannot decode the object sent.", &w, http.StatusBadRequest)
		return
	}

	//get the client
	client, err := db.CreateMongoClient()
	defer db.CloseClient(client)
	if err != nil {
		log.Printf("Cannot connect to db %q", err)
		WriteErrorOnResponse("Cannot connect to db.", &w, http.StatusInternalServerError)
		return
	}

	collection := (*client).Database("slyde").Collection("users")

	// TODO: add validation

	id, _ := uuid.NewUUID()

	userUuid := id.String()
	newUser := db.User{
		Id:           userUuid,
		Email:        creds.Email,
		SecurityInfo: db.SecurityInformation{Password: utils.HashAndSalt(creds.Password)},
	}

	insertResult, err := collection.InsertOne(context.TODO(), newUser)
	if err != nil {
		log.Printf("Cannot insert to db %q", err)
		WriteErrorOnResponse("Cannot insert to db.", &w, http.StatusInternalServerError)
		return
	}

	log.Println("Created a user: ", insertResult.InsertedID)

	//redact passwd
	newUser.SecurityInfo = db.SecurityInformation{}

	js, err := json.Marshal(newUser)
	if err != nil {
		log.Printf("Cannot convert user object to json %q", err)
		WriteErrorOnResponse("Cannot cannot convert the result to JSON.", &w, http.StatusInternalServerError)
		return
	}
	w.Write(js)
}
