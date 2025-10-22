package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
)

func main() {
	// Command-line flag for log file path
	logPath := flag.String("file", "", "Path to the .log file")
	topN := flag.Int("top", 20, "Number of top keywords to show")
	flag.Parse()

	if *logPath == "" {
		fmt.Println("Usage: go run main.go -file=your.log [-top=20]")
		os.Exit(1)
	}

	// Open the log file
	file, err := os.Open(*logPath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Regex for extracting words (customize as needed)
	re := regexp.MustCompile(`[a-zA-Z_]{3,}`) // words of length >= 3

	// Prepare for parallel processing
	wordFreq := make(map[string]int)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Read and process in chunks
	scanner := bufio.NewScanner(file)
	chunkSize := 1000
	lines := make([]string, 0, chunkSize)

	processChunk := func(chunk []string) {
		localFreq := make(map[string]int)
		for _, line := range chunk {
			words := re.FindAllString(strings.ToLower(line), -1)
			for _, word := range words {
				localFreq[word]++
			}
		}
		mu.Lock()
		for k, v := range localFreq {
			wordFreq[k] += v
		}
		mu.Unlock()
	}

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) == chunkSize {
			wg.Add(1)
			go func(chunk []string) {
				defer wg.Done()
				processChunk(chunk)
			}(lines)
			lines = make([]string, 0, chunkSize)
		}
	}
	if len(lines) > 0 {
		wg.Add(1)
		go func(chunk []string) {
			defer wg.Done()
			processChunk(chunk)
		}(lines)
	}
	wg.Wait()

	// Sort keywords by frequency
	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range wordFreq {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	// Print ASCII heatmap
	fmt.Printf("\nTop %d keywords in %s:\n\n", *topN, *logPath)
	maxBar := 40
	maxFreq := 1
	if len(sorted) > 0 {
		maxFreq = sorted[0].Value
	}
	for i := 0; i < *topN && i < len(sorted); i++ {
		barLen := sorted[i].Value * maxBar / maxFreq
		bar := strings.Repeat("❤︎ ", barLen)
		fmt.Printf("%-15s | %-5d %s\n", sorted[i].Key, sorted[i].Value, bar)
	}
}
