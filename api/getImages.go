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

	returnImages := UserImages{ImageList: []UserImage{
		UserImage{"test1", "image1"},
		UserImage{"test2", "image2"}}}

	js, err := json.Marshal(returnImages)
	if err != nil {
		log.Printf("Cannot convert images object to json %q", err)
		WriteErrorOnResponse("Cannot cannot convert the result to JSON.", &w, http.StatusInternalServerError)
		return
	}
	w.Write(js)
}
