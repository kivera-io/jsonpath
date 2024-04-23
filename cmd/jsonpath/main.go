package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kivera-io/jsonpath"
)

func quit(e error) {
	if e == nil {
		os.Exit(0)
	}
	fmt.Println(e)
	os.Exit(1)
}

func main() {
	var query string
	var input string
	var file string
	var set string

	flag.Usage = func() {
		fmt.Printf("An example implementation of the JSONPath package.\n\n")
		fmt.Printf("JSON data can be provided using the command line options or through stdin.\n\n")
		fmt.Printf("./jsonpath <query> [ -option, ... ]\n\n")
		fmt.Printf("options:\n\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&input, "data", "", "A JSON string to process")
	flag.StringVar(&file, "file", "", "A JSON file to process")
	flag.StringVar(&set, "set", "", "A value to set using the query")
	indent := flag.Int("indent", 0, "Indentation to use when printing the result")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		query = args[0]
		flag.CommandLine.Parse(args[1:])
	}

	if query == "" {
		quit(errors.New("no query provided"))
	}

	fi, err := os.Stdin.Stat()
	if err != nil {
		quit(err)
	}
	if fi.Mode()&os.ModeCharDevice == 0 {
		piped, err := io.ReadAll(os.Stdin)
		if err != nil {
			quit(err)
		}
		input = string(piped)
	}

	var data interface{}
	if input != "" {
		err := json.Unmarshal([]byte(input), &data)
		if err != nil {
			quit(err)
		}
	} else if file != "" {
		f, err := os.ReadFile(file)
		if err != nil {
			quit(err)
		}
		err = json.Unmarshal(f, &data)
		if err != nil {
			quit(err)
		}
	} else {
		quit(errors.New("no JSON input provided"))
	}

	c, err := jsonpath.Compile(query)
	if err != nil {
		quit(err)
	}

	var result interface{}
	if set != "" {
		var val interface{}
		err = json.Unmarshal([]byte(set), &val)
		if err != nil {
			val = set
		}
		err = c.Set(data, val)
		if err != nil {
			quit(err)
		}
		result = data
	} else {
		result, err = c.Get(data)
		if err != nil {
			quit(err)
		}
	}

	var output []byte
	if *indent != 0 {
		var prefix string
		for i := 0; i < *indent; i++ {
			prefix += " "
		}
		output, err = json.MarshalIndent(result, "", prefix)
		if err != nil {
			quit(err)
		}
	} else {
		output, err = json.Marshal(result)
		if err != nil {
			quit(err)
		}
	}

	fmt.Println(string(output))
}
