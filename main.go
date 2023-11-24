package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
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

type Cache struct {
	AseetDownloads map[string]map[string]int `json:"asset_downloads,omitempty"`
}

func (c *Cache) AssetsCount() int {
	var count int

	for _, v := range c.AseetDownloads {
		count += len(v)
	}

	return count
}

func (c *Cache) AssetReleasesCount() int {
	return len(c.AseetDownloads)
}

func (c *Cache) SetAssetDownloadCount(
	releaseName, assetName string,
	count int,
) {
	if c.AseetDownloads == nil {
		c.AseetDownloads = map[string]map[string]int{}
	}
	if c.AseetDownloads[releaseName] == nil {
		c.AseetDownloads[releaseName] = map[string]int{}
	}
	c.AseetDownloads[releaseName][assetName] = count
}

func (c *Cache) TotalAssetDownloads() int {
	var count int

	for _, v := range c.AseetDownloads {
		for _, v := range v {
			count += v
		}
	}

	return count
}

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
								Name:    "exclude",
								Aliases: []string{"e"},
								Usage:   "regexp asset filename pattern to exclude",
								Value:   cli.NewStringSlice(`.+\.sha\d+$`),
							},
							&cli.StringFlag{
								Name:    "output",
								Aliases: []string{"o"},
							},
							&cli.StringFlag{
								Name:    "cache",
								Aliases: []string{"c"},
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

	owner, repo, err := getOwnerAndRepo(c)
	if err != nil {
		return err
	}

	cache := &Cache{}
	cacheFile := c.String("cache")

	if cacheFile != "" {
		slog.Info("reading cache", slog.String("file", cacheFile))

		cache, err = readCache(cacheFile)
		if err != nil {
			return err
		}

		slog.Info("cache stats",
			slog.Int("downloads", cache.TotalAssetDownloads()),
			slog.Int("releases", cache.AssetReleasesCount()),
			slog.Int("assets", cache.AssetsCount()),
		)
	}

	lastPage := 1
	for page := 1; page > 0 && page <= lastPage; {
		slog.Info("fetching releases", slog.Int("page", page))

		var releases []*github.RepositoryRelease
		var resp *github.Response
		releases, resp, err = gh.Repositories.ListReleases(
			c.Context, owner, repo, &github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		)
		if err != nil {
			return err
		}

		updateAssetDownloadsCache(cache, releases, excludes)
		page = resp.NextPage
		lastPage = resp.LastPage
	}

	if cacheFile != "" {
		slog.Info("writing cache", slog.String("file", cacheFile))
		slog.Info(
			"cache stats",
			slog.Int("downloads", cache.TotalAssetDownloads()),
			slog.Int("releases", cache.AssetReleasesCount()),
			slog.Int("assets", cache.AssetsCount()),
		)

		err = writeCache(cacheFile, cache)
		if err != nil {
			return err
		}
	}

	count := cache.TotalAssetDownloads()
	var humanCount string
	if count >= 1000 {
		v, sym := humanize.ComputeSI(float64(count))
		humanCount = humanize.FormatFloat("#,###.#", v) + sym
	} else {
		humanCount = strconv.Itoa(count)
	}

	slog.Info("total downloads",
		slog.Int("count", count),
		slog.String("humanized_count", humanCount),
	)

	badge := &Badge{
		SchemaVersion: 1,
		Style:         "flat",
		Color:         "blue",
		LogoSVG:       downloadIcon,
		Label:         "total downloads",
		Message:       humanCount,
	}

	b, err := jsonPrettyMarshal(badge)
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
		err := os.WriteFile(filename, b, 0o644) //nolint:gosec
		if err != nil {
			return err
		}
	} else {
		fmt.Print(string(b))
	}

	return nil
}

func getOwnerAndRepo(c *cli.Context) (string, string, error) {
	repository := c.String("repository")
	if repository == "" {
		return "", "", fmt.Errorf(
			"No repository specified. Use --repository flag or set " +
				"GITHUB_REPOSITORY environment variable.",
		)
	}
	parts := strings.SplitN(repository, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf(
			"Invalid repository name. Expected format: owner/repo",
		)
	}

	return parts[0], parts[1], nil
}

func updateAssetDownloadsCache(
	cache *Cache,
	releases []*github.RepositoryRelease,
	excludes []*regexp.Regexp,
) {
	for _, release := range releases {
		for _, asset := range release.Assets {
			name := asset.GetName()
			if anyMatch(name, excludes) {
				continue
			}

			if v := asset.GetDownloadCount(); v > 0 {
				cache.SetAssetDownloadCount(release.GetName(), name, v)
			}
		}
	}
}

func anyMatch(s string, patterns []*regexp.Regexp) bool {
	for _, p := range patterns {
		if p.MatchString(s) {
			return true
		}
	}

	return false
}

func readCache(filename string) (*Cache, error) {
	cache := &Cache{AseetDownloads: map[string]map[string]int{}}

	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return cache, nil
		}

		return nil, err
	}
	defer func() {
		if e := f.Close(); e != nil {
			err = e
		}
	}()

	err = json.NewDecoder(f).Decode(&cache)
	if err != nil {
		return nil, err
	}

	return cache, nil
}

func writeCache(filename string, cache *Cache) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if e := f.Close(); e != nil {
			err = e
		}
	}()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(cache)
	if err != nil {
		return err
	}

	return nil
}

func jsonPrettyMarshal(v interface{}) ([]byte, error) {
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
