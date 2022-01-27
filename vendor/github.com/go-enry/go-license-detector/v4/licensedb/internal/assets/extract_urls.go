// +build make

package main

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
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
		seeAlso := data["seeAlso"]
		if seeAlso != nil {
			for _, url := range seeAlso.([]interface{}) {
				id := data["licenseId"].(string)
				strUrl := strings.TrimSpace(url.(string))
				strUrl = strUrl[strings.Index(strUrl, "://"):] // ignore http/https
				if strings.HasSuffix(strUrl, "/legalcode") && strings.HasPrefix(id, "CC") {
					strUrl = strUrl[:len(strUrl)-10]
				}
				writer.Write([]string{id, strUrl})
			}
		}
	}
	writer.Write([]string{"MIT", ".mit-license.org"})
}
