package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/f-person/fssg/parser"
	"github.com/f-person/fssg/utils"
)

const dateFormat = "02.01.2006, 15:04"

type Post struct {
	Title           string
	Content         string
	Date            time.Time
	PublicationDate time.Time
	Link            string
	Metadata        map[string]interface{}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// getMetadata return metadata in the same format, in which it's stored in a post file
func getMetadataAsInSource(metadata map[string]interface{}) (string, int) {
	var md strings.Builder
	md.Write([]byte("---\n"))

	delete(metadata, "contentStartsAt")
	contentStartsAt := 0

	for k, v := range metadata {
		line := fmt.Sprintf("%v: %v\n", k, v)
		md.WriteString(line)
		contentStartsAt += len(line)
	}

	md.Write([]byte("---\n\n"))
	contentStartsAt += 9
	metadata["contentStartsAt"] = contentStartsAt

	return md.String(), contentStartsAt
}

func processPost(postPath string, postInfo os.FileInfo, postTemplate *template.Template) Post {
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

	contentStartsAt, _ := metadata["contentStartsAt"].(int)
	parser.MD = parser.MD[contentStartsAt:]

	html, err := parser.ConvertMarkdownToHTML()
	check(err)

	post := Post{
		Title:    metadata["title"].(string),
		Content:  html,
		Metadata: metadata,
	}

	if metadata["published"] == nil {
		now := time.Now()
		metadata["published"] = now.Format(dateFormat)
		sourceMetadata, contentStartsAt := getMetadataAsInSource(metadata)
		ioutil.WriteFile(postPath, append([]byte(sourceMetadata), parser.MD...), 0755)
		metadata["contentStartsAt"] = contentStartsAt
		post.PublicationDate = now
	} else {
		post.PublicationDate, err = time.Parse(dateFormat, metadata["published"].(string))
		if err != nil {
			fmt.Println(err)
			post.PublicationDate = time.Now()
		}
	}

	if metadata["date"] != nil {
		date, err := time.Parse(dateFormat, metadata["date"].(string))
		if err != nil {
			fmt.Println(err)
			post.Date = time.Now()
		}

		post.Date = date
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

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := utils.CopyDir("static", "public/static")
		check(err)
	}()

	var posts []Post
	mu := &sync.Mutex{}

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
				defer wg.Done()

				post := processPost(path, info, postTemplate)

				mu.Lock()
				posts = append(posts, post)
				mu.Unlock()
			}()
		}

		return err
	})
	check(err)

	wg.Wait()

	sort.Slice(posts, func(i, j int) bool {
		iDate := posts[i].Date
		jDate := posts[j].Date

		if iDate.IsZero() {
			iDate = posts[i].PublicationDate
		}
		if jDate.IsZero() {
			jDate = posts[j].PublicationDate
		}

		return iDate.After(jDate)
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
