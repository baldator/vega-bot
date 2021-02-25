package main

import "github.com/ilyakaznacheev/cleanenv"

type ConfigVars struct {
	SocialServiceURL      string `yaml:"SocialServiceURL" env:"SOCIALSERVICEURL" env-default:"127.0.0.1"`
	SocialTwitterEnabled  bool   `yaml:"SocialTwitterEnabled" env:"TWITTER-ENABLED" env-default:"false"`
	SocialTelegramEnabled bool   `yaml:"SocialTelegramEnabled" env:"TELEGRAM-ENABLED" env-default:"false"`
	SocialDiscordEnabled  bool   `yaml:"SocialDiscordEnabled" env:"DISCORD-ENABLED" env-default:"false"`
	SocialSlackEnabled    bool   `yaml:"SocialSlackEnabled" env:"SLACK-ENABLE" env-default:"false"`
	SocialServiceKey      string `yaml:"SocialServiceKey" env:"SOCIALSERVICEKEY" env-default:""`
	SocialServiceSecret string `yaml:"SocialServiceSecret" env:"SocialServiceSecret" env-default:""`
	GrpcNodeURL           string `yaml:"GrpcNodeUrl" env:"GRPCNODEURL" env-default:"n06.testnet.vega.xyz:3002"`
}

// ReadConfig import config struct from yaml file
func ReadConfig(path string) (ConfigVars, error) {
	var cfg ConfigVars
	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
