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
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	// read data from body
	var payload createPostPayload
	if err := utils.ReadJson(w, r, &payload); err != nil {
		utils.WriteJsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  1, // TODO: get user id from context
	}
	ctx := r.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		utils.WriteJsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := utils.WriteJson(w, http.StatusOK, post); err != nil {
		utils.WriteJsonError(w, http.StatusInternalServerError, err.Error())
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	postId := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(postId, 10, 64)
	// id, err := strconv.Atoi(postId)
	if err != nil {
		utils.WriteJsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	post, err := app.store.Posts.GetById(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			utils.WriteJsonError(w, http.StatusNotFound, err.Error())
			return
		default:
			utils.WriteJsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if err := utils.WriteJson(w, http.StatusOK, post); err != nil {
		utils.WriteJsonError(w, http.StatusInternalServerError, err.Error())
	}
}
