package config

type Config struct {
	Debug bool
}

func Init() *Config {
	return &Config{
		Debug: false,
	}
}
