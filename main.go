package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var markdownTemplate = `| Slide | Notes |
| --- | --- |
{{ range . }}| ![]({{ . }}){ width=100% } |  |
{{ end }}

`

func main() {
	log.SetFlags(log.Lshortfile)

	checkBinaries()

	filenameIn := selectFile()
	filenameOut := strings.Replace(filenameIn, ".pdf", ".odt", -1) // TODO use regexp to make sure that at the end

	overwriteExistingFile(filenameOut)

	tmpPath := fmt.Sprintf("%s/%s-%d", os.TempDir(), filenameIn, time.Now().Unix())
	markdownFilePath := fmt.Sprintf("%s/index.md", tmpPath)
	imagesPath := fmt.Sprintf("%s/images", tmpPath)

	convertPDFToImages(filenameIn, imagesPath)
	createMarkdownFile(markdownFilePath, imagesPath)

	createODTDocument(filenameOut, markdownFilePath)

	// TODO delete tmpPath (first echo)
}

func createODTDocument(filenameOut, markdownFilePath string) {
	command := exec.Command("pandoc", "-f", "markdown", "-t", "odt", "-o", filenameOut, markdownFilePath) // TODO are the params escaped by Command?
	if err := command.Run(); err != nil {
		log.Fatalln(err)
	}
}

func createMarkdownFile(markdownFilePath, imagesPath string) {
	t, err := template.New("notes").Parse(markdownTemplate)
	if err != nil {
		log.Fatalln(err)
	}

	files, err := ioutil.ReadDir(imagesPath)
	if err != nil {
		log.Fatal(err)
	}

	filenames := []string{}

	for _, file := range files {
		filenames = append(filenames, fmt.Sprintf("%s/%s", imagesPath, file.Name()))
	}

	file, err := os.Create(markdownFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	defer file.Close()

	if err := t.Execute(file, filenames); err != nil {
		log.Fatalln(err)
	}
}

func checkBinaries() {
	if _, err := exec.LookPath("convert"); err != nil {
		log.Fatalln("convert not installed")
	}

	if _, err := exec.LookPath("pandoc"); err != nil {
		log.Fatalln("pandoc not installed")
	}
}

func convertPDFToImages(filenameIn, imagesPath string) {
	if err := os.MkdirAll(imagesPath, 0700); err != nil {
		log.Fatalln(err)
	}

	command := exec.Command("convert", filenameIn, imagesPath+"/%003d.jpg") // TODO are the params escaped by Command?
	if err := command.Run(); err != nil {
		log.Fatalln(err)
	}
}

func selectFile() string {
	fmt.Println("Available pdf files:")

	filenames, err := filepath.Glob("*.pdf")
	if err != nil {
		log.Fatalln(err)
	}

	if len(filenames) < 1 {
		fmt.Println("No files available.")
		os.Exit(1)
	}

	for index, filename := range filenames {
		fmt.Printf("[%d] %s\n", index, filename)
	}

	fmt.Print("\nSelect a file: ")

	var filenumber int
	count, err := fmt.Scanf("%d", &filenumber)
	if err != nil || count != 1 || filenumber > len(filenames) {
		fmt.Println("Invalid input. Please try again.")
		os.Exit(1)
	}

	return filenames[filenumber]
}

func overwriteExistingFile(filename string) {
	_, err := os.Stat(filename)
	if err == nil {
		fmt.Printf("\nFile %s already exists. Do you want to overwrite the file? [yes/no]\n", filename)

		var overwrite string
		_, err := fmt.Scanf("%s", &overwrite)
		if err != nil {
			log.Fatalln(err)
		}

		if overwrite != "yes" {
			os.Exit(0)
		}
	}

	return
}
