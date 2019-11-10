package config

type Yamler struct {
	original []byte // 配置的原始数据
	conf     *AppConf
}

func (t *Yamler) Init(original []byte) error {
	t.original = original
	t.conf = nil
	return nil
}

func (t *Yamler) Get() *AppConf {
	return t.conf
}
