// Package httpapi exposes the application use cases over HTTP. It owns the
// wire format, the status codes and the wording the client sees; it holds no
// business rules of its own.
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

const contentTypeJSON = "application/json; charset=utf-8"

// User facing wording lives here so every handler answers consistently.
const (
	messageUnexpected      = "ไม่สามารถดำเนินการได้ กรุณาลองใหม่"
	messageMalformed       = "รูปแบบข้อมูลไม่ถูกต้อง"
	messageUsernameTaken   = "ชื่อผู้ใช้งานนี้ถูกใช้แล้ว"
	messageSignInFailed    = "ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง"
	messageUnauthenticated = "กรุณาลงชื่อเข้าใช้งาน"
	messageValidationEN    = "One or more validation errors occurred."
)

// messageBody is the shape of every single line error response.
type messageBody struct {
	Message string `json:"message"`
}

// validationBody is the problem details shape, with one list of messages per
// offending field so a form can mark each input.
type validationBody struct {
	Title  string              `json:"title"`
	Status int                 `json:"status"`
	Errors map[string][]string `json:"errors"`
}

func writeJSON(writer http.ResponseWriter, logger *slog.Logger, status int, payload any) {
	writer.Header().Set("Content-Type", contentTypeJSON)
	// Authentication responses carry a token; no cache, shared or private,
	// should keep a copy of one.
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)

	if err := json.NewEncoder(writer).Encode(payload); err != nil {
		// The status line is already on the wire, so all that is left is to
		// make the truncated response visible in the logs.
		logger.Error("encode response failed", "error", err)
	}
}

func writeMessage(writer http.ResponseWriter, logger *slog.Logger, status int, message string) {
	writeJSON(writer, logger, status, messageBody{Message: message})
}
