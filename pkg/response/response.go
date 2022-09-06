package response

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type JSON map[string]interface{}

func Ok(w http.ResponseWriter) {
	w.WriteHeader(200)
	return
}

func Created(w http.ResponseWriter) {
	w.WriteHeader(201)
	return
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(204)
	return
}

func Internal(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte("Internal error"))
	return
}

func Json(logger *zap.SugaredLogger, w http.ResponseWriter, code int, content JSON) {

	w.Header().Add("Content-Type", "application/json")
	bytes, err := json.Marshal(content)
	if err != nil {
		logger.Error(err.Error())
		Internal(w)
		return
	}
	w.WriteHeader(code)
	w.Write(bytes)
	return
}
