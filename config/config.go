package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var service = map[string]interface{} {
	"host": 	"127.0.0.1",
	"port":		9100,
}

var log = map[string]interface{}{
	"level": 	logrus.InfoLevel,
	"json":     	true,
	"requests": 	true,
}

var metadata = map[string]interface{}{
	"host": 	"http://169.254.169.254",
	"interval": 	30000,
	"version":	"latest",
}

var tags = map[string]interface{}{
	"interval":	300000,
}

var propsd = map[string]interface{}{
	"conqueso": "http://localhost:9301/conqueso",
	"properties": "http://localhost:9301/properties",
}

// Defaults generates a set of default configuration options
func Defaults() {
	viper.SetDefault("service", service)
	viper.SetDefault("propsd", propsd)
	viper.SetDefault("log", log)
	viper.SetDefault("metadata", metadata)
	viper.SetDefault("tags", tags)
}
