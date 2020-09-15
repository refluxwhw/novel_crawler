package crawler

import (
	"fmt"
	"os"
	"strings"
	"time"

	"io/ioutil"
	"math/rand"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type Chapter struct {
	title string
	url   string
}

func Capture(mainUrl string, filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// main page
	chapters, err := parseChapters(mainUrl)
	if err != nil {
		return err
	}

	for _, c := range chapters {
		c.title = fixTitle(c.title)
		context := ""

		errTimes := 0
		for {
			context, err = parseContent(c.url)
			if err != nil {
				fmt.Println(err)
				errTimes++
				if errTimes > 10 {
					return err
				}

				n := rand.Uint32() % 10000
				fmt.Println("sleep ", n, "ms, and try again")
				time.Sleep(time.Millisecond * time.Duration(n))
				return err
			}
			break
		}

		fmt.Println(c.title, c.url, len(context))

		_, _ = file.WriteString(c.title)
		_, _ = file.WriteString("\n")
		_, _ = file.WriteString(context)
		_, _ = file.WriteString("\n\n")

		// 随机休眠一段时间，减少被检测抓包的几率
		n := rand.Uint32() % 500
		time.Sleep(time.Millisecond * time.Duration(n))
	}

	return nil
}

func fixTitle(title string) string {
	sl := strings.Split(title, " ")
	if sl[0] == "第一卷" && len(sl) > 3 {
		title = sl[2] + " " + sl[3]
		sl = sl[2:]
	}
	return strings.Join(sl, " ")
}

func parseChapters(url string) (chapters []*Chapter, err error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	chapters = make([]*Chapter, 0, 1000)

	doc.Find("#list").Find("dd a").EachWithBreak(func(i int, selection *goquery.Selection) bool {
		title := selection.Text()
		href, _ := selection.Attr("href")
		c := &Chapter{
			title: title,
			url:   url + href,
		}
		chapters = append(chapters, c)
		return true
	})

	return chapters, nil
}

func parseContent(url string) (content string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("response error, code=%d, message=%s", resp.StatusCode, string(b))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	node := doc.Find("#content")
	node.Find("br").ReplaceWithHtml("\n")
	content = node.Text()

	if len(content) == 0 {
		fmt.Println(doc.Text())
		return "", fmt.Errorf("content is empty")
	}

	return content, nil
}
