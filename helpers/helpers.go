package helpers

import (
	"log"
	"os"
	"labix.org/v2/mgo/bson"
	"encoding/base64"
	"math/rand"
	"strconv"
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

func RandomSmallPercent() string {
	return strconv.FormatFloat(rand.Float64()*10, 'f', 2, 64)
}

func RandomBigPercent() string {
	return strconv.FormatFloat(rand.Float64()*100, 'f', 2, 64)
}

func RandomScore() string {
	return strconv.FormatInt(int64(rand.Intn(200)+100), 10)
}