package main

import (
	"fmt"
	"time"

	"github.com/google/go-github/v35/github"
)

type Commit struct {
	Repo      *Repo      `yaml:"repo"`
	Ref       string     `yaml:"ref"`
	SHA       string     `yaml:"sha"`
	Date      *time.Time `yaml:"date"`
	Author    string     `yaml:"author"`
	Committer string     `yaml:"committer"`
	Message   string     `yaml:"message"`
}

func NewCommit(repo *Repo, ref string, rc *github.RepositoryCommit) *Commit {
	return &Commit{
		Repo: repo,
		Ref:  ref,
		SHA:  rc.GetSHA(),
		Date: rc.GetCommit().GetCommitter().Date,
		Author: fmt.Sprintf(
			"%s <%s>",
			rc.GetCommit().GetAuthor().GetName(),
			rc.GetCommit().GetAuthor().GetEmail(),
		),
		Committer: fmt.Sprintf(
			"%s <%s>",
			rc.GetCommit().GetCommitter().GetName(),
			rc.GetCommit().GetCommitter().GetEmail(),
		),
		Message: rc.GetCommit().GetMessage(),
	}
}

func (s *Commit) ShortSHA() string {
	return s.SHA[0:7]
}

func (s *Commit) DateString() string {
	return s.Date.Format("2006-01-02")
}
