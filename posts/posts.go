package posts

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"code.litriv.com/southerly/migrate/authors"
	"code.litriv.com/southerly/migrate/parser"
	"code.litriv.com/southerly/migrate/writer"
)

type Images []Image

type Image struct {
	HubspotUrl string
	Location   string
}

type Posts []*post

type postsWrapper struct {
	Limit   int
	Objects Posts
}

type post struct {
	Id                int
	Archived          bool
	Blog_author       authors.BlogAuthor
	Html_title        string
	Blog_author_id    int
	Category_id       int
	Comment_count     int
	Created           int64
	Deleted_at        int64
	Featured_image    string
	Is_draft          bool
	Keywords          interface{}
	Meta_description  string
	Meta_keywords     string
	Name              string
	Post_body         string
	Post_summary      string
	Processing_status string
	Publish_date      int64
	Published_url     string
	Rss_body          string
	State             string
	Subcategory       string
	Topic_ids         []int
	Url               string

	Images Images
}

type keyword struct {
	Keyword, Text string
}

func Parse() (authors.Authors, Posts) {
	pw := parser.Parse(&postsWrapper{}, "posts").(*postsWrapper)

	as := make([]authors.BlogAuthor, 100)
	// Do some sanity checking and collect authors
	var states = make(map[string]interface{}, 100)
	checkState := func(p *post) {
		if p.State != "DRAFT" && p.State != "PUBLISHED" {
			panic("Expected state to have values DRAFT or PUBLISHES")
		}
		if _, ok := states[p.State]; !ok {
			fmt.Println("state: ", p.State)
			states[p.State] = true
		}
	}
	checkWhetherNameMatchesHtmlTitle := func(p *post) {
		if p.Name != p.Html_title {
			panic("Not all posts' names match their html_title")
		}
	}
	for _, p := range pw.Objects {
		// If this check passed, it means we can safely ignore published_url
		//checkUrlIsSameAsPublishedUrl(p)
		//checkProcessingStatusIsPublished(p)
		checkState(p)
		checkWhetherNameMatchesHtmlTitle(p)
		p.Blog_author.SanityCheck()
		as = append(as, p.Blog_author)
	}
	fmt.Println("Posts sanity check passed.")

	return as, pw.Objects
}

