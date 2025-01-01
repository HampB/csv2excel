package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/HampB/csv2excel/internal/file"
	"github.com/spf13/cobra"
)

// mergeCmd represents the merge command
var (
	inputFiles  []string
	inputFolder string

	mergeCmd = &cobra.Command{
		Use:   "merge",
		Short: "Merge multiple CSV files into a single Excel file",
		Long: `The merge command allows you to combine multiple CSV files into a single Excel file.
You can specify individual CSV files or a folder containing CSV files. For example:

csv2excel merge --files file1.csv,file2.csv --output result.xlsx
csv2excel merge --folder /path/to/csvfiles --output result.xlsx`,
		Run: func(cmd *cobra.Command, args []string) {
			if delimiter == "" {
				fmt.Println("Delimiter cannot be empty")
				return
			}
			delimiterRune := []rune(delimiter)[0]

			if inputFolder != "" {
				var err error
				inputFiles, err = createFileList(inputFolder)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			f, err := processFiles(inputFiles, delimiterRune)

			if err != nil {
				fmt.Println(err)
				return
			}

			if convertTypes {
				f.ConvertColumnTypes()
			}
			if outputFile == "" && outputName == "" {
				outputFile = strings.Replace(inputFile, ".csv", ".xlsx", 1)
			}
			if outputName != "" {
				outputFile = filepath.Join(filepath.Dir(inputFile), outputName+".xlsx")
			}
			if _, err := os.Stat(filepath.Dir(outputFile)); os.IsNotExist(err) {
				fmt.Printf("Invalid output path: %s\n", filepath.Dir(outputFile))
				return
			}
			err = f.SaveAsExcel(outputFile, "Sheet1")
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Successfully converted %d records with %d columns to %s\n", len(f.Records), len(f.Headers), outputFile)
		},
	}
)

func init() {
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().StringSliceVarP(&inputFiles, "files", "f", []string{}, "List of CSV files to merge")
	mergeCmd.Flags().StringVarP(&inputFolder, "folder", "F", "", "Path to the folder containing CSV files")
	mergeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Path to the output Excel file")
	mergeCmd.Flags().StringVarP(&delimiter, "delimiter", "d", ",", "Delimiter for CSV file")
	mergeCmd.Flags().BoolVarP(&convertTypes, "convert", "c", false, "Convert column types to inferred types")

	mergeCmd.MarkFlagsOneRequired("files", "folder")
	mergeCmd.MarkFlagsMutuallyExclusive("files", "folder")
	mergeCmd.MarkFlagRequired("output")
}

// processFiles reads and processes multiple CSV files concurrently.
// It takes a slice of file paths and a delimiter as input, and returns a merged CSV file or an error.
func processFiles(filePaths []string, delimiter rune) (*file.CSV, error) {
	wg := sync.WaitGroup{}
	resultChannel := make(chan processResult)

	for _, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)
		if !strings.HasSuffix(filePath, ".csv") {
			return nil, fmt.Errorf("invalid input file format. Please provide a CSV file")
		}
		wg.Add(1)
		go func(filePath string, delimiter rune) {
			defer wg.Done()
			f := file.New(
				file.WithFilePath(filePath),
				file.WithDelimiter(delimiter),
			)
			err := f.Read()
			if err != nil {
				resultChannel <- processResult{err: err}
				return
			}
			resultChannel <- processResult{file: f}
		}(filePath, delimiter)
	}
	go func() {
		wg.Wait()
		close(resultChannel)
	}()
	var errors []error
	var files = make([]*file.CSV, 0, len(filePaths))
	for result := range resultChannel {
		if result.err != nil {
			errors = append(errors, result.err)
		} else {
			files = append(files, result.file)
		}
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("encountered errors while reading files: %v", errors)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no valid CSV files to merge")
	}
	return file.Merge(files...)

}

// processResult represents the result of processing a CSV file.
// It contains a pointer to the processed CSV file and an error, if any occurred during processing.
type processResult struct {
	file *file.CSV
	err  error
}

// createFileList scans the specified folder for files with a .csv extension
// and returns a slice of their file paths. If an error occurs during reading
// the directory, it returns the error.
func createFileList(folderPath string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".csv") {
			files = append(files, filepath.Join(folderPath, entry.Name()))
		}
	}
	return files, nil
}
