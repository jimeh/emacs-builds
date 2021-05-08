package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func checkCmd() *cli.Command {
	return &cli.Command{
		Name:      "check",
		Usage:     "Check if GitHub release and asset exists",
		UsageText: "github-release [global options] check [options]",
		Action:    actionHandler(checkAction),
	}
}

func checkAction(c *cli.Context, opts *globalOptions) error {
	gh := opts.gh
	repo := opts.repo
	plan, err := LoadPlan(opts.plan)
	if err != nil {
		return err
	}

	fmt.Printf(
		"==> Checking github.com/%s for release: %s\n",
		repo.String(), plan.Release,
	)

	release, resp, err := gh.Repositories.GetReleaseByTag(
		c.Context, repo.Owner, repo.Name, plan.Release,
	)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("release %s does not exist", plan.Release)
		} else {
			return err
		}
	}

	fmt.Println("    -> Release exists")

	filename := filepath.Base(plan.Archive)

	fmt.Printf("==> Checking release for asset: %s\n", filename)
	for _, a := range release.Assets {
		if a.Name != nil && filename == *a.Name {
			fmt.Println("    -> Asset exists")
			return nil
		}
	}

	return fmt.Errorf("release does contain asset: %s", filename)
}
