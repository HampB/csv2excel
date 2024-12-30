package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	file "github.com/HampB/csv2excel/internal"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	inputFile    string
	outputFile   string
	outputName   string
	delimiter    string
	convertTypes bool

	rootCmd = &cobra.Command{
		Use:   "csv2excel",
		Short: "Convert CSV files to Excel format",
		Long: `csv2excel is a CLI tool that allows you to convert CSV files to Excel format.
You can specify the input CSV file, output Excel file, and the delimiter used in the CSV file.`,
		Run: func(cmd *cobra.Command, args []string) {
			if delimiter == "" {
				fmt.Println("Delimiter cannot be empty")
				return
			}
			delimiterRune := []rune(delimiter)[0]
			if !strings.HasSuffix(inputFile, ".csv") {
				fmt.Println("Invalid input file format. Please provide a CSV file.")
				return
			}
			f := file.New(inputFile, delimiterRune)
			err := f.Read()
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

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Path to the input CSV file")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Path to the output Excel file")
	rootCmd.Flags().StringVarP(&outputName, "name", "n", "", "Name of the output Excel file")
	rootCmd.Flags().StringVarP(&delimiter, "delimiter", "d", ",", "Delimiter for CSV file")
	rootCmd.Flags().BoolVarP(&convertTypes, "convert", "c", false, "Convert column types to inferred types")

	rootCmd.MarkFlagRequired("input")
	rootCmd.MarkFlagFilename("input", "csv")
	rootCmd.MarkFlagsMutuallyExclusive("output", "name")
}
