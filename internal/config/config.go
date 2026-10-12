package config

import (
	"log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

type BtConfig struct {
	Padding         int     `mapstructure:"padding" validate:"gte=1,lte=256"`
	FilePreview     bool    `mapstructure:"file_preview"`
	HighlightIndent bool    `mapstructure:"highlight_indent"`
	InPlaceRender   bool    `mapstructure:"in_place_render"`
	ShowHidden 		bool    `mapstructure:"show_hidden"`
	PreviewRatio 	float64 `mapstructure:"preview_ratio" validate:"gte=0.1,lte=0.9"`
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

	err = validateConfig(&config)
	if err != nil {
		log.Fatalf("Invalid configuration values provided, %v", err)
	}

	return config
}

func validateConfig(config *BtConfig) error {
	validate = validator.New()
	return validate.Struct(config)
}
