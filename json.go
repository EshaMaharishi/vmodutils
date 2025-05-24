package vmodutils

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func ReadJSONFromFile(fn string, where any) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}

	jsonData, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, where)
}
