package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"real-time/backend/internal/repository"
)

// CategoryHandler provides HTTP handlers for categories
type CategoryHandler struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(categoryRepo repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{categoryRepo: categoryRepo}
}

func (h *CategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryRepo.FetchAllCategories()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		http.Error(w, `{"message": "Failed to fetch categories"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response := map[string][]repository.Category{"Categories": categories}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding categories response: %v", err)
		http.Error(w, `{"message": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}
