package httpapi

import (
	"fmt"

	"example.com/it02-auth/backend/internal/domain"
)

// jsonFieldNames maps a domain field to the JSON key the client sent, so the
// form can attach each message to the input that produced it. The domain does
// not know these names, and should not.
var jsonFieldNames = map[domain.Field]string{
	domain.FieldUsername:        "username",
	domain.FieldPassword:        "password",
	domain.FieldConfirmPassword: "confirmPassword",
}

// fieldLabels are how each field is named in a sentence shown to the user.
var fieldLabels = map[domain.Field]string{
	domain.FieldUsername:        "ชื่อผู้ใช้งาน",
	domain.FieldPassword:        "รหัสผ่าน",
	domain.FieldConfirmPassword: "ยืนยันรหัสผ่าน",
}

// validationMessages turns domain violations into the wording a client shows.
// This is the only place the codes become sentences, which is what lets the
// domain stay free of user facing text.
func validationMessages(violations []domain.Violation) map[string][]string {
	messages := make(map[string][]string, len(violations))
	for _, violation := range violations {
		key, ok := jsonFieldNames[violation.Field]
		if !ok {
			key = string(violation.Field)
		}
		messages[key] = append(messages[key], violationMessage(violation))
	}
	return messages
}

func violationMessage(violation domain.Violation) string {
	label, ok := fieldLabels[violation.Field]
	if !ok {
		label = string(violation.Field)
	}

	switch violation.Code {
	case domain.ViolationRequired:
		return fmt.Sprintf("กรุณากรอก%s", label)
	case domain.ViolationMinLength:
		return fmt.Sprintf("%sต้องมีอย่างน้อย %d ตัวอักษร", label, violation.Limit)
	case domain.ViolationMaxLength:
		return fmt.Sprintf("%sต้องไม่เกิน %d ตัวอักษร", label, violation.Limit)
	case domain.ViolationInvalidFormat:
		return fmt.Sprintf("%sใช้ได้เฉพาะตัวอักษร ตัวเลข และ . _ - เท่านั้น", label)
	case domain.ViolationMismatch:
		return "รหัสผ่านและยืนยันรหัสผ่านไม่ตรงกัน"
	default:
		return fmt.Sprintf("%sไม่ถูกต้อง", label)
	}
}
