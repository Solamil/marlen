package marlen

import (
	"fmt"
	"encoding/json"
	"github.com/beevik/etree"
	"strings"
	"sync"
	"time"
)

type Article struct {
	Author string `json:"author"`
	Title string `json:"title"`
	Description string `json:"description"`
	LinkSite string `json:"linkSite"`
	Date string `json:"date"`

}

type Feed struct {
	Title string `json:"title"`
	LinkSite string `json:"linkSite"`
	ArtList []Article `json:"articles"`
	Class string `json:"class"`
}

func AtomFeedRoutine(url string, answer chan Feed, wg *sync.WaitGroup) {
	defer wg.Done()
	answer <- AtomFeed(url)
}

func AtomFeed(url string) Feed {
	var result Feed
	signature := fmt.Sprintf(`%s:%s`, url, "rssFeed")
	if record, found := Get(signature); found && record.Value != "" {
		now := time.Now()
		d := record.Expiry
		d = d.Add(time.Hour * 6)
		json.Unmarshal([]byte(record.Value), &result)
		if d.After(now) {
			return result
		}
	}
	resp := NewRequest(url)
	if resp == "" {
		return result
	}
	doc := etree.NewDocument()

	if err := doc.ReadFromString(resp); err != nil {
		fmt.Println(err)
		return Feed{}
	}

	//	if err := doc.ReadFromFile("BwJLymVb.atom"); err != nil {
	//		fmt.Println(err)
	//		return ""
	//	}
	root := doc.SelectElement("feed")
	mainTitle := root.SelectElement("title").Text()
	linkSite := root.SelectElement("link").SelectAttrValue("href", "")
	var feed Feed = Feed{mainTitle, linkSite, []Article{}, ""}
	for _, e := range root.SelectElements("entry") {
		title := e.SelectElement("title").Text()
		author := e.SelectElement("author")
		name := author.SelectElement("name").Text()
		published := ""
		if e.SelectElement("published") != nil {
			published = e.SelectElement("published").Text()
		}
		link := e.SelectElement("link").SelectAttrValue("href", "")
		feed.ArtList = append(feed.ArtList, Article{name, title, "", link, published})
	}

	byteResult, _ := json.Marshal(feed)
	Store(signature, string(byteResult))
	return feed 
}

func RssCtkRoutine(url string, nTitles int, showDescription bool, answer chan Feed, wg *sync.WaitGroup) {
	defer wg.Done()
	answer <- RssCtk(url, nTitles, showDescription)
}

func RssCtk(url string, nTitles int, showDescription bool) Feed {
	var result Feed
	signature := fmt.Sprintf(`%s:%s`, url, "rssFeed")
	if record, found := Get(signature); found {
		now := time.Now()
		d := record.Expiry
		if record.Value != "" {
			d = d.Add(time.Hour * 2)
		} else {
			d = d.Add(time.Minute * 35)
		}
		json.Unmarshal([]byte(record.Value), &result)
		if d.After(now) {
			return result
		}
	}
	doc := etree.NewDocument()
	//	if err := doc.ReadFromFile("cr.rss"); err != nil {
	//		fmt.Println(err)
	//		return ""
	//	}
	resp := NewRequest(url)
	if resp == "" {
		return result
	}
	if err := doc.ReadFromString(resp); err != nil {
		fmt.Println(err)
		return Feed{}
	}

	root := doc.SelectElement("rss").SelectElement("channel")
	mainTitle := root.SelectElement("title").Text()
	linkSite := root.SelectElement("link").Text()
	if nTitles < 1 || nTitles > 100 {
		nTitles = 5
	}
	var feed Feed = Feed{mainTitle, linkSite, []Article{}, ""}
	var size int = nTitles
	for i, e := range root.SelectElements("item") {
		if i >= size {
			break
		}
		title := e.SelectElement("title").Text()
		published := ""
		if e.SelectElement("pubDate") != nil {
			published = e.SelectElement("pubDate").Text()
		}
		link := e.SelectElement("link").Text()
		//		t, _ := time.Parse(time.RFC3339, published)
		description := ""
		if showDescription {
			description = e.SelectElement("description").Text()
		} else {
			description = ""
		}
		name := ""
		feed.ArtList = append(feed.ArtList, Article{name, title, description, link, published})
	}
	byteResult, _ := json.Marshal(feed)
	Store(signature, string(byteResult))

	return feed 
}

