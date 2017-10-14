package searchengine

import (
	"github.com/QtRoS/acadebot2/shared"
	"github.com/QtRoS/acadebot2/shared/logu"
	"github.com/QtRoS/acadebot2/shared/netu"
)

const (
	IversityApiUrl         = "https://iversity.org/api/v1/courses"
	IversityCollectionName = "iversity"
)

type iversityResponse struct {
	Courses []iversityResult `json:"courses"`
}

type iversityResult struct {
	Url      string `json:"url"`
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
	return IversityAdapter(query, limit)
}

func IversityAdapter(query string, limit int) []shared.CourseInfo {
	data, err0 := netu.MakeRequest(IversityApiUrl, nil, nil)
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
		info := shared.CourseInfo{Name: e.Title, Headline: e.Subtitle, Link: e.Url, Art: e.Image}
		infos[i] = info
	}

	return infos
}
