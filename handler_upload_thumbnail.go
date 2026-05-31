package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	maxMemory := int64(10 << 20)
	r.ParseMultipartForm(maxMemory)

	file, header, _ := r.FormFile("thumbnail")
	contentType := header.Header.Get("Content-Type")
	mimeType, _, _ := mime.ParseMediaType(contentType)
	ext := strings.Split(mimeType, "/")[1]
	allowedExt := map[string]bool{
		"png":  true,
		"jpg":  true,
		"jpeg": true,
	}
	if _, ok := allowedExt[ext]; !ok {
		respondWithError(w, http.StatusBadRequest, "Invalid Media Type", err)
		return
	}
	var bytes []byte = make([]byte, 32)
	rand.Read(bytes)
	imgName := fmt.Sprintf(
		"%s.%s",
		base64.RawURLEncoding.EncodeToString(bytes),
		ext,
	)
	imgPath := filepath.Join(cfg.assetsRoot, imgName)
	f, err := os.Create(imgPath)
	byteCount, _ := io.Copy(f, file)
	defer file.Close()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update Video", err)
		return
	}

	if byteCount == 0 {
		respondWithError(w, http.StatusLengthRequired, "Invalid File", nil)
		return
	}

	v, err := cfg.db.GetVideo(videoID)

	if v.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Invalid Owner Prohibited", nil)
		return
	}

	var dataUrl string = fmt.Sprintf("http://localhost:8091/assets/%s.%s", base64.RawURLEncoding.EncodeToString(bytes), ext)

	v.ThumbnailURL = &dataUrl
	if err := cfg.db.UpdateVideo(v); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update Video", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, v)
}
