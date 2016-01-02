package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/metakeule/fmtdate"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	args := os.Args
	help := map[string]string{"search": "search repos in github.com",
		"cok": "search position in cok by name and kingdom ID"}

	if len(args) < 3 {
		fmt.Print("Cage is a personal tool write by golang.\n\n")
		fmt.Print("Usage:\n\n")
		fmt.Print("    cage <command> [<args>][<options>]\n\n")
		fmt.Print("Commands:\n\n")
		for k, v := range help {
			fmt.Printf("    %-15s\t%s\n", k, v)
		}
		fmt.Print("\n")
		return
	}

	command := args[1]

	switch command {
	case "search":
		search(args[2])
	case "cok":
		cok(args[2], args[3])
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
	decodeSearchJSON(body)
}

func cok(sid string, name string) {
	if sid == "" || name == "" {
		panic("invalid command parameters")
	}

	url := "http://cok.icrazy.me/search.php"
	data := fmt.Sprintf("sid=%s&name=%s", sid, name)
	out := make(chan string)

	go post(url, data, out)

	body := <-out
	decodeCokJSON(body)
}

func decodeSearchJSON(body string) {
	js, err := simplejson.NewJson([]byte(body))
	if err != nil {
		panic("json format error")
	}

	items := js.Get("items")
	index := 0

	fmt.Printf("\r%21s\t%6s\t%s\n", "language", "stars", "url")
	fmt.Println("--------------------------------------")

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

func decodeCokJSON(body string) {
	items, err := simplejson.NewJson([]byte(body))
	if err != nil {
		panic("json format error")
	}

	index := 0

	fmt.Printf("\r%9s\t%2s\t%10s\t%10s\t%12s\n", "x", "y", "name", "time", "sid")
	fmt.Println("-------------------------------------------------------------")

	for {
		item := items.GetIndex(index)

		xPosition := item.Get("x").MustInt()
		yPosition := item.Get("y").MustInt()
		name := item.Get("name").MustString()
		date := item.Get("lasttime").MustInt64()
		sid := item.Get("sid").MustInt()

		if name == "" && xPosition == 0 && yPosition == 0 {
			return
		}

		tm := time.Unix(date, 0)

		fmt.Printf("%2v - %5d\t%3d\t%8s\t%10s\t%5d\n",
			index+1,
			xPosition,
			yPosition,
			name,
			fmtdate.Format("YYYY-MM-DD hh:mm:ss", tm),
			sid)

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

func post(url string, data string, out chan string) {
	resp, err := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader(data))

	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	out <- string(body)
}
