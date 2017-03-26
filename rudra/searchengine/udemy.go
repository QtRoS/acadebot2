package searchengine

import (
	"github.com/QtRoS/acadebot2/shared"
	"github.com/QtRoS/acadebot2/shared/logu"
	"github.com/QtRoS/acadebot2/shared/netu"
	"strconv"
)

const (
	UdemyApiUrl  = "https://www.udemy.com/api-2.0/courses"
	AuthHeader   = "Basic MlloUmZ1TXpUSjJLMjJmZWZoSldTeVoyanVtOWx0dkdoWFhFUWZQaTpiNGRIUXhmUDdsODVWa3RHQlM4dUFpdU5ZclpyOEZWY3E3cFpTaWRXbVNMSTBuNm5mWGFyRUxSQ2xqdEtDbjZPcTR3ZkZwWjlqM0RsdU13aUhDN0UxVW1zS1YyQzRtSUlvR2ZEYXpNYVhtbDZjRGtHcjJmOHVqVzVkQ2J5VThaaw=="
	UdemyBaseUrl = "https://www.udemy.com"
)

type udemyResponse struct {
	Results []udemyResult `json:"results"`
}

type udemyResult struct {
	Id       int    `json:"id"`
	Url      string `json:"url"`
	Title    string `json:"title"`
	Headline string `json:"headline"`
	Image    string `json:"image_480x270"`
}

func UdemyAdapter(query string, limit int) []shared.CourseInfo {

	data, err0 := netu.MakeRequest(UdemyApiUrl,
		map[string]string{"search": query, "page_size": strconv.Itoa(limit), "ordering": "trending", "fields[course]": "@default,headline"},
		map[string]string{"Authorization": AuthHeader})

	if err0 != nil {
		logu.Error.Println("err0", err0)
		return nil
	}

	response := udemyResponse{}
	err1 := parseJson(data, &response)
	if err1 != nil {
		logu.Error.Println("err1", err1)
		return nil
	}

	logu.Info.Println("Results count", len(response.Results))

	var infos = make([]shared.CourseInfo, 0, limit)
	for _, e := range response.Results {
		link := UdemyBaseUrl + e.Url
		info := shared.CourseInfo{Name: e.Title, Headline: e.Headline, Link: link, Art: e.Image}
		// infos[i] = info
		infos = append(infos, info)
	}

	return infos
}
