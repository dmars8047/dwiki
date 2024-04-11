package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dmars8047/dwiki/pkg/dwiki"
)

func main() {

	var topic string

	// Look for the -topic flag
	if len(os.Args) > 1 {
		if os.Args[1] == "-topic" || os.Args[1] == "--topic" || os.Args[1] == "-t" || os.Args[1] == "--t" {
			// Get the rest of the arguments
			topic = strings.Join(os.Args[2:], "_")
		}
	}

	if topic == "" {
		fmt.Print("\nWelcome to the Wikipedia search tool!\n\n")
		// Get the topic from the user
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Enter the topic you want to search for: ")
		topic, _ = reader.ReadString('\n')
	}

	topic = strings.TrimSpace(topic)

	if topic == "" {
		fmt.Println("Error. You must enter a topic to search for.")
		return
	}

	fmt.Println()

	options, err := dwiki.GetMatchingArticles(topic, os.Stdout)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	if len(options) == 0 {
		return
	}

	fmt.Println()

	// Get the user's choice
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter the number of the article you want to read: ")
	choice, _ := reader.ReadString('\n')

	choice = strings.TrimSpace(choice)

	choiceInt := 0

	// Convert the choice to an integer
	if choice == "" {
		fmt.Println("Error. You must enter a valid number.")
		return
	}

	choiceInt, err = strconv.Atoi(choice)

	if err != nil {
		fmt.Println("Error. You must enter a valid number.")
		return
	}

	fmt.Println()

	selectedTitle, ok := options[choiceInt]

	if !ok {
		fmt.Println("Error. You must enter a valid number.")
		return
	}

	// Get the article summary
	err = dwiki.GetArticleSummary(selectedTitle, os.Stdout)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Print("\n\n")

}
