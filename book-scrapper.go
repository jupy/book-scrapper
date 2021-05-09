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

type Person struct {
	FirstName  string
	MiddleName string
	LastName   string
}

type Book struct {
	Type         string
	FileName     string
	Name         string
	InitName     string
	PosterUrl    string
	Year         string
	Genres       []string
	Authors      []Person
	Painters     []Person
	Editors      []Person
	Countries    []string
	Publisher    string
	Summary      string
	LabirintUrl  string
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
		fmt.Printf("Author:       %s, %s\n", d.LastName, d.FirstName)
	}
	for _, d := range book.Painters {
		fmt.Printf("Painter:      %s, %s\n", d.LastName, d.FirstName)
	}
	for _, d := range book.Editors {
		fmt.Printf("Editor:       %s, %s\n", d.LastName, d.FirstName)
	}
	fmt.Printf("Publisher:      %s\n", book.Publisher)
	for _, c := range book.Countries {
		fmt.Printf("Country:        %s\n", c)
	}

	fmt.Printf("Labirint:       %s\n", book.LabirintUrl)
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
		text += "[[" + person.LastName + ", " + person.FirstName + "]]"
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

	PrintPersonsList(w, "**author:**", book.Authors)
	PrintPersonsList(w, "**painter:**", book.Painters)
	PrintPersonsList(w, "**editor:**", book.Editors)
	_, err = fmt.Fprintf(w, "**publisher:** [[%s]]]\n", book.Publisher)
	check(err)
	PrintList(w, "**country:**", book.Countries)
	PrintList(w, "**tags:**", book.Genres)

	_, err = fmt.Fprintf(w, "**[labirint](%s)**\n", book.LabirintUrl)
	check(err)
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

func ParsePerson(str string) Person {
	var person Person
	v := strings.Split(str, " ")
	if len(v) == 2 {
		person.FirstName = v[1]
		person.LastName = v[0]
	}
	if len(v) == 3 {
		person.FirstName = v[1]
		person.MiddleName = v[2]
		person.LastName = v[0]
	}
	return person
}

func VisitLabirint(link string) Book {

	var book Book

	book.Type = "book"
	book.LabirintUrl = link

	fmt.Printf("Labirint: %s\n", link)

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

	/* 	c.OnHTML("#product-image img[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		book.PosterUrl, _ = url.QueryUnescape(link)
	}) */
	c.OnHTML("meta", func(e *colly.HTMLElement) {
		if e.Attr("property") == "og:image" {
			link := e.Attr("content")
			book.PosterUrl, _ = url.QueryUnescape(link)
		}
	})

	c.OnHTML(".authors", func(e *colly.HTMLElement) {
		if strings.HasPrefix(e.Text, "Автор: ") {
			s := strings.TrimPrefix(e.Text, "Автор: ")
			book.Authors = append(book.Authors, ParsePerson(s))
		}
		if strings.HasPrefix(e.Text, "Художник: ") {
			s := strings.TrimPrefix(e.Text, "Художник: ")
			book.Painters = append(book.Painters, ParsePerson(s))
		}
		if strings.HasPrefix(e.Text, "Редактор: ") {
			s := strings.TrimPrefix(e.Text, "Автор: ")
			book.Editors = append(book.Editors, ParsePerson(s))
		}
	})

	c.OnHTML(".publisher a", func(e *colly.HTMLElement) {
		book.Publisher = e.Text
	})

	c.OnHTML(".publisher", func(e *colly.HTMLElement) {
		r, _ := regexp.Compile("[0-9][0-9][0-9][0-9]")
		book.Year = r.FindString(e.Text)
	})

	c.Visit(book.LabirintUrl)

	var authors string
	l := len(book.Authors)
	if l > 0 {
		authors = book.Authors[0].LastName + ", " + book.Authors[0].FirstName
	}
	if l == 2 {
		authors += " и " + book.Authors[1].LastName + ", " + book.Authors[1].FirstName
	}
	if l > 2 {
		authors += " и др."
	}
	book.FileName = authors + " - " + book.Name + ".md"

	return book
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

	c.Visit(book.GoodreadsUrl)
	return book
}

