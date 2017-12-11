package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/sclevine/agouti"
)

const (
	endpoint = "https://prize.travel.rakuten.co.jp/frt/"
)

var (
	Mail       string
	Passwd     string
	compelinks []string
	alllinks   []string
	wg         sync.WaitGroup
)

func LinkScrape() {
	var links []string
	var largeurl string
	flag.StringVar(&largeurl, "url", "", "rakuten travel url")
	flag.Parse()
	if len(os.Args) != 2 {
		log.Fatalf("Please set url -url='*******'")
	}
	if largeurl == "" {
		log.Fatalf("url is nothing")
	}

	doc, _ := goquery.NewDocument(largeurl)
	doc.Find("#result > div.hotels > div").Each(func(_ int, s *goquery.Selection) {
		l, _ := s.Attr("onclick")
		links = append(links, l)
	})

	for i, _ := range links {
		compelinks = append(compelinks, endpoint+strings.Trim(links[i], "location.href=' '; return false"))
	}
}

func AllLinkExtraction(url string) {
	doc, _ := goquery.NewDocument(url)
	doc.Find("p.RthPresentBt > a").Each(func(_ int, s *goquery.Selection) {
		l, _ := s.Attr("href")
		alllinks = append(alllinks, l)
	})
}

func init() {
	Mail = os.Getenv("MY_MAIL")
	Passwd = os.Getenv("MY_PASSWORD")
	if Mail == "" || Passwd == "" {
		log.Fatalf("Please set the environment value MY_MAIL and MY_PASSWORD\n")
	}
}

func main() {
	LinkScrape()

	for i, _ := range compelinks {
		AllLinkExtraction(compelinks[i])
	}

	for i, _ := range alllinks {
		wg.Add(1)
		go func(url string, i int) {
			defer wg.Done()
			driver := agouti.ChromeDriver(agouti.Browser("chrome"))
			if err := driver.Start(); err != nil {
				log.Fatalf("Failed to start driver:%v", err)
			}
			defer driver.Stop()

			page, err := driver.NewPage()
			if err != nil {
				log.Fatalf("Failed to open page: %s", err)
				driver.Stop()
			}
			if err := page.Navigate(url); err != nil {
				log.Fatalf("Failed to navigate: %s", err)
				driver.Stop()
			}

			// Login
			mail := page.FindByName("u")
			pass := page.FindByName("p")
			mail.Fill(Mail)
			pass.Fill(Passwd)

			if err := page.Find("#login > #l_rakuten > #l_submit > a").Submit(); err != nil {
				log.Fatalf("login submit err: %s", err)
				driver.Stop()
			}

			// uncheck checkbox
			page.Find("td > table > tbody > tr:nth-child(4) > td > form").First("input[type=checkbox]:nth-child(7)").Uncheck()
			page.Find("td > table > tbody > tr:nth-child(4) > td > form").First("input[type=checkbox]:nth-child(8)").Uncheck()

			if err := page.Find("td > table > tbody > tr:nth-child(4) > td > form").First("input[type=submit]:nth-child(3)").Click(); err != nil {
				fmt.Printf("[aready done] %s\n", url)
				driver.Stop()
			}

			fmt.Printf("[Done] %s\n", url)

		}(alllinks[i], i)
	}
	wg.Wait()
}
