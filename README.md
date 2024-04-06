# dwiki
A utility for getting a summary (the first paragraph) of wikipedia articles. Search on a topic and the first result will be returned.

## Installation
```
go install github.com/dmars8047/dwiki/cmd/dwiki@latest
```

## Example usage
```
	err := GetWikiArticleSummary("golang", os.Stdout)

	if err != nil {
		fmt.Println("Error:", err)
	}
```
This will search for the term "golang" on Wikipedia and print a summary of the first search result to the console.
