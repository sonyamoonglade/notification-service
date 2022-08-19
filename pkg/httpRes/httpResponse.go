package httpRes

import "net/http"

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
