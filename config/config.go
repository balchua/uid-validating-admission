package config

import (
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//Configuration object for the pod-watcher application
// Applications holder
type Configuration struct {
	Kind       string `mapstructure:"kind"`
	APIVersion string `mapstructure:"apiVersion"`
	Spec       Spec   `mapstructure:"spec"`
}

type Spec struct {
	IgnoreOnFailure bool                `mapstructure:"ignoreOnFailure"`
	Excluded        []ExcludeNamespaces `mapstructure:"excludeNamespaces"`
	Included        []IncludeNamespaces `mapstructure:"includeNamespaces"`
}

type ExcludeNamespaces struct {
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
}

type IncludeNamespaces struct {
	Name string `mapstructure:"name"`
	Uids []Uids `mapstructure:"uids"`
}

type Uids struct {
	Uid int64 `mapstructure:"uid"`
}

func GetAppConfig(podPolicyFile string) Configuration {
	viper.SetConfigName("pod-uid-policy")

	viper.AddConfigPath(podPolicyFile)
	var appConfig Configuration

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(&appConfig)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	for _, excluded := range appConfig.Spec.Excluded {
		logrus.Infof("Excluded namespace: %s", excluded.Name)
	}

	return appConfig
}
