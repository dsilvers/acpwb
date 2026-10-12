package router

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"acpwb_edge/page"
)

// files serves root's contents for URIs under prefix — nginx `alias` when
// prefix is a location, `root` when it's "/". Directories get index.html
// when index is set (botseed.net's `index` + `try_files $uri $uri/ =404`)
// and nginx's 403 otherwise. headers, if set, adds caching headers to
// successful responses. os.Root keeps lookups (symlinks included) inside
// root whatever the URI says.
func files(root, prefix string, index bool, headers func(http.Header)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			page.Error(w, http.StatusMethodNotAllowed)
			return
		}
		if root == "" {
			page.Error(w, http.StatusNotFound)
			return
		}
		rel := strings.TrimPrefix(normalize(r.URL.Path), prefix)
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			rel = "."
		}
		dir, err := os.OpenRoot(root)
		if err != nil {
			page.Error(w, http.StatusNotFound)
			return
		}
		defer dir.Close()

		f, err := dir.Open(rel)
		if err != nil {
			fileError(w, err)
			return
		}
		defer f.Close()
		fi, err := f.Stat()
		if err != nil {
			fileError(w, err)
			return
		}
		if fi.IsDir() {
			if !index {
				page.Error(w, http.StatusForbidden)
				return
			}
			f.Close()
			rel = path.Join(rel, "index.html")
			if f, err = dir.Open(rel); err != nil {
				page.Error(w, http.StatusForbidden)
				return
			}
			defer f.Close()
			if fi, err = f.Stat(); err != nil || fi.IsDir() {
				page.Error(w, http.StatusForbidden)
				return
			}
		}
		if headers != nil {
			headers(w.Header())
		}
		http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
	})
}

func fileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		page.Error(w, http.StatusNotFound)
	case errors.Is(err, fs.ErrPermission):
		page.Error(w, http.StatusForbidden)
	default:
		// Includes os.Root refusing a path that escapes root.
		page.Error(w, http.StatusNotFound)
	}
}
