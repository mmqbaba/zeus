package config

type Tomler struct {
	original []byte // 配置的原始数据
	conf *AppConf
}

func (t *Tomler) Init(original []byte) error {
	t.original = original
	t.conf = nil
	return nil
}

func (t *Tomler) Get() *AppConf {
	return t.conf
}