package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"

	"gitlab.dg.com/BackEnd/deliver/tif/zeus/config"
	mgoc "gitlab.dg.com/BackEnd/deliver/tif/zeus/mongo"
)

type user struct {
	ID   string `json:"id,omitempty" bson:"id,omitempty"`
	Name string `json:"name,omitempty" bson:"name,omitempty"`
}

func main() {
	appConf := &config.AppConf{
		MongoDB: config.MongoDB{
			Name:            "default",
			Host:            "127.0.0.1:27017",
			User:            "root",
			Pwd:             "123456",
			MaxPoolSize:     20,
			MaxConnIdleTime: 10,
		},
	}
	mgoc.InitDefalut(&appConf.MongoDB)
	defer mgoc.DefaultClientRelease()
	cli, err := mgoc.DefaultClient()
	if err != nil {
		log.Println("mgoc.DefaultClient err: ", err)
		return
	}

	db := cli.DB("example")
	coll := db.Collection("user")
	result := &user{}
	filter := bson.M{"name": "mark"}
	err = coll.FindOne(context.Background(), filter).Decode(result)
	if err != nil {
		log.Println("findOne err: ", err)
		return
	}
	log.Println(result)
}
