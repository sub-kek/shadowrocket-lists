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

		fileName := file.Name()
		processSingleCategory(fileName)
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
		} else if entry.Prefix == "DOMAIN" {
            if !isSubdomain(entry.Value, suffixes) && !suffixes[entry.Value] {
                fullLine := entry.Prefix + "," + entry.Value
                if !others[fullLine] {
                    writer.WriteString(fullLine + "\n")
                    others[fullLine] = true
                    count++
                }
            }
        } else {
			fullLine := entry.Prefix + "," + entry.Value
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
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "include:") {
			incName := strings.TrimSpace(strings.TrimPrefix(line, "include:"))
			collectEntries(filepath.Join(inputDir, incName), entries, processed)
			continue
		}

		var p, v string
		switch {
		case strings.HasPrefix(line, "full:"):
			p, v = "DOMAIN", strings.TrimPrefix(line, "full:")
		case strings.HasPrefix(line, "keyword:"):
			p, v = "DOMAIN-KEYWORD", strings.TrimPrefix(line, "keyword:")
		case strings.HasPrefix(line, "regexp:"):
			continue
		// 	p, v = "DOMAIN-WILDCARD", strings.TrimPrefix(line, "regexp:")
		default:
			p, v = "DOMAIN-SUFFIX", strings.TrimPrefix(line, "domain:")
		}
		*entries = append(*entries, Entry{Prefix: p, Value: strings.TrimSpace(v)})
	}
	return scanner.Err()
}

func isSubdomain(domain string, suffixes map[string]bool) bool {
    for i := 0; i < len(domain); i++ {
        if domain[i] == '.' {
            parent := domain[i+1:]
            if suffixes[parent] {
                return true
            }
        }
    }
    return false
}
