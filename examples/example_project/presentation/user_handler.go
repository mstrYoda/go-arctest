package presentation

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mstrYoda/go-arctest/examples/example_project/application"
)

// UserHandler handles HTTP requests related to users
type UserHandler struct {
	userService *application.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService *application.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUser handles GET requests for user details
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// In a real app, you would get the ID from the URL or query parameters
	id := r.URL.Query().Get("id")

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting user: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// CreateUser handles POST requests for creating users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateUser(input.Username, input.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating user: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
