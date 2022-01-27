// +build make

package main

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func main() {
	dir := os.Args[1]
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("Listing %s: %v\n", dir, err)
	}
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()
	for _, file := range files {
		var data map[string]interface{}
		content, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			log.Fatalf("Reading %s: %v\n", file.Name(), err)
		}
		json.Unmarshal(content, &data)
		name := data["name"].(string)
		id := data["licenseId"].(string)
		writer.Write([]string{id, name})
	}
}
