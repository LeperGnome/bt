package config

import (
	"log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type BtConfig struct {
	Padding         int  `mapstructure:"padding"`
	FilePreview     bool `mapstructure:"file_preview"`
	HighlightIndent bool `mapstructure:"highlight_indent"`
	InPlaceRender   bool `mapstructure:"in_place_render"`
}

func GetConfig(flags *pflag.FlagSet) BtConfig {
	vp := viper.New()

	vp.BindPFlags(flags)

	vp.SetConfigName("conf")
	vp.SetConfigType("yaml")
	vp.AddConfigPath("$HOME/.config/bt")

	if err := vp.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("Error retreiving config file: %v", err)
		}
	}

	var config BtConfig
	err := vp.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	return config
}
