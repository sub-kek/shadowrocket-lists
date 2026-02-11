package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	inputDir  = "./v2fly-data/data"
	outputDir = "./shadowrocket"
)

type Entry struct {
	Prefix string
	Value  string
}

func main() {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	files, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		processSingleCategory(file.Name())
	}
}

func processSingleCategory(fileName string) {
	fmt.Printf("Processing: %s...\n", fileName)

	var allEntries []Entry
	processedFiles := make(map[string]bool)
	err := collectEntries(filepath.Join(inputDir, fileName), &allEntries, processedFiles)
	if err != nil {
		fmt.Printf("  [!] Error in %s: %v\n", fileName, err)
		return
	}

	outPath := filepath.Join(outputDir, fileName+".list")
	out, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("  [!] Failed to create file %s: %v\n", outPath, err)
		return
	}
	defer out.Close()

	writer := bufio.NewWriter(out)
	suffixes := make(map[string]bool)
	others := make(map[string]bool)

	count := 0
	for _, entry := range allEntries {
		if entry.Prefix == "DOMAIN-SUFFIX" {
			if !isSubdomain(entry.Value, suffixes) && !suffixes[entry.Value] {
				writer.WriteString(fmt.Sprintf("%s,%s\n", entry.Prefix, entry.Value))
				suffixes[entry.Value] = true
				count++
			}
		} else {
			fullLine := entry.Prefix + "," + entry.Value
			if entry.Prefix == "DOMAIN" && (suffixes[entry.Value] || isSubdomain(entry.Value, suffixes)) {
				continue
			}
			if !others[fullLine] {
				writer.WriteString(fullLine + "\n")
				others[fullLine] = true
				count++
			}
		}
	}
	writer.Flush()
	fmt.Printf("  [+] Done: %d entries\n", count)
}

func collectEntries(path string, entries *[]Entry, processed map[string]bool) error {
	if processed[path] {
		return nil
	}
	processed[path] = true

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.Split(line, "#")[0]

		line = strings.Split(line, "@")[0]
		line = strings.Split(line, "&")[0]
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "include:") {
			parts := strings.Fields(strings.TrimPrefix(line, "include:"))
			if len(parts) > 0 {
				incName := parts[0]
				collectEntries(filepath.Join(inputDir, incName), entries, processed)
			}
			continue
		}

		var p, v string
		switch {
		case strings.HasPrefix(line, "full:"):
			p, v = "DOMAIN", strings.TrimPrefix(line, "full:")
		case strings.HasPrefix(line, "keyword:"):
			p, v = "DOMAIN-KEYWORD", strings.TrimPrefix(line, "keyword:")
		case strings.HasPrefix(line, "regexp:"):
			// Ignore nor now
			continue
		case strings.HasPrefix(line, "domain:"):
			p, v = "DOMAIN-SUFFIX", strings.TrimPrefix(line, "domain:")
		default:
			p, v = "DOMAIN-SUFFIX", line
		}

		v = strings.TrimSpace(v)
		if v != "" {
			*entries = append(*entries, Entry{Prefix: p, Value: v})
		}
	}
	return scanner.Err()
}

func isSubdomain(domain string, suffixes map[string]bool) bool {
	parts := strings.Split(domain, ".")
	for i := 1; i < len(parts); i++ {
		parent := strings.Join(parts[i:], ".")
		if suffixes[parent] {
			return true
		}
	}
	return false
}
