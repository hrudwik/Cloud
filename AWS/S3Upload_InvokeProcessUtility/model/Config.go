package model

import "time"

type Config struct {
	Service struct {
		Name  string        `yaml:"name"`
		Port  string        `yaml:"port"`
		Timer time.Duration `yaml:"timer"`
	} `yaml:"service"`
	LogFile struct {
		Location string `yaml:"location"`
		Name     string `yaml:"name"`
	} `yaml:"logfile"`
	ConfigFile struct {
		Location string `yaml:"location"`
		Name     string `yaml:"name"`
	} `yaml:"configfile"`
	Queue struct {
		Name string `yaml:"name"`
	} `yaml:"queue"`
	DestinationLocation struct {
		Path string `yaml:"path"`
	} `yaml:"destinationLocation"`
}
