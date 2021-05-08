package main

import (
	"fmt"
	"strings"
)

type Repo struct {
	URL   string `yaml:"url"`
	Owner string `yaml:"owner"`
	Name  string `yaml:"name"`
}

func NewRepo(ownerAndRepo string) *Repo {
	parts := strings.SplitN(ownerAndRepo, "/", 2)

	return &Repo{
		URL:   fmt.Sprintf("https://github.com/%s/%s", parts[0], parts[1]),
		Owner: parts[0],
		Name:  parts[1],
	}
}

func (s *Repo) String() string {
	return s.Owner + "/" + s.Name
}
