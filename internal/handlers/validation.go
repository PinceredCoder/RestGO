package handlers

import (
	"strings"

	"github.com/PinceredCoder/RestGo/internal/errors"
)

func (h *TaskHandler) convertValidationError(err error) *errors.APIError {
	errorMsg := err.Error()
	lines := strings.Split(errorMsg, "\n")

	var details []errors.ValidationErrorDetail

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "invalid ") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fieldPart := strings.TrimPrefix(parts[0], "invalid ")
				fieldParts := strings.Split(fieldPart, ".")

				fieldName := fieldPart
				if len(fieldParts) > 1 {
					fieldName = fieldParts[len(fieldParts)-1]
				}

				message := strings.TrimSpace(parts[1])

				details = append(details, errors.ValidationErrorDetail{
					Field:   fieldName,
					Message: message,
				})
			}
		}
	}

	if len(details) == 0 {
		return errors.NewValidationError("Validation failed", map[string]string{
			"error": errorMsg,
		})
	}

	return errors.NewValidationError("Validation failed", details)
}
