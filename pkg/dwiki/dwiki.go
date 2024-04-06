/*
Package dwiki provides a function to get a summary of a Wikipedia article based on a search term.

The GetWikiArticleSummary function takes a search term and an io.Writer as arguments. It makes a request to the Wikipedia API to search for the given term and returns a summary of the first search result.

Example usage:

	err := GetWikiArticleSummary("golang", os.Stdout)

	if err != nil {
		fmt.Println("Error:", err)
	}

This will search for the term "golang" on Wikipedia and print a summary of the first search result to the console.
*/
package dwiki

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type searchResponse struct {
	Batchcomplete string `json:"batchcomplete"`
	Continue      struct {
		Sroffset int    `json:"sroffset"`
		Continue string `json:"continue"`
	} `json:"continue"`
	Query struct {
		Searchinfo struct {
			Totalhits int `json:"totalhits"`
		} `json:"searchinfo"`
		Search []struct {
			Ns        int    `json:"ns"`
			Title     string `json:"title"`
			Pageid    int    `json:"pageid"`
			Size      int    `json:"size"`
			Wordcount int    `json:"wordcount"`
			Snippet   string `json:"snippet"`
			Timestamp string `json:"timestamp"`
		} `json:"search"`
	} `json:"query"`
}

type extractResponse struct {
	Batchcomplete string `json:"batchcomplete"`
	Query         struct {
		Normalized []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"normalized"`
		Pages map[string]struct {
			Pageid  int    `json:"pageid"`
			Ns      int    `json:"ns"`
			Title   string `json:"title"`
			Extract string `json:"extract"`
		} `json:"pages"`
	} `json:"query"`
	Limits struct {
		Extracts int `json:"extracts"`
	} `json:"limits"`
}

func GetWikiArticleSummary(topic string, writer io.Writer) error {

	const url = "https://en.wikipedia.org/w/api.php"

	params := make(map[string]string)

	params["action"] = "query"
	params["list"] = "search"
	params["srsearch"] = topic
	params["format"] = "json"

	// Call the API
	httpClient := http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return err
	}

	q := req.URL.Query()

	for key, value := range params {
		q.Set(key, value)
	}

	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Write response to console
	responseBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var searchResponse searchResponse

	err = json.Unmarshal(responseBytes, &searchResponse)

	if err != nil {
		return err
	}

	// If there are no search results, print a message
	if len(searchResponse.Query.Search) == 0 {
		return errors.New("no search results found")
	}

	// Get the title of the first search result
	pgIdStr := fmt.Sprintf("%d", searchResponse.Query.Search[0].Pageid)

	title := searchResponse.Query.Search[0].Title

	// Replace spaces with underscores
	title = strings.Replace(title, " ", "_", -1)

	explainUrl := fmt.Sprintf("https://en.wikipedia.org/w/api.php?format=json&action=query&prop=extracts&exlimit=max&explaintext&exintro&titles=%s", title)

	req, err = http.NewRequest("GET", explainUrl, nil)

	if err != nil {
		return err
	}

	resp, err = httpClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	responseBytes, err = io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var extractResponse extractResponse

	err = json.Unmarshal(responseBytes, &extractResponse)

	if err != nil {
		return err
	}

	if extractResponse.Query.Pages[pgIdStr].Extract == "" {
		return errors.New("no extract found")
	}

	extract := extractResponse.Query.Pages[pgIdStr].Extract

	// Get the first 500 characters of the extract or the first paragraph. Whichever is shorter
	// Split the text into paragraphs
	paragraphs := strings.Split(extract, "\n")

	// Get the first paragraph
	firstParagraph := paragraphs[0]

	// If the first paragraph is longer than 1024 characters, truncate it
	if len(firstParagraph) > 1024 {
		firstParagraph = firstParagraph[:1024] + "..."
	}

	// Add the find out more link
	firstParagraph += fmt.Sprintf("\n\nFind out more: https://en.wikipedia.org/wiki/%s", title)

	io.WriteString(writer, firstParagraph)

	return nil
}
