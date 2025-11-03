package config

type Config struct {
	Modules Modules
}

type Modules struct {
	Path string
	Mirrorlib string
}
