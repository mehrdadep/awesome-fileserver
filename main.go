package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

const (
	PORT          = ":9393"
	maxUploadSize = 2000000 // 2 MB
	uploadPath    = "./upload"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	// Upload a file to an absolute path (uploadPath)
	// and return the name of the file
	if r.Method == http.MethodPost {
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			http.Error(w, fmt.Sprintf("Could not parse multipart form: %v\n", err), http.StatusInternalServerError)
			return
		}
		file, fileHeader, err := r.FormFile("uploadFile")
		if err != nil {
			http.Error(w, "Invalid file", http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileSize := fileHeader.Size
		if fileSize > maxUploadSize {
			http.Error(w, "File is too large", http.StatusBadRequest)
			return
		}
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, "File is invalid", http.StatusBadRequest)
			return
		}
		fileType := http.DetectContentType(fileBytes)
		if fileType != "image/jpeg" && fileType != "image/jpg" &&
			fileType != "image/gif" && fileType != "image/png" &&
			fileType != "application/pdf" {
			http.Error(w, "File type is not supported invalid", http.StatusBadRequest)
			return
		}
		fileName, err := generateToken(12)
		if err != nil {
			http.Error(w, "Failed generating rand token", http.StatusInternalServerError)
			return
		}
		fileEndings, err := mime.ExtensionsByType(fileType)
		if err != nil {
			http.Error(w, "Can not read file type", http.StatusInternalServerError)
			return
		}
		newPath := filepath.Join(uploadPath, fileName+fileEndings[0])

		newFile, err := os.Create(newPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer newFile.Close()
		if _, err := newFile.Write(fileBytes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(fmt.Sprintf("FileType: %s, File: %s\n", fileType, newPath)))
	} else {
		http.Error(w, fmt.Sprintf("Method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
	}
}

// Returns a unique token based on the provided name
func generateToken(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func main() {
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		_ = os.Mkdir(uploadPath, os.ModePerm)
	}
	http.HandleFunc("/upload", UploadFile)
	fs := http.FileServer(http.Dir(uploadPath))
	http.Handle("/files/", http.StripPrefix("/files", fs))
	log.Print(fmt.Sprintf("Server started on localhost%s, use /upload for uploading files and /files/{fileName} for downloading files.", PORT))
	log.Fatal(http.ListenAndServe(PORT, nil))
}
