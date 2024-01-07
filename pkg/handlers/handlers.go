package handler

import (
	"io/fs"
	"log"

	parsing "github.com/banhcanh/portfolio/pkg/parsing"
	"github.com/gorilla/mux"
)

// Function to update routes based on .md files
func UpdateRoutes(r *mux.Router, files []fs.DirEntry, dynamicRoutes []string) {
	// Log the start of the function
	log.Println("Updating routes...")

	// Remove existing routes
	for _, route := range dynamicRoutes {
		r.Handle(route, nil)
		log.Printf("Removed existing route: %s", route)
	}

	// Clear the routes slice
	dynamicRoutes = nil

	// Call parsing.CreatePosts with logging
	parsing.CreatePosts(r, files, dynamicRoutes)

	// Log the end of the function
	log.Println("Routes updated successfully.")
}
