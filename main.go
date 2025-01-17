// main.go

package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	osm "github.com/gnolang/gno/pkgs/os"
	"github.com/gorilla/mux"
	"github.com/gotuna/gotuna"

	"github.com/gnolang/www_gno_land/static" // for static files
)

var flags struct {
	bindAddr string
	viewsDir string
	pagesDir string
}

func init() {
	flag.StringVar(&flags.bindAddr, "bind", "127.0.0.1:8888", "server listening address")
	flag.StringVar(&flags.viewsDir, "views-dir", "./views", "views directory location")
	flag.StringVar(&flags.pagesDir, "pages-dir", "./pages", "pages directory location")
}

func main() {
	flag.Parse()

	app := gotuna.App{
		ViewFiles: os.DirFS(flags.viewsDir),
		Router:    gotuna.NewMuxRouter(),
		Static:    static.EmbeddedStatic,
		// StaticPrefix: "static/",
	}

	app.Router.Handle("/", handlerHome(app))
	app.Router.Handle("/about", handlerAbout(app))
	app.Router.Handle("/game-of-realms", handlerGor(app))
	app.Router.Handle("/r/{path:.*}", handlerRedirect(app))
	app.Router.Handle("/p/{path:.*}", handlerRedirect(app))
	app.Router.Handle("/static/{path:.+}", handlerStaticFile(app))
	app.Router.Handle("/favicon.ico", handlerFavicon(app))

	fmt.Printf("Running on http://%s\n", flags.bindAddr)
	err := http.ListenAndServe(flags.bindAddr, app.Router)
	if err != nil {
		fmt.Fprintf(os.Stderr, "HTTP server stopped with error: %+v\n", err)
	}
}

func handlerHome(app gotuna.App) http.Handler {
	md := filepath.Join(flags.pagesDir, "HOME.md")
	mainContent := osm.MustReadFile(md)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NewTemplatingEngine().
			Set("MainContent", string(mainContent)).
			Render(w, r, "home.html", "funcs.html")
	})
}

func handlerAbout(app gotuna.App) http.Handler {
	md := filepath.Join(flags.pagesDir, "ABOUT.md")
	mainContent := osm.MustReadFile(md)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NewTemplatingEngine().
			Set("MainContent", string(mainContent)).
			Render(w, r, "about.html", "funcs.html")
	})
}

func handlerGor(app gotuna.App) http.Handler {
	md := filepath.Join(flags.pagesDir, "GOR.md")
	mainContent := osm.MustReadFile(md)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NewTemplatingEngine().
			Set("MainContent", string(mainContent)).
			Render(w, r, "gor.html", "funcs.html")
	})
}

func handlerRedirect(app gotuna.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://test3.gno.land"+r.URL.Path, http.StatusFound)
	})
}

func handlerStaticFile(app gotuna.App) http.Handler {
	fs := http.FS(app.Static)
	fileapp := http.StripPrefix("/static", http.FileServer(fs))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fpath := filepath.Clean(vars["path"])
		f, err := fs.Open(fpath)
		if os.IsNotExist(err) {
			handleNotFound(app, fpath, w, r)
			return
		}
		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			handleNotFound(app, fpath, w, r)
			return
		}

		// TODO: ModTime doesn't work for embed?
		// w.Header().Set("ETag", fmt.Sprintf("%x", stat.ModTime().UnixNano()))
		// w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%s", "31536000"))
		fileapp.ServeHTTP(w, r)
	})
}

func handlerFavicon(app gotuna.App) http.Handler {
	fs := http.FS(app.Static)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fpath := "img/favicon.ico"
		f, err := fs.Open(fpath)
		if os.IsNotExist(err) {
			handleNotFound(app, fpath, w, r)
			return
		}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Header().Set("Cache-Control", "public, max-age=604800") // 7d
		io.Copy(w, f)
	})
}

func handleNotFound(app gotuna.App, path string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	app.NewTemplatingEngine().
		Set("title", "Not found").
		Set("path", path).
		Render(w, r, "404.html", "header.html")
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}