func VisitLitres(link string) []string {

	var tags []string

	c := colly.NewCollector(
		colly.AllowedDomains("www.litres.ru"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "www.litres.ru/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	//document.querySelector("#page-wrap > div.page > div:nth-child(2) > div > div.content_column.column > div > div.biblio_book_center.column > div.biblio_book_info > ul > li:nth-child(2) > a")
	/* 	c.OnHTML("div.biblio_book_center", func(e *colly.HTMLElement) {
		fmt.Printf("tag: %s\n", e.Text)
		if e.Text != "#" {
			tags = append(tags, e.Text)
		}
	}) */

	/* 	c.OnHTML("html", func(e *colly.HTMLElement) {
		fmt.Printf("tag: %s\n", e.Text)
		if e.Text != "#" {
			tags = append(tags, e.Text)
		}
	}) */

	c.OnResponse(func(res *colly.Response) {
		re, _ := regexp.Compile(`(?U)<div class="biblio_book_info">(.*)<\/div>`)
		m := re.FindSubmatch([]byte(res.Body))
		if m != nil {
			s := strings.TrimSpace(string(m[1]))
			fmt.Println("Parsed: ", s)
		}
	})
	/* 	c.OnHTML("div.p-book-info img.p-picture__image[src]", func(e *colly.HTMLElement) {
		picture = e.Attr("src")
	}) */

	c.Visit(link)
	return tags
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

/* func ScrapeBookInner(labirint []string, goodreads string, flibusta string, litres string) Book {
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
} */

func ScrapeBookInner(urls []string) []Book {
	var books []Book

	for _, w := range urls {
		if strings.HasPrefix(w, "https://www.labirint.ru/books/") {
			books = append(books, VisitLabirint(w))
		}
	}

	/* 	name := strings.TrimSpace(book.InitName)
	   	if name == "" {
	   		name = book.Name
	   	}
	   	book.FileName = name + " (" + book.Year + ").md" */
	return books
}

func ScrapeBook(query string) []Book {
	labirint := SearchGoogle10(query, "labirint.ru/books")
	/* 	time.Sleep(1 * time.Second)
	   	goodreads := SearchGoogle(query, "goodreads.com")
	   	time.Sleep(1 * time.Second)
	   	flibusta := SearchGoogle(query, "flibusta.is")
	   	time.Sleep(1 * time.Second)
	   	litres := SearchGoogle(query, "litres.ru")
	   	time.Sleep(1 * time.Second) */

	/* 	fmt.Println(goodreads)
	   	fmt.Println(flibusta)
	   	fmt.Println(litres) */

	/* return ScrapeBookInner(labirint, goodreads, flibusta, litres) */
	return ScrapeBookInner(labirint)
}

func main() {
	/* 	var books []Book

	   	reader := bufio.NewReader(os.Stdin)
	   	query := os.Args[1]
	   	books = ScrapeBook(query)
	   	fmt.Printf("=======\n")
	   	fmt.Printf("0. none \n")
	   	for i, book := range books {
	   		fmt.Printf("%d. \"%s\", publisher: %s [%s]\n", i+1, book.FileName, book.Publisher, book.Year)
	   	}
	   	text, _ := reader.ReadString('\n')
	   	text = strings.TrimSuffix(text, "\n")
	   	if text == "" {
	   		text = "0"
	   	}
	   	i, _ := strconv.Atoi(text)
	   	if i > 0 {
	   		time.Sleep(1 * time.Second)
	   		litres := SearchGoogle(query, "litres.ru")
	   		fmt.Printf("Litres: %s\n", litres)
	   		books[i-1].Genres = VisitLitres(litres)
	   		fmt.Printf("Genres: %v\n", books[i-1].Genres)
	   		books[i-1].PrintMarkdown()
	   		fmt.Println("file \"" + books[i-1].FileName + "\" created")
	   	} */
	litres := SearchGoogle(os.Args[1], "litres.ru")
	fmt.Printf("Litres: %s\n", litres)
	genres := VisitLitres(litres)
	fmt.Printf("Genres: %v\n", genres)
}
