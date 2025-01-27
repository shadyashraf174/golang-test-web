package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

// Item represents a simple data structure
type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

var (
	items = make(map[int]Item) // In-memory storage for items
	mu    sync.Mutex           // Mutex to handle concurrent access to the items map
	idSeq = 1                  // Sequence for generating unique IDs
)

func main() {
	// Serve the frontend HTML file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Define API routes
	http.HandleFunc("/items", handleItems) // GET and POST
	http.HandleFunc("/items/", handleItem) // GET, PUT, DELETE for specific item

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// handleItems handles GET and POST requests for /items
func handleItems(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getItems(w, r)
	case http.MethodPost:
		createItem(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleItem handles GET, PUT, and DELETE requests for /items/{id}
func handleItem(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getItem(w, r)
	case http.MethodPut:
		updateItem(w, r)
	case http.MethodDelete:
		deleteItem(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getItems returns a list of all items
func getItems(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Convert items map to a slice
	itemList := make([]Item, 0, len(items))
	for _, item := range items {
		itemList = append(itemList, item)
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(itemList)
}

// createItem adds a new item
func createItem(w http.ResponseWriter, r *http.Request) {
	var newItem Item
	err := json.NewDecoder(r.Body).Decode(&newItem)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Assign a unique ID and add to the map
	newItem.ID = idSeq
	items[idSeq] = newItem
	idSeq++

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newItem)
}

// getItem returns a specific item by ID
func getItem(w http.ResponseWriter, r *http.Request) {
	// Extract ID from the URL
	idStr := r.URL.Path[len("/items/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Find the item
	item, exists := items[id]
	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// updateItem updates an existing item
func updateItem(w http.ResponseWriter, r *http.Request) {
	// Extract ID from the URL
	idStr := r.URL.Path[len("/items/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var updatedItem Item
	err = json.NewDecoder(r.Body).Decode(&updatedItem)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Check if the item exists
	_, exists := items[id]
	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	// Update the item
	updatedItem.ID = id
	items[id] = updatedItem

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedItem)
}

// deleteItem deletes an item by ID
func deleteItem(w http.ResponseWriter, r *http.Request) {
	// Extract ID from the URL
	idStr := r.URL.Path[len("/items/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Check if the item exists
	_, exists := items[id]
	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	// Delete the item
	delete(items, id)

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}
