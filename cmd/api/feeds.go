package main

import (
	"github/hassanharga/go-social/internal/store"
	"github/hassanharga/go-social/utils"
	"net/http"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	// pagination,filter
	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	// log.Println("getUserFeedHandler", fq)

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := utils.Validate.Struct(fq); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	feed, err := app.store.Posts.GetFeed(ctx, int64(82), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err = app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
	}
}
