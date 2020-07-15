package zmysql

import (
	"github.com/jinzhu/gorm"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
)

type Mysql interface {
	Reload(cfg *config.Mysql)
	GetCli() *gorm.DB
}
