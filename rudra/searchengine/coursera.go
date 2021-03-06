package searchengine

import (
	"strconv"
	"strings"

	"github.com/QtRoS/acadebot2/shared"
	"github.com/QtRoS/acadebot2/shared/logu"
	"github.com/QtRoS/acadebot2/shared/netu"
)

const (
	courseraApiUrl  = "https://api.coursera.org/api/courses.v1"
	courseraBaseUrl = "http://www.coursera.org/learn/"
)

type courseraResponse struct {
	Elements []courseraElement `json:"elements"`
}

type courseraElement struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PhotoUrl    string `json:"photoUrl"`
	Link        string `json:"link"`
}

type courseraAdapter struct {
}

func (me *courseraAdapter) Name() string {
	return "Coursera"
}

func (me *courseraAdapter) Get(query string, limit int) []shared.CourseInfo {
	data, err0 := netu.MakeRequest(courseraApiUrl,
		map[string]string{"q": "search", "fields": "description,photoUrl", "query": query, "limit": strconv.Itoa(limit)}, nil)

	if err0 != nil {
		logu.Error.Println(err0)
		return nil
	}

	response := courseraResponse{}
	err1 := parseJSON(data, &response)
	if err1 != nil {
		logu.Error.Println(err1)
		return nil
	}

	logu.Info.Println("Results count", len(response.Elements))

	var infos = make([]shared.CourseInfo, 0, limit)
	for _, e := range response.Elements {
		link := courseraBaseUrl + e.Slug
		desc := strings.Split(e.Description, "\n")[0] //e.Description[:shared.Min(240, len(e.Description))]
		info := shared.CourseInfo{Name: e.Name, Headline: desc, Link: link, Art: e.PhotoUrl}
		infos = append(infos, info)
	}

	return infos
}
