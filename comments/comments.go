package comments

import (
	"fmt"

	"code.litriv.com/southerly/migrate/parser"
)

type comments struct {
	Limit   int
	Objects []comment
}

type comment struct {
	Akismet_passed          bool
	Approval_status         string
	Approval_status_label   string
	Comment                 string
	Content_id              int
	Created                 int64
	Deleted_at              int
	Id                      int
	Is_approved             bool
	Remote_ip               string
	Subscribe_to_follow_ups bool
	User_email              string
	User_name               string
	User_website            string
}

// Parse parses the comments.json
func Parse(basePath string) []comment {
	cs := parser.Parse(&comments{}, "comments").(comments)
	sanityCheck(cs.Objects)
	return cs.Objects
}

func sanityCheck(cs []comment) {
	for _, c := range cs {
		checkAkismetPassed(c)
		checkApprovalStatus(c)
		checkApprovalStatusLabel(c)
		checkCommentDeletedAt(c)
		checkIsApproved(c)
	}
	fmt.Println("Comments sanity check passed.")
}

func checkAkismetPassed(c comment) {
	if !c.Akismet_passed {
		panic(fmt.Sprint("Not all comments have akismet_passed = true, ", c.Id))
	}
}

func checkApprovalStatus(c comment) {
	if c.Approval_status != "approved-akismet" {
		panic(fmt.Sprint("Not all comments have approval_status = 'appoved-akismet'", c.Id))
	}
}

func checkApprovalStatusLabel(c comment) {
	if c.Approval_status_label != "Live" {
		panic("Not all comments have approval_status_label = 'Live'")
	}
}

func checkCommentDeletedAt(c comment) {
	if c.Deleted_at != 0 {
		panic("Not all comments has deleted_at = 0")
	}
}

func checkIsApproved(c comment) {
	if !c.Is_approved {
		panic("Not all comments have is_approved = true")
	}
}

func (cs comments) Upload() {
	// TODO implement this
}