func (ps Posts) Write() {
	fmt.Println("\nWriting posts...\n")

	// For link fixes (trims) and replacements
	type replacer struct {
		ex string
		r  string // replacement string
	}
	replace := func(s string, rs ...replacer) string {
		for _, r := range rs {
			re := regexp.MustCompile(r.ex)
			s = re.ReplaceAllString(s, r.r)
		}
		return s
	}

	imageOccursBeforeText := func(s string) ([]int, bool) {
		re := regexp.MustCompile(`<img [^>]+?>`)
		loc := re.FindStringIndex(s)
		return loc, loc != nil && loc[0] < 86
	}

	// Register images
	registerImage := func(p *post, hubspotUrl string) {
		downloadPath := filepath.Join("/Volumes/CaseSensitive/images")
		downloadeds, err := ioutil.ReadDir(downloadPath)
		if err != nil {
			panic(err)
		}
		fileName := filepath.Base(hubspotUrl)
		location := filepath.Join(downloadPath, fileName)
		downloaded := func(fileName string) bool {
			for _, d := range downloadeds {
				if d.Name() == fileName {
					return true
				}
			}
			return false
		}
		download := func(hubsportUrl string) {
			resp, err := http.Get(hubspotUrl)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			err = ioutil.WriteFile(location, data, 0644)
			if err != nil {
				panic(err)
			}
		}
		if !downloaded(fileName) {
			fmt.Println("Downloading image: ", hubspotUrl)
			download(fileName)
		} else {
			fmt.Println("Skipping downloaded file: " + fileName)
		}
		p.Images = append(p.Images, Image{hubspotUrl, fileName})
	}
	for _, p := range ps {
		fmt.Println(p.Name)
		p.Images = make(Images, 0)
		// if p.Featured_image != "" {
		// 	registerImage(p, p.Featured_image)
		// }
		re := regexp.MustCompile(`<img .*?src="(.*?)".*?>`)
		for _, s := range re.FindAllStringSubmatch(p.Post_body, -1) {
			fmt.Println("Image link found: ", s[1])
			// if s[1] == p.Featured_image {
			// 	continue
			// }
			registerImage(p, s[1])
		}
		if _, ok := imageOccursBeforeText(p.Post_body); ok {
			p.Featured_image = "has"
		} else {
			p.Featured_image = ""
		}
	}

	replaceSpecialChars := func(s string) string {
		s = strings.Replace(s, "\u2018", "", -1)
		s = strings.Replace(s, "\u2019", "", -1)
		s = strings.Replace(s, "%E2%80%98", "", -1)
		s = strings.Replace(s, "%E2%80%99", "", -1)
		return s
	}

	// Write posts images
	funcMap := map[string]interface{}{
		"funcName": func() string { return "hlsm_posts_images" },
		"globals":  func() string { return "$hlsm_uploaded_posts" },
	}
	writer.Execute("postsimages", ps, funcMap)

	// Write posts featured images
	funcMap = map[string]interface{}{
		"funcName": func() string { return "hlsm_posts_featured_images" },
		"globals":  func() string { return "$hlsm_uploaded_posts" },
	}
	writer.Execute("postsfeaturedimages", ps, funcMap)

	// Write posts without content
	funcMap = map[string]interface{}{
		"funcName": func() string { return "hlsm_posts" },
		"globals":  func() string { return "$hlsm_uploaded_authors, $hlsm_uploaded_topics" },
		"status": func(p post) string {
			if p.State == "DRAFT" {
				return "draft"
			}
			// It has to PUBLISHED in this case, because we checked in the sanity check
			return "publish"
		},
	}
	writer.Execute("posts", ps, funcMap)

	// Write posts urls
	funcMap = map[string]interface{}{
		"funcName":            func() string { return "hlsm_posts_urls" },
		"replaceSpecialChars": replaceSpecialChars,
		"test": func(s string) string {
			isJimBut := strings.Contains(s, "jim-but")
			if isJimBut {
				fmt.Println("jim-but!!!", s)
			}
			re := regexp.MustCompile(`(http://blog.hellosoutherly.com.*?)\x{2019}(.*?)`)
			if isJimBut && re.MatchString(s) {
				panic("Found jim-but!!!")
			}
			return s
		},
	}
	writer.Execute("postsurls", ps, funcMap)

	// Write posts content
	funcMap = map[string]interface{}{
		"funcName": func() string { return "hlsm_posts_content" },
		"globals":  func() string { return "$hlsm_uploaded_posts, $hlsm_uploaded_posts_by_url, $hlsm_uploaded_images" },
		"replaceBlogLinks": func(s string) string {
			// Trim spaces off end
			s = replace(s, replacer{
				`href="(http://blog\.hellosoutherly\.com.*?) +?"`,
				`href="${1}"`})

			s = strings.Replace(s, `href="http://blog.hellosoutherly.com"`, `href="http://www.hellosoutherly.com/blog"`, -1)

			prepCategory := func(s string) string {
				return strings.ToLower(strings.Replace(s, "+", "-", -1))
			}
			re := regexp.MustCompile(fmt.Sprintf(`href="http://blog\.hellosoutherly\.com/\?Tag=%v"`, "(.*?)"))
			ms := re.FindAllStringSubmatch(s, -1)
			for _, m := range ms {
				s = strings.Replace(s,
					fmt.Sprintf(`href="http://blog.hellosoutherly.com/?Tag=%v"`, m[1]),
					fmt.Sprintf(`href="http://www.hellosoutherly.com/blog/category/%v"`, prepCategory(m[1])), -1)
			}

			return replace(s,
				replacer{
					`href="(http://blog\.hellosoutherly\.com/.*?)"`,
					`href="'.idForHubspotId('${1}', $$hlsm_uploaded_posts_by_url, 'replacing blog link').'"`})
		},
		"replaceImageLinks": func(s string) string {
			// Remove first image if it occurs before any blog text (within first 100 chars)
			if loc, ok := imageOccursBeforeText(s); ok {
				s = s[:loc[0]] + s[loc[1]:]
			}

			return replace(s,
				// Trim spaces off end
				replacer{
					`<img (.*?)src="(.*?) +?"`,
					`<img ${1}src="${2}"`},
				// Wrap in span, with class post-featured-img, to make full width
				replacer{
					`(<img .*?>)`,
					`<span class="post-featured-img">${1}</span>`},
				replacer{
					`<img (.*?)src="(.*?)"`,
					`<img ${1}src="'.idForHubspotId('${2}', $$hlsm_uploaded_images, 'replacing image link').'"`})
		},
		"replaceCTAs": func(s string) string {
			return replace(s,
				replacer{
					`\{\{cta\(\\'.*?\\'\)\}\}`,
					`<a href="http://www.hellosoutherly.com/whitepaper-run-content-marketing-workshop-within-corporation"><span class="post-featured-img"><img src="http://www.hellosoutherly.com/wp-content/uploads/2014/06/Content-Marketing-Workshop.jpg"></span></a>`,
				},
			)
		},
		"replaceSpecialChars": func(s string) string {
			return regexp.
				MustCompile(`http://blog.hellosoutherly.com.*?"`).
				ReplaceAllStringFunc(s, replaceSpecialChars)
		},
		"cleanHtml": func(s string) string {
			s = replace(s,
				replacer{
					`(?s)<div.*?>(.*?)</div>`,
					`$1`},
				replacer{
					`<!--.*?-->`,
					``},
				replacer{
					`<span>(.*?)</span>`,
					`$1`},
				replacer{
					`<p>&nbsp;</p>`,
					``},
				replacer{
					`<div.*?>`,
					``},
				replacer{
					`</div>`,
					``},
				replacer{
					`<em> <br></em>`,
					``},
				replacer{
					`%MCEPASTEBIN%`,
					``})

			r := replacer{
				`<p>(.+?)<br>`,
				"<p>$1</p>\n<p>"}
			s = replace(s,
				r, r, r, r, r, r,
				replacer{
					`<br>(.+?)<br>`,
					"<p>$1</p>"},
				replacer{
					`</p><br>`,
					"</p>"},
				replacer{
					`<p><br>`,
					"<p>"},
				replacer{
					`</p><p>`,
					"</p>\n<p>"},
				replacer{
					`<p><p>`,
					`<p>`},
				replacer{
					`</p>(.*?)</p>`,
					"</p>\n<p>$1</p>"},
				replacer{
					`<p></p>`,
					``},
				replacer{
					`<p> *?</p>`,
					``},
				replacer{
					"<(.*?) class=\"Standard\">",
					"<$1>"},
			)

			r = replacer{
				"\n{2,}",
				"\n"}
			s = replace(s, r, r, r, r)
			return s
		},
	}
	writer.Execute("postscontent", ps, funcMap)

	// Write posts meta description
	funcMap = map[string]interface{}{
		"funcName": func() string { return "hlsm_posts_metadesc" },
		"globals":  func() string { return "$hlsm_uploaded_posts" },
	}
	writer.Execute("postsmetadesc", ps, funcMap)
}
