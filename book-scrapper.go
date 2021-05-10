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
	"strconv"
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
	ShortName    string
	Name         string
	InitName     string
	PosterUrl    string
	Year         string
	Genres       []string
	Tags         []string
	Authors      []Person
	Painters     []Person
	Editors      []Person
	Countries    []string
	Publisher    string
	Isbn         string
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
	for _, t := range book.Tags {
		fmt.Printf("Tag:            %s\n", t)
	}

	fmt.Printf("ISBN:           %s\n", book.Isbn)
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
	"Канада":  "Canada",
	"СССР":    "USSR",
	"США":     "USA",
	"1 класс": "#",
	"волшебные приключения":    "magic adventures",
	"государственная политика": "public policy",
	"детектив":         "detective",
	"детская классика": "children's classic",
	"драма":            "drama",
	"законы Вселенной": "laws of the universe",
	"зарубежная деловая литература":         "business",
	"зарубежная образовательная литература": "education",
	"квантовая физика":                      "the quantum physics",
	"книги по философии":                    "philosophy",
	"комедия":                               "comedy",
	"Латинская Америка":                     "Latin America",
	"мелодрама":                             "melodrama",
	"мультфильм":                            "cartoon",
	"мюзикл":                                "musical",
	"научная фантастика":                    "science fiction",
	"общая биология":                        "biology",
	"общая экономическая теория":            "general economic theory",
	"параллельные миры":                     "parallel worlds",
	"политология":                           "political science",
	"приключение":                           "adventures",
	"приключения":                           "adventures",
	"семейный":                              "family",
	"сказки":                                "fairy tale",
	"стимпанк":                              "steampunk",
	"терроризм":                             "terrorism",
	"физика":                                "physics",
	"физические теории":                     "physical theories",
	"фэнтези":                               "fantasy",
	"частная собственность":                 "private property",
	"экономическая политика":                "economic policy",
	"экономические реформы":                 "economic reforms",
	"экранизации":                           "film adaptation",
	"юридический триллер":                   "legal thriller",
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

	PrintList(w, "**genres:**", book.Genres)
	PrintPersonsList(w, "**author:**", book.Authors)
	PrintPersonsList(w, "**painter:**", book.Painters)
	PrintPersonsList(w, "**editor:**", book.Editors)
	_, err = fmt.Fprintf(w, "**publisher:** [[%s]]]\n", book.Publisher)
	check(err)
	PrintList(w, "**country:**", book.Countries)
	PrintList(w, "**tags:**", book.Tags)

	_, err = fmt.Fprintf(w, "**isbn:** %s\n", book.Isbn)
	check(err)
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

func (book *Book) InitFileName() {
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

	var name string
	if len(book.Name) <= 75 {
		name = book.Name
	} else if strings.Contains(book.Name, ".") {
		v := strings.Split(book.Name, ".")
		name = v[0]
	} else {
		name = book.Name[0:72] + "..."
	}

	book.FileName = authors + " - " + name + ".md"
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
	//fmt.Printf("person: %s\n", str)
	var person Person
	var v []string
	appendToLast := ""
	invert := false
	vec := strings.Split(str, " ")
	for _, item := range vec {
		if unicode.IsUpper(firstRune(item)) {
			v = append(v, item)
		} else {
			l := len(v)
			if l > 0 {
				v[l-1] += " " + item
				invert = true
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

	if len(v) == 2 {
		if invert {
			person.FirstName = v[0]
			person.LastName = v[1]
		} else {
			person.FirstName = v[1]
			person.LastName = v[0]
		}
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
			s := strings.TrimPrefix(e.Text, "Редактор: ")
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

	c.OnHTML(".isbn", func(e *colly.HTMLElement) {
		book.Isbn = strings.TrimPrefix(e.Text, "ISBN: ")
	})

	c.Visit(book.LabirintUrl)

	book.InitFileName()
	/* 	var authors string
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
	   	book.FileName = authors + " - " + book.Name + ".md" */

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
							trans := Translations[a.Text()]
							if len(trans) == 0 {
								fmt.Printf("can't translate: %s\n", a.Text())
								if title == "Жанр:" {
									book.Genres = append(book.Genres, a.Text())
								} else {
									book.Tags = append(book.Tags, a.Text())
								}
							} else if trans != "#" {
								if title == "Жанр:" {
									book.Genres = append(book.Genres, trans+"|"+a.Text())
								} else {
									book.Tags = append(book.Tags, trans+"|"+a.Text())
								}
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
				book.Authors = append(book.Authors, ParsePerson(s))
				book.InitFileName()
			}
		}
	})

	c.Visit(link)
	/* return genres, tags */
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
	}
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
	var books []Book

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
		/* 		books[i-1].Genres, books[i-1].Tags = VisitLitres(litres)
		   		fmt.Printf("Genres: %v\n", books[i-1].Genres)
		   		fmt.Printf("Tags:   %v\n", books[i-1].Tags) */
		VisitLitres(&books[i-1], litres)
		books[i-1].PrintMarkdown()
		fmt.Println("file \"" + books[i-1].FileName + "\" created")
	}
	/* 	litres := SearchGoogle(os.Args[1], "litres.ru")
	   	fmt.Printf("Litres: %s\n", litres)
	   	genres := VisitLitres(litres)
	   	fmt.Printf("Genres: %v\n", genres) */
}
