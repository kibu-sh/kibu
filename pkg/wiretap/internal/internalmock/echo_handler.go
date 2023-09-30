package internalmock

import (
	"encoding/json"
	"net/http"
	"time"
)

const (
	ScenarioHeader     = "x-scenario-override"
	ScenarioDisconnect = "disconnect"
	ScenarioTimeout    = "timeout"
)

// EchoHandler is a simple http handler that echos the request body back to the client
// if the x-scenario-override header is provided it will trigger a specific scenario
func EchoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scenario := r.Header.Get(ScenarioHeader)
		switch scenario {
		case ScenarioDisconnect:
			hj := w.(http.Hijacker)
			conn, _, _ := hj.Hijack()
			_ = conn.Close()
		case ScenarioTimeout:
			time.Sleep(time.Second * 10)
		default:
			JsonHandler(func(req *http.Request) (any, error) {
				return map[string]string{
					"message": "hello world",
				}, nil
			})(w, r)
		}
	}
}

func JsonHandler[T any](handler func(r *http.Request) (T, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, err := handler(r)
		if err != nil {
			panic(err)
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}

		_, err = w.Write(bodyBytes)
		if err != nil {
			panic(err)
		}
	}
}
