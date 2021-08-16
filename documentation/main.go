package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fagnercarvalho/redis-lsp/completer"
)

func main() {
	commands := completer.GetCommands()

	for _, c := range commands {
		name := strings.ToLower(strings.Replace(c, " ", "-", -1))
		resp, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/redis/redis-doc/master/commands/%v.md", name))
		if err != nil {
			panic(err)
		}

		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		currentDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		file, err := os.Create(filepath.Join(currentDir, "completer", "files", fmt.Sprintf("%v.%v", name, "md")))
		if err != nil {
			panic(err)
		}

		_, err = file.Write(bytes)
		if err != nil {
			panic(err)
		}

		file.Close()
	}
}
