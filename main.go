package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/h2non/filetype"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	outputMu sync.Mutex
	output   strings.Builder
	debug    bool
	generateHTML bool
)

const (
	outputFileName = "app_tree_prompt.txt"
	htmlFileName   = "app_tree.html"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "app-tree [directory]",
		Short: "Analyze and visualize directory structures",
		Long:  `app-tree is a CLI tool that analyzes and displays the structure of directories in a tree-like format. It can generate either a text output or an HTML file for easy viewing.`,
		Run:   runAnalysis,
	}

	rootCmd.Flags().BoolVarP(&generateHTML, "html", "", false, "Generate a static HTML file instead of text output")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runAnalysis(cmd *cobra.Command, args []string) {
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
		err = ioutil.WriteFile(outputFileName, []byte(output.String()), 0644)
		if err != nil {
			log.Printf("Error writing to file: %v\n", err)
			return
		}

		if debug {
			log.Printf("Output written to: %s\n", outputFileName)
		}

		fmt.Printf("\nAnalysis complete! Output written to: %s\n", outputFileName)
	}
}

func countItems(dir string) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v\n", path, err)
			return nil
		}
		count++
		return nil
	})
	return count
}

func traverseDirectory(dir, indent string, bar *progressbar.ProgressBar) {
	if debug {
		log.Printf("Traversing directory: %s\n", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Error reading directory %s: %v\n", dir, err)
		return
	}

	writeOutput(fmt.Sprintf("\nDIRECTORY: %s\n%s==========================\n", dir, indent))

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			traverseDirectory(path, indent+"  ", bar)
		} else {
			processFile(path, indent+"  ")
		}
		bar.Add(1)
		if debug {
			log.Printf("Processed: %s\n", path)
		}
	}
}

func processFile(file, indent string) {
	if debug {
		log.Printf("Processing file: %s\n", file)
	}

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
		log.Printf("Finished processing file: %s\n", file)
	}
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