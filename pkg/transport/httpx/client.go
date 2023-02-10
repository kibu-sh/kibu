package httpx

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func NewJSONRequest[T any](httpMethod string, url string, object *T) (req *http.Request, err error) {
	payload := new(bytes.Buffer)
	err = json.NewEncoder(payload).Encode(object)
	if err != nil {
		return
	}
	return http.NewRequest(httpMethod, url, payload)
}
