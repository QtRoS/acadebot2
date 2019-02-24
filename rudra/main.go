package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/QtRoS/acadebot2/rudra/searchengine"
	"github.com/QtRoS/acadebot2/shared/logu"
)

const (
	port         = ":19191"
	defaultLimit = 10
)

func serveCourses(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		limit = defaultLimit
	}
	searchResult := searchengine.Search(query, limit)
	fmt.Fprintln(w, searchResult)
}

func main() {
	http.HandleFunc("/courses", serveCourses)
	if err := http.ListenAndServe(port, nil); err != nil {
		logu.Error.Fatal(err)
	}
}
