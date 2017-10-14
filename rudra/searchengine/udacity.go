package searchengine

import (
	"github.com/QtRoS/acadebot2/shared"
	"github.com/QtRoS/acadebot2/shared/logu"
	"github.com/QtRoS/acadebot2/shared/netu"
)

const (
	UdacityApiUrl         = "https://www.udacity.com/public-api/v0/courses"
	UdacityCollectionName = "udacity"
)

type udacityResponse struct {
	Courses []udacityResult `json:"courses"`
}

type udacityResult struct {
	Key          string `json:"key"`
	Homepage     string `json:"homepage"`
	Title        string `json:"title"`
	ShortSummary string `json:"short_summary"`
	Image        string `json:"image"`
}

type udacityAdapter struct {
}

func (me *udacityAdapter) Name() string {
	return "Udacity"
}

func (me *udacityAdapter) Get(query string, limit int) []shared.CourseInfo {
	return UdacityAdapter(query, limit)
}

func UdacityAdapter(query string, limit int) []shared.CourseInfo {
	data, err0 := netu.MakeRequest(UdacityApiUrl, nil, nil)
	if err0 != nil {
		logu.Error.Println("err0", err0)
		return nil
	}

	response := udacityResponse{}
	err1 := parseJSON(data, &response)
	if err1 != nil {
		logu.Error.Println("err1", err1)
		return nil
	}

	uniqueSet := make(map[string]bool)
	var infos = make([]shared.CourseInfo, 0, len(response.Courses))
	for _, e := range response.Courses {
		// Check uniqueness.
		if uniqueSet[e.Homepage] {
			logu.Warning.Println("Result dublicate:", e.Homepage)
			continue
		} else {
			uniqueSet[e.Homepage] = true
		}

		info := shared.CourseInfo{Name: e.Title, Headline: e.ShortSummary, Link: e.Homepage, Art: e.Image}
		infos = append(infos, info)
	}

	logu.Info.Println("Results count", len(infos))

	return infos
}
