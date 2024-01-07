package parsing

// Import necessary packages
import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/banhcanh/portfolio/pkg/components"
	"github.com/gorilla/mux"
	"github.com/gosimple/slug"
	"github.com/yuin/goldmark"
)

// Unsafe is a function that creates a templ.Component from raw HTML content
func Unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return
	})
}

// parseMarkdownFile extracts metadata (Date, Title) and content from a Markdown file
func parseMarkdownFile(content []byte) components.Post {
	lines := strings.Split(string(content), "\n")
	var date time.Time
	var title string
	var contentBuilder strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "Date:") {
			dateStr := strings.TrimSpace(strings.TrimPrefix(line, "Date:"))
			date, _ = time.Parse("2006/01/02", dateStr)
		} else if strings.HasPrefix(line, "Title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
		} else {
			contentBuilder.WriteString(line + "\n")
		}
	}

	return components.Post{
		Date:    date,
		Title:   title,
		Content: contentBuilder.String(),
	}
}

// GetPosts retrieves a list of parsed posts sorted by date
func GetPosts(dir string) []components.Post {
	var posts []components.Post
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal("Error reading directory:", err)
		return posts
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".md") {
			filePath := filepath.Join(dir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Fatalf("Error reading file %s: %v\n", file.Name(), err)
				continue
			}

			post := parseMarkdownFile(content)
			posts = append(posts, post)
		}
	}
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	return posts
}

// CreatePosts creates dynamic routes for each Markdown file and handles the routing logic
func CreatePosts(r *mux.Router, dir string, dynamicRoutes []string) {
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
			post := parseMarkdownFile(fileContent)

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
			r.HandleFunc(postPath, func(w http.ResponseWriter, r *http.Request) {
				log.Println("Request:", r.Method, r.URL.Path)

				// Read content of the renamed file
				fileContent, err := os.ReadFile(newFilePath)
				if err != nil {
					log.Fatalf("Error reading file %s: %v\n", currentFile.Name(), err)
					return
				}

				// Parse Markdown content into a Post struct
				post := parseMarkdownFile(fileContent)

				// Convert Markdown content to HTML using goldmark
				var buf bytes.Buffer
				if err := goldmark.Convert([]byte(post.Content), &buf); err != nil {
					log.Fatalf("failed to convert markdown to HTML: %v", err)
				}

				// Create a templ.Component from HTML content
				content := Unsafe(buf.String())

				// Render the content page with post metadata and HTML content
				components.ContentPage(post.Title, post.Date.Format("2006/01/02"), content).Render(r.Context(), w)
			})

			// Append the dynamically created route to the list
			dynamicRoutes = append(dynamicRoutes, postPath)
			log.Printf("Created dynamic route: %s", postPath)
		}
	}
}
