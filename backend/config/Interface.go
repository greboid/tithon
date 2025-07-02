package config

import "github.com/csmith/config"

type Provider interface {
	Load(target interface{}) error
	Save(target interface{}) error
}

type DefaultConfigProvider struct {
	instance *config.Config
}

func NewDefaultConfigProvider() (Provider, error) {
	conf, err := config.New(config.DirectoryName(GetConfigDirName()), config.FileName(GetConfigFilename()))
	if err != nil {
		return nil, err
	}
	return &DefaultConfigProvider{
		instance: conf,
	}, nil
}

func (p *DefaultConfigProvider) Load(target interface{}) error {
	return p.instance.Load(target)
}

func (p *DefaultConfigProvider) Save(target interface{}) error {
	return p.instance.Save(target)
}
