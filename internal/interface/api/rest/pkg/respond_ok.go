package pkg

import (
	"encoding/json"
	"net/http"
)

func RespondOK(w http.ResponseWriter, data interface{}) {
	raw, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}
