package zmongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mmqbaba/zeus/utils"
)

type Mongo interface {
	utils.Releaser
	DB(name string, opts ...*options.DatabaseOptions) *mongo.Database
}
