package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type ErrorDecoder func(res *http.Response) error

func DefaultErrorDecoder(res *http.Response) error {
	defer func() {
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return errors.New(string(body))
}

func JSONErrorDecoder[T error](res *http.Response) error {
	var target T
	defer func() {
		_ = res.Body.Close()
	}()

	if err := json.NewDecoder(res.Body).Decode(&target); err != nil {
		return err
	}

	return target
}
