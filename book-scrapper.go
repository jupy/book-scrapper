package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"google.golang.org/api/customsearch/v1"

	/* 	"google.golang.org/api/googleapi" */
	"google.golang.org/api/option"
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

	_, err := fmt.Fprint(w, title)
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

	_, err := fmt.Fprint(w, title)
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

	_, err := fmt.Fprint(w, title)
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

func IsEnglish(s string) bool {
	if len(s) <= 0 {
		return false
	}
	r := s[0]
	if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
		return false
	}
	return true
}

func (book *Book) GetPrintAuthor() string {
	var author string
	sort.Slice(book.Authors, func(i, j int) bool {
		return book.Authors[i].LastName < book.Authors[j].LastName
	})
	/* fmt.Printf("author: %v\n", book.Authors) */
	l := len(book.Authors)
	if l > 0 {
		author = book.Authors[0].PrintName()
	}
	if l == 2 {
		a := book.Authors[1].PrintName()
		if IsEnglish(a) {
			author += " and " + a
		} else {
			author += " и " + a
		}
	}
	if l > 2 {
		if IsEnglish(author) {
			author += " et al"
		} else {
			author += " и др."
		}
	}
	return author
}

func (book *Book) InitFileName() {
	var name string
	if utf8.RuneCountInString(book.Name) <= 75 {
		name = book.Name
	} else if strings.Contains(book.Name, ".") {
		v := strings.Split(book.Name, ".")
		name = v[0]
	} else {
		name = book.Name[0:72] + "..."
	}

	name = book.GetPrintAuthor() + " - " + name + ".md"
	name = strings.ReplaceAll(name, "<", "")
	name = strings.ReplaceAll(name, ">", "")
	name = strings.ReplaceAll(name, ":", " -")
	name = strings.ReplaceAll(name, "«", "")
	name = strings.ReplaceAll(name, "»", "")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, "|", "-")
	name = strings.ReplaceAll(name, "?", ".")
	name = strings.ReplaceAll(name, "*", "")
	book.FileName = name
}

func firstRune(str string) (r rune) {
	for _, r = range str {
		return
	}
	return
}

