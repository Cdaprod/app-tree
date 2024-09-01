package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/h2non/filetype"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	outputMu sync.Mutex
	output   strings.Builder
	debug    bool
)

const (
	outputFileName = "app_tree_prompt.txt"
	htmlFileName   = "app_tree.html"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "app-tree",
		Short: "Analyze and visualize directory structures",
		Long:  `A CLI tool to analyze and display the structure of directories in a tree-like format.`,
	}

	var (
		generateHTML bool
	)

	var analyzeCmd = &cobra.Command{
		Use:   "analyze [directory]",
		Short: "Analyze the structure of a directory",
		Long:  `Analyze the structure of a directory and serve the result via a local web server or generate a static HTML file.`,
		Run:   runAnalysis,
	}

	analyzeCmd.Flags().BoolVarP(&generateHTML, "html", "", false, "Generate a static HTML file instead of serving via local server")
	analyzeCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	rootCmd.AddCommand(analyzeCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runAnalysis(cmd *cobra.Command, args []string) {
	generateHTML, _ := cmd.Flags().GetBool("html")

	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Printf("Error getting absolute path: %v\n", err)
		return
	}

	if debug {
		log.Printf("Analyzing directory: %s\n", absDir)
	}

	tempDir, err := ioutil.TempDir("", "app-tree")
	if err != nil {
		log.Printf("Error creating temporary directory: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	if debug {
		log.Printf("Temporary directory created: %s\n", tempDir)
	}

	fmt.Println("Counting items...")
	totalItems := countItems(absDir)
	fmt.Printf("Total items: %d\n", totalItems)

	fmt.Println("Processing files and directories...")
	bar := progressbar.Default(int64(totalItems))
	traverseDirectory(absDir, "", bar)

	if debug {
		log.Printf("Finished traversing directory\n")
	}

	if generateHTML {
		htmlContent := generateHTMLContent(output.String())
		err = ioutil.WriteFile(htmlFileName, []byte(htmlContent), 0644)
		if err != nil {
			log.Printf("Error writing to HTML file: %v\n", err)
			return
		}
		fmt.Printf("\nAnalysis complete! Open %s in your web browser to view the results.\n", htmlFileName)
	} else {
		outputPath := filepath.Join(tempDir, outputFileName)
		err = ioutil.WriteFile(outputPath, []byte(output.String()), 0644)
		if err != nil {
			log.Printf("Error writing to file: %v\n", err)
			return
		}

		if debug {
			log.Printf("Output written to: %s\n", outputPath)
		}

		serveResult(outputPath)
	}
}

func traverseDirectory(dir, indent string, bar *progressbar.ProgressBar) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Error reading directory %s: %v\n", dir, err)
		return
	}

	writeOutput(fmt.Sprintf("\nDIRECTORY: %s\n%s==========================\n", dir, indent))

	for _, entry := range entries {
		bar.Add(1)
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			traverseDirectory(path, indent+"  ", bar)
		} else {
			processFile(path, indent+"  ")
		}
	}
}

func processFile(file, indent string) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", file, err)
		return
	}

	kind, _ := filetype.Match(content)
	fileTypeStr := "unknown"
	if kind != filetype.Unknown {
		fileTypeStr = kind.MIME.Value
	}

	output := fmt.Sprintf("\nFILE: %s\nTYPE: %s\nSIZE: %d bytes\nCONTENT:\n%s==========================\n", file, fileTypeStr, len(content), indent)

	if strings.HasPrefix(fileTypeStr, "text") {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			output += indent + template.HTMLEscapeString(line) + "\n"
		}
	} else {
		output += indent + "[Binary file content not displayed]\n"
	}

	output += indent + "==========================\n"
	writeOutput(output)

	if debug {
		log.Printf("Processed file: %s\n", file)
	}
}

// ... (rest of the code remains the same)

func processFile(file, indent string) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", file, err)
		return
	}

	kind, _ := filetype.Match(content)
	fileTypeStr := "unknown"
	if kind != filetype.Unknown {
		fileTypeStr = kind.MIME.Value
	}

	output := fmt.Sprintf("\nFILE: %s\nTYPE: %s\nSIZE: %d bytes\nCONTENT:\n%s==========================\n", file, fileTypeStr, len(content), indent)

	if strings.HasPrefix(fileTypeStr, "text") {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			output += indent + template.HTMLEscapeString(line) + "\n"
		}
	} else {
		output += indent + "[Binary file content not displayed]\n"
	}

	output += indent + "==========================\n"
	writeOutput(output)
}

func writeOutput(content string) {
	outputMu.Lock()
	defer outputMu.Unlock()
	output.WriteString(content)
}

func generateHTMLContent(content string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>App Tree Analysis</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; padding: 20px; }
        h1 { color: #333; }
        h2 { color: #0066cc; }
        h3 { color: #009900; }
        pre { background-color: #f4f4f4; padding: 10px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>App Tree Analysis</h1>
    <pre>%s</pre>
</body>
</html>
`, template.HTMLEscapeString(content))
}