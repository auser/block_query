package utils

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"runtime"
)

// ReadFixture returns json from a fixture at tests/fixtures/filepath
func ReadFixture(filename string) ([]byte, error) {
	_, thisFile, _, _ := runtime.Caller(0)
	filepath := filepath.Join(path.Dir(thisFile), fmt.Sprintf("../test/fixtures/%s", filename))

	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// var result map[string]interface{}
	// json.Unmarshal(contents, &result)

	return contents, nil
}
