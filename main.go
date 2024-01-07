package main

import (
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

var dynamicRoutes []string

func main() {
	r := mux.NewRouter()
	server := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}
	dir := "./posts"

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

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal("Error reading directory:", err)
		return
	}
	parsing.CreatePosts(r, files, dynamicRoutes)

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
					handlers.UpdateRoutes(r, files, dynamicRoutes)
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

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("404 Not Found:", r.URL.Path)
		components.NotFoundComponent().Render(r.Context(), w)
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request:", r.Method, r.URL.Path)
		components.IndexPage(parsing.GetPosts()).Render(r.Context(), w)
	})

	log.Println("Server is listening on :3000")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting server: %v\n", err)
	}
}
