package main 

import (
	"kqnc/models"
	"kqnc/controllers"
	"thegoods.biz/httpbuf"
	"code.google.com/p/gorilla/pat"
	"code.google.com/p/gorilla/sessions"
	"encoding/gob"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
	"log"
)

var store sessions.Store
var session *mgo.Session
var database string
var router *pat.Router

type handler func(http.ResponseWriter, *http.Request, *models.Context) error

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  //create the context
  ctx, err := models.NewContext(r, store, session, database)
  if err != nil {
  	  //http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  defer ctx.Close()

  //run the handler and grab the error, and report it
  buf := new(httpbuf.Buffer)
  err = h(buf, r, ctx)
  if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
  }

  //save the session
  if err = ctx.Session.Save(r, buf); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
  }

  //apply the buffered response to the writer
  buf.Apply(w)
}

func init() {
  gob.Register(bson.ObjectId(""))
}

func main() {
	var err error
	session, err = mgo.Dial(os.Getenv("127.0.0.1"))
	if err != nil {
		panic(err)
	}
	
	database = session.DB("").Name
	if err := session.DB("").C("users").EnsureIndex(mgo.Index{
        Key:    []string{"email"},
        Unique: true,
    }); err != nil {
        log.Println("Ensuring unqiue index on users:", err)
    }

	store = sessions.NewCookieStore([]byte(os.Getenv("kqnc")))

	router = pat.New()
	controllers.Init(router)

	router.Add("GET", "/login", handler(controllers.LoginForm)).Name("login")
	router.Add("POST", "/login", handler(controllers.Login))
	router.Add("GET", "/logout", handler(controllers.Logout)).Name("logout")
	router.Add("GET", "/register", handler(controllers.RegisterForm)).Name("register")
	router.Add("POST", "/register", handler(controllers.Register))

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	router.Add("GET", "/", handler(controllers.Index)).Name("index")

	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}