package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github/hassanharga/go-social/internal/mailer"
	"github/hassanharga/go-social/internal/store"
	"github/hassanharga/go-social/utils"
	"net/http"

	// "github.com/golang-jwt/jwt/v5"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

// registerUserHandler godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/register [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := utils.ReadJson(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
		// Role: store.Role{
		// 	Name: "user",
		// },
	}

	// hash the user password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken := uuid.New().String()

	// hash the token for storage but keep the plain token for email
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.expiry)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestError(w, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}
	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)

	isProdEnv := app.config.env == "production"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	// send mail
	status, err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending welcome email", "error", err)

		// rollback user creation if email fails (SAGA pattern)
		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("error deleting user", "error", err)
		}

		app.internalServerError(w, r, err)
		return
	}

	app.logger.Infow("Email sent", "status code", status)

	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
	}
}

// type CreateUserTokenPayload struct {
// 	Email    string `json:"email" validate:"required,email,max=255"`
// 	Password string `json:"password" validate:"required,min=3,max=72"`
// }

// // createTokenHandler godoc
// //
// //	@Summary		Creates a token
// //	@Description	Creates a token for a user
// //	@Tags			authentication
// //	@Accept			json
// //	@Produce		json
// //	@Param			payload	body		CreateUserTokenPayload	true	"User credentials"
// //	@Success		200		{string}	string					"Token"
// //	@Failure		400		{object}	error
// //	@Failure		401		{object}	error
// //	@Failure		500		{object}	error
// //	@Router			/authentication/token [post]
// func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
// 	var payload CreateUserTokenPayload
// 	if err := readJSON(w, r, &payload); err != nil {
// 		app.badRequestError(w, r, err)
// 		return
// 	}

// 	if err := utils.Validate.Struct(payload); err != nil {
// 		app.badRequestError(w, r, err)
// 		return
// 	}

// 	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
// 	if err != nil {
// 		switch err {
// 		case store.ErrNotFound:
// 			app.unauthorizedError(w, r, err)
// 		default:
// 			app.internalServerError(w, r, err)
// 		}
// 		return
// 	}

// 	if err := user.Password.Compare(payload.Password); err != nil {
// 		app.unauthorizedError(w, r, err)
// 		return
// 	}

// 	claims := jwt.MapClaims{
// 		"sub": user.ID,
// 		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
// 		"iat": time.Now().Unix(),
// 		"nbf": time.Now().Unix(),
// 		"iss": app.config.auth.token.iss,
// 		"aud": app.config.auth.token.iss,
// 	}

// 	token, err := app.authenticator.GenerateToken(claims)
// 	if err != nil {
// 		app.internalServerError(w, r, err)
// 		return
// 	}

// if err := app.jsonResponse(w, http.StatusCreated, nil); err != nil {
// 	app.internalServerError(w, r, err)
// }
// }

// ActivateUser godoc
//
//	@Summary		Activates/Register a user
//	@Description	Activates/Register a user by invitation token
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	data := map[string]string{
		"message": "user activated",
	}

	if err := app.jsonResponse(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
	}
}
