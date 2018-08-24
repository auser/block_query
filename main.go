package main

import (
	"io/ioutil"
	"os"

	cmd "github.com/auser/block_query/cmd"
	"github.com/urfave/cli"
)

func main() {
	Run(os.Args)
}

// Run executes the program
func Run(args []string) {
	app := cli.NewApp()
	var version = readVersion()
	app.Name = "block_query"
	app.Version = version
	app.Usage = "block query"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "nocolor",
			Usage: "disable color",
		},
	}
	app.Commands = []cli.Command{cmd.QueryCmd}
	app.Run(args)
}

func readVersion() string {
	var version = "0.0.0"
	da, err := ioutil.ReadFile("./Version")
	if err == nil {
		version = string(da)
	}
	return version
}
