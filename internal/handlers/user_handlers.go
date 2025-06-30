package handlers

import (
	"encoding/json"
	"net/http"
	"periodic-api/internal/models"
	"periodic-api/internal/store"
	"strconv"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	store store.UserStore
}

// NewUserHandler creates a new handler with the given store
func NewUserHandler(store store.UserStore) *UserHandler {
	return &UserHandler{
		store: store,
	}
}

// HandleCreateUser handles POST requests to create a new user
// @Summary Create a user
// @Description Create a new user with the given details
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.User true "User to create"
// @Success 201 {object} models.User
// @Failure 400 {string} string "Bad request"
// @Router /users [post]
func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdUser := h.store.CreateUser(user)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
}

// HandleGetUser handles GET requests to retrieve a user by ID
// @Summary Get a user by ID
// @Description Get a specific user by their ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.User
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "User not found"
// @Router /users/{id} [get]
func (h *UserHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/users/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	user, exists := h.store.GetUser(id)
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// HandleGetAllUsers handles GET requests to retrieve all users
// @Summary Get all users
// @Description Retrieve all users from the store
// @Tags users
// @Produce json
// @Success 200 {array} models.User
// @Router /users [get]
func (h *UserHandler) HandleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users := h.store.GetAllUsers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// HandleUpdateUser handles PUT requests to update a user
// @Summary Update a user
// @Description Update a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body models.User true "Updated user data"
// @Success 200 {object} models.User
// @Failure 400 {string} string "Bad request"
// @Failure 404 {string} string "User not found"
// @Router /users/{id} [put]
func (h *UserHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/users/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedUser models.User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, exists := h.store.UpdateUser(id, updatedUser)
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// HandleDeleteUser handles DELETE requests to remove a user
// @Summary Delete a user
// @Description Delete a user by their ID
// @Tags users
// @Param id path int true "User ID"
// @Success 204 "No content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "User not found"
// @Router /users/{id} [delete]
func (h *UserHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/users/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if success := h.store.DeleteUser(id); !success {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetupRoutes configures the HTTP routes for users
func (h *UserHandler) SetupRoutes() {
	// User collection endpoints
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleGetAllUsers(w, r)
		case http.MethodPost:
			h.HandleCreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// User instance endpoints
	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleGetUser(w, r)
		case http.MethodPut:
			h.HandleUpdateUser(w, r)
		case http.MethodDelete:
			h.HandleDeleteUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
