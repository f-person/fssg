package main

import (
	"html"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/f-person/fssg/parser"
)

type Post struct {
	Title   string
	Content string
	Date    time.Time
	Link    string
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func processPost(postPath string, postInfo os.FileInfo, postTemplate *template.Template, wg *sync.WaitGroup) Post {
	defer wg.Done()

	file, err := os.Open(postPath)
	check(err)

	data, err := ioutil.ReadAll(file)
	check(err)

	metadata := parser.ParseMetadata(data)
	if metadata["title"] == "" || metadata["date"] == "" {
		panic("title and data should not be empty")
	}

	date, _ := time.Parse("02.01.2006, 15:04", metadata["date"])

	contentStartsAt, _ := strconv.Atoi(metadata["contentStartsAt"])
	data = data[contentStartsAt:]

	html, err := parser.ConvertMarkdownToHTML(data)
	check(err)

	post := Post{
		Title:   metadata["title"],
		Date:    date,
		Content: html,
	}

	pathParts := strings.Split(postPath, "/")
	filePath := strings.Join(pathParts[1:], "/")
	filePathParts := strings.Split(filePath, ".")

	// TODO decide from where i want to get dirPath: from the file or the title of the post
	dirPath := strings.Join(filePathParts[:len(filePathParts)-1], ".")
	publicDirPath := "public/" + dirPath + "/"
	err = os.Mkdir(publicDirPath, 0755)
	check(err)

	publicFile, err := os.Create(publicDirPath + "index.html")
	check(err)

	postTemplate.Execute(publicFile, post)
	publicFile.Close()

	post.Link = "/" + dirPath

	return post
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
		err := os.RemoveAll("public")
		check(err)

		err = os.Mkdir("public", 0755)
		check(err)
	}

	postTemplate := template.Must(template.ParseFiles("post.tmpl"))
	var wg sync.WaitGroup

	var posts []Post

	err := filepath.Walk("./posts/", func(path string, info os.FileInfo, err error) error {
		if path == "./posts/" {
			return nil
		}

		if info.IsDir() {
			dirParts := strings.Split(path, "posts/")[1:]
			dirPath := strings.Join(dirParts, "/")
			publicDirPath := "public/" + dirPath
			err := os.Mkdir(publicDirPath, 0755)
			check(err)
		} else {
			wg.Add(1)
			go func() {
				post := processPost(path, info, postTemplate, &wg)
				posts = append(posts, post)
			}()
		}

		return err
	})
	check(err)

	wg.Wait()

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	wg.Add(1)
	go func() {
		defer wg.Done()

		indexFile, err := os.Create("public/index.html")
		check(err)

		indexTemplate := template.Must(template.ParseFiles("post_index.tmpl"))
		err = indexTemplate.Execute(indexFile, posts)
		check(err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		feedFile, err := os.Create("public/feed.xml")
		check(err)

		//feedTemplate := template.Must(template.ParseFiles("feed.tmpl"))
		feedTemplate, err := template.New("feed.tmpl").
			Funcs(template.FuncMap{"safeHTML": html.EscapeString}).
			ParseFiles("feed.tmpl")
		check(err)
		err = feedTemplate.Execute(feedFile, posts)
		check(err)
	}()

	wg.Wait()

}
