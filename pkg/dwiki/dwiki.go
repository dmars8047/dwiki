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
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
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
			Ns              int    `json:"ns"`
			Title           string `json:"title"`
			Pageid          int    `json:"pageid"`
			Wordcount       int    `json:"wordcount"`
			CategorySnippet string `json:"categorysnippet"`
		} `json:"search"`
	} `json:"query"`
}

type categoryResponse struct {
	Batchcomplete string `json:"batchcomplete"`
	Query         struct {
		Pages map[string]struct {
			Pageid     int    `json:"pageid"`
			Ns         int    `json:"ns"`
			Title      string `json:"title"`
			Categories []struct {
				Ns    int    `json:"ns"`
				Title string `json:"title"`
			} `json:"categories,omitempty"`
			PageProps *struct {
				Disambiguation string `json:"disambiguation,omitempty"`
			} `json:"pageprops,omitempty"`
		} `json:"pages"`
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
			FullURL string `json:"fullurl"`
		} `json:"pages"`
	} `json:"query"`
	Limits struct {
		Extracts int `json:"extracts"`
	} `json:"limits"`
}

// GetMatchingArticles searches for articles matching the given topic and writes the results to the given writer.
// It returns a map of article titles with their corresponding index.
func GetMatchingArticles(topic string, writer io.Writer) (map[int]int, error) {
	const url = "https://en.wikipedia.org/w/api.php"

	options := make(map[int]int)

	params := make(map[string]string)

	params["action"] = "query"
	params["list"] = "search"
	params["srsearch"] = topic
	params["format"] = "json"
	params["srlimit"] = "20"
	params["srprop"] = "wordcount|categorysnippet"

	// Call the API
	httpClient := http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return options, err
	}

	q := req.URL.Query()

	for key, value := range params {
		q.Set(key, value)
	}

	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)

	if err != nil {
		return options, err
	}

	defer resp.Body.Close()

	// Write response to console
	responseBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return options, err
	}

	var searchResponse searchResponse

	err = json.Unmarshal(responseBytes, &searchResponse)

	if err != nil {
		return options, err
	}

	// If there are no search results, print a message
	if len(searchResponse.Query.Search) == 0 {
		writer.Write([]byte("No search results found.\n\n"))
		return options, nil
	}

	// Get the categories for the search results to eliminate disambiguation pages
	categoryQueryUrl := "https://en.wikipedia.org/w/api.php?action=query&prop=pageprops&ppprop=disambiguation&redirects&format=json&pageids="

	for _, result := range searchResponse.Query.Search {
		categoryQueryUrl += fmt.Sprintf("%d|", result.Pageid)
	}

	// Get rid of the last pipe character
	categoryQueryUrl = categoryQueryUrl[:len(categoryQueryUrl)-1]

	req, err = http.NewRequest("GET", categoryQueryUrl, nil)

	if err != nil {
		return options, err
	}

	resp, err = httpClient.Do(req)

	if err != nil {
		return options, err
	}

	defer resp.Body.Close()

	responseBytes, err = io.ReadAll(resp.Body)

	if err != nil {
		return options, err
	}

	var categoryResponse categoryResponse

	err = json.Unmarshal(responseBytes, &categoryResponse)

	if err != nil {
		return options, err
	}

	// Print the titles of the search results
	resultString := "Search results:\n"

	num := 1

	for _, result := range searchResponse.Query.Search {
		// Check if the article is a disambiguation page
		isDisambiguation := false

		categoryPage, ok := categoryResponse.Query.Pages[strconv.Itoa(result.Pageid)]

		if !ok {
			continue
		}

		if categoryPage.PageProps != nil {
			isDisambiguation = categoryPage.PageProps.Disambiguation == ""
		}

		if isDisambiguation {
			continue
		}

		resultString += fmt.Sprintf("%d. %s\n", num, result.Title)
		options[num] = result.Pageid
		num++

		if num > 10 {
			break
		}
	}

	if num < 1 {
		writer.Write([]byte("No valid search results found\n"))
		return options, nil
	}

	_, err = writer.Write([]byte(resultString))

	if err != nil {
		return make(map[int]int), err
	}

	return options, nil
}

func GetArticleSummary(pageId int, writer io.Writer) error {
	explainUrl := fmt.Sprintf("https://en.wikipedia.org/w/api.php?format=json&action=query&prop=info|extracts&exlimit=max&explaintext&exintro&pageids=%d&inprop=url", pageId)

	httpClient := http.Client{}

	req, err := http.NewRequest("GET", explainUrl, nil)

	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var extractResponse extractResponse

	err = json.Unmarshal(responseBytes, &extractResponse)

	if err != nil {
		return err
	}

	// Get the page ID
	var pgIdStr string

	for key := range extractResponse.Query.Pages {
		pgIdStr = key
		break
	}

	if extractResponse.Query.Pages[pgIdStr].Extract == "" {
		return errors.New("no extract found")
	}

	articleUrl := extractResponse.Query.Pages[pgIdStr].FullURL

	extract := extractResponse.Query.Pages[pgIdStr].Extract

	// Get the first 500 characters of the extract or the first paragraph. Whichever is shorter
	// Split the text into paragraphs
	paragraphs := strings.Split(extract, "\n")

	// Get the first paragraph
	summary := paragraphs[0]

	// If the first paragraph is longer than 1024 characters, truncate it
	if len(summary) > 1024 {
		summary = strings.TrimSpace(summary[:1024]) + "..."
	} else {
		// If there is a second paragraph, add it
		if len(paragraphs) > 1 {
			summary += "\n\n" + paragraphs[1]

			if len(summary) > 1024 {
				summary = strings.TrimSpace(summary[:1024]) + "..."
			}
		}
	}

	// Add the find out more link
	summary += fmt.Sprintf("\n\nFind out more: %s", articleUrl)

	io.WriteString(writer, summary)

	return nil
}

// GetWikiArticleSummary searches for the given topic on Wikipedia and writes a summary of the first search result to the given writer.
func GetWikiArticleSummary(topic string, writer io.Writer) error {
	options, err := GetMatchingArticles(topic, writer)

	if err != nil {
		return err
	}

	if len(options) == 0 {
		return errors.New("no search results found")
	}

	// Get the user's choice
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter the number of the article you want to read: ")
	choice, _ := reader.ReadString('\n')

	choice = strings.TrimSpace(choice)

	choiceInt := 0

	// Convert the choice to an integer
	if choice == "" {
		return errors.New("you must enter a valid number")
	}

	choiceInt, err = strconv.Atoi(choice)

	if err != nil {
		return errors.New("you must enter a valid number")
	}

	// Get the article summary
	err = GetArticleSummary(options[choiceInt], writer)

	if err != nil {
		return err
	}

	return nil
}
