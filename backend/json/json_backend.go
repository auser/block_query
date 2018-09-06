package json_backend

import (
	"encoding/json"
	"fmt"
	"log"
)

func Parse(str []byte) {
	var res map[string]interface{}
	err := json.Unmarshal(str, &res)

	if err != nil {
		log.Fatal("Error when parsing JSON: %s\n", err.Error())
	}

	fmt.Printf("res: %#v\n", res)
}
