package searchengine

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/QtRoS/acadebot2/shared"
	"github.com/QtRoS/acadebot2/shared/logu"
)

const (
	emptyResult = "[]"
)

type SourceAdapter interface {
	Get(query string, limit int) []shared.CourseInfo
	Name() string
}

var adapters = []SourceAdapter{
	&courseraAdapter{},
	newFuzzyFilteringAdapter(newCachingAdapter(&udacityAdapter{}, time.Hour*6)),
	&udemyAdapter{},
	newFuzzyFilteringAdapter(newCachingAdapter(&openlearningAdapter{}, time.Hour*6)),
	newFuzzyFilteringAdapter(newCachingAdapter(&iversityAdapter{}, time.Hour*6)),
}

// Search for courses in all services.
func Search(query string, perSourceLimit int) string {
	if query == "" || perSourceLimit <= 0 {
		return emptyResult
	}

	logu.Info.Println("Gonna search for:", query, " <------------------- ")

	mergedResults := callAdapters(query, perSourceLimit)
	jsonData, err := toJSON(mergedResults)
	if err != nil {
		logu.Error.Println(err)
		return emptyResult
	}
	// logu.Trace.Println(string(json))
	return string(jsonData)
}

func callAdapters(query string, perSourceLimit int) []shared.CourseInfo {
	results := make([]shared.CourseInfo, 0, perSourceLimit)

	adaptersChunks := make(chan []shared.CourseInfo)
	var wg sync.WaitGroup
	wg.Add(len(adapters) + 1)

	logu.Info.Println("Before calling adapters...")

	for _, adapter := range adapters {
		go func(adapt SourceAdapter) {
			defer wg.Done()
			adaptersChunks <- adapt.Get(query, perSourceLimit)
		}(adapter)
	}

	go func() {
		defer wg.Done()
		for i := 0; i < len(adapters); i++ {
			chunk := <-adaptersChunks
			results = append(results, chunk...)
		}
	}()

	wg.Wait()
	logu.Info.Println("Merged result len:", len(results))

	return results
}

// func callAdapters(query string, perSourceLimit int) []CourseInfo {
// 	var results []CourseInfo
// 	for _, adapter := range adapters {
// 		courses := adapter(query, perSourceLimit)
// 		results = append(results, courses...)
// 	}

// 	return results
// }

func toJSON(infos []shared.CourseInfo) ([]byte, error) {
	return json.Marshal(infos)
}

func parseJSON(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}
