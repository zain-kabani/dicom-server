package api

import (
	"httpserver/internal/api/handlers"
	"httpserver/internal/storage/db"
	"httpserver/internal/storage/file"
	"net/http"
)

type Server struct {
	imageHandler *handlers.ImageHandler
}

func NewServer(dbStore *db.Store, fileStore *file.Store) *Server {
	return &Server{
		imageHandler: handlers.NewImageHandler(fileStore, dbStore),
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/images", s.imageHandler.Upload)
	mux.HandleFunc("/api/v1/dicom/", s.imageHandler.HandleDicom)
	return mux
}
