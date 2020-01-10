package main

import (
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"
)

type Post struct {
	Title   string
	Content string
	Date    time.Time
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func processPost(postFilename string, postTemplate *template.Template, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.Open("posts/" + postFilename)
	check(err)

	data, err := ioutil.ReadAll(file)
	check(err)

	html, err := convertMarkdownToHTML(data)
	check(err)

	post := Post{
		Title:   "Hello world, this is my first post",
		Date:    time.Now(),
		Content: html,
	}

	postBaseName := strings.Split(postFilename, ".md")[0]
	postFile, err := os.Create("public/" + postBaseName + ".html")

	postTemplate.Execute(postFile, post)

	postFile.Close()
}

func main() {
	// Create directory "public" if it does not exist
	// otherwise create after recursively deleting.
	if _, err := os.Stat("public"); err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir("public", 0755)
			check(err)
		} else {
			panic(err)
		}
	} else {
		os.RemoveAll("public")
		os.Mkdir("public", 0755)
	}

	postTemplate := template.Must(template.ParseFiles("post.tmpl"))

	postsDir, err := os.Open("posts")
	check(err)
	postFilenames, err := postsDir.Readdirnames(-1)
	postsDir.Close()
	check(err)

	var wg sync.WaitGroup

	for _, postFilename := range postFilenames {
		wg.Add(1)
		go processPost(postFilename, postTemplate, &wg)
	}

	wg.Wait()
}
