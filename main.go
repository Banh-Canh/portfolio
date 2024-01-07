package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	components "github.com/banhcanh/portfolio/pkg/components"
	handlers "github.com/banhcanh/portfolio/pkg/handlers"
	parsing "github.com/banhcanh/portfolio/pkg/parsing"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/mux"
)

const postDir = "./posts"

var dynamicRoutes []string

func main() {
	// Create a new Gorilla mux router
	r := mux.NewRouter()

	// Create an HTTP server with the router
	server := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}
	dir := postDir // Directory to watch for file changes

	// Set up file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creating watcher:", err)
	}
	defer watcher.Close()

	// Add the 'posts' directory to the watcher
	if err := watcher.Add(dir); err != nil {
		log.Fatal("Error adding 'posts' directory to watcher:", err)
	}

	// create initial routes
	parsing.CreatePosts(r, dir, dynamicRoutes)

	// Goroutine to watch for file changes
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Remove == fsnotify.Remove {
					// Handle file changes (add or remove)
					log.Printf("File %s modified or added\n", event.Name)
					handlers.UpdateRoutes(r, dir, dynamicRoutes)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Fatal("Error watching files:", err)
			}
		}
	}()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-stop
		log.Printf("Received signal %v. Shutting down...\n", sig)

		// Stop the file watcher
		watcher.Close()

		// Gracefully shutdown the HTTP server
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("Error during server shutdown: %v\n", err)
		}
	}()

	// Serve 404 pages
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("404 Not Found:", r.URL.Path)
		components.NotFoundComponent().Render(r.Context(), w)
	})

	// Serve IndexPage
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request:", r.Method, r.URL.Path)
		var buf bytes.Buffer
		// Render the IndexPage and write the content into the buffer
		components.IndexPage(parsing.GetPosts(dir)).Render(context.Background(), &buf)
		// Get the content of the buffer as a string and store it in the variable indexPage
		indexPage := buf.String()
		// Now you can use indexPage as needed
		components.LoadingPage(indexPage).Render(r.Context(), w)
	})

	// Serve Static Assets
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	log.Println("Server is listening on :3000")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting server: %v\n", err)
	}
}
