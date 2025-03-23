package main

import (
	"errors"
	"github/hassanharga/go-social/internal/store"
	"github/hassanharga/go-social/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type createPostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	// read data from body
	var payload createPostPayload
	if err := utils.ReadJson(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  1, // TODO: get user id from context
	}

	// validate the payload
	ctx := r.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := utils.WriteJson(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	postId := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(postId, 10, 64)
	// id, err := strconv.Atoi(postId)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	post, err := app.store.Posts.GetById(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	comments, err := app.store.Comments.GetByPostId(ctx, id)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := utils.WriteJson(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}
