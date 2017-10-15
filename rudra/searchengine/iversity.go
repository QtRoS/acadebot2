package searchengine

import (
	"github.com/QtRoS/acadebot2/shared"
	"github.com/QtRoS/acadebot2/shared/logu"
	"github.com/QtRoS/acadebot2/shared/netu"
)

const (
	iversityAPIURL = "https://iversity.org/api/v1/courses"
)

type iversityResponse struct {
	Courses []iversityResult `json:"courses"`
}

type iversityResult struct {
	URL      string `json:"url"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Image    string `json:"image"`
}

type iversityAdapter struct {
}

func (me *iversityAdapter) Name() string {
	return "Iversity"
}

func (me *iversityAdapter) Get(query string, limit int) []shared.CourseInfo {
	data, err0 := netu.MakeRequest(iversityAPIURL, nil, nil)
	if err0 != nil {
		logu.Error.Println("err0", err0)
		return nil
	}

	response := iversityResponse{}
	err1 := parseJSON(data, &response)
	if err1 != nil {
		logu.Error.Println("err1", err1)
		return nil
	}

	logu.Info.Println("Results count", len(response.Courses))

	var infos = make([]shared.CourseInfo, len(response.Courses))
	for i, e := range response.Courses {
		info := shared.CourseInfo{Name: e.Title, Headline: e.Subtitle, Link: e.URL, Art: e.Image}
		infos[i] = info
	}

	return infos
}
