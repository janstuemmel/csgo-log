package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/janstuemmel/csgo-log"
)

// Usage
// go run main.go example.log
// cat example.log | go run main.go

func main() {

	var file *os.File
	var err error

	if len(os.Args) < 2 {
		file = os.Stdin
	} else {
		file, err = os.Open(os.Args[1])
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	r := bufio.NewReader(file)

	// read first line
	l, _, err := r.ReadLine()

	for err == nil {

		// parse
		m, errParse := csgolog.Parse(string(l))

		if errParse != nil {
			// print parse errors to stderr
			fmt.Fprintf(os.Stderr, "ERROR: %s", csgolog.ToJSON(m))
		} else {
			// print to stdout
			fmt.Fprintf(os.Stdout, "%s", csgolog.ToJSON(m))
		}

		// next line
		l, _, err = r.ReadLine()
	}
}
