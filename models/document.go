package models

import (
	"labix.org/v2/mgo/bson"
	"encoding/base64"
)

type Document struct {
	ID bson.ObjectId `bson:"_id,omitempty" col:"campaigns"`
	Name string
	User bson.ObjectId
	Content string
}

func (d Document) GetIDEncoded() string {
	return base64.URLEncoding.EncodeToString([]byte(d.ID))
}
