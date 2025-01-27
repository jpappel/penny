package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jpappel/penny/data"
)

const DB_FILE = "file:data.sqlite3"

func GetComments(w http.ResponseWriter, r *http.Request) {
	pdb := data.New(DB_FILE)
	pageUrl := r.PathValue("pageUrl")

	slog.Info("fetching coments for page", slog.Any("pageUrl", pageUrl))

	ctx := context.WithValue(r.Context(), "now", time.Now().UTC().Unix())

	page, err := pdb.GetPageComments(ctx, pageUrl)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to get Page Comments", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ch := make(chan data.Comment, 32)
	go func() {
		defer close(ch)
		page.Comments.Range(func(k any, v any) bool {
			commment, ok := v.(data.Comment)
			if !ok {
				return false
			}

			ch <- commment
			return true
		})
	}()

	err = tmpls.ExecuteTemplate(w, "comments.html", struct {
		Page     *data.Page
		Comments chan data.Comment
	}{page, ch})
	if err != nil {
		slog.ErrorContext(r.Context(), "An error occured while executing template", slog.Any("error", err))
	}

}

func PostComment(w http.ResponseWriter, r *http.Request) {
	// pdb := data.New(DB_FILE)
}

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/penny/comments/{pageUrl...}", GetComments)

	return mux
}
