package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"periodic-api/internal/models"
	"periodic-api/internal/store"
	"periodic-api/internal/utils"
	"strconv"
	"strings"
)

// ScheduledItemHandler handles HTTP requests for scheduled items
type ScheduledItemHandler struct {
	store     store.ScheduledItemStore
	awsClient *utils.AWSLLMClient
}

// NewScheduledItemHandler creates a new handler with the given store
func NewScheduledItemHandler(store store.ScheduledItemStore) *ScheduledItemHandler {
	// Initialize AWS client
	awsClient, err := utils.NewAWSLLMClient(context.Background())
	if err != nil {
		// Log error but don't fail - the endpoint will return errors if AWS is not configured
		awsClient = nil
	}

	return &ScheduledItemHandler{
		store:     store,
		awsClient: awsClient,
	}
}

// HandleCreateScheduledItem handles POST requests to create a new scheduled item
// @Summary Create a scheduled item
// @Description Create a new scheduled item with the given details
// @Tags scheduled-items
// @Accept json
// @Produce json
// @Param item body models.ScheduledItem true "Scheduled item to create"
// @Success 201 {object} models.ScheduledItem
// @Failure 400 {string} string "Bad request"
// @Router /scheduled-items [post]
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

	// Calculate next execution time
	item.NextExecutionAt = utils.CalculateNextExecution(
		item.StartsAt,
		item.Repeats,
		item.CronExpression,
		item.Expiration,
	)

	createdItem := h.store.CreateScheduledItem(item)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdItem)
}

// HandleGetScheduledItem handles GET requests to retrieve a scheduled item by ID
// @Summary Get a scheduled item by ID
// @Description Get a specific scheduled item by its ID
// @Tags scheduled-items
// @Produce json
// @Param id path int true "Scheduled item ID"
// @Success 200 {object} models.ScheduledItem
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Scheduled item not found"
// @Router /scheduled-items/{id} [get]
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
// @Summary Get all scheduled items
// @Description Retrieve all scheduled items from the store
// @Tags scheduled-items
// @Produce json
// @Success 200 {array} models.ScheduledItem
// @Router /scheduled-items [get]
func (h *ScheduledItemHandler) HandleGetAllScheduledItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	items := h.store.GetAllScheduledItems()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// HandleGetNextScheduledItems handles GET requests to retrieve next scheduled items by execution time
// @Summary Get next scheduled items
// @Description Retrieve the next scheduled items ordered by execution time
// @Tags scheduled-items
// @Produce json
// @Param limit query int false "Maximum number of items to return" default(10)
// @Success 200 {array} models.ScheduledItem
// @Failure 500 {string} string "Internal server error"
// @Router /scheduled-items/next [get]
func (h *ScheduledItemHandler) HandleGetNextScheduledItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse limit parameter, default to 10
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	items, err := h.store.GetNextScheduledItems(limit, 0)
	if err != nil {
		http.Error(w, "Failed to retrieve scheduled items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// HandleDeleteScheduledItem handles DELETE requests to remove a scheduled item
// @Summary Delete a scheduled item
// @Description Delete a scheduled item by its ID
// @Tags scheduled-items
// @Param id path int true "Scheduled item ID"
// @Success 204 "No content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Scheduled item not found"
// @Router /scheduled-items/{id} [delete]
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

// GeneratePromptRequest represents the request body for generating scheduled items
type GeneratePromptRequest struct {
	Prompt   string `json:"prompt" example:"Schedule a weekly team meeting every Tuesday at 2 PM"`
	Timezone string `json:"timezone" example:"America/New_York"`
}

// HandleGenerateScheduledItem handles POST requests to generate a scheduled item from a prompt
// @Summary Generate a scheduled item from a text prompt
// @Description Use AI to generate a scheduled item from a natural language prompt
// @Tags generation
// @Accept json
// @Produce json
// @Param request body GeneratePromptRequest true "Generation request with prompt and timezone"
// @Success 200 {object} models.ScheduledItem
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Failure 503 {string} string "AWS LLM service not available"
// @Router /generate-scheduled-item [post]
func (h *ScheduledItemHandler) HandleGenerateScheduledItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if AWS client is available
	if h.awsClient == nil {
		http.Error(w, "AWS LLM service not available", http.StatusServiceUnavailable)
		return
	}

	// Parse request body
	var req GeneratePromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		http.Error(w, "Prompt cannot be empty", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Timezone) == "" {
		http.Error(w, "Timezone is required", http.StatusBadRequest)
		return
	}

	// Generate JSON from AWS LLM
	generatedJSON, err := h.awsClient.GenerateScheduledItemJSON(r.Context(), req.Prompt, req.Timezone)
	if err != nil {
		http.Error(w, "Failed to generate scheduled item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate and parse the generated JSON into ScheduledItem
	var scheduledItem models.ScheduledItem
	if err := json.Unmarshal([]byte(generatedJSON), &scheduledItem); err != nil {
		http.Error(w, "Generated invalid JSON format: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the generated ScheduledItem as JSON (without storing it)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(scheduledItem)
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

	// Get next scheduled items
	http.HandleFunc("/scheduled-items/next", h.HandleGetNextScheduledItems)

	// Generate scheduled item from prompt
	http.HandleFunc("/generate-scheduled-item", h.HandleGenerateScheduledItem)

	// ScheduledItem instance endpoints
	http.HandleFunc("/scheduled-items/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleGetScheduledItem(w, r)
		case http.MethodDelete:
			h.HandleDeleteScheduledItem(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
