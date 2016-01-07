package helpers

import (
	"log"
	"os"
	"labix.org/v2/mgo/bson"
	"encoding/base64"
)

var logger = log.New(os.Stderr, "app: ", log.LstdFlags) //| log.Llongfile)

func CheckErr(err error, msg string) {
	if err != nil {
		logger.Println(msg, err)
	}
}

func ObjectIdFromString(encodedid string) (bson.ObjectId, error) {
	data, err := base64.URLEncoding.DecodeString(encodedid)
	if err != nil {
		log.Println("Error decoding object ID:", err)
		return bson.NewObjectId(), err
	}
	return bson.ObjectId(data), err
}