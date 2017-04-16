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
	OpenLearningApiUrl         = "https://www.openlearning.com/api/courses/list?type=free,paid"
	OpenLearningCollectionName = "openlearning"
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

func init() {
	refreshOpenlearningCache()
	ticker := time.NewTicker(time.Hour * 12)
	go func() {
		for range ticker.C {
			refreshOpenlearningCache()
		}
	}()
}

func refreshOpenlearningCache() {
	logu.Info.Println("Gonna refresh cache", OpenLearningCollectionName)

	data, err0 := netu.MakeRequest(OpenLearningApiUrl, nil, nil)
	if err0 != nil {
		logu.Error.Println("err0", err0)
		return
	}

	response := openlearningResponse{}
	err1 := parseJson(data, &response)
	if err1 != nil {
		logu.Error.Println("err1", err1)
		return
	}

	var infos = make([]shared.CourseInfo, len(response.Courses))
	for i, e := range response.Courses {
		headline := e.Summary[:shared.Min(240, len(e.Summary))]
		info := shared.CourseInfo{Name: e.Name, Headline: headline, Link: e.CourseUrl, Art: e.Image}
		infos[i] = info
	}

	logu.Info.Println("New cache size", len(infos))

	session := mongoutils.MongoSession.Copy()
	defer session.Close()

	coll := session.DB(mongoutils.SearchCacheDbName).C(OpenLearningCollectionName)
	coll.DropCollection()
	for _, i := range infos {
		err2 := coll.Insert(i)
		if err2 != nil {
			logu.Error.Println("err2", err2)
		}
	}

	logu.Info.Println("Saved to MongoDB")
}

func OpenLearningAdapter(query string, limit int) []shared.CourseInfo {
	var result = []shared.CourseInfo{}

	session := mongoutils.MongoSession.Copy()
	defer session.Close()

	coll := session.DB(mongoutils.SearchCacheDbName).C(OpenLearningCollectionName)
	iter := coll.Find(bson.M{"name": bson.RegEx{".*" + query + ".*", "i"}}).Limit(limit).Iter()

	err := iter.All(&result)
	if err != nil {
		logu.Error.Println(err)
		return nil
	}

	logu.Info.Println("Results count", len(result))

	return result
}
