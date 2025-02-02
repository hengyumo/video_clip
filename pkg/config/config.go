package config

import (
	"github.com/spf13/viper"
	"path/filepath"
)

type Config struct {
	VideoDir string
}

var AppConfig Config

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg.VideoDir = viper.GetString("video_dir")

	// 确保视频目录路径是绝对路径
	if !filepath.IsAbs(cfg.VideoDir) {
		absPath, err := filepath.Abs(cfg.VideoDir)
		if err != nil {
			return nil, err
		}
		cfg.VideoDir = absPath
	}

	return cfg, nil
}
