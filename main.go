package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nyudlts/go-aspace"
	"log"
	"os"
	"strings"
)

var (
	client      *aspace.ASClient
	handle      = "http://hdl.handle.net/2333.1/"
	config      string
	environment string
	test        bool
)

func init() {
	flag.StringVar(&config, "config", "", "location of go-aspace config file")
	flag.StringVar(&environment, "environment", "", "aspace environment")
	flag.BoolVar(&test, "test", false, "")
}

func main() {
	flag.Parse()
	outFile, _ := os.Create("handle-updates.txt")
	defer outFile.Close()
	log.SetOutput(outFile)

	var err error
	client, err = aspace.NewClient(config, environment, 20)
	if err != nil {
		panic(err)
	}

	for _, i := range []int{2, 3, 6} {
		updateHandles(i)
	}

}

func updateHandles(repoID int) {
	iresults, err := client.Search(repoID, "digital_object", "*", 1)
	if err != nil {
		panic(err)
	}

	for page := 1; page <= iresults.LastPage; page++ {
		fmt.Printf("Repository %d Page %d of %d\n", repoID, page, iresults.LastPage)

		pageResults, err := client.Search(repoID, "digital_object", "*", page)
		if err != nil {
			panic(err)
		}
		for _, result := range pageResults.Results {
			do := aspace.DigitalObject{}
			err = json.Unmarshal([]byte(fmt.Sprint(result["json"])), &do)
			if err != nil {
				panic(err)
			}
			if containsHandle(do.FileVersions) == true {
				oldFV := do.FileVersions
				newFV := []aspace.FileVersion{}
				for _, fv := range oldFV {
					if strings.Contains(fv.FileURI, handle) == true {
						fv.FileURI = strings.ReplaceAll(fv.FileURI, "http", "https")
						newFV = append(newFV, fv)
					} else {
						newFV = append(newFV, fv)
					}
				}
				do.FileVersions = newFV
				log.Printf("[INFO] %s original: %v updated %v", do.URI, oldFV, do.FileVersions)

				if test != true {
					repoID, doID, err := aspace.URISplit(do.URI)
					msg, err := client.UpdateDigitalObject(repoID, doID, do)
					if err != nil {
						log.Printf("[ERROR] %s", strings.ReplaceAll(err.Error(), "\n", ""))
					} else {
						log.Printf("[INFO] %s", strings.ReplaceAll(msg, "\n", ""))
					}
				} else {
					log.Printf("[INFO] test mode skipping")
				}
			}
		}
	}
}

func containsHandle(fileVersions []aspace.FileVersion) bool {
	for _, fv := range fileVersions {
		if strings.Contains(fv.FileURI, handle) == true {
			return true
		}
	}
	return false
}
