package parsing

// Import necessary packages
import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/banhcanh/portfolio/pkg/components"
)

// Unsafe is a function that creates a templ.Component from raw HTML content
func Unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return
	})
}

// parseMarkdownFile extracts metadata (Date, Title) and content from a Markdown file
func ParseMarkdownFile(content []byte) components.Post {
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

			post := ParseMarkdownFile(content)
			posts = append(posts, post)
		}
	}
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	return posts
}
