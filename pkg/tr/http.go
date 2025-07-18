package tr

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

func RespondOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}

func RespondError(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func RedirectError(w http.ResponseWriter, location *url.URL, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	q := location.Query()
	q.Set("status", strconv.Itoa(status))
	q.Set("error", message)
	location.RawQuery = q.Encode()

	w.Header().Set("Location", location.String())
	w.WriteHeader(http.StatusPermanentRedirect)
}

