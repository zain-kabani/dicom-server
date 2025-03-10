package handlers

import (
	"context"
	"fmt"
	"httpserver/internal/api/responses"
	"httpserver/internal/models"
	"httpserver/internal/storage/db"
	"httpserver/internal/storage/file"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"httpserver/internal/dcmutil"
	"path/filepath"
)

type ImageHandler struct {
	fileStore *file.Store
	dbStore   *db.Store
}

func NewImageHandler(fileStore *file.Store, dbStore *db.Store) *ImageHandler {
	return &ImageHandler{
		fileStore: fileStore,
		dbStore:   dbStore,
	}
}

func (h *ImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responses.WriteJSON(w, http.StatusMethodNotAllowed, responses.ErrorResponse{
			Error: "method not allowed",
		})
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		responses.WriteJSON(w, http.StatusBadRequest, responses.ErrorResponse{
			Error: "request too large",
		})
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		responses.WriteJSON(w, http.StatusBadRequest, responses.ErrorResponse{
			Error: "invalid file upload",
		})
		return
	}
	defer file.Close()

	// Save to staging
	stagingPath, err := h.fileStore.SaveToStaging(file)
	if err != nil {
		responses.WriteJSON(w, http.StatusInternalServerError, responses.ErrorResponse{
			Error: "failed to save file to staging",
		})
		return
	}
	defer h.fileStore.Cleanup(stagingPath)

	// Parse DICOM and extract metadata
	dicom_dataset, err := dcmutil.ExtractDICOMData(stagingPath)
	if err != nil {
		responses.WriteJSON(w, http.StatusBadRequest, responses.ErrorResponse{
			Error: fmt.Sprintf("invalid DICOM file: %v", err),
		})
		return
	}

	// Create hash and check if file already exists
	hash := dcmutil.CreateMetadataHash(dicom_dataset)
	finalDir := filepath.Join(h.fileStore.FinalDir(), hash)

	exists, err := h.dbStore.FileExists(r.Context(), finalDir)
	if err != nil {
		responses.WriteJSON(w, http.StatusInternalServerError, responses.ErrorResponse{
			Error: "failed to check file existence",
		})
		return
	}

	if exists {
		responses.WriteJSON(w, http.StatusConflict, responses.ErrorResponse{
			Error: "file with identical content already exists",
		})
		return
	}

	// Create directory for this study
	if _, err := h.fileStore.CreateFinalDirectory(hash); err != nil {
		responses.WriteJSON(w, http.StatusInternalServerError, responses.ErrorResponse{
			Error: "failed to create directory",
		})
		return
	}

	// Extract and save PNG
	pngData, err := dcmutil.GetImageData(dicom_dataset)
	if err != nil {
		h.fileStore.CleanupDirectory(hash)
		responses.WriteJSON(w, http.StatusInternalServerError, responses.ErrorResponse{
			Error: "failed to get image data",
		})
		return
	}

	_, err = h.fileStore.SavePNG(pngData, hash)
	if err != nil {
		h.fileStore.CleanupDirectory(hash)
		responses.WriteJSON(w, http.StatusInternalServerError, responses.ErrorResponse{
			Error: "failed to save PNG",
		})
		return
	}

	// Move DICOM file to final location
	finalPath, err := h.fileStore.MoveDICOM(stagingPath, hash)
	if err != nil {
		h.fileStore.CleanupDirectory(hash)
		responses.WriteJSON(w, http.StatusInternalServerError, responses.ErrorResponse{
			Error: "failed to store file",
		})
		return
	}

	metadataJSON, err := dcmutil.GetDicomMetadataAsJSON(dicom_dataset)
	if err != nil {
		responses.WriteJSON(w, http.StatusInternalServerError, responses.ErrorResponse{
			Error: "failed to process metadata",
		})
		return
	}

	// Save file metadata to database
	fileModel := &models.File{
		Filepath: finalPath,
		Size:     handler.Size,
		Metadata: metadataJSON,
	}

	id, err := h.dbStore.SaveFile(context.Background(), fileModel)
	if err != nil {
		responses.WriteJSON(w, http.StatusInternalServerError, responses.ErrorResponse{
			Error: "failed to store file metadata",
		})
		return
	}

	// Return success response
	responses.WriteJSON(w, http.StatusCreated, responses.SuccessResponse{
		Message:  "file uploaded successfully",
		Filename: handler.Filename,
		FileID:   id,
	})
}

func (h *ImageHandler) HandleDicom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responses.WriteJSON(w, http.StatusMethodNotAllowed, responses.ErrorResponse{
			Error: "method not allowed",
		})
		return
	}

	// Remove the prefix to get just the relevant part of the path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/dicom/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		responses.WriteJSON(w, http.StatusBadRequest, responses.ErrorResponse{
			Error: "invalid path format - expected /api/v1/dicom/{id}/{action}",
		})
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		responses.WriteJSON(w, http.StatusBadRequest, responses.ErrorResponse{
			Error: "invalid id format - must be a number",
		})
		return
	}

	switch parts[1] {
	case "header":
		h.handleDicomHeader(w, r, id)
	case "preview":
		h.handleDicomPreview(w, r, id)
	default:
		responses.WriteJSON(w, http.StatusNotFound, responses.ErrorResponse{
			Error: "unknown action - supported actions are 'header' and 'preview'",
		})
	}
}

func (h *ImageHandler) handleDicomHeader(w http.ResponseWriter, r *http.Request, id int64) {
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		responses.WriteJSON(w, http.StatusBadRequest, responses.ErrorResponse{
			Error: "tag parameter is required",
		})
		return
	}

	value, err := h.dbStore.GetDicomTag(r.Context(), id, tag)
	if err != nil {
		responses.WriteJSON(w, http.StatusNotFound, responses.ErrorResponse{
			Error: "could not find tag in file",
		})
		return
	}

	responses.WriteJSON(w, http.StatusOK, map[string]string{
		"tag":   tag,
		"value": value,
	})
}

func (h *ImageHandler) handleDicomPreview(w http.ResponseWriter, r *http.Request, id int64) {
	file, err := h.dbStore.GetFileByID(r.Context(), id)
	if err != nil {
		responses.WriteJSON(w, http.StatusNotFound, responses.ErrorResponse{
			Error: "file not found",
		})
		return
	}

	// Construct PNG path from DICOM path
	dicomDir := filepath.Dir(file.Filepath)
	pngPath := filepath.Join(dicomDir, "preview.png")

	// Open and serve the PNG file
	pngFile, err := os.Open(pngPath)
	if err != nil {
		responses.WriteJSON(w, http.StatusInternalServerError, responses.ErrorResponse{
			Error: "preview image not found",
		})
		return
	}
	defer pngFile.Close()

	// Set content type header
	w.Header().Set("Content-Type", "image/png")

	// Stream the file to the response
	if _, err := io.Copy(w, pngFile); err != nil {
		// At this point headers are already sent, so we can't send a JSON error response
		// Just log the error
		fmt.Printf("Error streaming preview image: %v\n", err)
	}
}
