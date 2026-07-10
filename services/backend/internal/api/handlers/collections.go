package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"video-archiver/internal/domain"
)

// CollectionsHandler exposes CRUD for collections (user-defined video sets)
// and their video memberships.
type CollectionsHandler struct {
	collections domain.CollectionRepository
}

func NewCollectionsHandler(collections domain.CollectionRepository) *CollectionsHandler {
	return &CollectionsHandler{collections: collections}
}

func (h *CollectionsHandler) RegisterRoutes(r chi.Router) {
	r.Route("/collections", func(r chi.Router) {
		r.Get("/", h.HandleList)
		r.Post("/", h.HandleCreate)
		r.Get("/for-video/{videoID}", h.HandleListForVideo)
		r.Get("/{id}", h.HandleGet)
		r.Put("/{id}", h.HandleUpdate)
		r.Delete("/{id}", h.HandleDelete)
		r.Get("/{id}/videos", h.HandleGetVideos)
		r.Post("/{id}/videos", h.HandleAddVideos)
		r.Delete("/{id}/videos/{videoID}", h.HandleRemoveVideo)
	})
}

// CollectionRequest is the body for creating or updating a collection.
type CollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// maxCollectionNameLength keeps names displayable; there is no meaningful use
// for longer ones.
const maxCollectionNameLength = 100

func (req *CollectionRequest) validate() (string, bool) {
	name := strings.Join(strings.Fields(req.Name), " ")
	if name == "" {
		return "", false
	}
	if runes := []rune(name); len(runes) > maxCollectionNameLength {
		name = strings.TrimSpace(string(runes[:maxCollectionNameLength]))
	}
	return name, true
}

func (h *CollectionsHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	collections, err := h.collections.List()
	if err != nil {
		log.WithError(err).Error("Failed to list collections")
		http.Error(w, "Failed to list collections", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: collections})
}

func (h *CollectionsHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req CollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	name, ok := req.validate()
	if !ok {
		http.Error(w, "Collection name is required", http.StatusBadRequest)
		return
	}

	now := time.Now()
	collection := &domain.Collection{
		ID:          uuid.New().String(),
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := h.collections.Create(collection); err != nil {
		log.WithError(err).Error("Failed to create collection")
		http.Error(w, "Failed to create collection", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, Response{Message: collection})
}

// getCollection loads the collection from the URL's {id}, writing the error
// response itself when the collection can't be served.
func (h *CollectionsHandler) getCollection(w http.ResponseWriter, r *http.Request) *domain.Collection {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing collection ID", http.StatusBadRequest)
		return nil
	}
	collection, err := h.collections.GetByID(id)
	if err != nil {
		log.WithError(err).Error("Failed to get collection")
		http.Error(w, "Failed to get collection", http.StatusInternalServerError)
		return nil
	}
	if collection == nil {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return nil
	}
	return collection
}

func (h *CollectionsHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	collection := h.getCollection(w, r)
	if collection == nil {
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: collection})
}

func (h *CollectionsHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	collection := h.getCollection(w, r)
	if collection == nil {
		return
	}

	var req CollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	name, ok := req.validate()
	if !ok {
		http.Error(w, "Collection name is required", http.StatusBadRequest)
		return
	}

	collection.Name = name
	collection.Description = strings.TrimSpace(req.Description)
	if err := h.collections.Update(collection); err != nil {
		log.WithError(err).Error("Failed to update collection")
		http.Error(w, "Failed to update collection", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: collection})
}

func (h *CollectionsHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	collection := h.getCollection(w, r)
	if collection == nil {
		return
	}
	if err := h.collections.Delete(collection.ID); err != nil {
		log.WithError(err).Error("Failed to delete collection")
		http.Error(w, "Failed to delete collection", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: "Collection deleted successfully"})
}

func (h *CollectionsHandler) HandleGetVideos(w http.ResponseWriter, r *http.Request) {
	collection := h.getCollection(w, r)
	if collection == nil {
		return
	}
	videos, err := h.collections.GetVideos(collection.ID)
	if err != nil {
		log.WithError(err).Error("Failed to get collection videos")
		http.Error(w, "Failed to get collection videos", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: videos})
}

// AddCollectionVideosRequest is the body for adding videos to a collection.
type AddCollectionVideosRequest struct {
	VideoIDs []string `json:"video_ids"`
}

func (h *CollectionsHandler) HandleAddVideos(w http.ResponseWriter, r *http.Request) {
	collection := h.getCollection(w, r)
	if collection == nil {
		return
	}

	var req AddCollectionVideosRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if len(req.VideoIDs) == 0 {
		http.Error(w, "At least one video ID is required", http.StatusBadRequest)
		return
	}

	if err := h.collections.AddVideos(collection.ID, req.VideoIDs); err != nil {
		// Referencing a job that is not a downloaded video is a client error.
		log.WithError(err).Warn("Failed to add videos to collection")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, err := h.collections.GetByID(collection.ID)
	if err != nil || updated == nil {
		log.WithError(err).Error("Failed to reload collection")
		http.Error(w, "Failed to reload collection", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: updated})
}

func (h *CollectionsHandler) HandleRemoveVideo(w http.ResponseWriter, r *http.Request) {
	collection := h.getCollection(w, r)
	if collection == nil {
		return
	}
	videoID := chi.URLParam(r, "videoID")
	if videoID == "" {
		http.Error(w, "Missing video ID", http.StatusBadRequest)
		return
	}
	if err := h.collections.RemoveVideo(collection.ID, videoID); err != nil {
		log.WithError(err).Error("Failed to remove video from collection")
		http.Error(w, "Failed to remove video from collection", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: "Video removed from collection"})
}

// HandleListForVideo returns the IDs of the collections containing a video,
// so the frontend can mark them in the add-to-collection picker.
func (h *CollectionsHandler) HandleListForVideo(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "videoID")
	if videoID == "" {
		http.Error(w, "Missing video ID", http.StatusBadRequest)
		return
	}
	ids, err := h.collections.ListForVideo(videoID)
	if err != nil {
		log.WithError(err).Error("Failed to list collections for video")
		http.Error(w, "Failed to list collections for video", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: ids})
}
