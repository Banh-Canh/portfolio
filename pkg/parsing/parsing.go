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
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type Post struct {
	Date    time.Time
	Title   string
	Content templ.Component
}

// Unsafe is a function that creates a templ.Component from raw HTML content
func Unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return
	})
}

// mdToHTML converts Markdown content to HTML
func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

// parseMarkdownFile extracts metadata (Date, Title) and content from a Markdown file
func ParseMarkdownFile(content []byte) Post {
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

	md := []byte(contentBuilder.String())
	html := mdToHTML(md)
	htmlString := string(html)

	return Post{
		Date:    date,
		Title:   title,
		Content: writeComponent(htmlString),
	}
}

// GetPosts retrieves a list of parsed posts sorted by date
func GetPosts(dir string) []Post {
	var posts []Post
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

// writeComponent is a helper function that creates a templ.Component from HTML string
func writeComponent(htmlString string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, htmlString)
		return err
	})
}
