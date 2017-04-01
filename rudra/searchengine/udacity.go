package searchengine

import (
	"github.com/QtRoS/acadebot2/rudra/searchengine/mongoutils"
	"github.com/QtRoS/acadebot2/shared"
	"github.com/QtRoS/acadebot2/shared/logu"
	"github.com/QtRoS/acadebot2/shared/netu"
	"gopkg.in/mgo.v2/bson"
	"time"
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

func init() {
	refreshUdacityCache()
	ticker := time.NewTicker(time.Hour * 12)
	go func() {
		for _ = range ticker.C {
			refreshUdacityCache()
		}
	}()
}

func refreshUdacityCache() {
	logu.Info.Println("Gonna refresh cache", UdacityCollectionName)

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

	logu.Info.Println("New cache size", len(infos))

	session := mongoutils.MongoSession.Copy()
	defer session.Close()

	coll := session.DB(mongoutils.SeachCasheDbName).C(UdacityCollectionName)
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

	session := mongoutils.MongoSession.Copy()
	defer session.Close()

	coll := session.DB(mongoutils.SeachCasheDbName).C(UdacityCollectionName)
	iter := coll.Find(bson.M{"name": bson.RegEx{".*" + query + ".*", "i"}}).Limit(limit).Iter()

	err := iter.All(&result)
	if err != nil {
		logu.Error.Println(err)
		return nil
	}

	logu.Info.Println("Results count", len(result))

	return result
}
