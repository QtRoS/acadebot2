package searchengine

import (
	"../../shared"
	"../../shared/logu"
	"../../shared/netu"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const (
	UdacityApiUrl         = "https://www.udacity.com/public-api/v0/courses"
	SeachCangeDbName      = "search_cache"
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

var session *mgo.Session // https://godoc.org/gopkg.in/mgo.v2#Database.C

func init() {

	session, _ = mgo.Dial("localhost")
	//refreshCache() // TODO BUG

	ticker := time.NewTicker(time.Hour * 12)
	go func() {
		for _ = range ticker.C {
			refreshCache()
		}
	}()
}

func refreshCache() {
	logu.Info.Println("Gonna refresh cache")

	data, err0 := netu.MakeRequest(UdacityApiUrl, nil, nil)
	if err0 != nil {
		logu.Error.Println("err0", err0)
		return
	}

	response := udacityResponse{}
	err1 := parseJson(data, &response)
	if err1 != nil {
		logu.Error.Println("err1", err1)
		return
	}

	var infos = make([]shared.CourseInfo, len(response.Courses))
	for i, e := range response.Courses {
		info := shared.CourseInfo{Name: e.Title, Headline: e.ShortSummary, Link: e.Homepage, Art: e.Image}
		infos[i] = info
	}

	logu.Info.Println("New cache size", len(infos))

	coll := session.DB(SeachCangeDbName).C(UdacityCollectionName)
	coll.DropCollection()
	for _, i := range infos {
		err2 := coll.Insert(i)
		if err2 != nil {
			logu.Error.Println("err2", err2)
		}
	}

	logu.Info.Println("Saved to MongoDB")
}

func UdacityAdapter(query string, limit int) []shared.CourseInfo {
	var result = []shared.CourseInfo{}

	coll := session.DB(SeachCangeDbName).C(UdacityCollectionName)
	iter := coll.Find(bson.M{"name": bson.RegEx{".*" + query + ".*", "i"}}).Limit(limit).Iter()

	err := iter.All(&result)
	if err != nil {
		logu.Error.Println(err)
		return nil
	}

	logu.Info.Println("Results count", len(result))

	return result
}
