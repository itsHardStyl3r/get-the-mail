package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/goccy/go-yaml"
)

type Source struct {
	Name      string `yaml:"name"`
	RepoURL   string `yaml:"repo_url"`
	DataURI   string `yaml:"data_uri"`
	Type      string `yaml:"type"`
	Whitelist bool   `yaml:"whitelist"`
}

type Config struct {
	Input []Source `yaml:"input"`
}

var domainRegex = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)

func main() {
	configFile, err := os.ReadFile("config.yml")
	if err != nil {
		fmt.Printf("Error reading config.yml: %v\n", err)
		return
	}

	var config Config
	if err := yaml.Unmarshal(configFile, &config); err != nil {
		fmt.Printf("Error parsing config.yml: %v\n", err)
		return
	}

	var wg sync.WaitGroup
	blacklistDomains := make(map[string]struct{})
	whitelistDomains := make(map[string]struct{})
	var muB sync.Mutex
	var muW sync.Mutex

	fmt.Println("Processing sources from config...")
	for _, source := range config.Input {
		wg.Add(1)
		go func(src Source) {
			defer wg.Done()

			var reader io.ReadCloser
			if src.Type == "repo" {
				resp, err := http.Get(src.DataURI)
				if err != nil {
					fmt.Printf("Failed to download %s: %v\n", src.Name, err)
					return
				}
				reader = resp.Body
			} else if src.Type == "local" {
				file, err := os.Open(src.DataURI)
				if err != nil {
					fmt.Printf("Failed to open local file %s: %v\n", src.Name, err)
					return
				}
				reader = file
			} else {
				fmt.Printf("Unknown type for source %s\n", src.Name)
				return
			}
			defer reader.Close()

			scanner := bufio.NewScanner(reader)
			localCount := 0
			for scanner.Scan() {
				if src.Whitelist {
					if processLine(scanner.Text(), whitelistDomains, &muW) {
						localCount++
					}
				} else {
					if processLine(scanner.Text(), blacklistDomains, &muB) {
						localCount++
					}
				}
			}
			fmt.Printf("Successfully processed %d domain names from %s.\n", localCount, src.Name)
		}(source)
	}
	wg.Wait()

	saveToFile(blacklistDomains, "output/blacklist.txt")

	graylistDomains := make(map[string]struct{})
	for domain := range blacklistDomains {
		if _, found := whitelistDomains[domain]; !found {
			graylistDomains[domain] = struct{}{}
		}
	}
	saveToFile(graylistDomains, "output/graylist.txt")
}

func processLine(line string, storage map[string]struct{}, mu *sync.Mutex) bool {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
		return false
	}
	domain := strings.ToLower(line)
	if domainRegex.MatchString(domain) {
		mu.Lock()
		storage[domain] = struct{}{}
		mu.Unlock()
		return true
	}
	return false
}

func saveToFile(domains map[string]struct{}, filename string) {
	dir := filepath.Dir(filename)
	_ = os.MkdirAll(dir, 0755)

	sortedDomains := make([]string, 0, len(domains))
	for d := range domains {
		sortedDomains = append(sortedDomains, d)
	}
	sort.Strings(sortedDomains)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to save to file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, d := range sortedDomains {
		_, _ = writer.WriteString(d + "\n")
	}
	writer.Flush()

	fmt.Printf("Successfully saved %d domain names to %s.\n", len(sortedDomains), filename)
}
