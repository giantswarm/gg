package config

import (
	"io/ioutil"
	"os/user"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

func Colour(colour bool) bool {
	v := struct {
		Colour *bool `yaml:"colour"`
	}{}

	fromFile(&v)

	if v.Colour == nil {
		return colour
	}

	return *v.Colour
}

func Group(group string) string {
	v := struct {
		Group *string `yaml:"group"`
	}{}

	fromFile(&v)

	if v.Group == nil {
		return group
	}

	return *v.Group
}

func Time(time string) string {
	v := struct {
		Time *string `yaml:"time"`
	}{}

	fromFile(&v)

	if v.Time == nil {
		return time
	}

	return *v.Time
}

func fromFile(v interface{}) {
	b, err := ioutil.ReadFile(name())
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(b, v)
	if err != nil {
		panic(err)
	}
}

// name returns the config file name as absolute path according to the current
// OS User known to the running process.
//
//     /Users/xh3b4sd/.config/gg/config.yaml
//
func name() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	return filepath.Join(u.HomeDir, ".config/gg/config.yaml")
}
