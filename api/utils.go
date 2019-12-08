package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteErrorOnResponse(error string, w *http.ResponseWriter, status int) {
	(*w).WriteHeader(http.StatusInternalServerError)
	errorResp := errorResponse{Description: error}
	js, err := json.Marshal(errorResp)
	if err != nil {
		log.Fatal(err)
		return
	}
	(*w).Write(js)
}
