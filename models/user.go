package models

import (
	"labix.org/v2/mgo/bson"
	"code.google.com/p/go.crypto/bcrypt"
)

type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Email string
	FirstName string
	LastName string
	Password []byte
}

//SetPassword takes a plaintext password and hashes it with bcrypt and sets the
//password field to the hash.
func (u *User) SetPassword(password string) {
	hpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err) //this is a panic because bcrypt errors on invalid costs
	}
	u.Password = hpass
}

//Login validates and returns a user object if they exist in the database.
func Login(ctx *Context, email, password string) (u *User, err error) {
	err = ctx.C("users").Find(bson.M{"email": email}).One(&u)
	if err != nil {
		return
	}

	err = bcrypt.CompareHashAndPassword(u.Password, []byte(password))
	if err != nil {
		u = nil
	}
	return
}