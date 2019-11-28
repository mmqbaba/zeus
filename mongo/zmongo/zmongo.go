package zmongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo interface {
	DB(name string, opts ...*options.DatabaseOptions) *mongo.Database
}
