package marlen

import (
	"fmt"
	"sync"
	"time"
	"strings"
	"github.com/beevik/etree"
)

func AtomFeed(url string, answer chan string, wg* sync.WaitGroup) string {
	defer wg.Done()
	var result string = ""
	signature := fmt.Sprintf(`%s:%s`, url, "rssFeed")
	if record, found := Get(signature); found && record.Value != "" {
		now := time.Now()
		d := record.Expiry
		d = d.Add(time.Hour * 6)
		result = record.Value
		if d.After(now) {
			answer <- result
			return result
		}
	}
	resp := NewRequest(url)
	if resp == "" {
		answer <- result
		return result
	}
	doc := etree.NewDocument()

	if err := doc.ReadFromString(resp); err != nil {
		fmt.Println(err)
		answer <- ""
		return ""
	}

	//	if err := doc.ReadFromFile("BwJLymVb.atom"); err != nil {
	//		fmt.Println(err)
	//		return ""
	//	}
	root := doc.SelectElement("feed")
	mainTitle := root.SelectElement("title").Text()
	linkSite := root.SelectElement("link").SelectAttrValue("href", "")
	result = fmt.Sprintf("<h3><a href=\"%s\" target=\"_blank\">%s</a></h3>\n<ul>", linkSite, mainTitle)
	for _, e := range root.SelectElements("entry") {
		title := e.SelectElement("title").Text()
		author := e.SelectElement("author")
		name := author.SelectElement("name").Text()
		date := ""
		if e.SelectElement("published") != nil {
			published := e.SelectElement("published").Text()
			date = fmt.Sprintf("<span class=\"date\">%s</span>", published)

		}
		link := e.SelectElement("link").SelectAttrValue("href", "")
		//		t, _ := time.Parse(time.RFC3339, published)
		// 	✏️ &#9999;📜&#128220;
		line := fmt.Sprintf(`<li><a href="%s" target="_blank">%s &#9999;%s &#128220;%s</a></li>`, link, date, name, title)
		result = fmt.Sprintf("%s\n%s", result, line)

	}
	result = fmt.Sprintf("%s\n</ul>", result)
	Store(signature, result)
	answer <- result
	return result
}


func RssCtk(url string, nTitles int, showDescription bool, answer chan string, wg* sync.WaitGroup) string {
	defer wg.Done()
	var result string = ""
	signature := fmt.Sprintf(`%s:%s`, url, "rssFeed")
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
	//	if err := doc.ReadFromFile("cr.rss"); err != nil {
	//		fmt.Println(err)
	//		return ""
	//	}
	resp := NewRequest(url)
	if resp == "" {
		Store(signature, result)
		answer <- result
		return result
	}
	if err := doc.ReadFromString(resp); err != nil {
		fmt.Println(err)
		answer <- ""
		return ""
	}

	root := doc.SelectElement("rss").SelectElement("channel")
	mainTitle := root.SelectElement("title").Text()
	linkSite := root.SelectElement("link").Text()
	result = fmt.Sprintf("<div>\n<h3><a href=\"%s\" target=\"_blank\">%s</a></h3>\n<ul>", linkSite, mainTitle)
	if nTitles < 1 || nTitles > 100 {
		nTitles = 5
	}
	var size int = nTitles
	for i, e := range root.SelectElements("item") {
		if i >= size {
			break
		}
		title := e.SelectElement("title").Text()
		date := ""
		if e.SelectElement("pubDate") != nil {
			published := e.SelectElement("pubDate").Text()
			date = fmt.Sprintf("<span class=\"date\">%s</span>", published)
		}
		link := e.SelectElement("link").Text()
		//		t, _ := time.Parse(time.RFC3339, published)
		// 	✏️ &#9999;📜&#128220;
		var line string = ""
		if showDescription {
			description := e.SelectElement("description").Text()
			line = fmt.Sprintf("<li><h4><a href=\"%s\" target=\"_blank\" class=\"ctk\">%s &#128220;%s"+
				"</a></h4>\n"+
				"<p>%s<p>\n"+
				"</li>", link, date, title, description)
		} else {
			line = fmt.Sprintf("<li><a href=\"%s\" target=\"_blank\">%s &#128220;%s</a>\n"+
				"</li>", link, date, title)
		}
		result = fmt.Sprintf("%s\n%s", result, line)
	}
	result = fmt.Sprintf("%s\n</ul></div>", result)
	Store(signature, result)

	answer <- result
	return result
}


