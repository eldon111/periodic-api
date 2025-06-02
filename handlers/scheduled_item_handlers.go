package handlers

import (
	"awesomeProject/models"
	"awesomeProject/store"
	"encoding/json"
	"net/http"
	"strconv"
)

// ScheduledItemHandler handles HTTP requests for scheduled items
type ScheduledItemHandler struct {
	store store.ScheduledItemStore
}

// NewScheduledItemHandler creates a new handler with the given store
func NewScheduledItemHandler(store store.ScheduledItemStore) *ScheduledItemHandler {
	return &ScheduledItemHandler{
		store: store,
	}
}

// HandleCreateScheduledItem handles POST requests to create a new scheduled item
func (h *ScheduledItemHandler) HandleCreateScheduledItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var item models.ScheduledItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdItem := h.store.CreateScheduledItem(item)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdItem)
}

// HandleGetScheduledItem handles GET requests to retrieve a scheduled item by ID
func (h *ScheduledItemHandler) HandleGetScheduledItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/scheduled-items/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, exists := h.store.GetScheduledItem(id)
	if !exists {
		http.Error(w, "Scheduled item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// HandleGetAllScheduledItems handles GET requests to retrieve all scheduled items
func (h *ScheduledItemHandler) HandleGetAllScheduledItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	items := h.store.GetAllScheduledItems()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// HandleUpdateScheduledItem handles PUT requests to update a scheduled item
func (h *ScheduledItemHandler) HandleUpdateScheduledItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/scheduled-items/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var item models.ScheduledItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedItem, exists := h.store.UpdateScheduledItem(id, item)
	if !exists {
		http.Error(w, "Scheduled item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedItem)
}

// HandleDeleteScheduledItem handles DELETE requests to remove a scheduled item
func (h *ScheduledItemHandler) HandleDeleteScheduledItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/scheduled-items/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if success := h.store.DeleteScheduledItem(id); !success {
		http.Error(w, "Scheduled item not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetupRoutes configures the HTTP routes for scheduled items
func (h *ScheduledItemHandler) SetupRoutes() {
	// ScheduledItem collection endpoints
	http.HandleFunc("/scheduled-items", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleGetAllScheduledItems(w, r)
		case http.MethodPost:
			h.HandleCreateScheduledItem(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ScheduledItem instance endpoints
	http.HandleFunc("/scheduled-items/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleGetScheduledItem(w, r)
		case http.MethodPut:
			h.HandleUpdateScheduledItem(w, r)
		case http.MethodDelete:
			h.HandleDeleteScheduledItem(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
