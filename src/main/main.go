package main

import (
	"crawler"
	"fmt"
)

const (
	kMainUrl  = "http://www.biquge.info/1_1245/"
	kFilename = "剑来.txt"
)

func main() {
	err := crawler.Capture(kMainUrl, kFilename)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("capture success!")
	}
}
