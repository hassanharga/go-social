package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github/hassanharga/go-social/internal/store"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (app *application) authTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// read auth header
		reqHeader := r.Header.Get("Authorization")
		if reqHeader == "" {
			// if no header, return unauthorized
			app.unauthorizedError(w, r, fmt.Errorf("missing authorization header"))
			return
		}

		//parse it
		parts := strings.Split(reqHeader, " ")
		if len(parts) < 2 || parts[0] != "Bearer" {
			app.unauthorizedError(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		token := parts[1]
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedError(w, r, fmt.Errorf("invalid token"))
			return
		}

		claims := jwtToken.Claims.(jwt.MapClaims)

		userId, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedError(w, r, fmt.Errorf("invalid token"))
			return
		}

		ctx := r.Context()

		// fetch user data
		user, err := app.getUser(ctx, userId)
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, userCtxKey, user)

		// Call the next handler in the chain
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) basicMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// read auth header
			reqHeader := r.Header.Get("Authorization")
			if reqHeader == "" {
				// if no header, return unauthorized
				app.unauthorizedBasicError(w, r, fmt.Errorf("missing authorization header"))
				return
			}

			//parse it
			parts := strings.Split(reqHeader, " ")
			if len(parts) < 2 || parts[0] != "Basic" {
				app.unauthorizedBasicError(w, r, fmt.Errorf("authorization header is malformed"))
				return
			}

			// decode it
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicError(w, r, err)
				return
			}

			username := app.config.auth.basic.user
			password := app.config.auth.basic.password

			// check the credentials
			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != password {
				app.unauthorizedBasicError(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			// Call the next handler in the chain
			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) checkPostOwnership(requiredRole store.RoleKeys, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromCtx(r)
		post := getPostFromCtx(r)

		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenError(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, roleName store.RoleKeys) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

func (app *application) getUser(ctx context.Context, userId int64) (*store.User, error) {
	if !app.config.cache.enabled {
		return app.store.Users.GetById(ctx, userId)
	}

	// fetch user from cache
	user, err := app.cacheStorage.Users.Get(ctx, userId)
	if err != nil {
		return nil, err
	}

	// fetch the user from the database
	if user == nil {
		user, err = app.store.Users.GetById(ctx, userId)
		if err != nil {
			return nil, err
		}

		// update the cache
		err := app.cacheStorage.Users.Set(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}
