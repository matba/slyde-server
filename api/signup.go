package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/matba/slyde-server/internals/cacher"
	"github.com/matba/slyde-server/internals/db"
	"github.com/matba/slyde-server/internals/email"
	"github.com/matba/slyde-server/internals/utils"
	"go.mongodb.org/mongo-driver/bson"
)

const signupSerivce = "SIGN_UP"
const verifyService = "VERIFY_EMAIL"
const registrationReqCacheKey = "REGISTRATION_REQUEST_"
const registrationCodeCacheKey = "REGISTRATION_CODE_"
const registrationTriesCacheKey = "REGISTRATION_TRIES_"

// SignUp Rest API handler for sign up
func SignUp(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)

	// the object for keeping the request in serialized form
	var serializedRequest string
	// read the info from request
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	serializedRequest = buf.String()

	// the object for keeping the deserialized request
	var request signupRequest
	// deserialze the request into object
	err := json.Unmarshal([]byte(serializedRequest), &request)
	if err != nil {
		log.Printf(errLogTemplate, errLogCannotDecode, signupSerivce, "", err.Error())
		WriteErrorOnResponse(errCannotDecode, &w, http.StatusBadRequest)
		return
	}

	// Validate the request
	err = validateSignupRequest(request)
	if err != nil {
		log.Printf(errLogTemplate, errLogValidation, signupSerivce, request.Email, err.Error())
		WriteErrorOnResponse(err.Error(), &w, http.StatusBadRequest)
		return
	}

	// Make the sure the request is not already pending verification
	_, err = cacher.GetCache().GetKeyValue(registrationReqCacheKey + strings.ToLower(request.Email))
	if err != nil && err != cacher.NotFound {
		// An error has occured accessing the cache
		log.Printf(errLogTemplate, errLogCacheFailure, signupSerivce, request.Email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
	}

	// If the error is null it means the user is already in pending email state
	if err == nil {
		log.Printf(errLogTemplate, errLogAlreadyExists, signupSerivce, request.Email, err.Error())
		js, _ := json.Marshal(signupResponse{
			AlreadyRequested: true,
		})
		w.Write(js)
		return
	}

	// We put the request in the cache and look it up after email verification finished
	err = cacher.GetCache().AddKeyValue(registrationReqCacheKey+strings.ToLower(request.Email),
		serializedRequest, twentyfourHours)

	if err != nil {
		log.Printf(errLogTemplate, errLogCacheFailure, signupSerivce, request.Email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	// The confirmation code that will be sent to email
	confirmationCode := GenerateVerificationKey(6)

	err = cacher.GetCache().AddKeyValue(registrationCodeCacheKey+strings.ToLower(request.Email),
		confirmationCode, twentyfourHours)

	if err != nil {
		log.Printf(errLogTemplate, errLogCacheFailure, signupSerivce, request.Email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		// roll back cache inserts
		cacher.GetCache().DeleteKey(registrationReqCacheKey + strings.ToLower(request.Email))
		cacher.GetCache().DeleteKey(registrationCodeCacheKey + strings.ToLower(request.Email))
		return
	}

	log.Printf("The verification code for user %q is %q", request.Email, confirmationCode)

	// Sent the verification code
	err = email.GetEmailSender().
		SendEmail(request.Email, emailSubject, emailBody+confirmationCode)
	if err != nil {
		log.Printf(errLogTemplate, errLogEmailFailure, signupSerivce, request.Email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		cacher.GetCache().DeleteKey(registrationReqCacheKey + strings.ToLower(request.Email))
		cacher.GetCache().DeleteKey(registrationCodeCacheKey + strings.ToLower(request.Email))
		return
	}

	log.Printf("Verification email sent to %q", request.Email)

	response := signupResponse{
		AlreadyRequested: false,
	}
	js, _ := json.Marshal(response)
	w.Write(js)
	return
}

// VerifyEmail Rest API handler for verifying the email
func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)
	var request verifyRequest
	// Get the JSON body and decode into verify request
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf(errLogTemplate, errLogCannotDecode, verifyService, "", err.Error())
		WriteErrorOnResponse(errCannotDecode, &w, http.StatusBadRequest)
		return
	}

	// Verify the email is a valid email
	emailMatched, _ := regexp.MatchString(emailRegex, request.Email)
	if !emailMatched {
		log.Printf(errLogTemplate, errLogValidation, verifyService, request.Email, errInvalidEmailFormat)
		WriteErrorOnResponse(errInvalidEmailFormat, &w, http.StatusBadRequest)
		return
	}

	// Make sure the user have not exceed valid number of tries
	triesNo := 0
	tries, err := cacher.GetCache().GetKeyValue(registrationTriesCacheKey + strings.ToLower(request.Email))
	if err == nil {
		triesNo, _ = strconv.Atoi(tries)
		if triesNo > 2 {
			log.Printf(errLogTemplate, errLogTooManyTries, verifyService, request.Email, tries)
			WriteErrorOnResponse(errTooMayTries, &w, http.StatusBadRequest)
			return
		}
	}
	if err != nil && err != cacher.NotFound {
		log.Printf(errLogTemplate, errLogCacheFailure, verifyService, request.Email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	code, err := cacher.GetCache().GetKeyValue(registrationCodeCacheKey + strings.ToLower(request.Email))
	if err != nil && err != cacher.NotFound {
		log.Printf(errLogTemplate, errLogCacheFailure, verifyService, request.Email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	// If the value does not exist there is no such registration waiting for email confirmation
	if err == cacher.NotFound {
		log.Printf(errLogTemplate, errLogCacheFailure, verifyService, request.Email, err.Error())
		WriteErrorOnResponse(errRegistrationNotFound, &w, http.StatusBadRequest)
		return
	}

	// If the the verification code is wrong increment the number of tries
	if code != strings.ToUpper(request.VerificationCode) {
		cacher.GetCache().AddKeyValue(registrationTriesCacheKey+strings.ToLower(request.Email),
			strconv.Itoa(triesNo+1), twelveHours)
		log.Printf(errLogTemplate, errLogWrongVerificationCode, verifyService, request.Email, "")
		WriteErrorOnResponse(errWrongVerificationCode, &w, http.StatusInternalServerError)
		return
	}

	// The verification has succeeded at this point
	// Retrieve the initial request from cache as save it to db
	initialRequest, err := cacher.GetCache().GetKeyValue(registrationReqCacheKey + strings.ToLower(request.Email))
	if err != nil && err != cacher.NotFound {
		log.Printf(errLogTemplate, errLogCacheFailure, verifyService, request.Email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}
	if err == cacher.NotFound {
		log.Printf(errLogTemplate, errLogInitialRequestLost, verifyService, request.Email, "")
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	// Decode intial request into an object
	var signupReq signupRequest
	err = json.Unmarshal([]byte(initialRequest), &signupReq)
	if err != nil {
		log.Printf(errLogTemplate, errLogCannotDecode, verifyService, request.Email, err.Error())
		WriteErrorOnResponse(errCannotDecode, &w, http.StatusBadRequest)
		return
	}

	// Generate a UUID for the user
	id, _ := uuid.NewUUID()
	userUUID := id.String()

	// Create a db insert object
	newUser := db.User{
		ID:           userUUID,
		Email:        strings.ToLower(signupReq.Email),
		SecurityInfo: db.SecurityInformation{Password: utils.HashAndSalt(signupReq.Password)},
		Name:         signupReq.Name,
		CreationDate: time.Now(),
		ImageQuota:   10,
		Images:       []db.ImageInfo{},
	}

	//get the client
	client, err := db.CreateMongoClient()
	defer db.CloseClient(client)
	if err != nil {
		log.Printf(errLogTemplate, errLogCannotConnectToDb, verifyService, request.Email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	// insert to DB
	collection := (*client).Database(db.MainDbName).Collection(db.UsersCollection)
	insertResult, err := collection.InsertOne(context.TODO(), newUser)
	if err != nil {
		log.Printf(errLogTemplate, errLogCannotInsertToDb, verifyService, request.Email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	log.Println("Created a user: ", insertResult.InsertedID, request.Email)

	//redact passwd
	newUser.SecurityInfo = db.SecurityInformation{}

	js, _ := json.Marshal(newUser)
	w.Write(js)
	// delete the stuff from the cache
	cacher.GetCache().DeleteKey(registrationReqCacheKey + strings.ToLower(request.Email))
	cacher.GetCache().DeleteKey(registrationCodeCacheKey + strings.ToLower(request.Email))
	cacher.GetCache().DeleteKey(registrationTriesCacheKey + strings.ToLower(request.Email))
}

func validateSignupRequest(info signupRequest) error {
	nameMatched, _ := regexp.MatchString(nameRegex, strings.ToLower(info.Name))
	if !nameMatched {
		return errors.New("name is not valid")
	}

	emailMatched, _ := regexp.MatchString(emailRegex, info.Email)
	if !emailMatched {
		return errors.New(errInvalidEmailFormat)
	}

	if len(info.Password) < 11 {
		return errors.New("password is not valid")
	}

	//Make sure the user does not already exists
	client, err := db.CreateMongoClient()
	defer db.CloseClient(client)
	if err != nil {
		return errors.New("user creation failed because of internal error")
	}

	collection := (*client).Database("slyde").Collection("users")

	filter := bson.D{{"email", info.Email}}
	var matchUser db.User
	err = collection.FindOne(context.TODO(), filter).Decode(&matchUser)
	if err == nil {
		return errors.New("the email already been used")
	}

	return nil
}
