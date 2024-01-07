package handlers

import (
	"log"

	parsing "github.com/banhcanh/portfolio/pkg/parsing"
	"github.com/gorilla/mux"
)

// Function to update routes based on .md files
func UpdateRoutes(r *mux.Router, dir string, dynamicRoutes []string) {
	// Log the start of the function
	log.Println("Updating routes...")

	// Call parsing.CreatePosts with logging
	parsing.CreatePosts(r, dir, dynamicRoutes)

	// Log the end of the function
	log.Println("Routes updated successfully.")
}
