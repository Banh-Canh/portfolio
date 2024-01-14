package server

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/banhcanh/portfolio/pkg/components"
	"github.com/banhcanh/portfolio/pkg/parsing"
	"github.com/gorilla/mux"
	"github.com/gosimple/slug"
)

type Server struct {
	Port   string
	server *http.Server
	Router *mux.Router
}

func NewServer(port string) *Server {
	// Create a new Gorilla mux router
	r := mux.NewRouter()

	// Create an HTTP server with the router
	server := &http.Server{
		Addr:    port,
		Handler: r,
	}

	return &Server{Port: port, server: server, Router: r}
}

// Start initializes and starts the HTTP server.
func (s *Server) Start() {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting server: %v\n", err)
	}
}

func (s *Server) Stop() {
	if err := s.server.Shutdown(context.Background()); err != nil {
		log.Printf("Error during server shutdown: %v\n", err)
	}
}

// CreatePosts creates dynamic routes for each Markdown file and handles the routing logic
func (s *Server) SetupRoutes(dir string) {
	var dynamicRoutes []string
	// Serve 404 pages
	s.Router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("404 Not Found:", r.URL.Path)
		components.NotFoundComponent().Render(r.Context(), w)
	})

	// Serve IndexPage
	s.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request:", r.Method, r.URL.Path)
		var buf bytes.Buffer
		// Render the IndexPage and write the content into the buffer
		components.IndexPage(parsing.GetPosts(dir)).Render(context.Background(), &buf)
		// Get the content of the buffer as a string and store it in the variable indexPage
		indexPage := buf.String()
		// Now you can use indexPage as needed
		components.LoadingPage(indexPage).Render(r.Context(), w)
	})
	// Server Assets
	s.Router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal("Error reading directory:", err)
		return
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".md") {
			// Create a new variable scoped within the loop
			currentFile := file

			// Read content of the Markdown file
			filePath := filepath.Join(dir, currentFile.Name())
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				log.Fatalf("Error reading file %s: %v\n", currentFile.Name(), err)
				return
			}

			// Parse Markdown content into a Post struct
			post := parsing.ParseMarkdownFile(fileContent)

			// Create a new filename based on the post date
			newFileName := post.Date.Format("2006-01-02") + "-" + slug.Make(post.Title) + ".md"
			newFilePath := filepath.Join(dir, newFileName)

			// Rename the file to include the formatted date
			if err := os.Rename(filePath, newFilePath); err != nil {
				log.Fatalf("Error renaming file %s to %s: %v\n", currentFile.Name(), newFileName, err)
				return
			}

			// Define the route path for the post
			postPath := "/" + path.Join(post.Date.Format("2006/01/02"), slug.Make(post.Title))

			// Handle requests for the post route
			s.Router.HandleFunc(postPath, func(w http.ResponseWriter, r *http.Request) {
				log.Println("Request:", r.Method, r.URL.Path)

				// Read content of the renamed file
				fileContent, err := os.ReadFile(newFilePath)
				if err != nil {
					log.Fatalf("Error reading file %s: %v\n", currentFile.Name(), err)
					return
				}

				// Parse Markdown content into a Post struct
				post := parsing.ParseMarkdownFile(fileContent)
				var buf bytes.Buffer
				components.ContentPage(post.Title, post.Date.Format("2006/01/02"), post.Content).Render(context.Background(), &buf)
				contentPage := buf.String()

				// Render the content page with post metadata and HTML content
				components.LoadingPage(contentPage).Render(r.Context(), w)
			})

			// Append the dynamically created route to the list
			dynamicRoutes = append(dynamicRoutes, postPath)
			log.Printf("Created dynamic route: %s", postPath)
		}
	}
}
