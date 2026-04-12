package apperror

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("not found")
)

type AppError struct {
	Err              error  `json:"-"`
	Message          string `json:"message"`
	DeveloperMessage string `json:"developer_message"`
	Code             string `json:"code"`
}

func New(err error, code string, message string, developerMessage string) *AppError {
	if message != "" && err == nil {
		err = errors.New(message)
	}

	return &AppError{
		Err:              err,
		Code:             code,
		Message:          message,
		DeveloperMessage: developerMessage,
	}
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *AppError) Marshal() []byte {
	data, _ := json.Marshal(e)
	return data
}

func BadRequestError(message string) error {
	if message == "" {
		message = "bad request"
	}

	return New(errors.New(message), "NS-40000", message, message)
}

func SystemError(err error) error {
	if err == nil {
		err = errors.New("system error")
	}

	return New(err, "NS-50000", "system error", err.Error())
}

func NotFoundError(message string) error {
	if message == "" {
		message = "not found"
	}

	return New(ErrNotFound, "NS-40400", message, message)
}
