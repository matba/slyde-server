package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func GetImages(w http.ResponseWriter, r *http.Request) {
	response := GetUser(w, r)

	if response == "" {
		return
	}

	returnImages := images{ImageList: []image{
		image{"test1", "http://ls3.rnet.ryerson.ca/people/mahdi/images/my_picture.jpg"},
		image{"test2", "http://ls3.rnet.ryerson.ca/wp-content/uploads/2013/01/mtacceexterior1.jpg"}}}

	js, err := json.Marshal(returnImages)
	if err != nil {
		log.Printf("Cannot convert images object to json %q", err)
		WriteErrorOnResponse("Cannot cannot convert the result to JSON.", &w, http.StatusInternalServerError)
		return
	}
	w.Write(js)
}
