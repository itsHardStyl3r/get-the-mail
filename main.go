package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

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

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Failed to download %s: %v\n", url, err)
				return
			}
			defer resp.Body.Close()

			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())

				if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
					continue
				}

				mu.Lock()
				domains[strings.ToLower(line)] = struct{}{}
				mu.Unlock()
			}
		}(url)
	}

	wg.Wait()

	file, err := os.Create("output/blacklist.txt")
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	writer := bufio.NewWriter(file)
	for domain := range domains {
		writer.WriteString(domain + "\n")
	}
	writer.Flush()

	fmt.Printf("Finished! Collected %d unique domains.\n", len(domains))
}
