package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

func main() {
	// Read the markdown file
	files := []string{"client", "common", "logger", "util", "vaultapi"}

	for _, filename := range files {
        file := "docs/" + filename + ".md"
		data, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		// fmt.Println(data)
		// Convert byte data to string
		markdown := string(data)

		markdown = "{% env enable=\"goSdkRef\" %}\n\n" + markdown 

		// Remove import statement
		markdown = removeImportStatement(markdown)

		// Remove copyright statement
		markdown = removeCopyrightStatement(markdown)

		// Remove functions with Deprecated description
		markdown = removeDeprecatedOrInternal(markdown)

		// Represent type as h2
		markdown = changeTypeHeading(markdown)

		// Represent functions as h3
		markdown = changeFuncHeading(markdown)

		// Remove multiple empty lines
		markdown = removeMultipleEmptyLines(markdown)

		markdown = markdown + "\n{% /env %}"

		// Write the modified markdown back to the file
		err = ioutil.WriteFile(file, []byte(markdown), 0644)
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
    }

	fmt.Println("Markdown files updated successfully.")

}

func removeImportStatement(markdown string) string {
	// Use regex to match and remove import statement
	importRegex := regexp.MustCompile(`--\n\s*import\s+"."\n`)
	markdown = importRegex.ReplaceAllString(markdown, "")

	return markdown
}

func removeCopyrightStatement(markdown string) string {
	// Use regex to match and remove copyright statement
	copyrightRegex := regexp.MustCompile(`\s*Copyright.*`)
	markdown = copyrightRegex.ReplaceAllString(markdown, "")

	return markdown
}

// func removeDeprecatedFunctions(markdown string) string {
// 	// Regular expression pattern to match the deprecated functions
// 	pattern := regexp.MustCompile(`(?s)(?m)#### func\s*((?!#### func).)*?Deprecated.*?\n\n`)
// 	return pattern.ReplaceAllString(markdown, "")
// }

func sum(arr []int) int {
    sum := 0
    for _, valueInt := range arr {
        sum += valueInt
    }
    return sum
}

func removeDeprecatedOrInternal(markdown string) string {
	lines := strings.Split(markdown, "\n")
	var result []string
	var temp []int
	startIndex := 0
	endIndex := 0
	for i, line := range lines {
		if strings.HasPrefix(line, "#### func") {
			startIndex = i
		} else if strings.HasPrefix(line, "#### type") {
			startIndex = i
		} else if strings.HasPrefix(line, "Deprecated") {
			endIndex = i
		} else if strings.HasPrefix(line, "Internal") {
			endIndex = i
		}

		result = append(result, line)
		if startIndex > 0 && endIndex > 0 {
			result = append(result[:startIndex-(sum(temp)+1)])
			temp = append(temp, (endIndex-startIndex)+1)
			startIndex = 0
			endIndex = 0
		}
	}

	return strings.Join(result, "\n")
}

func changeTypeHeading(markdown string) string {
	// Regular expression pattern to match type headings (h4)
	pattern := regexp.MustCompile(`#### type`)

	// Replace type heading (h4) with h2
	return pattern.ReplaceAllStringFunc(markdown, func(match string) string {
		return strings.Replace(match, "####", "##", 1)
	})
}

func changeFuncHeading(markdown string) string {
	// Regular expression pattern to match func headings (h4)
	pattern := regexp.MustCompile(`#### func`)

	// Replace func heading (h4) with h3
	return pattern.ReplaceAllStringFunc(markdown, func(match string) string {
		return strings.Replace(match, "####", "###", 1)
	})
}

func removeMultipleEmptyLines(markdown string) string {
	// Regular expression pattern to match multiple empty lines
	pattern := regexp.MustCompile(`\n{2,}`)

	// Replace multiple empty lines with a single empty line
	return pattern.ReplaceAllString(markdown, "\n\n")
}