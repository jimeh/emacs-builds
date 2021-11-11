package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/google/go-github/github"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

const downloadIcon = `` +
	`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 14 14">` +
	`<defs>` +
	`<style>` +
	`.a{fill:none;stroke:#fff;stroke-linecap:round;stroke-miterlimit:10;}` +
	`</style>` +
	`</defs>` +
	`<line class="a" x1="2.5" y1="12.65" x2="11.5" y2="12.65"/>` +
	`<line class="a" x1="7" y1="1" x2="7" y2="9"/>` +
	`<line class="a" x1="11" y1="5.25" x2="7" y2="9.25"/>` +
	`<line class="a" x1="3" y1="5.25" x2="7" y2="9.25"/>` +
	`</svg>`

type Badge struct {
	SchemaVersion int    `json:"schemaVersion,omitempty"`
	Label         string `json:"label,omitempty"`
	Message       string `json:"message,omitempty"`
	Color         string `json:"color,omitempty"`
	LabelColor    string `json:"labelColor,omitempty"`
	IsError       bool   `json:"isError,omitempty"`
	NamedLogo     string `json:"namedLogo,omitempty"`
	LogoSVG       string `json:"logoSvg,omitempty"`
	LogoColor     string `json:"logoColor,omitempty"`
	LogoWidth     string `json:"logoWidth,omitempty"`
	LogoPosition  string `json:"logoPosition,omitempty"`
	Style         string `json:"style,omitempty"`
	CacheSeconds  int    `json:"cacheSeconds,omitempty"`
}

func NewApp() *cli.App {
	return &cli.App{
		Name: "meta-updater",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "repository",
				Aliases: []string{"repo", "r"},
				EnvVars: []string{"GITHUB_REPOSITORY"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "badges",
				Aliases: []string{"badge"},
				Subcommands: []*cli.Command{
					{
						Name: "downloads",
						Flags: []cli.Flag{
							&cli.StringSliceFlag{
								Name:  "exclude",
								Usage: "regexp asset filename pattern to exclude",
								Value: cli.NewStringSlice(`.+\.sha\d+$`),
							},
							&cli.StringFlag{
								Name:    "output",
								Aliases: []string{"o"},
							},
						},
						Action: badgesDownloadsAction,
					},
				},
			},
		},
	}
}

func NewGH(ctx context.Context) *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return github.NewClient(nil)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func badgesDownloadsAction(c *cli.Context) error {
	gh := NewGH(c.Context)

	var excludes []*regexp.Regexp
	for _, s := range c.StringSlice("exclude") {
		excludes = append(excludes, regexp.MustCompile(s))
	}

	parts := strings.SplitN(c.String("repository"), "/", 2)
	owner := parts[0]
	repo := parts[1]

	count := 0

	for page := 1; page > 0; {
		releases, resp, err := gh.Repositories.ListReleases(
			c.Context, owner, repo, &github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		)
		if err != nil {
			return err
		}

		count += releaseAssetsDownlaodCount(releases, excludes)
		page = resp.NextPage
	}

	var humanCount string
	if count >= 1000 {
		v, sym := humanize.ComputeSI(float64(count))
		humanCount = humanize.FormatFloat("#,###.#", v) + sym
	} else {
		humanCount = strconv.Itoa(count)
	}

	badge := &Badge{
		SchemaVersion: 1,
		Style:         "flat",
		Color:         "blue",
		LogoSVG:       downloadIcon,
		Label:         "total downloads",
		Message:       humanCount,
	}

	b, err := jsonMarshal(badge)
	if err != nil {
		return err
	}

	if filename := c.String("output"); filename != "" {
		dir := filepath.Dir(filename)
		if dir != "" && dir != "." {
			err := os.MkdirAll(dir, 0o755)
			if err != nil {
				return err
			}
		}
		err := ioutil.WriteFile(filename, b, 0o644) //nolint:gosec
		if err != nil {
			return err
		}
	} else {
		fmt.Print(string(b))
	}

	return nil
}

func releaseAssetsDownlaodCount(
	releases []*github.RepositoryRelease,
	excludes []*regexp.Regexp,
) int {
	count := 0

	for _, release := range releases {
		for _, asset := range release.Assets {
			for _, exclude := range excludes {
				if exclude.MatchString(asset.GetName()) {
					continue
				}
			}

			if v := asset.GetDownloadCount(); v > 0 {
				count += v
			}
		}
	}

	return count
}

func jsonMarshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func main() {
	app := NewApp()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
