package zmysql

import (
	"github.com/jinzhu/gorm"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
)

type Mysql interface {
	Reload(cfg *config.Mysql)
	GetCli() *gorm.DB
	ZFind(out interface{}, where ...interface{}) *gorm.DB
	ZCreate(value interface{}) *gorm.DB
	ZUpdate(attrs ...interface{}) *gorm.DB
	ZDelete(value interface{}, where ...interface{}) *gorm.DB
}
