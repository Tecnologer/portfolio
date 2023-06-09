package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/alecthomas/chroma/quick"
)

var (
	root, _      = getDir()
	sourcePath   = filepath.Join(root, "about")
	templatePath = filepath.Join(root, "template", "template.html")
	pagePath     = filepath.Join(root, "page")

	format = []struct {
		symbols []string
		format  string
	}{
		{
			symbols: []string{
				`(<span class="nx">Experience<\/span><span class="p">:<\/span>\s*<span class="nf">)(ListExperience)(<\/span>)`,
			},
			format: `$1<a href="experience.html" title="Expand experience list" class="nf">$2</a>$3`,
		},
		{
			symbols: []string{
				`(<span class="nx">ContactOptions<\/span><span class="p">:<\/span>\s*<span class="nf">)(ListContactOptions)(<\/span>)`,
			},
			format: `$1<a href="contact.html" title="See contact options" class="nf">$2</a>$3`,
		},
		{
			symbols: []string{
				"(rdominguez@tecnologer.net)",
			},
			format: `<a href="mailto:${1}" class="s">${1}</a>`,
		},
		{
			symbols: []string{
				`(<span class="s">&#34;)(http(s)?:\/\/.+)(&#34;<\/span>)`,
			},
			format: `$1<a href="$2" class="s">${2}</a>$4`,
		},
	}
)

type source struct {
	name string
	path string
}

func main() {
	version := time.Now().Format("2006.0102.1504")

	files, err := getSourceFiles(sourcePath)
	if err != nil {
		log.Fatal(err)
	}

	template, err := os.ReadFile(templatePath)
	if err != nil {
		log.Fatal(err)
	}

	var (
		goBack    []byte
		rules     map[string][]byte
		titlePage string
		htmlPath  string
	)

	for _, file := range files {
		content, err := os.ReadFile(file.path)
		if err != nil {
			log.Fatal(err)
		}

		goBack = []byte{}
		htmlPath = "index"

		if file.name != "me" {
			goBack = homeIcon()
			htmlPath = file.name
		}

		titlePage = strings.ToUpper(string(file.name[0])) + strings.ToLower(string(file.name[1:]))

		rules = map[string][]byte{
			"{{code}}":    formatText(content),
			"{{file}}":    []byte(file.name),
			"{{go_back}}": goBack,
			"{{version}}": []byte(version),
			"{{title}}":   []byte("Tecnologer | " + titlePage),
		}

		htmlPath = fmt.Sprintf("%s.html", filepath.Join(pagePath, htmlPath))

		err = writeFile(htmlPath, insertText(rules, template))
		if err != nil {
			log.Printf("writing %s. Err: %v\n", htmlPath, err)
		}

		log.Printf("file %s writed correctly.\n", htmlPath)
	}
}

func formatText(fileContent []byte) []byte {
	var buf bytes.Buffer

	err := quick.Highlight(&buf, string(fileContent), "Go", "html", "dracula")
	if err != nil {
		log.Println(err)
		return fileContent
	}

	fileContent = buf.Bytes()

	var reg *regexp.Regexp

	for _, f := range format {
		for _, word := range f.symbols {
			reg, err = regexp.Compile(fmt.Sprintf("(?mi)%s", word))
			if err != nil {
				log.Println(err)
				continue
			}
			fileContent = reg.ReplaceAll(fileContent, []byte(f.format))
			// fileContent = bytes.ReplaceAll(fileContent, []byte(word), []byte(fmt.Sprintf(f.format, word)))
		}
	}

	return fileContent
}

func writeFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0777)
}

func insertText(rules map[string][]byte, origin []byte) []byte {
	for key, value := range rules {
		origin = bytes.ReplaceAll(origin, []byte(key), value)
	}

	return origin
}

func getDir() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("get dir for connection file")
	}

	return path.Dir(filename), nil
}

func homeIcon() []byte {
	return []byte(`<a href="index.html" class="icon-link" title="return to home">
		<svg xmlns="http://www.w3.org/2000/svg" x="0px" y="0px" width="30" height="30" viewBox="0 -5 50 50">
    <path d="M 25 1.0507812 C 24.7825 1.0507812 24.565859 1.1197656 24.380859 1.2597656 L 1.3808594 19.210938 C 0.95085938 19.550938 0.8709375 20.179141 1.2109375 20.619141 C 1.5509375 21.049141 2.1791406 21.129062 2.6191406 20.789062 L 4 19.710938 L 4 46 C 4 46.55 4.45 47 5 47 L 19 47 L 19 29 L 31 29 L 31 47 L 45 47 C 45.55 47 46 46.55 46 46 L 46 19.710938 L 47.380859 20.789062 C 47.570859 20.929063 47.78 21 48 21 C 48.3 21 48.589063 20.869141 48.789062 20.619141 C 49.129063 20.179141 49.049141 19.550938 48.619141 19.210938 L 25.619141 1.2597656 C 25.434141 1.1197656 25.2175 1.0507812 25 1.0507812 z M 35 5 L 35 6.0507812 L 41 10.730469 L 41 5 L 35 5 z"></path>
</svg>
	</a>`)
}

func getSourceFiles(path string) ([]source, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	files := make([]source, 0)
	for _, entry := range entries {
		name := entry.Name()
		if filepath.Ext(name) != ".go" {
			continue
		}

		files = append(files, source{
			name: name[:len(name)-3],
			path: filepath.Join(path, name),
		})
	}

	return files, nil
}
