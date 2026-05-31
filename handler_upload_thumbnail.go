package main

import (
	"fmt"
	"io"
	"net/http"

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
	data, e := io.ReadAll(file)
	if e != nil {
		respondWithError(w, http.StatusNotAcceptable, "Invalid File", nil)
	}

	defer file.Close()

	v, err := cfg.db.GetVideo(videoID)

	if v.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Invalid Owner Prohibited", nil)
		return
	}

	videoThumbnails[videoID] = thumbnail{
		data,
		contentType,
	}

	s := fmt.Sprintf("http://localhost:%d/api/thumbnails/%s", 8091, videoID)
	v.ThumbnailURL = &s
	if err := cfg.db.UpdateVideo(v); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to updated Video", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, v)
}
