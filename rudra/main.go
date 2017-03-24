package main

import (
	"fmt"
	// "github.com/gorilla/mux"
	"./searchengine"
	"net/http"
	"strconv"
)

const (
	Port = ":19191"
)

func serveCourses(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		limit = 10
	}
	searchResult := searchengine.Search(query, limit)
	fmt.Fprintln(w, searchResult)
}

func main() {
	http.HandleFunc("/courses", serveCourses)
	http.ListenAndServe(Port, nil)
}
