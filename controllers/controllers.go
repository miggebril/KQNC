package controllers

import (
	"kqnc/models"
  	"net/http"
  	"code.google.com/p/gorilla/pat"
  	"html/template"
  	"path/filepath"
  	"labix.org/v2/mgo/bson"
  	"sync"
  	"fmt"
)

var cachedTemplates = map[string]*template.Template{}
var cachedMutex sync.Mutex

var router *pat.Router

func Init(r *pat.Router) {
	router = r
}

func reverse(name string, things ...interface{}) string {
	//convert the things to strings
	strs := make([]string, len(things))
	for i, th := range things {
		strs[i] = fmt.Sprint(th)
	}
	//grab the route
	u, err := router.GetRoute(name).URL(strs...)
	if err != nil {
		panic(err)
	}
	return u.Path
}

var funcs = template.FuncMap{
	"reverse": reverse,
}

func T(name string, useBase bool) *template.Template {
	cachedMutex.Lock()
	defer cachedMutex.Unlock()

	if t, ok := cachedTemplates[name]; ok {
		return t
	}

	var base string
	if useBase {
		base = "_base.html"
	} else {
		base = "content"
	}

	t := template.Must(template.New(base).Funcs(funcs).ParseGlob("templates/partials/*"))
	t = template.Must(t.ParseFiles(
		"templates/_base.html",
		filepath.Join("templates", name),
	))

	//cachedTemplates[name] = t

	return t
}

func LoginForm(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	return T("login.html", false).Execute(w, map[string]interface{}{
		"ctx": ctx,
	})
}

func Login(w http.ResponseWriter, r *http.Request, ctx *models.Context) error {
	email, password := r.FormValue("email"), r.FormValue("password")

	user, e := models.Login(ctx, email, password)
	if e != nil {
		ctx.Session.AddFlash("Invalid Email/Password")
		return LoginForm(w, r, ctx)
	}

	//store the user id in the values and redirect to index
	ctx.Session.Values["user"] = user.ID
	http.Redirect(w, r, reverse("index"), http.StatusSeeOther)
	return nil
}

func Logout(w http.ResponseWriter, r *http.Request, ctx *models.Context) error {
	delete(ctx.Session.Values, "user")
	http.Redirect(w, r, reverse("index"), http.StatusSeeOther)
	return nil
}

func RegisterForm(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	return T("register.html", false).Execute(w, map[string]interface{}{
		"ctx": ctx,
		"email": r.FormValue("email"),
		"first": r.FormValue("first"),
		"last": r.FormValue("last"),
	})
}

func Register(w http.ResponseWriter, r *http.Request, ctx *models.Context) error {
	email, password, password_confirm := r.FormValue("email"), r.FormValue("password"), r.FormValue("password_confirm")
	first, last := r.FormValue("first"), r.FormValue("last")

	if len(password) < 8 {
		ctx.Session.AddFlash("Password must be at least 8 characters long.", "danger")
		return RegisterForm(w, r, ctx)
	}

	if password != password_confirm {
		ctx.Session.AddFlash("Password confirmation does not match.", "danger")
		return RegisterForm(w, r, ctx)
	}

	u := &models.User{
		Email: 	  email,
		FirstName: first,
		LastName: last,
		ID:       bson.NewObjectId(),
	}
	u.SetPassword(password)

	if err := ctx.C("users").Insert(u); err != nil {
		ctx.Session.AddFlash("Please choose a different username.", "danger")
		r.Form.Del("email")
		return RegisterForm(w, r, ctx)
	}

	//store the user id in the values and redirect to index
	ctx.Session.Values["user"] = u.ID
	http.Redirect(w, r, reverse("index"), http.StatusSeeOther)
	return nil
}

func Index(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	//execute the template
	if ctx.User == nil || ctx.Session.Values["user"] == "" {
		return T("welcome.html", false).Execute(w, map[string]interface{}{
			"ctx": ctx,
		})
	}

	return T("index.html", true).Execute(w, map[string]interface{}{
		"documents": ctx.GetDocuments(),
		"ctx":     ctx,
	})
}