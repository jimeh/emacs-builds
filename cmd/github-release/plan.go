package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Plan struct {
	Commit  *Commit `yaml:"commit"`
	OS      *OSInfo `yaml:"os"`
	Release string  `yaml:"release"`
	Archive string  `yaml:"archive"`
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
