package httpapi

import (
	"time"

	"example.com/it02-auth/backend/internal/domain"
	"example.com/it02-auth/backend/internal/service"
)

// registerRequest is the IT 02-2 form. The confirmation is checked on the
// server as well as in the browser: a client side check is a convenience, not
// a guarantee, because anyone can call this endpoint directly.
type registerRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

// loginRequest is the IT 02-1 form.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// userResponse is how an account is described to a client. There is no field
// for the password or its hash, so neither can be exposed by accident.
type userResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"createdAt"`
}

// sessionResponse is the answer to a successful sign in.
type sessionResponse struct {
	AccessToken string       `json:"accessToken"`
	TokenType   string       `json:"tokenType"`
	ExpiresAt   string       `json:"expiresAt"`
	User        userResponse `json:"user"`
}

func toUserResponse(user domain.User) userResponse {
	return userResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toSessionResponse(session service.Session) sessionResponse {
	return sessionResponse{
		AccessToken: session.Token.Value,
		// Named so the client knows to send "Authorization: Bearer <token>"
		// without that format being hard coded in two places.
		TokenType: "Bearer",
		ExpiresAt: session.Token.ExpiresAt.UTC().Format(time.RFC3339),
		User:      toUserResponse(session.User),
	}
}