func RssCrashnet(url string, firstTitle string, linkSite string, nTitles int, answer chan string, wg* sync.WaitGroup) string {
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
	mainTitle := fmt.Sprintf("&#128220;%s", firstTitle)		// root.SelectElement("title").Text() 
	// linkSite := "https://crash.net"				// root.SelectElement("link").Text()
	result = fmt.Sprintf("<div class=\"articles\" style=\"margin:5px;\">\n<h4><a href=\"%s\" target=\"_blank\">%s</a></h4><ul>\n", linkSite, mainTitle)
	if nTitles < 1 || nTitles > 100 {
		nTitles = 5
	}

	for i, e := range root.SelectElements("item") {
		if i >=	nTitles {
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

func RssLocalplaceRoutine(url string, nTitles int, tannoy, showDescription bool, answer chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	answer <-RssLocalplace(url, nTitles, tannoy, showDescription)	
}

func RssLocalplace(url string, nTitles int, tannoy, showDescription bool) string {
	var result string = ""
	var signature string = fmt.Sprintf(`%s:%s`, url, "rssArticles")
	if tannoy {
		signature = fmt.Sprintf(`%s:%s`, url, "rssTannoy")
	}
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
			return result
		}
	}
	doc := etree.NewDocument()
	//	if err := doc.ReadFromFile("cr.rss"); err != nil {
	//		fmt.Println(err)
	//		return ""
	//	}
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
		Store(signature, result)
		return result
	}
	if err := doc.ReadFromString(resp); err != nil {
		fmt.Println(err)
		return ""
	}

	root := doc.SelectElement("rss")
	if !tannoy {
		mainTitle := "📜Články města"
		linkSite := "https://www.mnhradiste.cz/"
		result = fmt.Sprintf("<div class=\"articles\" style=\"margin: 5px;\">\n<h4><a href=\"%s\" target=\"_blank\">%s</a></h4>\n"+
			"<ul>", linkSite, mainTitle)
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
			if found {
				continue
			}
			link := e.SelectElement("link").Text()
			var line string = ""
			if showDescription {
				//	description := e.SelectElement("description").Text()
				line = fmt.Sprintf("<li><a href=\"%s\" target=\"_blank\">%s</a></li>"+
					"\n", link, title)
			} else {
				line = fmt.Sprintf("<a href=\"%s\" target=\"_blank\" style=\"display: block;\">%s</a>\n",
					link, new_title)
			}
			result = fmt.Sprintf("%s\n%s", result, line)
			i++
		}
		result = fmt.Sprintf("%s\n</ul></div>", result)
	} else {
		mainTitle := "📣Hlášení rozhlasu"
		linkSite := "https://www.mnhradiste.cz/radnice/komunikace-s-obcany/hlaseni-rozhlasu"
		result = fmt.Sprintf("<div class=\"tannoy\" style=\"margin:5px;\">\n<h4><a href=\"%s\" target=\"_blank\">%s</a></h4>\n", linkSite, mainTitle)
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
			link := e.SelectElement("link").Text()
			var line string = ""
			if showDescription {
				description := e.SelectElement("description").Text()
				line = fmt.Sprintf("<details style=\"margin-left:30px;\"><summary>%s</summary>"+
					"\n"+
					"<p>%s</p>\n"+
					"</details>", new_title, description)
			} else {
				line = fmt.Sprintf("<a href=\"%s\" target=\"_blank\" style=\"display: block;\">%s</a>\n",
					link, new_title)
			}
			result = fmt.Sprintf("%s\n%s", result, line)
			i++
		}
		result = fmt.Sprintf("%s\n</div>", result)

	}
	Store(signature, result)
	return result
}
