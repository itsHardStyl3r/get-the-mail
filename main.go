package main

import (
	"bufio"
	"fmt"
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
	Name    string `yaml:"name"`
	RepoURL string `yaml:"repo_url"`
	DataURL string `yaml:"data_url"`
	Type    string `yaml:"type"`
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
	domains := make(map[string]struct{})
	var mu sync.Mutex

	fmt.Println("Downloading domain names from", len(config.Input), "sources...")

	for _, source := range config.Input {
		wg.Add(1)
		go func(src Source) {
			defer wg.Done()

			resp, err := http.Get(src.DataURL)
			if err != nil {
				fmt.Printf("Failed to download %s: %v\n", src.Name, err)
				return
			}
			defer resp.Body.Close()

			scanner := bufio.NewScanner(resp.Body)
			localCount := 0

			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())

				if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
					continue
				}

				domain := strings.ToLower(line)

				if domainRegex.MatchString(domain) {
					mu.Lock()
					domains[domain] = struct{}{}
					mu.Unlock()
					localCount++
				}
			}
			fmt.Printf("Downloaded and corrected %d domain names from %s.\n", localCount, src.Name)
		}(source)
	}

	wg.Wait()

	outputPath := "output/blacklist.txt"
	saveToFile(domains, outputPath)
}

func saveToFile(domains map[string]struct{}, filename string) {
	dir := filepath.Dir(filename)

	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Failed to create directories: %v\n", err)
		return
	}

	fmt.Println("Sorting...")
	sortedDomains := make([]string, 0, len(domains))
	for d := range domains {
		sortedDomains = append(sortedDomains, d)
	}
	sort.Strings(sortedDomains)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to save to file: %v\n", err)
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
