package searchengine

import (
	"strconv"

	"github.com/QtRoS/acadebot2/shared"
	"github.com/QtRoS/acadebot2/shared/logu"
	"github.com/QtRoS/acadebot2/shared/netu"
)

const (
	udemyAPIURL  = "https://www.udemy.com/api-2.0/courses"
	authHeader   = "Basic MlloUmZ1TXpUSjJLMjJmZWZoSldTeVoyanVtOWx0dkdoWFhFUWZQaTpiNGRIUXhmUDdsODVWa3RHQlM4dUFpdU5ZclpyOEZWY3E3cFpTaWRXbVNMSTBuNm5mWGFyRUxSQ2xqdEtDbjZPcTR3ZkZwWjlqM0RsdU13aUhDN0UxVW1zS1YyQzRtSUlvR2ZEYXpNYVhtbDZjRGtHcjJmOHVqVzVkQ2J5VThaaw=="
	udemyBaseUrl = "https://www.udemy.com"
)

type udemyResponse struct {
	Results []udemyResult `json:"results"`
}

type udemyResult struct {
	ID       int    `json:"id"`
	URL      string `json:"url"`
	Title    string `json:"title"`
	Headline string `json:"headline"`
	Image    string `json:"image_480x270"`
}

type udemyAdapter struct {
}

func (me *udemyAdapter) Name() string {
	return "Udemy"
}

func (me *udemyAdapter) Get(query string, limit int) []shared.CourseInfo {
	data, err0 := netu.MakeRequest(udemyAPIURL,
		map[string]string{"search": query, "page_size": strconv.Itoa(limit), "ordering": "trending", "fields[course]": "@default,headline"},
		map[string]string{"Authorization": authHeader})

	if err0 != nil {
		logu.Error.Println("err0", err0)
		return nil
	}

	response := udemyResponse{}
	err1 := parseJSON(data, &response)
	if err1 != nil {
		logu.Error.Println("err1", err1)
		return nil
	}

	logu.Info.Println("Results count", len(response.Results))

	var infos = make([]shared.CourseInfo, 0, limit)
	for _, e := range response.Results {
		link := udemyBaseUrl + e.URL
		info := shared.CourseInfo{Name: e.Title, Headline: e.Headline, Link: link, Art: e.Image}
		infos = append(infos, info)
	}

	return infos
}
