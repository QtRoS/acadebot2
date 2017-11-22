package searchengine

import (
	"strings"

	"github.com/QtRoS/acadebot2/shared"
	"github.com/QtRoS/acadebot2/shared/logu"
	"github.com/QtRoS/acadebot2/shared/netu"
)

const (
	openLearningAPIURL = "https://www.openlearning.com/api/courses/list?type=free,paid"
)

type openlearningResponse struct {
	Courses []openlearningResult `json:"courses"`
}

type openlearningResult struct {
	CourseUrl string `json:"courseUrl"`
	Image     string `json:"image"`
	Name      string `json:"name"`
	Summary   string `json:"summary"`
}

type openlearningAdapter struct {
}

func (me *openlearningAdapter) Name() string {
	return "OpenLearning"
}

func (me *openlearningAdapter) Get(query string, limit int) []shared.CourseInfo {
	data, err0 := netu.MakeRequest(openLearningAPIURL, nil, nil)
	if err0 != nil {
		logu.Error.Println("err0", err0)
		return nil
	}

	response := openlearningResponse{}
	err1 := parseJSON(data, &response)
	if err1 != nil {
		logu.Error.Println("err1", err1)
		return nil
	}

	logu.Info.Println("Results count", len(response.Courses))

	var infos = make([]shared.CourseInfo, len(response.Courses))
	for i, e := range response.Courses {
		headline := strings.Split(e.Summary, "\n")[0] //e.Summary[:shared.Min(240, len(e.Summary))]
		info := shared.CourseInfo{Name: e.Name, Headline: headline, Link: e.CourseUrl, Art: e.Image}
		infos[i] = info
	}

	return infos
}
