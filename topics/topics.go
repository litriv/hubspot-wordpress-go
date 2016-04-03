package topics

import "code.litriv.com/southerly/migrate/parser"
import "code.litriv.com/southerly/migrate/writer"
import "fmt"

type Topics []Topic

type topicsWrapper struct {
	Objects Topics
}

type Topic struct {
	Created     int64
	Deleted_at  int
	Description string
	Id          int
	Name        string
	Slug        string
}

func Parse() Topics {
	tw := parser.Parse(&topicsWrapper{}, "topics").(*topicsWrapper)
	return tw.Objects
}

func (ts Topics) Write() {
	// TODO implement this
	fmt.Println("\nWriting topics...\n")
	for _, t := range ts {
		fmt.Println(t.Name)
	}
	funcMap := map[string]interface{}{
		"funcName": func() string { return "hlsm_topics" },
	}
	writer.Execute("topics", ts, funcMap)
}
