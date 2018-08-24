package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	bq "github.com/auser/block_query/grammar"
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
		cli.StringFlag{
			Name:  "query",
			Usage: "query",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "debug statements",
		},
	}
	app.Action = handleQuery
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func handleQuery(c *cli.Context) error {
	query := c.String("query")
	debugging := c.Bool("debug")

	if query == "" {
		log.Fatal("No query defined. Must be passed in")
	}

	q := &bq.BlockQuery{Buffer: query, Pretty: debugging}
	q.Init()
	if err := q.Parse(); err != nil {
		log.Fatal(err)
	}
	q.Execute()

	if debugging {
		q.PrintSyntaxTree()
	}

	fmt.Printf("%v\n", q.ExprStack.String())

	return nil
}

func readVersion() string {
	var version = "0.0.0"
	da, err := ioutil.ReadFile("./Version")
	if err == nil {
		version = string(da)
	}
	return version
}
