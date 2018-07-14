package main

import (
	"fmt"
	"github.com/ghodss/yaml"
	"google.golang.org/appengine"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var availableFileExtensions = []string{".yaml", ".yml"}
var sourceDir string
var htmlLang string

var httpPaths = make([]string, 0)

func main() {

	readEnv()
	err := walkSources()
	if err != nil {
		panic(err)
	}
	handleRoot()
	appengine.Main()
}

func readEnv() {

	sourceDir = os.Getenv("SOURCE_DIR")
	if sourceDir == "" {
		sourceDir = "."
	}

	htmlLang = os.Getenv("HTML_LANG")
	if htmlLang == "" {
		htmlLang = "en"
	}
}

func walkSources() error {

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		var availableFile = false
		for _, ext := range availableFileExtensions {
			if strings.HasSuffix(path, ext) {
				availableFile = true
				break
			}
		}

		if !availableFile {
			return nil
		}

		openPath, e := filepath.Abs(path)
		if e != nil {
			return nil
		}

		yamlFile, e := os.Open(openPath)
		if e != nil {
			return nil
		}
		defer yamlFile.Close()

		yamlData, e := ioutil.ReadAll(yamlFile)
		if e != nil {
			return nil
		}

		jsonString, e := yaml.YAMLToJSON(yamlData)
		if e != nil {
			return nil
		}

		var html = fmt.Sprintf(htmlTemplate, htmlLang, jsonString)

		listenPath := strings.Replace(path, filepath.Clean(sourceDir), "", 1)

		httpPaths = append(httpPaths, listenPath)

		http.HandleFunc(listenPath, func(w http.ResponseWriter, r *http.Request) {

			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "text/html")

			fmt.Fprintf(w, "%s", html)
		})

		return nil
	})
}

func handleRoot() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "text/html")

		if len(httpPaths) == 0 {
			fmt.Fprintf(w, "no documents")
			return
		}

		var links string
		for _, path := range httpPaths {
			links += "<a href=" + path + ">" + path + "</a><br>\n"
		}
		fmt.Fprintf(w, topPageTemplate, htmlLang, links)
	})
}
