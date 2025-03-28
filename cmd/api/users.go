package main

import (
	"context"
	"errors"
	"github/hassanharga/go-social/internal/store"
	"github/hassanharga/go-social/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtxKey userKey = "user"

type FollowedUser struct {
	UserID int64 `json:"user_id"`
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followedUser := getUserFromCtx(r)
	// TODO: get user from auth
	var payload FollowedUser

	if err := utils.ReadJson(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	err := app.store.Followers.Follow(ctx, followedUser.ID, payload.UserID)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			app.conflictError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unfollowedUser := getUserFromCtx(r)
	// TODO: get user from auth
	var payload FollowedUser

	if err := utils.ReadJson(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	err := app.store.Followers.Unfollow(ctx, unfollowedUser.ID, payload.UserID)
	if err != nil {

		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := chi.URLParam(r, "id")

		id, err := strconv.ParseInt(userId, 10, 64)
		// id, err := strconv.Atoi(postId)
		if err != nil {
			app.badRequestError(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.store.Users.GetById(ctx, id)
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

		ctx = context.WithValue(ctx, userCtxKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromCtx(r *http.Request) *store.User {
	user, ok := r.Context().Value(userCtxKey).(*store.User)
	if !ok {
		return nil
	}
	return user
}
