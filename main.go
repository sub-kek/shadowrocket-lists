package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings",
	"sort"
)

const (
	inputDir  = "./v2fly-data/data"
	outputDir = "./shadowrocket"
)

type Rule struct {
	Prefix       string
	Value        string
	Attributes   map[string]bool
	Affiliations []string
	Inclusions   []Inclusion
}

type Inclusion struct {
	Target string
	Plus   []string // @attr
	Minus  []string // @-attr
}

var allFilesData = make(map[string][]Rule)

func main() {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	files, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Printf("Error reading data: %v\n", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		allFilesData[file.Name()] = parseFile(filepath.Join(inputDir, file.Name()))
	}

	for fileName := range allFilesData {
		processTarget(fileName)
	}
}

func parseFile(path string) []Rule {
	var rules []Rule
	file, err := os.Open(path)
	if err != nil {
		return rules
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.Split(line, "#")[0]

		rule := Rule{Attributes: make(map[string]bool)}
		parts := strings.Fields(line)
		
		mainPart := ""
		for _, p := range parts {
			if strings.HasPrefix(p, "@") {
				rule.Attributes[strings.TrimPrefix(p, "@")] = true
			} else if strings.HasPrefix(p, "&") {
				rule.Affiliations = append(rule.Affiliations, strings.TrimPrefix(p, "&"))
			} else if mainPart == "" {
				mainPart = p
			}
		}

		if strings.HasPrefix(mainPart, "include:") {
			inc := Inclusion{Target: strings.TrimPrefix(mainPart, "include:")}
			for _, p := range parts {
				if strings.HasPrefix(p, "@-") {
					inc.Minus = append(inc.Minus, strings.TrimPrefix(p, "@-"))
				} else if strings.HasPrefix(p, "@") {
					inc.Plus = append(inc.Plus, strings.TrimPrefix(p, "@"))
				}
			}
			rule.Inclusions = append(rule.Inclusions, inc)
		} else {
			switch {
			case strings.HasPrefix(mainPart, "full:"):
				rule.Prefix, rule.Value = "DOMAIN", strings.TrimPrefix(mainPart, "full:")
			case strings.HasPrefix(mainPart, "keyword:"):
				rule.Prefix, rule.Value = "DOMAIN-KEYWORD", strings.TrimPrefix(mainPart, "keyword:")
			case strings.HasPrefix(mainPart, "regexp:"):
				continue
			case strings.HasPrefix(mainPart, "domain:"):
				rule.Prefix, rule.Value = "DOMAIN-SUFFIX", strings.TrimPrefix(mainPart, "domain:")
			default:
				rule.Prefix, rule.Value = "DOMAIN-SUFFIX", mainPart
			}
		}
		rules = append(rules, rule)
	}
	return rules
}

func processTarget(targetName string) {
    fmt.Printf("Building target: %s...\n", targetName)
    
    resultMap := make(map[string]string)
    visited := make(map[string]bool)
	
    collect(targetName, []string{}, []string{}, resultMap, visited)

    for _, rules := range allFilesData {
        for _, r := range rules {
            for _, aff := range r.Affiliations {
                if aff == targetName && r.Value != "" {
                    resultMap[r.Value] = r.Prefix
                }
            }
        }
    }

    if len(resultMap) == 0 {
        return
    }

    suffixes := make(map[string]bool)
    for val, pref := range resultMap {
        if pref == "DOMAIN-SUFFIX" {
            suffixes[val] = true
        }
    }
	
    keys := make([]string, 0, len(resultMap))
    for val := range resultMap {
        keys = append(keys, val)
    }

    sort.Strings(keys)

    outPath := filepath.Join(outputDir, targetName+".list")
    out, err := os.Create(outPath)
    if err != nil {
        fmt.Printf("  [!] Error creating file: %v\n", err)
        return
    }
    defer out.Close()
    
    writer := bufio.NewWriter(out)
    count := 0

    for _, val := range keys {
        pref := resultMap[val]
        
        if pref == "DOMAIN" && isSubdomainOfAny(val, suffixes) {
            continue
        }
        
        _, err := writer.WriteString(fmt.Sprintf("%s,%s\n", pref, val))
        if err != nil {
            break
        }
        count++
    }
    
    writer.Flush()
    fmt.Printf("  [+] Done: %d entries\n", count)
}

func collect(target string, plus, minus []string, res map[string]string, visited map[string]bool) {
	visitKey := fmt.Sprintf("%s|%v|%v", target, plus, minus)
	if visited[visitKey] {
		return
	}
	visited[visitKey] = true

	rules, ok := allFilesData[target]
	if !ok {
		return
	}

	for _, r := range rules {
		match := true
		for _, p := range plus {
			if !r.Attributes[p] {
				match = false; break
			}
		}
		if !match { continue }
		for _, m := range minus {
			if r.Attributes[m] {
				match = false; break
			}
		}
		if !match { continue }

		if r.Value != "" {
			res[r.Value] = r.Prefix
		}

		for _, inc := range r.Inclusions {
			collect(inc.Target, inc.Plus, inc.Minus, res, visited)
		}
	}
}

func isSubdomainOfAny(domain string, suffixes map[string]bool) bool {
	parts := strings.Split(domain, ".")
	for i := 1; i < len(parts); i++ {
		parent := strings.Join(parts[i:], ".")
		if suffixes[parent] {
			return true
		}
	}
	return false
}
