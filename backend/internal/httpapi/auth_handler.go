package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"

	"example.com/it02-auth/backend/internal/domain"
	"example.com/it02-auth/backend/internal/service"
)

// maxRequestBodyBytes bounds how much a client can send, so a hostile body
// cannot exhaust memory.
const maxRequestBodyBytes = 1 << 20

type authHandler struct {
	auth   *service.AuthService
	logger *slog.Logger
}

func newAuthHandler(auth *service.AuthService, logger *slog.Logger) *authHandler {
	return &authHandler{auth: auth, logger: logger}
}

// register handles POST /api/auth/register, the IT 02-2 form.
func (h *authHandler) register(writer http.ResponseWriter, request *http.Request) {
	var payload registerRequest
	if !h.decode(writer, request, &payload) {
		return
	}

	registration, err := domain.NewRegistration(payload.Username, payload.Password, payload.ConfirmPassword)
	if err != nil {
		h.fail(writer, request, "register", err)
		return
	}

	user, err := h.auth.Register(request.Context(), registration)
	if err != nil {
		h.fail(writer, request, "register", err)
		return
	}

	writeJSON(writer, h.logger, http.StatusCreated, toUserResponse(user))
}

// login handles POST /api/auth/login, the IT 02-1 form.
func (h *authHandler) login(writer http.ResponseWriter, request *http.Request) {
	var payload loginRequest
	if !h.decode(writer, request, &payload) {
		return
	}

	attempt, err := domain.NewLoginAttempt(payload.Username, payload.Password)
	if err != nil {
		h.fail(writer, request, "login", err)
		return
	}

	session, err := h.auth.Login(request.Context(), attempt)
	if err != nil {
		h.fail(writer, request, "login", err)
		return
	}
	writeJSON(writer, h.logger, http.StatusOK, toSessionResponse(session))
}

// me handles GET /api/auth/me, which is what IT 02-3 calls to prove the token
// is still valid and to learn whose name to greet.
func (h *authHandler) me(writer http.ResponseWriter, request *http.Request) {
	claims, ok := claimsFrom(request.Context())
	if !ok {
		// Unreachable while the route is wrapped in authenticate; handled
		// rather than assumed, so a future routing mistake cannot turn into
		// an anonymous read.
		unauthenticated(writer, h.logger, "Bearer")
		return
	}

	user, err := h.auth.CurrentUser(request.Context(), claims.UserID)
	if err != nil {
		h.fail(writer, request, "current user", err)
		return
	}
	writeJSON(writer, h.logger, http.StatusOK, toUserResponse(user))
}

// decode reads a JSON body into target, answering the client itself and
// returning false when the body cannot be used.
func (h *authHandler) decode(writer http.ResponseWriter, request *http.Request, target any) bool {
	if contentType := request.Header.Get("Content-Type"); contentType != "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || mediaType != "application/json" {
			writeMessage(writer, h.logger, http.StatusUnsupportedMediaType, messageMalformed)
			return false
		}
	}

	request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(request.Body)
	// An unknown field is usually a client sending the wrong shape, and
	// silently ignoring it hides the mistake until it matters.
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		writeMessage(writer, h.logger, http.StatusBadRequest, messageMalformed)
		return false
	}
	// A valid body contains exactly one JSON value followed by EOF.
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeMessage(writer, h.logger, http.StatusBadRequest, messageMalformed)
		return false
	}
	return true
}

// fail is the single place where an error from an inner layer becomes a status
// code, so every route reports the same failure the same way.
func (h *authHandler) fail(
	writer http.ResponseWriter,
	request *http.Request,
	operation string,
	err error,
) {
	var validation *domain.ValidationError

	switch {
	case errors.As(err, &validation):
		writeJSON(writer, h.logger, http.StatusBadRequest, validationBody{
			Title:  messageValidationEN,
			Status: http.StatusBadRequest,
			Errors: validationMessages(validation.Violations),
		})

	case errors.Is(err, domain.ErrUsernameTaken):
		// 409 rather than 400: the request is well formed, it conflicts with
		// the current state of the server.
		writeMessage(writer, h.logger, http.StatusConflict, messageUsernameTaken)

	case errors.Is(err, domain.ErrInvalidCredentials):
		writeMessage(writer, h.logger, http.StatusUnauthorized, messageSignInFailed)

	case errors.Is(err, domain.ErrUserNotFound):
		// Only reachable from /me, where the token is valid but its subject no
		// longer exists. That is a dead session, not a missing page, so the
		// client is told to sign in again.
		unauthenticated(writer, h.logger, `Bearer error="invalid_token"`)

	default:
		// Only unexpected failures are logged; the cases above are ordinary
		// client mistakes and would be noise.
		h.logger.ErrorContext(request.Context(), operation+" failed", "error", err)
		writeMessage(writer, h.logger, http.StatusInternalServerError, messageUnexpected)
	}
}
