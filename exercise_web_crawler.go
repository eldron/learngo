package main

import (
	"fmt"
	"sync"
	//"time"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

var visited map[string]bool
var mux sync.Mutex


// use WaitGroup to wait for sub goroutines to exit
func wgcrawl(url string, depth int, fetcher Fetcher, wg *sync.WaitGroup, ch chan string){
	defer wg.Done()
	if depth > 0 {
		mux.Lock()
		_, found := visited[url]
		mux.Unlock()

		if !found {
			mux.Lock()
			visited[url] = true
			mux.Unlock()

			body, urls, err := fetcher.Fetch(url)
			if err == nil {
				ch <- fmt.Sprintf("found %s %q\n", url, body)
				var subwg sync.WaitGroup
				subwg.Add(len(urls))
				for _, u := range urls {
					wgcrawl(u, depth - 1, fetcher, &subwg, ch)
				}
				subwg.Wait()
			}
		}
	}
}
// use channel to wait for sub goroutines to exit
// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, ch chan string) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	defer close(ch)
	if depth <= 0 {
		return
	}
	
	mux.Lock()
	_, found := visited[url]
	mux.Unlock()
	
	if !found {
		mux.Lock()
		visited[url] = true
		mux.Unlock()
		
		body, urls, err := fetcher.Fetch(url)
		if err != nil {
			//fmt.Println(err)
			return
		}
		//fmt.Printf("found: %s %q\n", url, body)
		ch <- fmt.Sprintf("found: %s %q\n", url, body)

		chs := make([]chan string, len(urls))
		for i := range urls {
			chs[i] = make(chan string)
			go Crawl(urls[i], depth - 1, fetcher, chs[i])
		}
		for i := range chs {
			for content := range chs[i] {
				ch <- content
			}
		}
		return
	}
}

func main() {
	// chs := make([]chan string, 5)
	// if chs[0] == nil {
	// 	fmt.Println("nil channel")
	// } else {
	// 	fmt.Println("not nil channel")
	// }
	visited = make(map[string]bool)
	ch := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	//go Crawl("https://golang.org/", 4, fetcher, ch)
	go wgcrawl("https://golang.org/", 4, fetcher, &wg, ch)
	go func() {
		for content := range ch {
			fmt.Println(content)
		}
	}() 
	wg.Wait()
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}

