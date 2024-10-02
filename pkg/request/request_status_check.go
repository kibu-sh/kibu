package request

import (
	"errors"
	"fmt"
	"net/http"
)

type StatusCheckFunc func(status string, code int) error

func NewExactStatusCheckFunc(expectedCode int) StatusCheckFunc {
	return func(status string, code int) error {
		if code == expectedCode {
			return nil
		}
		return errors.Join(
			ErrStatusCheckFailed,
			fmt.Errorf("expected status code %d, got %d", expectedCode, code),
		)
	}
}

func NewStatusRangeCheckFunc(min, max int) StatusCheckFunc {
	return func(status string, code int) error {
		if code >= min && code < max {
			return nil
		}
		return errors.Join(
			ErrStatusCheckFailed,
			fmt.Errorf("expected status code in range %d-%d, got %d", min, max, code),
		)
	}
}

func NewOkayRangeCheckFunc() StatusCheckFunc {
	return NewStatusRangeCheckFunc(
		http.StatusOK,
		http.StatusMultipleChoices,
	)
}

func NewBasicOkayCheckFunc() StatusCheckFunc {
	return NewExactStatusCheckFunc(http.StatusOK)
}
