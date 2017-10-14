package searchengine

import (
	"strings"

	"github.com/QtRoS/acadebot2/shared"
	"github.com/QtRoS/acadebot2/shared/logu"
)

type filteringAdapter struct {
	sourceAdapter SourceAdapter
}

func NewFilteringAdapter(adapter SourceAdapter) *filteringAdapter {
	return &filteringAdapter{adapter}
}

func (me *filteringAdapter) Name() string {
	return me.sourceAdapter.Name() + " (Filtered)"
}

func (me *filteringAdapter) Get(query string, limit int) []shared.CourseInfo {
	courses := me.sourceAdapter.Get(query, limit)

	logu.Error.Println(me.Name(), "before filter", len(courses))

	infos := make([]shared.CourseInfo, 0, limit)
	for i := 0; i < len(courses) && len(infos) < limit; i++ {
		ci := &courses[i]
		if strings.Contains(ci.Name, query) || strings.Contains(ci.Headline, query) {
			infos = append(infos, *ci)
		}
	}

	logu.Error.Println(me.Name(), "after filter", len(infos))

	return infos
}
