package mutcask

type Config struct {
	Path    string
	CaskNum uint32
}

func defaultConfig() *Config {
	return &Config{
		CaskNum: 256,
	}
}

type Option func(cfg *Config)

func ConfCaskNum(caskNum int) Option {
	return func(cfg *Config) {
		cfg.CaskNum = uint32(caskNum)
	}
}

func ConfPath(dir string) Option {
	return func(cfg *Config) {
		cfg.Path = dir
	}
}
