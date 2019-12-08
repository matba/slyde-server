package api

import (
	"log"
	"net/http"

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
	c.active = true

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/signup", SignUp)
	router.HandleFunc("/signin", Signin)
	router.HandleFunc("/images", GetImages)
	log.Fatal(http.ListenAndServe(":8080", router))
}
