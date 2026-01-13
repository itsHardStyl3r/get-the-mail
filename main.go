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
)

var domainRegex = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)

func main() {
	urls := []string{
		"https://raw.githubusercontent.com/disposable-email-domains/disposable-email-domains/main/disposable_email_blocklist.conf",
		"https://raw.githubusercontent.com/disposable/disposable-email-domains/master/domains_strict.txt",
		"https://raw.githubusercontent.com/TheDahoom/disposable-email/main/blacklist.txt",
		"https://raw.githubusercontent.com/eser/sanitizer-svc/main/disposable_email_blocklist.conf",
		"https://raw.githubusercontent.com/GeroldSetz/emailondeck.com-domains/master/emailondeck.com_domains_from_bdea.cc.txt",
		"https://raw.githubusercontent.com/groundcat/disposable-email-domain-list/master/domains.txt",
		"https://raw.githubusercontent.com/jespernissen/disposable-maildomain-list/master/disposable-maildomain-list.txt",
		"https://raw.githubusercontent.com/kslr/disposable-email-domains/master/list.txt",
		"https://raw.githubusercontent.com/MattKetmo/EmailChecker/master/res/throwaway_domains.txt",
		"https://raw.githubusercontent.com/unkn0w/disposable-email-domain-list/main/domains.txt",
		"https://raw.githubusercontent.com/7c/fakefilter/main/txt/data.txt",
		"https://raw.githubusercontent.com/wesbos/burner-email-providers/master/emails.txt",
		"https://raw.githubusercontent.com/FGRibreau/mailchecker/master/list.txt",
		"https://raw.githubusercontent.com/willwhite/freemail/master/data/free.txt",
		"https://raw.githubusercontent.com/sublime-security/static-files/master/disposable_email_providers.txt",
	}

	var wg sync.WaitGroup
	domains := make(map[string]struct{})
	var mu sync.Mutex

	fmt.Println("Downloading domain names from", len(urls), "sources...")

	for _, url := range urls {
		wg.Add(1)
		go func(targetUrl string) {
			defer wg.Done()

			resp, err := http.Get(targetUrl)
			if err != nil {
				fmt.Printf("Failed to download %s: %v\n", url, err)
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
			fmt.Printf("Downloaded and corrected %d domain names from %s\n", localCount, targetUrl)
		}(url)
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

	fmt.Printf("Successfully saved %d domain names to %s\n", len(sortedDomains), filename)
}
