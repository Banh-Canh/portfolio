package watcher

import (
	"log"

	"github.com/fsnotify/fsnotify"

	"github.com/banhcanh/portfolio/pkg/server"
)

type Watcher struct {
	watcher *fsnotify.Watcher
}

func NewWatcher() *Watcher {
	// Set up file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creating watcher:", err)
	}
	return &Watcher{watcher: watcher} // Use the actual watcher object, not its value
}

func (w *Watcher) StopWatcher() {
	// Set up file watcher
	watcher := w.watcher
	watcher.Close()
}

func (w *Watcher) WatchDirectoryAndUpdateRoutes(dir string, s *server.Server) {
	// Add the 'posts' directory to the watcher
	watcher := w.watcher
	if err := watcher.Add(dir); err != nil {
		log.Fatal("Error adding 'posts' directory to watcher:", err)
	}
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
					s.SetupRoutes(dir)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Fatal("Error watching files:", err)
			}
		}
	}()
}
