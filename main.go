package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"
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

	for _, postFilename := range postFilenames {
		file, err := os.Open("posts/" + postFilename)
		check(err)

		data, err := ioutil.ReadAll(file)
		check(err)

		post := Post{
			Title:   "Hello world, this is my first post",
			Date:    time.Now(),
			Content: string(data),
		}

		postBaseName := strings.Split(postFilename, ".md")[0]
		postFile, err := os.Create("public/" + postBaseName + ".html")
		fmt.Println(postFile)

		postTemplate.Execute(postFile, post)

		postFile.Close()
	}
}
