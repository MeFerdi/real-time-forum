package api

import "net/http"

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "405 method not allowed.", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "frontend/index.html")
}
