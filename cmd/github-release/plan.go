package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ReleaseType string

type Release struct {
	Name  string `yaml:"name"`
	Title string `yaml:"title,omitempty"`
	Draft bool   `yaml:"draft,omitempty"`
	Pre   bool   `yaml:"prerelease,omitempty"`
}

type Plan struct {
	Commit  *Commit  `yaml:"commit"`
	OS      *OSInfo  `yaml:"os"`
	Release *Release `yaml:"release"`
	Archive string   `yaml:"archive"`
}

func LoadPlan(filename string) (*Plan, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	plan := &Plan{}
	err = yaml.Unmarshal(b, plan)

	return plan, err
}
