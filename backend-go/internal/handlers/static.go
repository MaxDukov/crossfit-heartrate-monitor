package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// NewSPAHandler создаёт раздачу статики с SPA-fallback.
func NewSPAHandler(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		clean := filepath.Clean("/" + r.URL.Path)
		if strings.HasPrefix(clean, "/api/") || clean == "/api" {
			http.NotFound(w, r)
			return
		}

		name := strings.TrimPrefix(clean, "/")
		if name == "" {
			name = "index.html"
		}

		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(name))); err == nil {
			fs.ServeHTTP(w, r)
			return
		}

		// SPA-fallback: клиентские роуты (/athletes, /sensors) отдают index.html.
		index := filepath.Join(dir, "index.html")
		if _, err := os.Stat(index); err != nil {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, index)
	})
}
