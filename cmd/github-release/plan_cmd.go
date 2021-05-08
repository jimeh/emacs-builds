package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var nonAlphaNum = regexp.MustCompile(`[^\w-_]+`)

func planCmd() *cli.Command {
	wd, err := os.Getwd()
	if err != nil {
		wd = ""
	}

	return &cli.Command{
		Name:      "plan",
		Usage:     "Plan if GitHub release and asset exists",
		UsageText: "github-release [global options] plan [<branch/tag>]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "emacs-mirror-repo",
				Usage:   "Github owner/repo to get Emacs commit info from",
				Aliases: []string{"e"},
				EnvVars: []string{"EMACS_MIRROR_REPO"},
				Value:   "emacs-mirror/emacs",
			},
			&cli.StringFlag{
				Name:  "work-dir",
				Usage: "Github owner/repo to get Emacs commit info from",
				Value: wd,
			},
			&cli.StringFlag{
				Name:  "sha",
				Usage: "Override commit SHA of specified git branch/tag",
			},
		},
		Action: actionHandler(planAction),
	}
}

func planAction(c *cli.Context, opts *globalOptions) error {
	gh := opts.gh
	planFile := opts.plan
	repo := NewRepo(c.String("emacs-mirror-repo"))
	buildsDir := filepath.Join(c.String("work-dir"), "builds")

	ref := c.Args().Get(0)
	if ref == "" {
		ref = "master"
	}

	lookupRef := ref
	if s := c.String("sha"); s != "" {
		lookupRef = s
	}

	repoCommit, _, err := gh.Repositories.GetCommit(
		c.Context, repo.Owner, repo.Name, lookupRef,
	)
	if err != nil {
		return err
	}

	rb, _ := yaml.Marshal(repoCommit)
	fmt.Printf("commit:\n---\n%s\n", string(rb))

	commit := NewCommit(repo, ref, repoCommit)
	osInfo, err := NewOSInfo()
	if err != nil {
		return err
	}

	cleanRef := sanitizeString(ref)
	cleanOS := sanitizeString(osInfo.Name + "-" + osInfo.ShortVersion())
	cleanArch := sanitizeString(osInfo.Arch)

	releaseName := fmt.Sprintf(
		"Emacs.%s.%s.%s",
		commit.DateString(), commit.ShortSHA(), cleanRef,
	)
	archiveName := fmt.Sprintf(
		"Emacs.%s.%s.%s.%s.%s.tbz",
		commit.DateString(), commit.ShortSHA(), cleanRef, cleanOS, cleanArch,
	)

	plan := &Plan{
		Commit:  commit,
		OS:      osInfo,
		Release: releaseName,
		Archive: filepath.Join(buildsDir, archiveName),
	}

	buf := bytes.Buffer{}
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	err = enc.Encode(plan)
	if err != nil {
		return err
	}

	return os.WriteFile(planFile, buf.Bytes(), 0666)
}

func sanitizeString(s string) string {
	return nonAlphaNum.ReplaceAllString(s, "-")
}
