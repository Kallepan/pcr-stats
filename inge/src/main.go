package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sqweek/dialog"
)

var wg sync.WaitGroup

type ExportData struct {
	code    string
	message string // First message with the code
	count   int
}

type Data struct {
	date    string
	kind    string
	code    string
	message string
}

func (d *Data) filterWithCodes(codes []string) bool {
	for _, code := range codes {
		if d.code == code {
			return true
		}
	}
	return false
}

var codesToBeFiltered = []string{
	"20042",
	"20039",
	"20009",
	"20008",
	"20045",
	"20031",
	"2506",
	"20034",
	"20005",
	"20045",
	"20012",
	"20043",
	"20040",
	"2002",
	"20054",
	"2006",
	"9007",
	"20055",
	"20041",
	"20038",
	"20025",
	"20013",
}

func countOccurenceOfCodes(dataArr []Data) map[string]int {
	counts := make(map[string]int)
	for _, data := range dataArr {
		counts[data.code]++
	}
	return counts
}

func getSourceDir() string {
	if runtime.GOOS != "windows" {
		return path.Join("inge", "data") // for testing on linux
	}
	dir, err := dialog.Directory().Title("Ordner mit den Logs auswählen").Browse()
	if err != nil {
		slog.Error("Error selecting directory", "error", err)
	}

	return dir
}

func isLogFile(file os.DirEntry) bool {
	return strings.HasSuffix(file.Name(), ".log")
}

func main() {
	start := time.Now()
	slog.SetLogLoggerLevel(slog.LevelInfo)

	// Read directory
	filePath := getSourceDir()
	if filePath == "" {
		slog.Error("No directory selected")
		return
	}
	files, err := os.ReadDir(filePath)
	if err != nil {
		slog.Error("Error reading directory", "error", err)
	}

	// set up regex for log line
	lineStartRegex := regexp.MustCompile(`^\d{2}\/\d{2}\/\d{4} \d{2}:\d{2}:\d{2}.\d{3}`)
	var allLines []string
	linesCh := make(chan []string, len(files))

	// filter for log files
	for i, file := range files {
		if !isLogFile(file) {
			slog.Warn(fmt.Sprintf("Skipping file: %s", file.Name()))
			continue
		}

		wg.Add(1)
		go func(file os.DirEntry) {
			defer wg.Done()
			slog.Info(fmt.Sprintf("Processing file %d/%d: %s", i+1, len(files), file.Name()))

			f, err := os.Open(path.Join(filePath, file.Name()))
			if err != nil {
				slog.Error("Error opening file", "error", err)
			}
			defer f.Close()

			var lines []string
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				if line == "" || line == " " || line == "\t" || line == "\n" {
					continue
				}

				if !lineStartRegex.MatchString(line) {
					lines[len(lines)-1] = lines[len(lines)-1] + " " + strings.Trim(line, " ")
					continue
				}

				lines = append(lines, line)
			}

			linesCh <- lines
		}(file)
	}

	// wait for all files to be processed
	go func() {
		wg.Wait()
		close(linesCh)
	}()

	// collect all lines
	for lines := range linesCh {
		allLines = append(allLines, lines...)
	}

	// cleanup and Parsing
	var dataArr []Data
	var codeToMessageMap = make(map[string]string)
	for _, row := range allLines {
		d := strings.Replace(row, "\t\t", "\t", -1)

		// remove [Spur1] to [Spur12] from message
		d = regexp.MustCompile(`\[(Spur\d)\]`).ReplaceAllString(d, "")

		s := strings.Split(d, "\t")

		if len(s) < 4 {
			continue
		}

		date := s[0]
		kind := s[2]
		code := s[3]
		message := s[len(s)-1]

		if _, ok := codeToMessageMap[code]; !ok {
			codeToMessageMap[code] = message
		}

		dataArr = append(dataArr, Data{date, kind, code, message})
	}

	// Filtering
	var filteredDataArr []Data
	for _, data := range dataArr {
		if data.filterWithCodes(codesToBeFiltered) {
			filteredDataArr = append(filteredDataArr, data)
		}
	}

	// Counting
	counts := countOccurenceOfCodes(filteredDataArr)

	// Exporting
	var exportDataArr []ExportData
	for code, count := range counts {
		exportDataArr = append(exportDataArr, ExportData{code, codeToMessageMap[code], count})
	}

	// Write to file
	f, err := os.Create(fmt.Sprintf("export_%s.csv", time.Now().Format("2006-01-02_15-04-05")))
	if err != nil {
		slog.Error("Error creating file", "error", err)
	}
	defer f.Close()

	f.WriteString("Code,Message,Count\n")
	for _, data := range exportDataArr {
		f.WriteString(fmt.Sprintf("%s,\"%s\",%d\n", data.code, data.message, data.count))
	}

	slog.Info(fmt.Sprintf("Exported %d rows", len(exportDataArr)))
	slog.Debug(fmt.Sprintf("Time taken: %s", time.Since(start)))
}
