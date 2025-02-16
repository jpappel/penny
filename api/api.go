package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jpappel/penny/auth"
	"github.com/jpappel/penny/data"
)

const DB_FILE = "file:data.sqlite3"

func ListPages(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "now", time.Now().Unix())
	pdb := data.New(DB_FILE)

	pageInfos, err := pdb.GetPagesInfo(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get info on pages", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "<h1>Internal Server Error</h1>")
		return
	}

	err = tmpls.ExecuteTemplate(w, "pages.html", pageInfos)
	if err != nil {
		slog.ErrorContext(ctx, "An error occured while executing template", slog.String("error", err.Error()))
	}
}

func GetComments(w http.ResponseWriter, r *http.Request) {
	// TODO: reuse db connection
	pdb := data.New(DB_FILE)
	pageUrl := r.PathValue("pageUrl")

	slog.Info("fetching coments for page", slog.Any("pageUrl", pageUrl))

	ctx := context.WithValue(r.Context(), "now", time.Now().Unix())

	page, err := pdb.GetPageComments(ctx, pageUrl)
	if err == data.ErrNoPage {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "<h1>Error 404</h1><p>Comments for page %s not found</p>\n", pageUrl)
		return
	} else if err != nil {
		slog.ErrorContext(ctx, "Failed to get Page Comments", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "<h1>Internal Server Error</h1>")
		return
	}

	err = tmpls.ExecuteTemplate(w, "comments.html", page)
	if err != nil {
		slog.ErrorContext(r.Context(), "An error occured while executing template", slog.Any("error", err))
	}

}

// TODO: parse form data
func PostComment(w http.ResponseWriter, r *http.Request) {
	// pageUrl := r.PathValue("pageUrl")
	// TODO: extract user from request and add to context

	r.ParseForm()
	// comment := r.Form.Get("commentText")
	// TODO: run filters over text

	// pdb := data.New(DB_FILE)
	// pdb.PostComment(r.Context(), pageUrl, "", comment, nil)
}

func NewComment(w http.ResponseWriter, r *http.Request) {
	// TODO: extract user from request and add to context
	d := struct {
		User      *auth.User
		Providers []auth.Provider
	}{
		User: &auth.User{Name: "JP Appel", Email: "jp@jpappel.xyz"},
		Providers: []auth.Provider{
			{Name: "GitHub", Url: "https://github.com"},
		},
	}
	err := tmpls.ExecuteTemplate(w, "new_comment.html", d)
	if err != nil {
		slog.ErrorContext(r.Context(), "An error occured while executing template", slog.Any("error", err))
	}
}

func NewMux(baseUrl string) *http.ServeMux {
	var base string
	if baseUrl == "" {
		base = ""
	} else {
		base = fmt.Sprint("/", baseUrl)
	}
	mux := http.NewServeMux()

	logger := slog.Default()

	mux.HandleFunc(fmt.Sprint("/", baseUrl), ListPages)
	mux.HandleFunc(fmt.Sprint(base, "/comments/{pageUrl...}"), GetComments)
	mux.Handle(fmt.Sprintf("GET %s/new/comments/{pageUrl...}", base), Log(http.HandlerFunc(NewComment), logger))
	mux.Handle(fmt.Sprintf("POST %s/new/comments/{pageUrl...}", base), Log(http.HandlerFunc(PostComment), logger))

	return mux
}
