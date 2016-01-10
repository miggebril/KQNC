package controllers

import (
	"net/http"
	"kqnc/models"
	"kqnc/helpers"
	"github.com/gorilla/schema"
	"labix.org/v2/mgo/bson"
	"log"
	"kqnc/lib/store"
)

func NewDocumentForm(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	return T("documents/new.html", true).Execute(w, map[string]interface{}{
		"ctx":     ctx,
	})
}

func NewDocument(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	err = r.ParseForm()
	helpers.CheckErr(err, "Could not parse form")

	document := models.Document{}
	document.ID = bson.NewObjectId()
	document.User = ctx.User.ID

	decoder := schema.NewDecoder()
	err = decoder.Decode(&document, r.PostForm)
	helpers.CheckErr(err, "Failed to decode form.")
	log.Println(document)

	if err := ctx.C("documents").Insert(document); err != nil {
		ctx.Session.AddFlash("Problem creating new document.", "danger")
		helpers.CheckErr(err, "Failed to create new document.")
		return NewDocumentForm(w, r, ctx)
	}
	//doc, _ := store.NewDocument("hello")
	//log.Println("Creating doc with ID:", document.GetIDEncoded())
	if _, err := ctx.Curator.CreateDocument(ctx.User.GetIDEncoded(), document.GetIDEncoded(), store.Document{document.GetIDEncoded(), "lol"}); err != nil {
		ctx.Session.AddFlash("Problem creating new document.", "danger")
		helpers.CheckErr(err, "Failed to create new document.")
		return NewDocumentForm(w, r, ctx)
	}

	ctx.Session.AddFlash("Document created successfully.", "success")
	http.Redirect(w, r, "/documents/"+document.GetIDEncoded(), http.StatusFound)
	return nil
}

func DocumentForm(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	id, err := helpers.ObjectIdFromString(r.URL.Query().Get(":id"))

	if err != nil {
		log.Println(err)
		return nil
	}

	coll := ctx.C("documents")
	query := coll.Find(bson.M{"_id":id}).Sort("-timestamp")

	var document models.Document
	if err = query.One(&document); err != nil {
		log.Println(err)
		return nil
	}

	return T("documents/view.html", true).Execute(w, map[string]interface{}{
		"ctx":     ctx,
		"document": document,
	})
}

func DocumentIndexForm(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	return T("documents/index.html", true).Execute(w, map[string]interface{}{
		"ctx":     ctx,
	//	"documents": ctx.GetCampaigns(),
	})
}
