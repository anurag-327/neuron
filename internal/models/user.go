package models

import (
	"context"
	"fmt"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuthProvider string
type RoleType string

const (
	AuthProviderGoogle AuthProvider = "google"
	AuthProviderGithub AuthProvider = "github"
	AuthProviderEmail  AuthProvider = "email"
	RoleTypeUser       RoleType     = "user"
	RoleTypeAdmin      RoleType     = "Admin"

	DefaultSignupCredits int64 = 100
)

type User struct {
	mgm.DefaultModel `bson:",inline"`
	Email            string   `bson:"email" json:"email"`
	Password         string   `bson:"password,omitempty" json:"-"`
	Username         string   `bson:"username" json:"username"`
	Name             string   `bson:"name" json:"name"`
	ImageUrl         *string  `bson:"imageUrl" json:"imageUrl"`
	Role             RoleType `bson:"role,omitempty" json:"role" default:"user"`
	Verified         bool     `bson:"verified" json:"verified" default:"false"`
	AuthProvider     string   `bson:"authProvider,omitempty" json:"authProvider,omitempty"`

	Credits int64 `bson:"credits" json:"credits"`
}

func (u *User) CollectionName() string {
	return "users"
}

func CreateUserIndexes() error {
	coll := mgm.Coll(&User{})
	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
		},
		Options: options.Index().SetUnique(true).SetName("unique_email"),
	}
	_, err := coll.Indexes().CreateOne(context.Background(), indexModel)
	fmt.Println("User indexes created successfully")
	return err
}
