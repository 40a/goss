package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aelsabbahy/goss"
	"github.com/aelsabbahy/goss/outputs"
	"github.com/codegangsta/cli"
	//"time"
)

var version string

func main() {
	startTime := time.Now()
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Version = version
	app.Name = "goss"
	app.Usage = "Quick and Easy server validation"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "gossfile, g",
			Value:  "./goss.yaml",
			Usage:  "Goss file to read from / write to",
			EnvVar: "GOSS_FILE",
		},
		cli.StringFlag{
			Name:  "package",
			Usage: "Package type to use [rpm, deb, apk, pacman]",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "validate",
			Aliases: []string{"v"},
			Usage:   "Validate system",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "format, f",
					Value:  "rspecish",
					Usage:  fmt.Sprintf("Format to output in, valid options: %s", outputs.Outputers()),
					EnvVar: "GOSS_FMT",
				},
				cli.BoolFlag{
					Name:   "no-color",
					Usage:  "Force color off",
					EnvVar: "GOSS_NOCOLOR",
				},
			},
			Action: func(c *cli.Context) {
				goss.Validate(c, startTime)
			},
		},
		{
			Name:    "render",
			Aliases: []string{"r"},
			Usage:   "render gossfile after imports",
			Action: func(c *cli.Context) {
				fmt.Print(goss.RenderJSON(c.GlobalString("gossfile")))
			},
		},
		{
			Name:    "autoadd",
			Aliases: []string{"aa"},
			Usage:   "automatically add all matching resource to the test suite",
			Action: func(c *cli.Context) {
				goss.AutoAddResource(c.GlobalString("gossfile"), c.Args().First(), c)
			},
		},
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add a resource to the test suite",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "exclude-attr",
					Usage: "Exclude the following attributes when adding a new resource",
				},
			},
			Subcommands: []cli.Command{
				{
					Name:  "package",
					Usage: "add new package",
					Action: func(c *cli.Context) {
						goss.AddResource(c.GlobalString("gossfile"), "Package", c.Args().First(), c)
					},
				},
				{
					Name:  "file",
					Usage: "add new file",
					Action: func(c *cli.Context) {
						goss.AddResource(c.GlobalString("gossfile"), "File", c.Args().First(), c)
					},
				},
				{
					Name:  "addr",
					Usage: "add new remote address:port - ex: google.com:80",
					Flags: []cli.Flag{
						cli.DurationFlag{
							Name:  "timeout",
							Value: 500 * time.Millisecond,
						},
					},
					Action: func(c *cli.Context) {
						goss.AddResource(c.GlobalString("gossfile"), "Addr", c.Args().First(), c)
					},
				},
				{
					Name:  "port",
					Usage: "add new listening [protocol]:port - ex: 80 or udp:123",
					Action: func(c *cli.Context) {
						goss.AddResource(c.GlobalString("gossfile"), "Port", c.Args().First(), c)
					},
				},
				{
					Name:  "service",
					Usage: "add new service",
					Action: func(c *cli.Context) {
						goss.AddResource(c.GlobalString("gossfile"), "Service", c.Args().First(), c)
					},
				},
				{
					Name:  "user",
					Usage: "add new user",
					Action: func(c *cli.Context) {
						goss.AddResource(c.GlobalString("gossfile"), "User", c.Args().First(), c)
					},
				},
				{
					Name:  "group",
					Usage: "add new group",
					Action: func(c *cli.Context) {
						goss.AddResource(c.GlobalString("gossfile"), "Group", c.Args().First(), c)
					},
				},
				{
					Name:  "command",
					Usage: "add new command",
					Flags: []cli.Flag{
						cli.DurationFlag{
							Name:  "timeout",
							Value: 10 * time.Second,
						},
					},
					Action: func(c *cli.Context) {
						goss.AddResource(c.GlobalString("gossfile"), "Command", c.Args().First(), c)
					},
				},
				{
					Name:  "dns",
					Usage: "add new dns",
					Flags: []cli.Flag{
						cli.DurationFlag{
							Name:  "timeout",
							Value: 500 * time.Millisecond,
						},
					},
					Action: func(c *cli.Context) {
						goss.AddResource(c.GlobalString("gossfile"), "DNS", c.Args().First(), c)
					},
				},
				{
					Name:  "process",
					Usage: "add new process name",
					Action: func(c *cli.Context) {
						goss.AddResource(c.GlobalString("gossfile"), "Process", c.Args().First(), c)
					},
				},
				{
					Name:  "goss",
					Usage: "add new goss file",
					Action: func(c *cli.Context) {
						goss.AddResource(c.GlobalString("gossfile"), "Gossfile", c.Args().First(), c)
					},
				},
			},
		},
	}

	app.Run(os.Args)

}
