package main

import (
	"fmt"

	_authors "code.litriv.com/southerly/migrate/authors"
	_posts "code.litriv.com/southerly/migrate/posts"
	_topics "code.litriv.com/southerly/migrate/topics"
)

func main() {

	// Does parsing of files, constructing objects and do some individual sanity checking
	authors := _authors.Parse()
	topics := _topics.Parse()
	postAuthors, posts := _posts.Parse()
	//comments := parseComments()

	// Do some cross sanity checking
	checkThatPostsAuthorsMatchesParsedAuthors(append(authors, postAuthors...))
	fmt.Println("Cross sanity check passed.")

	// Create upload script
	authors.Write()
	topicsInPosts(posts, topics).Write()
	posts.Write()
	fmt.Println("\nFinished :-)")
}

func checkThatPostsAuthorsMatchesParsedAuthors(as []_authors.BlogAuthor) {
	for _, a := range as {
		a.CheckThatAuthorMatchesDuplicates(as)
	}
}

func topicsInPosts(ps _posts.Posts, ts _topics.Topics) _topics.Topics {
	filtered := make([]_topics.Topic, 0)
	// Construct the set of topic ids that are in posts
	inPosts := make(map[int]struct{}, 0)
	for _, p := range ps {
		for _, tid := range p.Topic_ids {
			inPosts[tid] = struct{}{}
		}
	}
	// Add all topics that has an id that is in the set constructed above, to the return array
	for _, t := range ts {
		if _, ok := inPosts[t.Id]; ok {
			filtered = append(filtered, t)
		} else {
			fmt.Println("Skipping topic: ", t.Name)
		}
	}
	return filtered
}
