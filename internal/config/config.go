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

	vp.SetConfigName("btconfig")
	vp.SetConfigType("yaml")
	vp.AddConfigPath("$HOME/.config/bt")
	vp.AddConfigPath(".")

	err := vp.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
		panic(err)
	}

	var config BtConfig
	// Unmarshal the config file into the AppConfig struct
	err = vp.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
		panic(err)
	}

	return config
}
