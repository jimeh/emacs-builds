package main

import (
	"debug/macho"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/xlab/treeprint"
)

var app = &cli.App{
	Name:      "dylib-tree",
	Usage:     "recursive list shared-libraries as a tree",
	UsageText: "dylib-tree [options] <binary-file> [<binary-file>]",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "depth",
			Usage: "max depth of tree (default: 0 = unlimited)",
			Value: 0,
		},
		&cli.StringSliceFlag{
			Name:  "ignore",
			Usage: "path patterns to ignore",
		},

		&cli.BoolFlag{
			Name:  "ignore-system",
			Usage: "ignore system libraries",
		},
		&cli.BoolFlag{
			Name: "real-path",
			Usage: "show resolved full paths instead of @executable_path " +
				"and @rpath",
		},
	},
	Action: actionHandler(LinkTreeCmd),
}

type Context struct {
	Root           string
	ExecutablePath string
	Depth          int
	MaxDepth       int
	RealPath       bool
	Ignore         []string
	RPaths         []string
}

func (s *Context) WithFile(filename string) *Context {
	ctx := *s
	ctx.Root = filename
	ctx.ExecutablePath = filepath.Dir(filename)
	return &ctx
}

func (s *Context) WithDepth(depth int) *Context {
	ctx := *s
	ctx.Depth = depth
	return &ctx
}

func (s *Context) WithIgnore(ignore []string) *Context {
	ctx := *s
	ctx.Ignore = append(ctx.Ignore, ignore...)
	return &ctx
}

func (s *Context) WithRpaths(rpaths []string) *Context {
	ctx := *s
	ctx.RPaths = append(ctx.RPaths, rpaths...)
	return &ctx
}

func actionHandler(
	f func(*cli.Context, *Context) error,
) func(*cli.Context) error {
	return func(c *cli.Context) error {
		ignore := c.StringSlice("ignore")
		if c.Bool("ignore-system") {
			ignore = append(
				ignore,
				"/System/Library/*",
				"*/libSystem.*.dylib",
				"*/libobjc.*.dylib",
			)
		}

		ctx := &Context{
			Ignore:   ignore,
			MaxDepth: c.Int("depth"),
			RealPath: c.Bool("real-path"),
		}

		return f(c, ctx)
	}
}

func LinkTreeCmd(c *cli.Context, ctx *Context) error {
	for _, filename := range c.Args().Slice() {
		ctx := ctx.WithFile(filename)

		treeRoot := treeprint.New()
		tree := treeRoot.AddBranch(filename)

		err := processBinary(&tree, ctx, filename)
		if err != nil {
			return err
		}

		fmt.Println(treeRoot.String())
	}
	return nil
}

func processBinary(
	parent *treeprint.Tree,
	ctx *Context,
	filename string,
) error {
	f, err := macho.Open(filename)
	if err != nil {
		return err
	}

	ctx = ctx.WithDepth(ctx.Depth + 1).WithRpaths(getRpaths(f))

	if ctx.MaxDepth > 0 && ctx.Depth > ctx.MaxDepth {
		return nil
	}

	libs, err := f.ImportedLibraries()
	if err != nil {
		return err
	}

	for _, lib := range libs {
		skip, err := ignoreLib(ctx, lib)
		if err != nil {
			return err
		}
		if skip {
			continue
		}

		filename, err := resolveLibFilename(ctx, lib)
		if err != nil {
			(*parent).AddBranch(lib)
			continue
		}

		skip, err = ignoreLib(ctx, filename)
		if err != nil {
			return err
		}
		if skip {
			continue
		}

		displayName := lib
		if ctx.RealPath {
			displayName = filename
		}

		tree := (*parent).AddBranch(displayName)

		err = processBinary(&tree, ctx, filename)
		if err != nil {
			return err
		}
	}

	return nil
}

func ignoreLib(ctx *Context, lib string) (bool, error) {
	for _, pattern := range ctx.Ignore {
		m, err := ignoreRegexp(pattern)
		if err != nil {
			return false, err
		}

		if m.MatchString(lib) {
			return true, nil
		}
	}

	return false, nil
}

func ignoreRegexp(p string) (*regexp.Regexp, error) {
	rp := "^" + regexp.QuoteMeta(p) + "$"
	rp = strings.ReplaceAll(rp, `\*`, ".*")
	rp = strings.ReplaceAll(rp, `\?`, ".")

	return regexp.Compile(rp)
}

func resolveLibFilename(ctx *Context, lib string) (string, error) {
	filename := lib

	if strings.HasPrefix(lib, "@executable_path") {
		filename = filepath.Join(ctx.ExecutablePath, lib[16:])
	} else if strings.HasPrefix(lib, "@rpath") {
		for _, r := range ctx.RPaths {
			if strings.HasPrefix(r, "@executable_path") {
				r = filepath.Join(ctx.ExecutablePath, r[16:])
			}

			rfile := filepath.Join(r, lib[6:])
			_, err := os.Stat(rfile)
			if err != nil {
				continue
			}

			return rfile, nil
		}

		return "", fmt.Errorf("could not find %s", lib)
	}

	_, err := os.Stat(filename)

	return filename, err
}

func getRpaths(f *macho.File) []string {
	paths := []string{}

	for _, i := range f.Loads {
		if r, ok := i.(*macho.Rpath); ok {
			paths = append(paths, r.Path)
		}
	}

	return paths
}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}
