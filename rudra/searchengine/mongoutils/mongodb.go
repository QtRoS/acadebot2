package mongoutils

import (
	"github.com/QtRoS/acadebot2/shared/logu"
	"gopkg.in/mgo.v2"
)

const SeachCasheDbName = "search_cache"

var MongoSession *mgo.Session // https://godoc.org/gopkg.in/mgo.v2#Database.C

func init() {
	var err error
	MongoSession, err = mgo.Dial("localhost")
	if err != nil {
		logu.Error.Println("CreateSession: %s\n", err)
	} else {
		logu.Info.Println("Successfully opened MongoDB session")
	}
	// mongoSession.SetMode(mgo.Monotonic, true)
}
