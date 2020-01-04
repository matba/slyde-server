package api

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// Controller represent the controller which handles http connections
type Controller interface {
	IsActive() bool
	Start()
}

type controllerInstance struct {
	active bool
}

func newControllerInstance() *controllerInstance {
	return &controllerInstance{active: false}
}

func (c *controllerInstance) IsActive() bool {
	return c.active
}

func (c *controllerInstance) Start() {
	// check if upload image directory exists
	if _, err := os.Stat(filesDirectory); os.IsNotExist(err) {
		log.Fatal("The files directory does not exist: " + filesDirectory)
	}

	if _, err := os.Stat(filesDirectory + userDirectory); os.IsNotExist(err) {
		err := os.Mkdir(filesDirectory+userDirectory, 0777)
		if err != nil {
			log.Fatal("The files directory does not have proper permissions: " + filesDirectory)
		}
	}

	c.active = true

	router := mux.NewRouter().StrictSlash(true)
	// the endpoint for registering new users
	router.HandleFunc("/signup", SignUp)
	// the endpoint for verifying email for new users
	router.HandleFunc("/verify", VerifyEmail)
	// the endpoint for sigining in
	router.HandleFunc("/signin", Signin)
	// the endpoint for signing out
	router.HandleFunc("/signout", Signout)
	// the end point for for getting uploaded images
	router.HandleFunc("/images", HandleImage)
	// the end point for for getting user information
	router.HandleFunc("/user", HandleUser)
	log.Fatal(http.ListenAndServe(":8080", router))
}
