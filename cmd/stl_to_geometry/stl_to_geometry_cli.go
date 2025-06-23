package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/erh/vmodutils/smtools"
)

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}

func realMain() error {
	fn := os.Args[1]
	g, err := smtools.STLFileToGeometry(fn)
	if err != nil {
		return err
	}

	raw, err := json.Marshal(g)
	if err != nil {
		return err
	}

	fmt.Println(string(raw))
	return nil
}
