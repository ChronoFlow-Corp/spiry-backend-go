package pkg

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/dto/response"
)

func RespondError(w http.ResponseWriter, code int, message response.Error) {
	raw, err := json.Marshal(message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(raw)
}

func RedirectError(w http.ResponseWriter, location *url.URL, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	q := location.Query()
	q.Set("code", strconv.Itoa(status))
	q.Set("error", message)
	location.RawQuery = q.Encode()

	w.Header().Set("Location", location.String())
	w.WriteHeader(http.StatusPermanentRedirect)
}