func RssCrashnet(url string, firstTitle string, linkSite string, nTitles int, answer chan string, wg *sync.WaitGroup) string {
	defer wg.Done()
	var result string = ""
	var signature string = fmt.Sprintf(`%s:%s`, url, "rssCrashnet")
	if record, found := Get(signature); found {
		now := time.Now()
		d := record.Expiry
		if record.Value != "" {
			d = d.Add(time.Hour * 2)
		} else {
			d = d.Add(time.Minute * 35)
		}
		result = record.Value
		if d.After(now) {
			answer <- result
			return result
		}
	}
	doc := etree.NewDocument()

	resp := NewRequest(url)
	if resp == "" {
		answer <- result
		Store(signature, result)
		return result
	}
	if err := doc.ReadFromString(resp); err != nil {
		fmt.Println(err)
		answer <- result
		return result
	}

	root := doc.SelectElement("rss").SelectElement("channel")
	mainTitle := fmt.Sprintf("&#128220;%s", firstTitle) // root.SelectElement("title").Text()
	// linkSite := "https://crash.net"				// root.SelectElement("link").Text()
	result = fmt.Sprintf("<div class=\"articles\" style=\"margin:5px;\">\n<h4><a href=\"%s\" target=\"_blank\">%s</a></h4><ul>\n", linkSite, mainTitle)
	if nTitles < 1 || nTitles > 100 {
		nTitles = 5
	}

	for i, e := range root.SelectElements("item") {
		if i >= nTitles {
			break
		}
		date := ""
		if e.SelectElement("pubDate") != nil {
			published := e.SelectElement("pubDate").Text()
			date = fmt.Sprintf("<span class=\"date\">%s</span>", published)
		}
		title := e.SelectElement("title").Text()
		link := e.SelectElement("link").Text()
		line := fmt.Sprintf("<li><a href=\"%s\" target=\"_blank\" style=\"display: block;\">%s &#128220;%s</a></li>\n",
			link, title, date)

		result = fmt.Sprintf("%s\n%s", result, line)
	}
	result = fmt.Sprintf("%s\n</ul></div>", result)
	Store(signature, result)

	answer <- result
	return result
}

func RssLocalplaceRoutine(url string, nTitles int, tannoy, showDescription bool, answer chan []Article, wg *sync.WaitGroup) {
	defer wg.Done()
	answer <- RssLocalplace(url, nTitles, tannoy, showDescription)
}

func RssLocalplace(url string, nTitles int, tannoy, showDescription bool) []Article {
	var signature string = fmt.Sprintf(`%s:%s`, url, "rssArticle")
	if tannoy {
		signature = fmt.Sprintf(`%s:%s`, url, "rssTannoy")
	}
	var result []Article
	if record, found := Get(signature); found {
		now := time.Now()
		d := record.Expiry
		if record.Value != "" {
			d = d.Add(time.Hour * 2)
		} else {
			d = d.Add(time.Minute * 35)
		}
		json.Unmarshal([]byte(record.Value), &result)
		if d.After(now) {
			return result
		}
	}
	doc := etree.NewDocument()
	var resp string = ""
	signatureResp := fmt.Sprintf(`%s:%s`, url, "rssResp")
	if record, found := Get(signatureResp); found {
		now := time.Now()
		d := record.Expiry
		if record.Value != "" {
			d = d.Add(time.Hour * 2)
		} else {
			d = d.Add(time.Minute * 35)
		}
		if d.After(now) {
			resp = record.Value
		} else {
			resp = NewRequest(url)
			Store(signatureResp, resp)
		}
	} else {
		resp = NewRequest(url)
		Store(signatureResp, resp)
	}
	if resp == "" {
		Store(signature, "")
		return []Article{}
	}
	if err := doc.ReadFromString(resp); err != nil {
		fmt.Println(err)
		return []Article{}
	}

	var artList []Article
	root := doc.SelectElement("rss")
	if !tannoy {
		if nTitles < 1 || nTitles > 10 {
			nTitles = 3
		}
		var size int = nTitles
		var i int = 0
		for _, e := range root.SelectElements("item") {
			if i >= size {
				break
			}
			title := e.SelectElement("title").Text()
			_, _, found := strings.Cut(title, "Hlášení rozhlasu")
			if found {
				continue
			}
			link := e.SelectElement("link").Text()
			artList = append(artList, Article{"", title, "", link, ""})
			i++
		}
	} else {
		if nTitles < 1 || nTitles > 10 {
			nTitles = 3
		}
		var size int = nTitles
		var i int = 0
		for _, e := range root.SelectElements("item") {
			if i >= size {
				break
			}
			title := e.SelectElement("title").Text()
			_, new_title, found := strings.Cut(title, "Hlášení rozhlasu")
			if !found {
				continue
			}
			var description string = ""
			if showDescription {
				description = e.SelectElement("description").Text()
			}
			artList = append(artList, Article{"", new_title, description, "", ""})
			i++
		}

	}
	byteResult, _ := json.Marshal(artList)

	Store(signature, string(byteResult))
	return artList 
}
