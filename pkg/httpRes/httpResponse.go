package httpRes

import "net/http"

func NoSubscribers(w http.ResponseWriter) {
	msg := "no subscribers for event to fire"
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(200)
	w.Write([]byte(msg))
	return
}

func Ok(w http.ResponseWriter) {
	w.WriteHeader(200)
	return
}

func Created(w http.ResponseWriter) {
	w.WriteHeader(201)
	return
}
