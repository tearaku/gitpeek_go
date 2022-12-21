package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.design/x/clipboard"
)

type GitEntry struct {
	Path      string
	GitBranch string
}

type AppContext struct {
	Exclude map[string]bool
	Limit   int
}

func (ctx *AppContext) Init(cCtx *cli.Context) {
	excludeMap := make(map[string]bool)
	for _, name := range cCtx.StringSlice("exclude") {
		excludeMap[name] = true
	}
	ctx.Limit = cCtx.Int("limit")
	ctx.Exclude = excludeMap
}

func setUpFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "exclude",
			Aliases: []string{"e"},
			Usage:   "Comma separated list of directory names to exclude from the search",
			Value:   cli.NewStringSlice("node_modules"),
		},
		&cli.IntFlag{
			Name:    "limit",
			Aliases: []string{"l"},
			Usage:   "Integer, specifying maximum limit on recursive search (starting from current directory)",
			Value:   5,
		},
	}
}

func getBranchName(path string) (string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if file.Name() != "HEAD" {
			continue
		}
		gitHead, err := os.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return "", err
		}
		bNames := strings.Split(string(gitHead), "ref: refs/heads/")
		return strings.TrimSpace(bNames[1]), nil
	}
	return "", fmt.Errorf("cannot find 'HEAD' file in %s", path)
}

// `root`: root from which git directories are recursively searched
func findGitDir(ctx AppContext, root string) ([]GitEntry, error) {
	result := make([]GitEntry, 0)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		// Depth check (-l flag)
		if len(strings.Split(path, "/")) > ctx.Limit {
			return filepath.SkipDir
		}
		// Name check (-e flag)
		if ctx.Exclude[d.Name()] {
			return nil
		}

		if d.Name() == ".git" {
			bName, err := getBranchName(path)
			if err != nil {
				return err
			}
			result = append(result, GitEntry{Path: filepath.Dir(path), GitBranch: bName})
			return nil
		}
		return nil
	})
	return result, err
}

func main() {
	if err := clipboard.Init(); err != nil {
		log.Fatal(err)
	}
	app := &cli.App{
		Usage: "Display folders and their git branches",
		Flags: setUpFlags(),
		Action: func(ctx *cli.Context) error {
			appCtx := AppContext{}
			appCtx.Init(ctx)

			result, err := findGitDir(appCtx, ".")
			if err != nil {
				return err
			}

			prompt, err := PromptMenu(result)
			if err != nil {
				return err
			}
			idx, _, err := prompt.Run()
			if err != nil {
				fmt.Println(err)
				return err
			}

			cdCmd := "cd " + result[idx].Path
			clipboard.Write(clipboard.FmtText, []byte(cdCmd))
			fmt.Printf("Command copied to clipboard!: %v\n", result[idx].Path)

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
