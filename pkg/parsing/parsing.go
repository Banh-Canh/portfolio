package parsing

import (
	"bytes"
	"context"
	"io"
	"io/fs"
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

func Unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return
	})
}

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

func GetPosts() []components.Post {
	var posts []components.Post
	dir := "./posts" // Change this to the path of your directory
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

func CreatePosts(r *mux.Router, files []fs.DirEntry, dynamicRoutes []string) {
	dir := "./posts"

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".md") {
			// Create a new variable scoped within the loop
			currentFile := file

			filePath := filepath.Join(dir, currentFile.Name())
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				log.Fatalf("Error reading file %s: %v\n", currentFile.Name(), err)
				return
			}
			post := parseMarkdownFile(fileContent)
			newFileName := post.Date.Format("2006-01-02") + ".md"
			newFilePath := filepath.Join(dir, newFileName)
			// Rename the file
			if err := os.Rename(filePath, newFilePath); err != nil {
				log.Fatalf("Error renaming file %s to %s: %v\n", currentFile.Name(), newFileName, err)
				return
			}

			postPath := "/" + path.Join(post.Date.Format("2006/01/02"), slug.Make(post.Title))

			// Use a function that captures the currentFile variable
			r.HandleFunc(postPath, func(w http.ResponseWriter, r *http.Request) {
				log.Println("Request:", r.Method, r.URL.Path)
				fileContent, err := os.ReadFile(newFilePath)
				if err != nil {
					log.Fatalf("Error reading file %s: %v\n", currentFile.Name(), err)
					return
				}

				post := parseMarkdownFile(fileContent)
				var buf bytes.Buffer
				if err := goldmark.Convert([]byte(post.Content), &buf); err != nil {
					log.Fatalf("failed to convert markdown to HTML: %v", err)
				}

				content := Unsafe(buf.String())
				components.ContentPage(post.Title, post.Date.Format("2006/01/02"), content).Render(r.Context(), w)
			})
			dynamicRoutes = append(dynamicRoutes, postPath)
			log.Printf("Created dynamic route: %s", postPath)
		}
	}
}
