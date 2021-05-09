package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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
				Value:   false,
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

	assetBaseName := filepath.Base(plan.Archive)
	assetSumFile := plan.Archive + ".sha256"

	if _, err := os.Stat(assetSumFile); os.IsNotExist(err) {
		fmt.Printf("==> Generating SHA256 sum for %s\n", assetBaseName)
		assetSum, err := fileSHA256(plan.Archive)
		if err != nil {
			return err
		}

		content := fmt.Sprintf("%s  %s", assetSum, assetBaseName)
		err = os.WriteFile(assetSumFile, []byte(content), 0666)
		if err != nil {
			return err
		}

		fmt.Printf("    -> Done: %s\n", assetSum)
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

	assetFiles := []string{plan.Archive, assetSumFile}

	for _, fileName := range assetFiles {
		fileIO, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer fileIO.Close()

		fileInfo, err := fileIO.Stat()
		if err != nil {
			return err
		}

		fileBaseName := filepath.Base(fileName)
		assetExists := false

		fmt.Printf("==> Checking asset %s\n", fileBaseName)

		for _, a := range release.Assets {
			if a.GetName() != fileBaseName {
				continue
			}

			if a.GetSize() == int(fileInfo.Size()) {
				fmt.Println("    -> Asset already exists")
				assetExists = true
			} else {
				fmt.Println(
					"    -> Asset exists with wrong file size, deleting...",
				)
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
				&github.UploadOptions{Name: fileBaseName},
				fileIO,
			)
			if err != nil {
				return err
			}
			fmt.Println("       -> Done")
		}

	}

	fmt.Printf("==> Release available at: %s\n", release.GetHTMLURL())

	return nil
}

func fileSHA256(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
