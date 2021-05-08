package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type Book struct {
	Type         string
	FileName     string
	Name         string
	InitName     string
	PosterUrl    string
	Year         string
	Genres       []string
	Authors      []string
	Countries    []string
	Publisher    []string
	Summary      string
	GoodreadsUrl string
	FlibustaUrl  string
	LitresUrl    string
}

func (book *Book) Print() {
	fmt.Printf("Name:           %s\n", book.Name)
	fmt.Printf("Original Title: %s\n", book.InitName)
	fmt.Printf("Picture:        %s\n", book.PosterUrl)
	for _, a := range book.Genres {
		fmt.Printf("Genre:          %s\n", a)
	}
	for _, d := range book.Authors {
		fmt.Printf("Author:       %s\n", d)
	}
	fmt.Printf("Publisher:      %s\n", book.Publisher)
	for _, c := range book.Countries {
		fmt.Printf("Country:        %s\n", c)
	}
	fmt.Printf("Goodreads:      %s\n", book.GoodreadsUrl)
	fmt.Printf("Flibusta:       %s\n", book.FlibustaUrl)
	fmt.Printf("Litres:         %s\n", book.LitresUrl)
	fmt.Printf("Summary:\n")
	fmt.Printf("%s\n", book.Summary)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var Translations = map[string]string{
	"Канада":              "Canada",
	"СССР":                "USSR",
	"США":                 "USA",
	"детектив":            "detective",
	"драма":               "drama",
	"комедия":             "comedy",
	"мелодрама":           "melodrama",
	"мультфильм":          "cartoon",
	"мюзикл":              "musical",
	"научная фантастика":  "science fiction",
	"приключение":         "adventures",
	"приключения":         "adventures",
	"семейный":            "family",
	"сказка":              "fairy tale",
	"стимпанк":            "steampunk",
	"фэнтези":             "fantasy",
	"экранизация":         "film adaptation",
	"юридический триллер": "legal thriller",
}

func PrintList(w *bufio.Writer, title string, lst []string) {
	if len(lst) == 0 {
		return
	}

	_, err := fmt.Fprintf(w, title)
	check(err)

	text := ""
	for _, str := range lst {
		if len(text) > 0 {
			text += ", "
		}
		text += "[[" + str + "]]"
	}

	_, err = fmt.Fprintf(w, " %s", text)
	check(err)
	_, err = fmt.Fprintf(w, "\n")
	check(err)
}

func (book *Book) PrintMarkdown() {
	f, err := os.Create(book.FileName)
	check(err)
	defer f.Close()

	w := bufio.NewWriter(f)

	_, err = fmt.Fprintf(w, "---\n")
	check(err)
	_, err = fmt.Fprintf(w, "created: %s\n", time.Now().Format("2006-01-02 15:04"))
	check(err)
	_, err = fmt.Fprintf(w, "alias: \"%s (%s)\"\n", book.Name, book.Year)
	check(err)
	_, err = fmt.Fprintf(w, "---\n\n")
	check(err)

	_, err = fmt.Fprintf(w, "<div style=\"float:right; padding: 10px\"><img width=200px src=\"%s\"/></div>\n\n", book.PosterUrl)
	check(err)
	_, err = fmt.Fprintf(w, "![[book.png|50]]\n")
	check(err)
	_, err = fmt.Fprintf(w, "# %s\n", book.Name)
	check(err)

	_, err = fmt.Fprintf(w, "**original name:** %s\n", book.InitName)
	check(err)
	_, err = fmt.Fprintf(w, "**year:** #y%s\n", book.Year)
	check(err)
	_, err = fmt.Fprintf(w, "**type:** #%s\n", book.Type)
	check(err)
	_, err = fmt.Fprintf(w, "**status:** #inbox\n")
	check(err)
	_, err = fmt.Fprintf(w, "**rate:**\n")
	check(err)

	PrintList(w, "**author:**", book.Authors)
	_, err = fmt.Fprintf(w, "**publisher:** [[%s]]]\n", book.Publisher)
	check(err)
	PrintList(w, "**country:**", book.Countries)
	PrintList(w, "**tags:**", book.Genres)

	_, err = fmt.Fprintf(w, "**[goodreads](%s)**\n", book.GoodreadsUrl)
	check(err)
	_, err = fmt.Fprintf(w, "**[flibusta](%s)**\n", book.FlibustaUrl)
	check(err)
	_, err = fmt.Fprintf(w, "**[litres](%s)**\n", book.LitresUrl)
	check(err)

	_, err = fmt.Fprintf(w, "\n---\n\n")
	check(err)

	_, err = fmt.Fprintf(w, "## Summary\n")
	check(err)
	_, err = fmt.Fprintf(w, "%s\n\n", book.Summary)
	check(err)
	_, err = fmt.Fprintf(w, "## Review\n\n")
	check(err)
	_, err = fmt.Fprintf(w, "## What attracted attention\n\n")
	check(err)
	_, err = fmt.Fprintf(w, "## Who might be interested\n\n")
	check(err)

	_, err = fmt.Fprintf(w, "## Links\n\n")
	check(err)

	w.Flush()
}

func firstRune(str string) (r rune) {
	for _, r = range str {
		return
	}
	return
}

func ParseList(html string) []string {
	var list []string
	lines := strings.Split(html, "<br/>")
	for _, line := range lines {
		dom, err := goquery.NewDocumentFromReader(strings.NewReader(line))
		if err == nil {
			dom.Each(func(i int, sel *goquery.Selection) {
				div := regexp.MustCompile(`[;,:\[\]]`)
				for _, item := range div.Split(sel.Text(), -1) {
					item = strings.Trim(item, " \t")
					if len(item) == 0 {
						continue
					} else if unicode.IsUpper(firstRune(item)) {
						list = append(list, item)
					}
				}
			})
		}
	}
	//fmt.Printf("item: %v\n", list)
	return list
}

func VisitGoodreads(link string) Book {

	var book Book

	book.Type = "book"
	book.GoodreadsUrl = link

	c := colly.NewCollector(
		colly.AllowedDomains("goodreads.com"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "goodreads.com/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	c.OnHTML(".infobox tbody tr:nth-child(1)", func(e *colly.HTMLElement) {
		if len(book.Name) == 0 {
			book.Name = e.Text
		}
	})

	/* 	c.OnHTML(".infobox tbody tr:nth-child(2)", func(e *colly.HTMLElement) {
	   		s := e.Text
	   		s = strings.TrimPrefix(s, "англ.")
	   		s = strings.TrimLeftFunc(s, func(r rune) bool {
	   			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	   		})
	   		if strings.HasPrefix(s, "A") && unicode.IsSpace([]rune(s)[1]) {
	   			s = s[1:]
	   			s = strings.TrimLeftFunc(s, func(r rune) bool {
	   				return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	   			})
	   			s = s + ", A"
	   		}
	   		if strings.HasPrefix(s, "The") && unicode.IsSpace([]rune(s)[3]) {
	   			s = s[3:]
	   			s = strings.TrimLeftFunc(s, func(r rune) bool {
	   				return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	   			})
	   			s = s + ", The"
	   		}
	   		if book.InitName == "" {
	   			book.InitName = strings.TrimSpace(s)
	   		}
	   	})

	   	c.OnHTML(".infobox-image a img[srcset]", func(e *colly.HTMLElement) {
	   		link := e.Attr("srcset")
	   		decodedLink, _ := url.QueryUnescape(link)
	   		v := strings.Split(decodedLink, " ")
	   		if len(book.PosterUrl) == 0 {
	   			book.PosterUrl = "https:" + v[0]
	   		}
	   	})

	   	c.OnHTML(".infobox tbody tr", func(e *colly.HTMLElement) {
	   		title := e.ChildText("th")
	   		if title == "Жанр" {
	   			e.ForEach("td a[href]", func(i int, a *colly.HTMLElement) {
	   				if strings.HasPrefix(a.Text, "[") {
	   					return
	   				}
	   				if a.Text == "экранизация" {
	   					return
	   				}
	   				trans := Translations[a.Text]
	   				if len(trans) == 0 {
	   					fmt.Printf("can't translate: %s\n", a.Text)
	   					book.Genres = append(book.Genres, a.Text)
	   				} else {
	   					book.Genres = append(book.Genres, trans+"|"+a.Text)
	   				}
	   			})
	   		}
	   		if title == "Сезонов" {
	   			book.Type = "serial"
	   		}
	   		if title == "Режиссёр" {
	   			e.ForEach("td span", func(i int, a *colly.HTMLElement) {
	   				html, _ := a.DOM.Html()
	   				book.Directors = ParseList(html)
	   			})
	   		}
	   		if title == "Продюсер" {
	   			e.ForEach("td span", func(i int, a *colly.HTMLElement) {
	   				html, _ := a.DOM.Html()
	   				book.Producers = ParseList(html)
	   			})
	   		}
	   		if strings.Contains(title, "Автор") && strings.Contains(title, "сценария") {
	   			e.ForEach("td span", func(i int, a *colly.HTMLElement) {
	   				html, _ := a.DOM.Html()
	   				book.Screenwriters = ParseList(html)
	   			})
	   		}
	   		if title == "Кинокомпания" || title == "Студия" {
	   			e.ForEach("td span", func(i int, a *colly.HTMLElement) {
	   				html, _ := a.DOM.Html()
	   				book.Companies = ParseList(html)
	   			})
	   		}
	   		if title == "Страна" {
	   			e.ForEach("td a", func(i int, a *colly.HTMLElement) {
	   				if a.Text != "" {
	   					trans := Translations[a.Text]
	   					if len(trans) == 0 {
	   						fmt.Printf("can't translate: %s\n", a.Text)
	   						book.Countries = append(book.Countries, a.Text)
	   					} else {
	   						book.Countries = append(book.Countries, trans+"|"+a.Text)
	   					}
	   				}
	   			})
	   		}
	   		if title == "Год" || title == "Премьера" {
	   			r, _ := regexp.Compile("[0-9][0-9][0-9][0-9]")
	   			e.ForEachWithBreak("td", func(i int, a *colly.HTMLElement) bool {
	   				s := r.FindString(a.Text)
	   				if s != "" {
	   					book.Year = s
	   					return false
	   				}
	   				return true
	   			})
	   		}
	   		if title == "На экранах" {
	   			r, _ := regexp.Compile("[0-9][0-9][0-9][0-9]")
	   			e.ForEachWithBreak("td", func(i int, a *colly.HTMLElement) bool {
	   				s := r.FindString(a.Text)
	   				if s != "" {
	   					book.Year = s
	   					return false
	   				}
	   				return true
	   			})
	   		}
	   	}) */

	c.Visit(book.GoodreadsUrl)
	return book
}

func VisitLitres(link string) (string, string) {

	var summary string
	var picture string

	c := colly.NewCollector(
		colly.AllowedDomains("litres.ru"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "litres.ru/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	c.OnHTML("div.p-book-info__content p", func(e *colly.HTMLElement) {
		summary = e.Text
	})

	/* 	c.OnHTML("div.p-book-info img.p-picture__image[src]", func(e *colly.HTMLElement) {
		picture = e.Attr("src")
	}) */

	c.Visit(link)
	return summary, picture
}

func SearchGoogle(query string, site string) string {
	s := "-w " + site
	out, err := exec.Command("googler", "-n 1", "--np", "--json", s, query).Output()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	var dat []map[string]interface{}

	if err := json.Unmarshal(out, &dat); err != nil {
		panic(err)
	}

	var str string
	for _, d := range dat {
		str, _ = url.QueryUnescape(d["url"].(string))
	}
	return str
}

func ScrapeBookInner(goodreads string, flibusta string, litres string) Book {
	var book Book
	var pic string

	w, _ := url.QueryUnescape(goodreads)
	book = VisitGoodreads(w)
	book.GoodreadsUrl = w
	book.FlibustaUrl = flibusta
	book.LitresUrl = litres
	book.Summary, pic = VisitLitres(litres)
	if len(book.PosterUrl) == 0 {
		book.PosterUrl = pic
	}

	name := strings.TrimSpace(book.InitName)
	if name == "" {
		name = book.Name
	}
	book.FileName = name + " (" + book.Year + ").md"
	return book
}

func ScrapeBook(query string) Book {
	goodreads := SearchGoogle(query, "goodreads.com")
	time.Sleep(1 * time.Second)
	flibusta := SearchGoogle(query, "flibusta.is")
	time.Sleep(1 * time.Second)
	litres := SearchGoogle(query, "litres.ru")
	time.Sleep(1 * time.Second)

	fmt.Println(goodreads)
	fmt.Println(flibusta)
	fmt.Println(litres)

	return ScrapeBookInner(goodreads, flibusta, litres)
}

func main() {
	var book Book

	reader := bufio.NewReader(os.Stdin)
	book = ScrapeBook(os.Args[1])
	book.Print()
	fmt.Printf("=======\n")
	fmt.Printf("Save markdown file \"" + book.FileName + "\"? (yes/no)> [yes]")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSuffix(text, "\n")
	if text == "" || text == "yes" {
		book.PrintMarkdown()
		fmt.Println("file \"" + book.FileName + "\" created")
	}
}