func (book *Book) SaveMarkdown() {
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
	if book.LivelibUrl != "" {
		_, err = fmt.Fprintf(w, "**[livelib](%s)**\n", book.LivelibUrl)
		check(err)
	}

	// **{{shell: open-library-folder "/Lib/ru/С/Сото, Эрнандо де"}}**
	folder := "/Lib/"
	author := book.GetPrintAuthor()
	if IsEnglish(author) {
		folder += "en/"
	} else {
		folder += "ru/"
	}
	folder += author[0:1] + "/" + author
	_, err = fmt.Fprintf(w, "**{{shell: open-library-folder \"%s\"}}**\n", folder)
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
		s := strings.TrimSpace(item)
		if s == "" {
			continue
		}
		r := []rune(s)
		count := utf8.RuneCountInString(s)
		if unicode.IsUpper(firstRune(s)) {
			if (count == 1) ||
				(count == 2 && r[1] == '.') ||
				(count == 4 && r[1] == '.' && r[3] == '.') {
				person.Initials += s
			} else {
				v = append(v, s)
			}
		} else {
			l := len(v)
			if l > 0 {
				v[l-1] += " " + s
				invert = !invert
			} else {
				appendToLast = s
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
		if invert {
			person.FirstName = v[0]
			person.MiddleName = v[1]
			person.LastName = v[2]
		} else {
			person.FirstName = v[1]
			person.MiddleName = v[2]
			person.LastName = v[0]
		}
	}
	/* fmt.Printf("person: %v\n", person) */
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

func RemoveNumPrefix(text string) string {
	s := text
	re, _ := regexp.Compile(`№\d* в (.*)`)
	m := re.FindSubmatch([]byte(s))
	if m != nil {
		s = strings.TrimSpace(string(m[1]))
	}
	return s
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
		e.ForEach("a[href].bc-author__link", func(i int, a *colly.HTMLElement) {
			str := strings.TrimSpace(a.Text)
			book.Authors = append(book.Authors, ParsePerson(str, true))
		})
	})

	c.OnHTML("#main-image-book", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		book.PosterUrl, _ = url.QueryUnescape(link)
	})

	c.OnHTML("a.bc-edition__link", func(e *colly.HTMLElement) {
		book.Publisher = strings.TrimSpace(e.Text)
	})

	c.OnHTML(".bc-genre", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(i int, a *colly.HTMLElement) {
			book.AppendGenre(RemoveNumPrefix(a.Text))
		})
	})

	c.OnHTML(".bc-info__wrapper div p", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "Жанры:") {
			e.ForEach("a[href]", func(i int, a *colly.HTMLElement) {
				book.AppendGenre(RemoveNumPrefix(a.Text))
			})
		}
	})

	c.OnHTML(".bc-info div p", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "ISBN: ") {
			r, _ := regexp.Compile("ISBN: (.*)")
			m := r.FindSubmatch([]byte(e.Text))
			if m != nil {
				book.Isbn = strings.TrimSpace(string(m[1]))
			}
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

	c.OnHTML("div#lenta-card__text-edition-full", func(e *colly.HTMLElement) {
		book.Summary = e.Text
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

	c := colly.NewCollector(
		colly.AllowedDomains("www.goodreads.com"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "www.goodreads.com/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	c.OnHTML("div.BookPageTitleSection__title", func(e *colly.HTMLElement) {
		if len(book.Name) == 0 {
			book.Name = strings.TrimSpace(e.Text)
		}
	})

	c.OnHTML("div.BookPageMetadataSection__contributor", func(e *colly.HTMLElement) {
		e.ForEach(".ContributorLink__name", func(i int, d *colly.HTMLElement) {
			str := strings.TrimSpace(d.Text)
			if str != "" {
				book.Authors = append(book.Authors, ParsePerson(str, true))
			}
		})
	})

	c.OnHTML("img#coverImage", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		book.PosterUrl, _ = url.QueryUnescape(link)
	})

	c.OnHTML(".EditionDetails", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "Published") {
			r, _ := regexp.Compile("[0-9][0-9][0-9][0-9]")
			s := r.FindString(e.Text)
			if s != "" {
				book.Year = s
			}
			r, _ = regexp.Compile(".* by (.*)")
			m := r.FindSubmatch([]byte(e.Text))
			if m != nil {
				book.Publisher = strings.TrimSpace(string(m[1]))
			}
		}
	})

	c.OnHTML("#description span:nth-child(1)", func(e *colly.HTMLElement) {
		book.Summary = strings.TrimSpace(e.Text)
	})

	c.OnHTML(".elementList div.left", func(e *colly.HTMLElement) {
		e.ForEach("a.actionLinkLite.bookPageGenreLink", func(i int, a *colly.HTMLElement) {
			genre := strings.ToLower(a.Text)
			book.Genres[genre] = ""
		})
	})

	c.OnHTML("#bookDataBox div.clearFloats", func(e *colly.HTMLElement) {
		e.ForEach("span", func(i int, s *colly.HTMLElement) {
			if s.Attr("itemprop") == "isbn" {
				if len(s.Text) == 13 {
					book.Isbn = s.Text[0:3] + "-" + s.Text[3:4] + "-" + s.Text[4:9] + "-" + s.Text[9:12] + "-" + s.Text[12:13]
				}
			}
		})
	})

	book.GoodreadsUrl = link
	c.Visit(book.GoodreadsUrl)
	book.InitFileName()
	return book
}

func VisitLitres(link string) Book {

	book := NewBook()

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
		pos := strings.LastIndex(link, "chitat-onlayn")
		if pos > 0 {
			link = link[0:pos]
		}
		book.LitresUrl = link
		c.Visit(book.LitresUrl)
	}

	book.InitFileName()
	return book
}

