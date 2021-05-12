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
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

var translator TagTranslator

type Person struct {
	FirstName  string
	MiddleName string
	LastName   string
	Initials   string
}

func (person *Person) PrintName() string {
	if person.FirstName == "" && person.Initials != "" {
		return person.LastName + " " + person.Initials
	} else {
		return person.LastName + ", " + person.FirstName
	}
}

type Book struct {
	Type         string
	FileName     string
	ShortName    string
	Name         string
	InitName     string
	PosterUrl    string
	Year         string
	Genres       map[string]string
	Tags         map[string]string
	Series       string
	Authors      []Person
	Painters     []Person
	Editors      []Person
	Translators  []Person
	Countries    []string
	Publisher    string
	Isbn         string
	Summary      string
	LabirintUrl  string
	GoodreadsUrl string
	FlibustaUrl  string
	LitresUrl    string
	OzonUrl      string
	LivelibUrl   string
}

func NewBook() Book {
	var book Book
	book.Type = "book"
	book.Genres = make(map[string]string)
	book.Tags = make(map[string]string)
	return book
}

func (book *Book) Print() {
	fmt.Printf("Name:           %s\n", book.Name)
	fmt.Printf("Original Title: %s\n", book.InitName)
	fmt.Printf("Year:           %s\n", book.Year)
	fmt.Printf("Picture:        %s\n", book.PosterUrl)
	for _, a := range book.Genres {
		fmt.Printf("Genre:          [%s]\n", book.GetGenre(a))
	}
	for _, d := range book.Authors {
		fmt.Printf("Author:       %s, %s\n", d.LastName, d.FirstName)
	}
	for _, d := range book.Painters {
		fmt.Printf("Painter:      %s, %s\n", d.LastName, d.FirstName)
	}
	for _, d := range book.Editors {
		fmt.Printf("Editor:       %s, %s\n", d.LastName, d.FirstName)
	}
	for _, d := range book.Translators {
		fmt.Printf("Translators:  %s, %s\n", d.LastName, d.FirstName)
	}
	fmt.Printf("Publisher:      %s\n", book.Publisher)
	for _, c := range book.Countries {
		fmt.Printf("Country:        %s\n", c)
	}
	for _, t := range book.Tags {
		fmt.Printf("Tag:            [%s]\n", book.GetTag(t))
	}

	fmt.Printf("Series:         %s\n", book.Series)
	fmt.Printf("ISBN:           %s\n", book.Isbn)
	fmt.Printf("Labirint:       %s\n", book.LabirintUrl)
	fmt.Printf("Goodreads:      %s\n", book.GoodreadsUrl)
	fmt.Printf("Flibusta:       %s\n", book.FlibustaUrl)
	fmt.Printf("Litres:         %s\n", book.LitresUrl)
	if book.OzonUrl != "" {
		fmt.Printf("Ozon:         %s\n", book.OzonUrl)
	}
	if book.LivelibUrl != "" {
		fmt.Printf("Livelib:      %s\n", book.LivelibUrl)
	}

	fmt.Printf("Summary:\n")
	fmt.Printf("%s\n", book.Summary)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func PrintMap(w *bufio.Writer, title string, lst map[string]string) {
	if len(lst) == 0 {
		return
	}

	_, err := fmt.Fprintf(w, title)
	check(err)

	text := ""
	for k, v := range lst {
		if len(text) > 0 {
			text += ", "
		}

		if v == "" || v == "#" {
			text += "[[" + k + "]]"
		} else {
			text += "[[" + v + "|" + k + "]]"
		}
	}

	_, err = fmt.Fprintf(w, " %s", text)
	check(err)
	_, err = fmt.Fprintf(w, "\n")
	check(err)
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

func PrintPersonsList(w *bufio.Writer, title string, lst []Person) {
	if len(lst) == 0 {
		return
	}

	_, err := fmt.Fprintf(w, title)
	check(err)

	text := ""
	for _, person := range lst {
		if len(text) > 0 {
			text += ", "
		}
		text += "[[" + person.PrintName() + "]]"
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

	PrintMap(w, "**genres:**", book.Genres)
	PrintPersonsList(w, "**author:**", book.Authors)
	PrintPersonsList(w, "**painter:**", book.Painters)
	PrintPersonsList(w, "**editor:**", book.Editors)
	PrintPersonsList(w, "**translators:**", book.Translators)
	_, err = fmt.Fprintf(w, "**publisher:** [[%s]]\n", book.Publisher)
	check(err)
	PrintList(w, "**country:**", book.Countries)
	if book.Series != "" {
		_, err = fmt.Fprintf(w, "**series:** [[%s]]\n", book.Series)
		check(err)
	}
	PrintMap(w, "**tags:**", book.Tags)

	_, err = fmt.Fprintf(w, "**isbn:** %s\n", book.Isbn)
	check(err)
	if book.LabirintUrl != "" {
		_, err = fmt.Fprintf(w, "**[labirint](%s)**\n", book.LabirintUrl)
		check(err)
	}
	if book.GoodreadsUrl != "" {
		_, err = fmt.Fprintf(w, "**[goodreads](%s)**\n", book.GoodreadsUrl)
		check(err)
	}
	if book.FlibustaUrl != "" {
		_, err = fmt.Fprintf(w, "**[flibusta](%s)**\n", book.FlibustaUrl)
		check(err)
	}
	if book.LitresUrl != "" {
		_, err = fmt.Fprintf(w, "**[litres](%s)**\n", book.LitresUrl)
		check(err)
	}
	if book.OzonUrl != "" {
		_, err = fmt.Fprintf(w, "**[ozon](%s)**\n", book.OzonUrl)
		check(err)
	}
	if book.LivelibUrl != "" {
		_, err = fmt.Fprintf(w, "**[livelib](%s)**\n", book.LivelibUrl)
		check(err)
	}

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

func (book *Book) InitFileName() {
	var authors string
	sort.Slice(book.Authors, func(i, j int) bool {
		return book.Authors[i].LastName < book.Authors[j].LastName
	})
	/* fmt.Printf("authors: %v\n", book.Authors) */
	l := len(book.Authors)
	if l > 0 {
		authors = book.Authors[0].PrintName()
	}
	if l == 2 {
		authors += " и " + book.Authors[1].PrintName()
	}
	if l > 2 {
		authors += " и др."
	}

	var name string
	if utf8.RuneCountInString(book.Name) <= 75 {
		name = book.Name
	} else if strings.Contains(book.Name, ".") {
		v := strings.Split(book.Name, ".")
		name = v[0]
	} else {
		name = book.Name[0:72] + "..."
	}

	book.FileName = authors + " - " + name + ".md"
}

func (book *Book) AppendGenre(genre string) {
	genre = strings.ToLower(genre)
	trans := translator.Translate(genre)
	if trans == "" {
		fmt.Printf("can't translate: %s\n", genre)
	}
	book.Genres[genre] = trans
}

func (book *Book) GetGenre(genre string) string {
	trans := book.Genres[genre]
	if trans == "" && trans != "#" {
		return trans + "|" + genre
	} else {
		return genre
	}
}

func (book *Book) AppendTag(tag string) {
	tag = strings.ToLower(tag)
	trans := translator.Translate(tag)
	if trans == "" {
		fmt.Printf("can't translate: %s\n", tag)
	}
	book.Tags[tag] = trans
}

func (book *Book) GetTag(tag string) string {
	trans := book.Tags[tag]
	if trans == "" && trans != "#" {
		return trans + "|" + tag
	} else {
		return tag
	}
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

func ParsePerson(str string, invert bool) Person {
	/* fmt.Printf("person: %s\n", str) */
	var person Person
	var v []string
	appendToLast := ""
	vec := strings.Split(str, " ")
	for _, item := range vec {
		r := []rune(item)
		count := utf8.RuneCountInString(item)
		if unicode.IsUpper(firstRune(item)) {
			if (count == 1) ||
				(count == 2 && r[1] == '.') ||
				(count == 4 && r[1] == '.' && r[3] == '.') {
				person.Initials += item
			} else {
				v = append(v, item)
			}
		} else {
			l := len(v)
			if l > 0 {
				v[l-1] += " " + item
				invert = !invert
			} else {
				appendToLast = item
			}
		}
	}

	if appendToLast != "" {
		l := len(v)
		if l > 0 {
			v[l-1] += " " + appendToLast
		}
	}

	if len(v) == 1 {
		person.LastName = v[0]
	} else if len(v) == 2 {
		if invert {
			person.FirstName = v[0]
			person.LastName = v[1]
		} else {
			person.FirstName = v[1]
			person.LastName = v[0]
		}
	} else if len(v) == 3 {
		person.FirstName = v[1]
		person.MiddleName = v[2]
		person.LastName = v[0]
	}
	return person
}

func VisitLabirint(link string) Book {

	book := NewBook()
	book.LabirintUrl = link

	c := colly.NewCollector(
		colly.AllowedDomains("www.labirint.ru"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "www.labirint.ru/books/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	c.OnHTML("#product-about h2", func(e *colly.HTMLElement) {
		if book.Name == "" {
			s := strings.TrimSpace(e.Text)
			s = strings.TrimPrefix(s, "Аннотация к книге \"")
			s = strings.TrimSuffix(s, "\"")
			book.Name = s
		}
	})

	c.OnHTML("#product-about p", func(e *colly.HTMLElement) {
		if book.Summary == "" {
			book.Summary = e.Text
		}
	})

	c.OnHTML("meta", func(e *colly.HTMLElement) {
		if e.Attr("property") == "og:image" {
			link := e.Attr("content")
			book.PosterUrl, _ = url.QueryUnescape(link)
		}
	})

	c.OnHTML(".authors", func(e *colly.HTMLElement) {
		if strings.HasPrefix(e.Text, "Автор: ") {
			e.ForEach("a[href]", func(i int, a *colly.HTMLElement) {
				/* fmt.Printf("author: %s\n", a.Text) */
				book.Authors = append(book.Authors, ParsePerson(a.Text, false))
			})
		}
		if strings.HasPrefix(e.Text, "Художник: ") {
			e.ForEach("a[href]", func(i int, a *colly.HTMLElement) {
				book.Painters = append(book.Painters, ParsePerson(a.Text, false))
			})
		}
		if strings.HasPrefix(e.Text, "Редактор: ") {
			e.ForEach("a[href]", func(i int, a *colly.HTMLElement) {
				book.Editors = append(book.Editors, ParsePerson(a.Text, false))
			})
		}
		if strings.HasPrefix(e.Text, "Переводчик: ") {
			e.ForEach("a[href]", func(i int, a *colly.HTMLElement) {
				book.Translators = append(book.Translators, ParsePerson(a.Text, false))
			})
		}
	})

	c.OnHTML(".publisher a", func(e *colly.HTMLElement) {
		book.Publisher = e.Text
	})

	c.OnHTML(".publisher", func(e *colly.HTMLElement) {
		r, _ := regexp.Compile("[0-9][0-9][0-9][0-9]")
		book.Year = r.FindString(e.Text)
	})

	c.OnHTML(".series a", func(e *colly.HTMLElement) {
		book.Series = e.Text
	})

	c.OnHTML(".isbn", func(e *colly.HTMLElement) {
		book.Isbn = strings.TrimPrefix(e.Text, "ISBN: ")
	})

	c.Visit(book.LabirintUrl)

	book.InitFileName()

	return book
}

func VisitOzon(link string) Book {

	book := NewBook()
	book.OzonUrl = link

	c := colly.NewCollector(
		colly.AllowedDomains("www.ozon.ru"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "www.ozon.ru/product/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Printf("Error: %v\n", err)
		/* log.Println("Something went wrong:", err) */
	})

	c.OnHTML("h1", func(e *colly.HTMLElement) {
		fmt.Printf("h1: %s\n", e.Text)
		if book.Name == "" {
			v := strings.Split(e.Text, " | ")
			if len(v) > 0 {
				book.Name = strings.TrimSpace(v[0])
			}
			if len(v) > 1 {
				book.Authors = append(book.Authors, ParsePerson(v[1], false))
			}
			fmt.Printf("title: %v\n", v)
		}
	})

	c.OnHTML(".a8n3 img", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		book.PosterUrl, _ = url.QueryUnescape(link)
	})

	c.OnHTML(".db8", func(e *colly.HTMLElement) {
		/* 		fmt.Printf("db8   : %s\n", e.Text)
		   		fmt.Printf("db8 dt: %s\n", e.ChildText("dt")) */
		if strings.Contains(e.ChildText("dt"), "Издательство") {
			book.Publisher = e.ChildText("dd")
		}
		if strings.Contains(e.ChildText("dt"), "Год выпуска") {
			book.Year = e.ChildText("dd")
		}
		if strings.Contains(e.ChildText("dt"), "Переводчик") {
			e.ForEach("dd a[href]", func(i int, a *colly.HTMLElement) {
				book.Translators = append(book.Translators, ParsePerson(a.Text, false))
			})
		}
		if strings.Contains(e.ChildText("dt"), "ISBN") {
			fmt.Printf("isbn: %s\n", e.Text)
			book.Isbn = e.ChildText("dd")
		}
	})

	/* 	c.OnResponse(func(res *colly.Response) {
		fmt.Printf("%v\n", res.StatusCode)
		fmt.Printf("%s\n", res.Body)
	}) */

	fmt.Printf("ozon: %s\n", book.OzonUrl)
	c.Visit(book.OzonUrl)

	book.InitFileName()

	return book
}

func VisitLivelib(link string) Book {

	book := NewBook()
	book.LivelibUrl = link

	c := colly.NewCollector(
		colly.AllowedDomains("www.livelib.ru"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "www.livelib.ru/book/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Printf("Error: %v\n", err)
	})

	c.OnHTML("h1", func(e *colly.HTMLElement) {
		if book.Name == "" {
			book.Name = strings.TrimSpace(e.Text)
		}
	})

	c.OnHTML("h2.bc-author", func(e *colly.HTMLElement) {
		e.ForEach("a[href].author-item", func(i int, a *colly.HTMLElement) {
			str := strings.TrimSpace(a.Text)
			fmt.Printf("author: %v\n", a.Text)
			book.Authors = append(book.Authors, ParsePerson(str, true))
		})
	})

	c.OnHTML("#main-image-book", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		book.PosterUrl, _ = url.QueryUnescape(link)
	})

	c.OnHTML("table.bc-edition tr td", func(e *colly.HTMLElement) {
		/* 		fmt.Printf("db8   : %s\n", e.Text)
		   		fmt.Printf("db8 dt: %s\n", e.ChildText("dt")) */
		if e.Attr("itemprop") == "publisher" {
			book.Publisher = strings.TrimSpace(e.Text)
		}
	})

	c.OnHTML(".bc-genre", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(i int, a *colly.HTMLElement) {
			book.AppendGenre(a.Text)
		})
	})

	c.OnHTML(".bc-info__wrapper div p", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "Жанры:") {
			e.ForEach("a[href]", func(i int, a *colly.HTMLElement) {
				book.AppendGenre(a.Text)
			})
		}
		/* 		else if strings.Contains(e.Text, "Теги:") {
			e.ForEach("a[href]", func(i int, a *colly.HTMLElement) {
				book.AppendGenre(a.Text)
			})
		} */
	})

	c.OnHTML(".bc-info div p span", func(e *colly.HTMLElement) {
		if e.Attr("itemprop") == "isbn" {
			book.Isbn = strings.TrimSpace(e.Text)
		}
	})

	c.OnHTML(".bc-info div p", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "Год издания:") {
			r, _ := regexp.Compile("[0-9][0-9][0-9][0-9]")
			s := r.FindString(e.Text)
			if s != "" {
				book.Year = s
			}
		}
	})

	c.OnHTML("#lenta-card__text-edition-escaped p", func(e *colly.HTMLElement) {
		if e.Attr("itemprop") == "description" {
			book.Summary = e.Text
		}
	})

	/* 	c.OnResponse(func(res *colly.Response) {
		fmt.Printf("%v\n", res.StatusCode)
		fmt.Printf("%s\n", res.Body)
	}) */

	c.Visit(book.LivelibUrl)

	book.InitFileName()

	return book
}

