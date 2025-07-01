package handlers

import (
	"encoding/json"
	"net/http"
	"periodic-api/internal/models"
	"periodic-api/internal/store"
	"strconv"
)

// TodoItemHandler handles HTTP requests for todo items
type TodoItemHandler struct {
	store store.TodoItemStore
}

// NewTodoItemHandler creates a new handler with the given store
func NewTodoItemHandler(store store.TodoItemStore) *TodoItemHandler {
	return &TodoItemHandler{
		store: store,
	}
}

// HandleCreateTodoItem handles POST requests to create a new todo item
// @Summary Create a todo item
// @Description Create a new todo item with the given details
// @Tags todo-items
// @Accept json
// @Produce json
// @Param item body models.TodoItem true "Todo item to create"
// @Success 201 {object} models.TodoItem
// @Failure 400 {string} string "Bad request"
// @Router /todo-items [post]
func (h *TodoItemHandler) HandleCreateTodoItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var item models.TodoItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdItem := h.store.CreateTodoItem(item)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdItem)
}

// HandleGetTodoItem handles GET requests to retrieve a todo item by ID
// @Summary Get a todo item by ID
// @Description Get a specific todo item by its ID
// @Tags todo-items
// @Produce json
// @Param id path int true "Todo item ID"
// @Success 200 {object} models.TodoItem
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Todo item not found"
// @Router /todo-items/{id} [get]
func (h *TodoItemHandler) HandleGetTodoItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/todo-items/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, exists := h.store.GetTodoItem(id)
	if !exists {
		http.Error(w, "Todo item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// HandleGetAllTodoItems handles GET requests to retrieve all todo items
// @Summary Get all todo items
// @Description Retrieve all todo items from the store
// @Tags todo-items
// @Produce json
// @Success 200 {array} models.TodoItem
// @Router /todo-items [get]
func (h *TodoItemHandler) HandleGetAllTodoItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	items := h.store.GetAllTodoItems()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// HandleUpdateTodoItem handles PUT requests to update a todo item
// @Summary Update a todo item
// @Description Update a todo item by its ID
// @Tags todo-items
// @Accept json
// @Produce json
// @Param id path int true "Todo item ID"
// @Param item body models.TodoItem true "Updated todo item"
// @Success 200 {object} models.TodoItem
// @Failure 400 {string} string "Bad request"
// @Failure 404 {string} string "Todo item not found"
// @Router /todo-items/{id} [put]
func (h *TodoItemHandler) HandleUpdateTodoItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/todo-items/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedItem models.TodoItem
	if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item, exists := h.store.UpdateTodoItem(id, updatedItem)
	if !exists {
		http.Error(w, "Todo item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// HandleDeleteTodoItem handles DELETE requests to remove a todo item
// @Summary Delete a todo item
// @Description Delete a todo item by its ID
// @Tags todo-items
// @Param id path int true "Todo item ID"
// @Success 204 "No content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Todo item not found"
// @Router /todo-items/{id} [delete]
func (h *TodoItemHandler) HandleDeleteTodoItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/todo-items/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if success := h.store.DeleteTodoItem(id); !success {
		http.Error(w, "Todo item not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetupRoutes configures the HTTP routes for todo items
func (h *TodoItemHandler) SetupRoutes() {
	// TodoItem collection endpoints
	http.HandleFunc("/todo-items", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleGetAllTodoItems(w, r)
		case http.MethodPost:
			h.HandleCreateTodoItem(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// TodoItem instance endpoints
	http.HandleFunc("/todo-items/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleGetTodoItem(w, r)
		case http.MethodPut:
			h.HandleUpdateTodoItem(w, r)
		case http.MethodDelete:
			h.HandleDeleteTodoItem(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}