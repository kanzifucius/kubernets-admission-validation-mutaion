package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type validationConfig struct {
	RequiredAnnotations []string
	RequiredLabels      []string
	RequiredImageTags   []string
}

type ingress struct {
	OldSuffix    string
	NewSuffix    string
	MutationType string
	Enabled      bool
}

type Mutation struct {
	Ingress ingress
}

type HookConfig struct {
	Validation validationConfig
	Mutation   Mutation
}

func LoadConfig(configmapLocation string, configName string) HookConfig {

	viper.SetConfigType("yml")
	viper.SetConfigName(configName)
	viper.AddConfigPath(configmapLocation)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})

	values := values()
	//fmt.Printf("%+v\n", values )
	return values
}

func values() HookConfig {
	return HookConfig{
		Validation: validationConfig{
			RequiredAnnotations: viper.GetStringSlice("validation.requiredAnnotations"),
			RequiredLabels:      viper.GetStringSlice("validation.requiredLabels"),
			RequiredImageTags:   viper.GetStringSlice("validation.requiredImageTags"),
		},
		Mutation: Mutation{
			Ingress: ingress{
				OldSuffix:    viper.GetString("mutation.ingress.oldSuffix"),
				NewSuffix:    viper.GetString("mutation.ingress.newSuffix"),
				MutationType: viper.GetString("mutation.ingress.mutationType"),
				Enabled:      viper.GetBool("mutation.ingress.enabled"),
			},
		},
	}
}
