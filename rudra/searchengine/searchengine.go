package searchengine

import (
	"../../shared"
	"../../shared/logu"
	"encoding/json"
	"sync"
)

const (
	EmptyResult = "[]"
)

// type CourseInfo struct {
// 	Name     string `json:"name"`
// 	Headline string `json:"headline"`
// 	Link     string `json:"link"`
// 	Art      string `json:"art"`
// }

type InfoSourceAdapter func(query string, limit int) []shared.CourseInfo

var adapters []InfoSourceAdapter

func init() {
	adapters = []InfoSourceAdapter{
		CourseraAdapter,
		UdacityAdapter,
		UdemyAdapter,
	}
}

func Search(query string, perSourceLimit int) string {
	if query == "" || perSourceLimit <= 0 {
		return EmptyResult
	}

	logu.Info.Println("Gonna search for:", query)

	mergedResults := callAdapters(query, perSourceLimit)
	json, error := toJson(mergedResults)
	if error != nil {
		logu.Error.Println(error)
		return EmptyResult
	}
	// logu.Trace.Println(string(json))
	return string(json)
}

func callAdapters(query string, perSourceLimit int) []shared.CourseInfo {
	// var results []CourseInfo
	results := make([]shared.CourseInfo, 0, perSourceLimit)

	adaptersChunks := make(chan []shared.CourseInfo)
	var wg sync.WaitGroup
	wg.Add(len(adapters) + 1)

	for _, adapter := range adapters {
		go func(adapt InfoSourceAdapter) {
			defer wg.Done()
			adaptersChunks <- adapt(query, perSourceLimit)
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

func toJson(infos []shared.CourseInfo) ([]byte, error) {
	return json.Marshal(infos)
}

func parseJson(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}
