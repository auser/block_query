package cmd

import (
	"fmt"
	"log"

	bq "github.com/auser/block_query/grammar"
	"github.com/urfave/cli"
)

var QueryCmd = cli.Command{
	Name:   "query",
	Usage:  "query",
	Action: handleQuery,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "query",
		},
	},
}

func handleQuery(c *cli.Context) {
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
}
