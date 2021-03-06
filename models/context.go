package models

import (
	"net/http"
	"code.google.com/p/gorilla/sessions"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
  "log"

  "kqnc/lib"
)

type Context struct {
  Database *mgo.Database
  Session  *sessions.Session
  User     *User
  Curator  *lib.Curator
}

func (c *Context) Close() {
	c.Database.Session.Close()
}

//C is a convenience function to return a collection from the context database.
func (c *Context) C(name string) *mgo.Collection {
	return c.Database.C(name)
}

func NewContext(req *http.Request, store sessions.Store, session *mgo.Session, database string, cur *lib.Curator) (*Context, error) {
  sess, err := store.Get(req, "gostbook")
  ctx := &Context{
      Database: session.Clone().DB(database),
      Session:  sess,
      Curator:  cur,
  }
  if err != nil {
      return ctx, err
  }

  //try to fill in the user from the session
  if uid, ok := sess.Values["user"].(bson.ObjectId); ok {
      err = ctx.C("users").Find(bson.M{"_id": uid}).One(&ctx.User)
  }

  return ctx, err
}

func (c *Context) GetDocuments() []Document {
  coll := c.C("documents")
  query := coll.Find(bson.M{}).Sort("-timestamp")

  var documents []Document
  if err := query.All(&documents); err != nil {
    log.Println(err)
    //return nil
  }
  return documents
}