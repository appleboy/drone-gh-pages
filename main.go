package main

import (
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

var (
	version = "unknown"
)

func main() {
	app := cli.NewApp()
	app.Name = "gh-pages plugin"
	app.Usage = "gh-pages plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "upstream-name",
			Usage:  "git upstream to target",
			EnvVar: "PLUGIN_UPSTREAM_NAME,INPUT_UPSTREAM_NAME",
			Value:  "origin",
		},
		cli.StringFlag{
			Name:   "target-branch",
			Usage:  "git branch to target",
			EnvVar: "PLUGIN_TARGET_BRANCH,INPUT_TARGET_BRANCH",
			Value:  "gh-pages",
		},
		cli.StringFlag{
			Name:   "temporary-base",
			Usage:  "temporary directory for pages pull",
			EnvVar: "PLUGIN_TEMPORARY_BASE,INPUT_TEMPORARY_BASE",
			Value:  ".tmp",
		},
		cli.StringFlag{
			Name:   "pages-directory",
			Usage:  "directory of content to publish",
			EnvVar: "PLUGIN_PAGES_DIRECTORY,INPUT_PAGES_DIRECTORY",
			Value:  "docs",
		},
		cli.StringFlag{
			Name:   "target-directory",
			Usage:  "directory of content to sync",
			EnvVar: "PLUGIN_TARGET_DIRECTORY,INPUT_TARGET_DIRECTORY",
		},
		cli.StringFlag{
			Name:   "ssh-key",
			Usage:  "private ssh key",
			EnvVar: "PLUGIN_SSH_KEY,GIT_PUSH_SSH_KEY,SSH_KEY,INPUT_SSH_KEY",
		},
		cli.StringFlag{
			Name:   "commit.author.name",
			Usage:  "git author name",
			EnvVar: "PLUGIN_USER_NAME,DRONE_COMMIT_AUTHOR,INPUT_COMMIT_AUTHOR",
		},
		cli.StringFlag{
			Name:   "commit.author.email",
			Usage:  "git author email",
			EnvVar: "PLUGIN_USER_EMAIL,DRONE_COMMIT_AUTHOR_EMAIL,INPUT_COMMIT_AUTHOR_EMAIL",
		},
		cli.StringFlag{
			Name:   "remote",
			Usage:  "git remote url",
			EnvVar: "PLUGIN_REMOTE_URL,DRONE_REMOTE_URL,INPUT_REMOTE_URL",
		},
		cli.StringFlag{
			Name:   "path",
			Usage:  "git clone path",
			EnvVar: "PLUGIN_WORKSPACE,DRONE_WORKSPACE,INPUT_WORKSPACE",
		},
		cli.StringFlag{
			Name:   "netrc.machine",
			Usage:  "netrc machine",
			EnvVar: "PLUGIN_NETRC_MACHINE,DRONE_NETRC_MACHINE,INPUT_NETRC_MACHINE",
			Value:  "github.com",
		},
		cli.StringFlag{
			Name:   "netrc.username",
			Usage:  "netrc username",
			EnvVar: "PLUGIN_USERNAME,DRONE_NETRC_USERNAME,GH_PAGES_USERNAME,GITHUB_USERNAME,INPUT_USERNAME",
		},
		cli.StringFlag{
			Name:   "netrc.password",
			Usage:  "netrc password",
			EnvVar: "PLUGIN_PASSWORD,DRONE_NETRC_PASSWORD,GH_PAGES_PASSWORD,GITHUB_PASSWORD,INPUT_PASSWORD",
		},
		cli.StringSliceFlag{
			Name:   "rsync.exclude",
			Usage:  "exclude files matching PATTERN",
			EnvVar: "PLUGIN_EXCLUDE,RSYNC_EXCLUDE,INPUT_EXCLUDE",
		},
		cli.BoolFlag{
			Name:   "github.actions",
			Usage:  "GitHub Actions CI",
			EnvVar: "GITHUB_ACTIONS",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	plugin := Plugin{
		Repo: Repo{
			Clone: c.String("remote"),
		},

		Build: Build{
			Path: c.String("path"),
		},

		Commit: Commit{
			Author: Author{
				Name:  c.String("commit.author.name"),
				Email: c.String("commit.author.email"),
			},
		},

		Netrc: Netrc{
			Login:    c.String("netrc.username"),
			Machine:  c.String("netrc.machine"),
			Password: c.String("netrc.password"),
		},
		Config: Config{
			Key:             c.String("ssh-key"),
			UpstreamName:    c.String("upstream-name"),
			TargetBranch:    c.String("target-branch"),
			TemporaryBase:   c.String("temporary-base"),
			PagesDirectory:  c.String("pages-directory"),
			TargetDirectory: c.String("target-directory"),
			ExcludeList:     c.StringSlice("rsync.exclude"),
		},
	}

	if !filepath.IsAbs(plugin.Config.TemporaryBase) {
		plugin.Config.TemporaryBase = filepath.Join(
			plugin.Build.Path,
			plugin.Config.TemporaryBase,
		)
	}

	if !filepath.IsAbs(plugin.Config.PagesDirectory) {
		plugin.Config.PagesDirectory = filepath.Join(
			plugin.Build.Path,
			plugin.Config.PagesDirectory,
		)
	}

	plugin.Config.WorkDirectory = filepath.Join(
		plugin.Config.TemporaryBase,
		filepath.Base(plugin.Config.PagesDirectory),
	)

	if c.Bool("github.actions") {
		u, err := url.Parse(plugin.Repo.Clone)
		if err != nil {
			return err
		}

		u.User = url.UserPassword(plugin.Netrc.Login, plugin.Netrc.Password)
		plugin.Repo.Clone = u.String()
	}

	return plugin.Exec()
}
