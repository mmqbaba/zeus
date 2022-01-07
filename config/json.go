package config

import (
	"errors"
	"log"
	"time"

	"github.com/mmqbaba/zeus/utils"
)

type Jsoner struct {
	original []byte // 配置的原始数据
	conf     *AppConf
}

func (j *Jsoner) Init(original []byte) (err error) {
	if utils.IsEmptyString(string(original)) {
		msg := "Jsoner.Init failed, 配置原始数据不能为空"
		log.Println(msg)
		err = errors.New(msg)
		return
	}
	var conf AppConf
	if err = utils.Unmarshal([]byte(original), &conf); err != nil {
		log.Println(err)
		return
	}
	j.original = original
	j.conf = &conf
	j.conf.UpdateTime = time.Now()
	return nil
}

func (j *Jsoner) Get() *AppConf {
	return j.conf
}