func VisitGoodreads(link string) Book {

	book := NewBook()
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

	c.Visit(book.GoodreadsUrl)
	return book
}

func VisitLitres(book *Book, link string) {

	c := colly.NewCollector(
		colly.AllowedDomains("www.litres.ru"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "www.litres.ru/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	c.OnResponse(func(res *colly.Response) {
		re, _ := regexp.Compile(`(?U)<div class="biblio_book_info">(.*)<\/div>`)
		m := re.FindSubmatch([]byte(res.Body))
		if m != nil {
			s := strings.TrimSpace(string(m[1]))
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
			if err != nil {
				log.Fatal(err)
			}
			doc.Find("li").Each(func(i int, s *goquery.Selection) {
				title := s.Find("strong").Text()
				if title == "Теги:" || title == "Жанр:" {
					s.Find("a").Each(func(i int, a *goquery.Selection) {
						href, _ := a.Attr("href")
						if href != "" && href != "#" {
							if title == "Жанр:" {
								book.AppendGenre(a.Text())
							} else {
								book.AppendTag(a.Text())
							}
						}
					})
				}
			})
		}
		if len(book.Authors) == 0 {
			re, _ := regexp.Compile(`(?U)author: \"(.*)\",`)
			m := re.FindSubmatch([]byte(res.Body))
			if m != nil {
				s := strings.TrimSpace(string(m[1]))
				book.Authors = append(book.Authors, ParsePerson(s, false))
				book.InitFileName()
			}
		}
		if book.Summary == "" {
			re, _ := regexp.Compile(`(?U)<div itemprop=\"description\" class=\"biblio_book_descr_publishers\">(.*)<\/div>`)
			m := re.FindSubmatch([]byte(res.Body))
			if m != nil {
				book.Summary = strings.TrimSpace(string(m[1]))
			}
		}
		if book.Isbn == "" {
			re, _ := regexp.Compile(`(?U)<span itemprop=\"isbn\">(.*)<\/span>`)
			m := re.FindSubmatch([]byte(res.Body))
			if m != nil {
				book.Isbn = strings.TrimSpace(string(m[1]))
			}
		}
	})

	if strings.Contains(link, "litres.ru") {
		book.LitresUrl = strings.TrimSuffix(link, "chitat-onlayn/")
		c.Visit(book.LitresUrl)
	}
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

func SearchGoogle10(query string, site string) []string {
	s := "-w " + site
	out, err := exec.Command("googler", "-n 10", "--np", "--json", s, query).Output()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	var dat []map[string]interface{}

	if err := json.Unmarshal(out, &dat); err != nil {
		panic(err)
	}

	var urls []string
	for _, d := range dat {
		str, _ := url.QueryUnescape(d["url"].(string))
		urls = append(urls, str)
	}
	return urls
}

func ScrapeBookInner(urls []string) []Book {
	var books []Book

	for _, w := range urls {
		if strings.HasPrefix(w, "https://www.labirint.ru/books/") {
			books = append(books, VisitLabirint(w))
		}
		if strings.HasPrefix(w, "https://www.ozon.ru/product/") {
			//books = append(books, VisitOzon(w))
		}
		if strings.HasPrefix(w, "https://www.livelib.ru/book/") {
			books = append(books, VisitLivelib(w))
		}
	}
	return books
}

func ScrapeLabirint(query string) []Book {
	labirint := SearchGoogle10(query, "labirint.ru/books")
	//fmt.Printf("Labirint: %v\n", labirint)
	return ScrapeBookInner(labirint)
}

func ScrapeOzon(query string) []Book {
	ozon := SearchGoogle10(query, "ozon.ru/product")
	return ScrapeBookInner(ozon)
}

func ScrapeLivelib(query string) []Book {
	livelib := SearchGoogle10(query, "livelib.ru/book")
	//fmt.Printf("Livelib: %v\n", livelib)
	return ScrapeBookInner(livelib)
}

func ScrapeBook(query string) []Book {
	books := ScrapeLabirint(query)
	if len(books) == 0 {
		books = ScrapeLivelib(query)
	}
	return books
}

func main() {
	var books []Book
	var i int

	translator.Load()

	reader := bufio.NewReader(os.Stdin)
	query := os.Args[1]
	books = ScrapeBook(query)
	if len(books) > 1 {
		fmt.Printf("=======\n")
		fmt.Printf("0. none \n")
		for i, book := range books {
			u := book.LabirintUrl
			if u == "" {
				u = book.LivelibUrl
			}
			fmt.Printf("%d. \"%s\", publisher: %s [%s] url: %s\n", i+1, book.FileName, book.Publisher, book.Year, u)
		}
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "0"
		}
		i, _ = strconv.Atoi(text)
	} else if len(books) == 1 {
		i = 1
	} else if len(books) == 0 {
		fmt.Println("Nothing found")
	}

	if i > 0 {
		i = i - 1
		time.Sleep(1 * time.Second)
		litres := SearchGoogle(query, "litres.ru")
		fmt.Printf("Litres: %s\n", litres)
		VisitLitres(&books[i], litres)
		books[i].PrintMarkdown()
		fmt.Println("file \"" + books[i].FileName + "\" created")
	}

	translator.Save()
}
