package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	// "reflect"
)

func main() {
	args := os.Args

	if len(args) < 3 {
		fmt.Println("invalid args' count")
		fmt.Println("example: cage search <package's name>")
		return
	}

	if len(args) > 3 {
		fmt.Println("unused args", args[3:])
	}

	command := args[1]
	target := args[2]

	switch command {
	case "search":
		search(target)
	default:
		fmt.Println("invalid command", command)
	}
}

func search(target string) {
	if target == "" {
		return
	}

	url := fmt.Sprintf("https://api.github.com/search/repositories?q=%s", target)
	out := make(chan string)

	go get(url, out)

	body := <-out
	decodeJSON(body)
}

func loading(stop chan bool) {
	index := 0
	arr := [4]string{"|", "/", "-", "\\"}
	for {
		fmt.Printf("\rOn %s", arr[index])
		index++
		time.Sleep(1000)
	}
}

func decodeJSON(body string) {
	js, err := simplejson.NewJson([]byte(body))
	if err != nil {
		panic("json format error")
	}

	items := js.Get("items")
	index := 0

	for {
		item := items.GetIndex(index).MustMap()
		if item == nil {
			break
		}

		language := item["language"]
		stars := item["stargazers_count"]
		name := item["full_name"]

		if language == nil {
			language = "unknow"
		}

		fmt.Printf("%2v - %16s\t%6s\tgithub.com/%s\n",
			index+1,
			language,
			stars,
			name)

		index++
	}
}

func get(url string, out chan string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
		return
	}

	req.Header.Add("content-type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return
	}

	out <- string(body)
}
