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
	}
	app.Action = handleQuery
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func handleQuery(c *cli.Context) error {
	fmt.Printf("Hi?")
	query := c.String("query")

	if query == "" {
		log.Fatal("No query defined. Must be passed in")
	}

	// buffer := bufio.NewReader(strings.NewReader(query))
	fmt.Printf("Handling query: %s\n", query)
	q := &bq.BlockQuery{Buffer: query, Pretty: true}
	q.Init()
	if err := q.Parse(); err != nil {
		log.Fatal(err)
	}
	// bq.Execute()
	q.PrintSyntaxTree()
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
