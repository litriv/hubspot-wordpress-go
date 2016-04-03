package authors

import (
	"fmt"
	"strings"

	"code.litriv.com/southerly/migrate/parser"
	"code.litriv.com/southerly/migrate/writer"
)

// Authors is a collection of BlogAuthor
type Authors []BlogAuthor

type authorsWrapper struct {
	Limit   int
	Objects Authors
}

type BlogAuthor struct {
	Avatar      string
	Bio         string
	Body        authorBody
	Created     int64
	Deleted_at  int64
	Email       string
	Facebook    string
	Full_name   string
	Google_plus string
	Id          int
	Linkedin    string
	Twitter     string
	Updated     int64
	User_id     int
	Username    string
	Website     string
}

type authorBody struct {
	Avatar      string
	Bio         string
	Facebook    string
	Google_plus string
	Linkedin    string
	Twitter     string
	Website     string
}

// Parse parses authors.json
func Parse() Authors {
	aw := parser.Parse(&authorsWrapper{}, "authors").(*authorsWrapper)
	aw.Objects.sanityCheck()
	return aw.Objects
}

func (as Authors) sanityCheck() {
	for _, a := range as {
		a.SanityCheck()
	}
	fmt.Println("Authors sanity checks passed.")
}

// SanityCheck checks that data in BlogAuthor matches that in its authorBody and panics if it doesn't
func (a BlogAuthor) SanityCheck() {
	// When this check passes, it means that we can safely ignore the author body
	a.shouldMatchAuthorBody()
	a.shouldHaveAuthorDeletedAtZero()
}

// shouldMatchAuthorBody does a sanity check that the data in the author body of the author matches that in the author and panics if it doesn't
func (a BlogAuthor) shouldMatchAuthorBody() {
	hasNonNilBodyFields := a.Avatar != "" || a.Bio != "" || a.Facebook != "" || a.Google_plus != "" || a.Linkedin != "" || a.Twitter != "" || a.Website != ""
	blankAuthorBody := authorBody{}
	if a.Body == blankAuthorBody && hasNonNilBodyFields {
		panic("Some authors have empty bodies, but non-empty counterpart fields.")
	}
	if a.Avatar != a.Body.Avatar {
		panic("Not all authors avatars match their body avatars.")
	}
	if a.Bio != a.Body.Bio {
		panic("Not all author bios match their body bios.")
	}
	if a.Facebook != a.Body.Facebook {
		panic("Not all author facebooks math their body facebooks.")
	}
	if a.Google_plus != a.Body.Google_plus {
		panic("Not all author google_pluses math their body google_plus.")
	}
	if a.Linkedin != a.Body.Linkedin {
		panic("Not all author linkedin match their body linkedin.")
	}
	if a.Twitter != a.Body.Twitter {
		panic("Not all author twitter match their body twitter.")
	}
	if a.Website != a.Body.Website {
		panic("Not all author websites match their body websites.")
	}
}

// shouldHaveAuthorDeletedAtZero panics if the value of deleted_at is not zero
func (a BlogAuthor) shouldHaveAuthorDeletedAtZero() {
	if a.Deleted_at != 0 {
		panic("Not all author have deleted_at = 0")
	}
}

// CheckThatAuthorMatchesDuplicates checks that the BlogAuthor matches those in the parameter
func (a BlogAuthor) CheckThatAuthorMatchesDuplicates(as []BlogAuthor) {
	for _, da := range as {
		if a.Id == da.Id && a != da {
			panic("Author does not match another field-by-field with same id")
		}
	}
}

// Write generates the PHP array as authors.php, by using the template authors.tmpl
func (as Authors) Write() {
	for _, a := range as {
		fmt.Println(a.Full_name)
	}
	names := func(fullName string) []string { return strings.Split(fullName, " ") }
	firstName := func(fullName string) string { return names(fullName)[0] }
	funcMap := map[string]interface{}{
		"funcName": func() string { return "hlsm_authors" },
		"userLogin": func(fullName string) string {
			return strings.ToLower(firstName(fullName))
		},
		"firstName": firstName,
		"lastName":  func(fullName string) string { return names(fullName)[1] },
	}
	writer.Execute("authors", as, funcMap)
}
