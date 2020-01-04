package api

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/edwvee/exiffix"
	"github.com/google/uuid"
	"github.com/matba/slyde-server/internals/db"
	"github.com/nfnt/resize"
	"go.mongodb.org/mongo-driver/bson"
)

const imageUploadService = "IMAGE_UPLOAD"
const imageDeleteService = "IMAGE_DELETE"
const imageGetService = "IMAGE_GET"

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
	log.Printf("Incoming call for getting images")
	email := GetUser(w, r)
	if email == "" {
		return
	}
	user, err := GetUserByEmail(w, email, imageGetService)
	if err != nil {
		log.Printf(errLogTemplate, errLogDb, imageGetService, email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	if len(r.FormValue("id")) == 0 {
		// this is a request for getting all user images
		returnUserImages(user, &w, r)
	} else {
		isThumbnail := len(r.FormValue("thumbnail")) > 0
		id := r.FormValue("id")

		// go through the user images to see if such image exist
		for _, img := range user.Images {
			if img.ID == id {
				fp := getUserImagePath(user.ID, id, isThumbnail)

				imgLength := img.Width
				if img.Width < img.Height {
					imgLength = img.Height
				}
				lengthStr := r.FormValue("width")

				if len(lengthStr) > 0 {
					log.Printf("Width of %q is provided.", lengthStr)
				}

				length, err := strconv.Atoi(lengthStr)

				if err == nil && !isThumbnail {
					lengthu := uint(length)
					log.Printf("Checking if resize is required for image %q and requested width %q.", img.Name, lengthStr)
					if float32(lengthu) < float32(imgLength)*0.9 {
						log.Printf("The image %q needs to resized to match requested width %q.", img.Name, lengthStr)
						resizeRatio := float32(lengthu) * 10 / float32(imgLength)

						resizeRatioTenth := uint(math.Round(float64(resizeRatio)))

						if resizeRatioTenth < 1 {
							resizeRatioTenth = 1
						}

						fpr := getUserResizedImagePath(user.ID, id, resizeRatioTenth)

						if _, err := os.Stat(fpr); os.IsNotExist(err) {
							newWidth := uint(float32(img.Width) * (float32(resizeRatioTenth) / 10.0))
							newHeight := uint(float32(img.Height) * (float32(resizeRatioTenth) / 10.0))

							log.Printf("Resizing image %q for serving", img.Name)

							imgfile, _ := os.Open(fp)
							defer imgfile.Close()

							imgObj, _, err := image.Decode(imgfile)
							if err != nil {
								log.Printf(errLogTemplate, errLogImageValidationError, imageGetService, email, err.Error())
								WriteErrorOnResponse(errUnsupportedImage, &w, http.StatusInternalServerError)
								return
							}
							resizedImage := resize.Resize(newWidth, newHeight, imgObj, resize.Lanczos3)

							file, err := os.Create(fpr)
							if err != nil {
								log.Printf(errLogTemplate, errLogIoError, imageGetService, email, err.Error())
								WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
								return
							}
							err = jpeg.Encode(file, resizedImage, &jpeg.Options{Quality: 75})
							if err != nil {
								log.Printf(errLogTemplate, errLogIoError, imageGetService, email, err.Error())
								WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
								return
							}
						}
						http.ServeFile(w, r, fpr)

					} else {
						http.ServeFile(w, r, fp)
					}
				} else {
					http.ServeFile(w, r, fp)
				}
				return
			}
		}

		log.Printf(errLogTemplate, errNotFound, imageUploadService, email, "Image not found")
		WriteErrorOnResponse(errNotFound, &w, http.StatusNotFound)
	}
}

func returnUserImages(user *db.User, w *http.ResponseWriter, r *http.Request) {
	imList := []UserImage{}
	for _, img := range user.Images {
		imList = append(imList, UserImage{
			ID:     img.ID,
			Name:   img.Name,
			Width:  img.Width,
			Height: img.Height,
		})
	}

	returnImages := UserImages{ImageList: imList}

	js, _ := json.Marshal(returnImages)
	(*w).Write(js)
}

func handleImagePost(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)
	log.Printf("Incoming call for uploading images")
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

	fileName := r.FormValue("name")
	if len(fileName) == 0 {
		log.Printf(errLogTemplate, errLogMissingField, imageUploadService, email, "File name not provided.")
		WriteErrorOnResponse(errBadRequest, &w, http.StatusBadRequest)
		return
	}

	if user.ImageQuota <= len(user.Images) {
		log.Printf(errLogTemplate, errLogNotFound, imageUploadService, email, "")
		WriteErrorOnResponse(errBadRequest, &w, http.StatusBadRequest)
		return
	}

	err = createUserDirectories(user.ID)

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `uploadedImg`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("image")
	if err != nil {
		log.Printf(errLogTemplate, errLogImageUploadError, imageUploadService, email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	var uploadedImage image.Image
	uploadedImage, _, err = exiffix.Decode(file)
	if err != nil {
		log.Printf(errLogTemplate, errLogImageValidationError, imageUploadService, email, err.Error())
		WriteErrorOnResponse(errUnsupportedImage, &w, http.StatusInternalServerError)
		return
	}
	width := uploadedImage.Bounds().Dx()
	height := uploadedImage.Bounds().Dy()

	if width > maxImageDimension || height > maxImageDimension {
		log.Printf(errLogTemplate, errLogImageValidationError, imageUploadService, email, "Image too big")
		WriteErrorOnResponse(errImageTooBig, &w, http.StatusInternalServerError)
		return
	}
	if width < minImageDimension || height < minImageDimension {
		log.Printf(errLogTemplate, errLogImageValidationError, imageUploadService, email, "Image too small")
		WriteErrorOnResponse(errImageTooSmall, &w, http.StatusInternalServerError)
		return
	}

	// Generate a UUID for the user
	id, _ := uuid.NewUUID()
	imageUUID := id.String()

	// save the image
	imgW, imgH, err := saveImage(uploadedImage, user.ID, imageUUID, false)
	if err != nil {
		log.Printf(errLogTemplate, errLogImageSavingError, imageUploadService, email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	// save the thumbnail
	_, _, err = saveImage(uploadedImage, user.ID, imageUUID, true)
	if err != nil {
		log.Printf(errLogTemplate, errLogImageSavingError, imageUploadService, email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	// user the user in DB
	client, err := db.CreateMongoClient()
	defer db.CloseClient(client)
	if err != nil {
		log.Printf(errLogTemplate, errLogCannotConnectToDb, imageUploadService, email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	collection := (*client).Database(db.MainDbName).Collection(db.UsersCollection)
	filter := bson.D{{"id", user.ID}}
	update := bson.D{
		{"$push", bson.D{
			{"images", db.ImageInfo{
				ID:         imageUUID,
				Width:      imgW,
				Height:     imgH,
				UploadDate: time.Now(),
				Name:       fileName,
			}},
		}},
	}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Printf(errLogTemplate, errLogCannotUpdateTheDb, imageUploadService, email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	newImageJSON := UserImage{
		ID:     imageUUID,
		Name:   fileName,
		Width:  imgW,
		Height: imgH,
	}

	js, _ := json.Marshal(newImageJSON)
	w.Write(js)
}

func handleImageDel(w http.ResponseWriter, r *http.Request) {
	SetJsonContentType(w)
	log.Printf("Incoming call for deleting images")
	email := GetUser(w, r)
	if email == "" {
		return
	}
	user, err := GetUserByEmail(w, email, imageUploadService)
	if err != nil {
		log.Printf(errLogTemplate, errLogDb, imageDeleteService, email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	var request ImageDeleteRequest
	// Get the JSON body and decode into credentials
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf(errLogTemplate, errLogCannotDecode, imageDeleteService, "", err.Error())
		WriteErrorOnResponse(errCannotDecode, &w, http.StatusBadRequest)
		return
	}

	deleteRequestSet := make(map[string]bool)
	for _, imgID := range request.ImageIds {
		deleteRequestSet[imgID] = true
	}

	deletedImages := []db.ImageInfo{}
	// go through the user images to see if such image exist
	deleteCntr := 0
	for _, img := range user.Images {
		if deleteRequestSet[img.ID] {
			// delete image from db
			deletedImages = append(deletedImages, img)
			deleteCntr++
		}
	}

	client, err := db.CreateMongoClient()
	defer db.CloseClient(client)
	if err != nil {
		log.Printf(errLogTemplate, errLogCannotConnectToDb, imageUploadService, email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}
	collection := (*client).Database(db.MainDbName).Collection(db.UsersCollection)
	filter := bson.D{{"id", user.ID}}
	update := bson.D{
		{"$pull", bson.D{
			{"images", bson.D{
				{
					"$in", deletedImages,
				}},
			}},
		}}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Printf(errLogTemplate, errLogCannotUpdateTheDb, imageUploadService, email, err.Error())
		WriteErrorOnResponse(errInternalError, &w, http.StatusInternalServerError)
		return
	}

	for _, img := range deletedImages {
		fpt := getUserImagePath(user.ID, img.ID, true)
		os.Remove(fpt)
		fp := getUserImagePath(user.ID, img.ID, false)
		os.Remove(fp)
	}

	js, _ := json.Marshal(ImageDeleteResponse{
		NumberDeleted: deleteCntr,
	})
	w.Write(js)
}

func saveImage(image image.Image, userID string, imageUUID string, isThumbnail bool) (uint, uint, error) {
	width := uint(image.Bounds().Dx())
	height := uint(image.Bounds().Dy())

	tWidth, tHeigth := getImageResizeSize(width, height, isThumbnail)

	resizedImage := image
	if tWidth != width || tHeigth != height {
		resizedImage = resize.Resize(tWidth, tHeigth, image, resize.Lanczos3)
	}

	file, err := os.Create(getUserImagePath(userID, imageUUID, isThumbnail))
	if err != nil {
		return 0, 0, err
	}
	err = jpeg.Encode(file, resizedImage, nil)
	if err != nil {
		return 0, 0, err
	}
	return tWidth, tHeigth, nil
}

func getUserImagePath(userID string, imageUUID string, isThumbnail bool) string {
	directory := imagesDiretory
	if isThumbnail {
		directory = thumbnailsDirectory
	}
	return path.Join(getCurUserDirectory(userID)+directory, imageUUID+jpegExtension)
}

func getUserResizedImagePath(userID string, imageUUID string, resizeTenth uint) string {
	return path.Join(getCurUserDirectory(userID)+imagesDiretory, imageUUID+"-"+strconv.Itoa(int(resizeTenth))+jpegExtension)
}

func getImageResizeSize(width uint, height uint, isThumbnail bool) (uint, uint) {
	isPortrait := (height > width)
	length := width
	if isPortrait {
		length = height
	}
	maximumAllowedLength := resizeSize
	if isThumbnail {
		maximumAllowedLength = thumbnailsSize
	}

	if !isThumbnail && length < resizeSize {
		return width, height
	}

	resizeRatio := float32(maximumAllowedLength) / float32(length)

	return uint(float32(width) * resizeRatio), uint(float32(height) * resizeRatio)
}

func getCurUserDirectory(userID string) string {
	return filesDirectory + userDirectory + "/" + userID
}

func createUserDirectories(userID string) error {
	curUserDirectory := getCurUserDirectory(userID)
	if _, err := os.Stat(curUserDirectory); os.IsNotExist(err) {
		err := os.Mkdir(curUserDirectory, 0777)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(curUserDirectory + thumbnailsDirectory); os.IsNotExist(err) {
		err := os.Mkdir(curUserDirectory+thumbnailsDirectory, 0777)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(curUserDirectory + imagesDiretory); os.IsNotExist(err) {
		err := os.Mkdir(curUserDirectory+imagesDiretory, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}
