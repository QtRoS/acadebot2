package searchengine

import (
	"strings"

	"github.com/QtRoS/acadebot2/rudra/searchengine/fuzzy"
	"github.com/QtRoS/acadebot2/shared"
)

type fuzzyFilteringAdapter struct {
	sourceAdapter SourceAdapter
}

func newFuzzyFilteringAdapter(adapter SourceAdapter) *fuzzyFilteringAdapter {
	return &fuzzyFilteringAdapter{adapter}
}

func (me *fuzzyFilteringAdapter) Name() string {
	return me.sourceAdapter.Name() + " (Fuzzy)"
}

func (me *fuzzyFilteringAdapter) Get(query string, limit int) []shared.CourseInfo {
	courses := me.sourceAdapter.Get(query, limit)

	//logu.Info.Println(me.Name(), "before filter", len(courses))
	queryLower := strings.ToLower(query)
	infos := make([]shared.CourseInfo, 0, limit)
	for i := 0; i < len(courses) && len(infos) < limit; i++ {
		ci := &courses[i]
		if fuzzy.Match(queryLower, strings.ToLower(ci.Name)) ||
			fuzzy.Match(queryLower, strings.ToLower(ci.Headline)) {
			infos = append(infos, *ci)
		}
	}
	//logu.Info.Println(me.Name(), "after filter", len(infos))

	return infos
}
