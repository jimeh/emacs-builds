package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v35/github"
	"github.com/urfave/cli/v2"
)

func publishCmd() *cli.Command {
	return &cli.Command{
		Name:      "publish",
		Usage:     "publish a release",
		UsageText: "github-release [global-options] publish [options]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "release-sha",
				Aliases: []string{"s"},
				Usage:   "Git SHA of repo to create release on",
				EnvVars: []string{"GITHUB_SHA"},
			},
			&cli.BoolFlag{
				Name:    "prerelease",
				Usage:   "Git SHA of repo to create release on",
				EnvVars: []string{"RELEASE_PRERELEASE"},
				Value:   true,
			},
		},
		Action: actionHandler(publishAction),
	}
}

func publishAction(c *cli.Context, opts *globalOptions) error {
	gh := opts.gh
	repo := opts.repo
	plan, err := LoadPlan(opts.plan)
	if err != nil {
		return err
	}

	releaseSHA := c.String("release-sha")
	prerelease := c.Bool("prerelease")

	assetFile, err := os.Open(plan.Archive)
	if err != nil {
		return err
	}
	assetInfo, err := assetFile.Stat()
	if err != nil {
		return err
	}

	fmt.Printf("==> Checking release %s\n", plan.Release)

	release, resp, err := gh.Repositories.GetReleaseByTag(
		c.Context, repo.Owner, repo.Name, plan.Release,
	)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			fmt.Println("    -> Release not found, creating...")
			release, _, err = gh.Repositories.CreateRelease(
				c.Context, repo.Owner, repo.Name, &github.RepositoryRelease{
					Name:            &plan.Release,
					TagName:         &plan.Release,
					TargetCommitish: &releaseSHA,
					Prerelease:      &prerelease,
				},
			)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if release.GetPrerelease() != prerelease {
		release.Prerelease = &prerelease

		release, _, err = gh.Repositories.EditRelease(
			c.Context, repo.Owner, repo.Name, release.GetID(), release,
		)
		if err != nil {
			return err
		}
	}

	assetFilename := plan.ReleaseAsset()
	assetExists := false

	fmt.Printf("==> Checking asset %s\n", assetFilename)

	for _, a := range release.Assets {
		if a.GetName() != assetFilename {
			continue
		}

		if a.GetSize() == int(assetInfo.Size()) {
			fmt.Println("    -> Asset already exists")
			assetExists = true
		} else {
			fmt.Println("    -> Asset exists with wrong file size, deleting...")
			_, err := gh.Repositories.DeleteReleaseAsset(
				c.Context, repo.Owner, repo.Name, a.GetID(),
			)
			if err != nil {
				return err
			}
			fmt.Println("       -> Done")
		}

	}

	if !assetExists {
		fmt.Println("    -> Asset missing, uploading...")
		_, _, err = gh.Repositories.UploadReleaseAsset(
			c.Context, repo.Owner, repo.Name, release.GetID(),
			&github.UploadOptions{Name: assetFilename},
			assetFile,
		)
		if err != nil {
			return err
		}
		fmt.Println("       -> Done")
	}

	fmt.Printf("==> Release available at: %s\n", release.GetHTMLURL())

	return nil
}
