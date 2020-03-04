package main

import (
	"fmt"
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
	"github.com/f-person/fssg/utils"
)

type Post struct {
	Title    string
	Content  string
	Date     time.Time
	Link     string
	Metadata map[string]interface{}
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

	parser := parser.Parser{MD: data}

	pathParts := strings.Split(postPath, "/")
	filePath := strings.Join(pathParts[1:], "/")
	filePathParts := strings.Split(filePath, ".")
	dirPath := strings.Join(filePathParts[:len(filePathParts)-1], ".")

	metadata := parser.ParseMetadata()
	if metadata["title"] == nil {
		metadata["title"] = dirPath
	}

	contentStartsAt, _ := strconv.Atoi(metadata["contentStartsAt"].(string))
	parser.MD = parser.MD[contentStartsAt:]

	html, err := parser.ConvertMarkdownToHTML()
	check(err)

	post := Post{
		Title:    metadata["title"].(string),
		Content:  html,
		Metadata: metadata,
	}

	if metadata["date"] != nil {
		date, err := time.Parse("02.01.2006, 15:04", metadata["date"].(string))
		if err != nil {
			fmt.Println(err)
			post.Date = time.Now()
		}

		post.Date = date
	} else {
		post.Date = time.Now()
	}

	publicDirPath := "public/" + dirPath + "/"
	err = os.Mkdir(publicDirPath, 0755)
	check(err)

	publicFile, err := os.Create(publicDirPath + "index.html")
	check(err)

	err = postTemplate.Execute(publicFile, post)
	if err != nil {
		fmt.Println(err)
	}
	publicFile.Close()

	post.Link = "/" + dirPath

	return post
}

func main() {
	err := utils.CreateDir("public")
	check(err)

	postTemplate := template.Must(template.ParseFiles("theme/post.tmpl"))
	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		err := utils.CopyDir("static", "public/static")
		check(err)

		defer wg.Done()
	}()

	var posts []Post

	err = filepath.Walk("./posts/", func(path string, info os.FileInfo, err error) error {
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

		indexTemplate := template.Must(template.ParseFiles("theme/post_index.tmpl"))
		err = indexTemplate.Execute(indexFile, posts)
		check(err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		feedFile, err := os.Create("public/feed.xml")
		check(err)

		feedTemplate, err := template.New("feed.tmpl").
			Funcs(template.FuncMap{"safeHTML": html.EscapeString}).
			ParseFiles("theme/feed.tmpl")
		check(err)
		err = feedTemplate.Execute(feedFile, posts)
		check(err)
	}()

	wg.Wait()

}
