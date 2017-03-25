package searchengine

import (
	"../../shared"
	"../../shared/logu"
	"../../shared/netu"
	"./mongoutils"
	"gopkg.in/mgo.v2/bson"
	"time"
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

func init() {
	refreshIversityCache()
	ticker := time.NewTicker(time.Hour * 12)
	go func() {
		for _ = range ticker.C {
			refreshIversityCache()
		}
	}()
}

func refreshIversityCache() {
	logu.Info.Println("Gonna refresh cache", IversityCollectionName)

	data, err0 := netu.MakeRequest(IversityApiUrl, nil, nil)
	if err0 != nil {
		logu.Error.Println("err0", err0)
		return
	}

	response := iversityResponse{}
	err1 := parseJson(data, &response)
	if err1 != nil {
		logu.Error.Println("err1", err1)
		return
	}

	var infos = make([]shared.CourseInfo, len(response.Courses))
	for i, e := range response.Courses {
		info := shared.CourseInfo{Name: e.Title, Headline: e.Subtitle, Link: e.Url, Art: e.Image}
		infos[i] = info
	}

	logu.Info.Println("New cache size", len(infos))

	session := mongoutils.MongoSession.Copy()
	defer session.Close()

	coll := session.DB(mongoutils.SeachCasheDbName).C(IversityCollectionName)
	coll.DropCollection()
	for _, i := range infos {
		err2 := coll.Insert(i)
		if err2 != nil {
			logu.Error.Println("err2", err2)
		}
	}

	logu.Info.Println("Saved to MongoDB")
}

func IversityAdapter(query string, limit int) []shared.CourseInfo {
	var result = []shared.CourseInfo{}

	session := mongoutils.MongoSession.Copy()
	defer session.Close()

	coll := session.DB(mongoutils.SeachCasheDbName).C(IversityCollectionName)
	iter := coll.Find(bson.M{"name": bson.RegEx{".*" + query + ".*", "i"}}).Limit(limit).Iter()

	err := iter.All(&result)
	if err != nil {
		logu.Error.Println(err)
		return nil
	}

	logu.Info.Println("Results count", len(result))

	return result
}