// Search google for a given query on a given site
func SearchGoogle(query string, site string, n int64) []string {

	ctx := context.Background()
	apiKey := "AIzaSyA7Tc319mUHNiqj5JnBFbDIddTHrM-x1vk"
	cx := "5190c921a536d4411"

	svc, err := customsearch.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create custom search service: %v", err)
	}

	resp, err := svc.Cse.List().Cx(cx).Num(n).SiteSearch(site).ExactTerms(query).Do()
	if err != nil {
		log.Fatalf("Failed to search: %v", err)
	}

	if len(resp.Items) == 0 {
		time.Sleep(2 * time.Second)
		resp, err = svc.Cse.List().Cx(cx).Num(n).SiteSearch(site).Q(query).Do()
		if err != nil {
			log.Fatalf("Failed to search: %v", err)
		}
	}

	var urls []string
	for _, item := range resp.Items {
		urls = append(urls, item.Link)
	}
	return urls
}

func ScrapeLabirint(query string) []Book {
	var books []Book
	for _, w := range SearchGoogle(query, "labirint.ru/books", 5) {
		fmt.Printf("%v\n", w)
		books = append(books, VisitLabirint(w))
	}
	return books
}

func ScrapeLivelib(query string) []Book {
	var books []Book
	for _, w := range SearchGoogle(query, "livelib.ru/book", 5) {
		fmt.Printf("%v\n", w)
		books = append(books, VisitLivelib(w))
	}
	return books
}

func ScrapeLitres(query string) []Book {
	var books []Book
	for _, w := range SearchGoogle(query, "litres.ru", 5) {
		fmt.Printf("%v\n", w)
		books = append(books, VisitLitres(w))
	}
	return books
}

func ScrapeGoodreads(query string) []Book {
	var books []Book
	for _, w := range SearchGoogle(query, "goodreads.com/book", 5) {
		fmt.Printf("%v\n", w)
		books = append(books, VisitGoodreads(w))
	}
	for _, w := range SearchGoogle(query, "goodreads.com/en/book", 5) {
		fmt.Printf("%v\n", w)
		books = append(books, VisitGoodreads(w))
	}
	return books
}

func SelectBook(books []Book) *Book {
	var i int
	reader := bufio.NewReader(os.Stdin)

	if len(books) > 0 {
		fmt.Printf("=======\n")
		fmt.Printf("0. none \n")
		for i, book := range books {
			u := book.LabirintUrl
			if u == "" {
				u = book.LivelibUrl
			}
			if u == "" {
				u = book.GoodreadsUrl
			}
			fmt.Printf("%d. \"%s\" [%s] publisher: %s\n", i+1, book.FileName, book.Year, book.Publisher)
			fmt.Printf("        %s\n", u)
		}
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "0"
		}
		i, _ = strconv.Atoi(text)
	} else if len(books) == 0 {
		fmt.Println("Nothing found")
	}

	if i == 0 {
		return nil
	}
	return &books[i-1]
}

func main() {
	var book *Book
	translator.Load()

	VisitLivelib("https://www.livelib.ru/book/1000551620-grabezh-po-zakonu-frederik-bastia")

	// check if there is a command line argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: book <query>")
		os.Exit(1)
	}

	query := os.Args[1]
	fmt.Printf("query: %s\n", query)

	if IsEnglish(query) {
		book = SelectBook(ScrapeGoodreads(query))
	} else {
		book = SelectBook(ScrapeLabirint(query))
		if book == nil {
			time.Sleep(2 * time.Second)
			book = SelectBook(ScrapeLivelib(query))
		}

		if book == nil {
			time.Sleep(2 * time.Second)
			book = SelectBook(ScrapeLitres(query))

			/* 			litres := SearchGoogle(query, "litres.ru", 1)
			   			if len(litres) > 0 {
			   				VisitLitres(book, litres[0])
			   				fmt.Printf("Litres: %s\n", book.LitresUrl)
			   			} */
		}
	}

	if book != nil {
		book.SaveMarkdown()
		fmt.Println("file \"" + book.FileName + "\" created")
	}

	translator.Save()
}
