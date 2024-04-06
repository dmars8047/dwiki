package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dmars8047/dwiki/pkg/dwiki"
)

func main() {

	var topic string

	// Look for the -topic flag
	if len(os.Args) > 1 {
		if os.Args[1] == "-topic" || os.Args[1] == "--topic" || os.Args[1] == "-t" || os.Args[1] == "--t" {
			topic = os.Args[2]
		}
	}

	if topic == "" {
		// Get the topic from the user
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Enter the topic you want to search for: ")
		topic, _ = reader.ReadString('\n')
	}

	topic = strings.TrimSpace(topic)

	fmt.Println()

	dwiki.GetWikiArticleSummary(topic, os.Stdout)

	fmt.Print("\n\n")

}
