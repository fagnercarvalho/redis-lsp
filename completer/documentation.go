package completer

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed files
var files embed.FS

func GetDocumentation(command string) ([]byte, error) {
	name := strings.ToLower(strings.Replace(command, " ", "-", -1))
	return files.ReadFile(fmt.Sprintf("files/%v.md", name))
}
