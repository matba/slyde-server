package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const imageUploadService = "IMAGE_UPLOAD"

// HandleImage handles API calls for images
func HandleImage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleImageGet(w, r)
	case "POST":
		handleImagePost(w, r)
	case "DELETE":
		handleImageDel(w, r)
	default:
		WriteErrorOnResponse(errUnsupportedOperation, &w, http.StatusBadRequest)
		return
	}
}

func handleImageGet(w http.ResponseWriter, r *http.Request) {

}

func handleImagePost(w http.ResponseWriter, r *http.Request) {
	email := GetUser(w, r)
	if email == "" {
		return
	}
	user, err := GetUserByEmail(w, email, imageUploadService)
	if err != nil {
		log.Printf(errLogTemplate, errLogDb, imageUploadService, email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	if user.ImageQuota <= user.QuotaUsed {
		log.Printf(errLogTemplate, errLogQuotaExceeded, imageUploadService, email, strconv.Itoa(user.ImageQuota))
		WriteErrorOnResponse(errQuotaExceeded, &w, http.StatusBadRequest)
		return
	}

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `uploadedImg`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("uploadedImg")
	if err != nil {
		log.Printf(errLogTemplate, errLogImageUploadError, imageUploadService, email, err)
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("temp-images", "upload-*.png")
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)
}

func handleImageDel(w http.ResponseWriter, r *http.Request) {

}
