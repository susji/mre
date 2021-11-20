package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/susji/mre"
)

func main() {
	var sre = flag.String("re", "", "Regular expression to evaluate")
	var dump = flag.Bool("d", false, "Dump expression matcher tree")
	flag.Parse()

	if len(*sre) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	re, err := mre.Compile(*sre)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to compile: %v", err)
		os.Exit(1)
	}
	if *dump {
		fmt.Println("# matcher tree")
		fmt.Println(re.Dump())
	}
	r := bufio.NewReader(os.Stdin)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Departing...\n")
			break
		}
		line = strings.TrimRight(line, "\n")
		matched := re.Match(line)
		fmt.Printf("``%s''", line)
		if matched {
			fmt.Printf(" -> matched: %#v\n", re.Captures())
		} else {
			fmt.Printf(" -> did not match.\n")
		}
	}
}
