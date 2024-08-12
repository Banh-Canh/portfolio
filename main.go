package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/banhcanh/portfolio/pkg/server"
	"github.com/banhcanh/portfolio/pkg/watcher"
)

const postDir = "./posts"

func main() {
	log.Println("Server is listening on :3000")
	s := server.NewServer(":3000")

	dir := postDir // Directory to watch for file changes
	w := watcher.NewWatcher()
	defer w.StopWatcher()
	// create initial routes
	s.SetupRoutes(dir)
	w.WatchDirectoryAndUpdateRoutes(dir, s)
	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-stop
		log.Printf("Received signal %v. Shutting down...\n", sig)
		// Stop the file watcher
		w.StopWatcher()
		// Gracefully shutdown the HTTP server
		s.Stop()
	}()
	// Serve Static Assets
	log.Println("Server is starting")
	s.Start()
}
