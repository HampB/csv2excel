// Package file provides functionality for reading, processing, and saving CSV files.
// It includes features for inferring column data types and saving the data to Excel files.
package file

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/xuri/excelize/v2"
)

const defaultTypeInferanceRows = 20

type ColumnType int

const (
	StringType ColumnType = iota + 1
	FloatType
	IntegerType
)

// Column represents a column in the CSV file, including its name and inferred data type.
type Column struct {
	// Name is the name of the column, typically read from the header row.
	Name string
	// Type is the inferred data type of the column.
	Type ColumnType
}

// CSV represents a CSV file and its parsed data.
type CSV struct {
	// FilePath is the path to the CSV file.
	FilePath string
	// Delimiter is the character used to separate fields in the CSV file.
	Delimiter rune
	// Columns is a slice of Column information.
	Headers []Column
	// Records is a slice of slices, where each inner slice represents a row of data.
	// The data type of the elements within the inner slices can vary based on type inference.
	Records [][]interface{}
}

// New creates a new CSV struct with the specified options.
func New(options ...func(*CSV)) *CSV {
	csv := &CSV{}
	for _, option := range options {
		option(csv)
	}
	return csv
}

// WithFilePath sets the file path for the CSV struct.
func WithFilePath(filePath string) func(*CSV) {
	return func(c *CSV) {
		c.FilePath = filePath
	}
}

// WithDelimiter sets the delimiter for the CSV struct.
func WithDelimiter(delimiter rune) func(*CSV) {
	return func(c *CSV) {
		c.Delimiter = delimiter
	}
}

// WithHeaders sets the column headers for the CSV struct.
func WithHeaders(columns []Column) func(*CSV) {
	return func(c *CSV) {
		c.Headers = columns
	}
}

// WithRecords sets the data records for the CSV struct.
func WithRecords(records [][]interface{}) func(*CSV) {
	return func(c *CSV) {
		c.Records = records
	}
}

// Read reads the CSV file, parses its contents, and populates the CSV struct.
// It infers column names from the first row and stores the data in the Records field.
// Returns an error if the file cannot be opened or read.
func (c *CSV) Read() error {
	if c.FilePath == "" {
		return fmt.Errorf("file path is empty, a valid file path is required")
	}
	file, err := os.Open(c.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = c.Delimiter

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	noOfRecords := len(records)
	if noOfRecords == 0 {
		return fmt.Errorf("no records found in %s", c.FilePath)
	}

	for _, column := range records[0] {
		c.Headers = append(c.Headers, Column{
			Name: column,
			Type: StringType,
		})
	}

	if noOfRecords > 1 {
		c.Records = make([][]interface{}, len(records)-1)
		for i, record := range records[1:] {
			c.Records[i] = make([]interface{}, len(record))
			for j, value := range record {
				c.Records[i][j] = value
			}
		}
	}
	return nil
}

// ConvertColumnTypes attempts to convert string values in the Records to their inferred types (float or integer).
// This function relies on the inferColumnTypes method to determine the appropriate type for each column.
func (c *CSV) ConvertColumnTypes() {
	c.inferColumnTypes()
	for _, record := range c.Records {
		for i := range c.Headers {
			if stringValue, ok := record[i].(string); ok {
				switch c.Headers[i].Type {
				case FloatType:
					if parsedValue, err := strconv.ParseFloat(stringValue, 64); err == nil {
						record[i] = parsedValue
					}
				case IntegerType:
					if parsedValue, err := strconv.ParseInt(stringValue, 10, 64); err == nil {
						record[i] = parsedValue
					}
				}
			}
		}
	}
}

// inferColumnTypes analyzes a sample of rows to infer the data type of each column.
// It checks if the values in a column can be parsed as float or integer.
// The number of rows to inspect is determined by the defaultTypeInferanceRows constant.
func (c *CSV) inferColumnTypes() {
	rangeToCheck := min(defaultTypeInferanceRows, len(c.Records))
	for i := range c.Headers {
		floatCount, intCount := 0, 0
		for _, record := range c.Records[:rangeToCheck] {
			if stringValue, ok := record[i].(string); ok {
				if _, err := strconv.ParseFloat(stringValue, 64); err == nil {
					floatCount++
				} else if _, err := strconv.ParseInt(stringValue, 10, 64); err == nil {
					intCount++
				}
			}
		}
		if floatCount == rangeToCheck {
			c.Headers[i].Type = FloatType
		} else if intCount == rangeToCheck {
			c.Headers[i].Type = IntegerType
		}
	}
}

// SaveAsExcel saves the CSV data to an Excel file.
// It creates a new Excel file and writes the column names and data records to the specified sheet.
// Returns an error if the file cannot be created or written to.
func (c *CSV) SaveAsExcel(filePath string, sheetName string) error {

	f := excelize.NewFile()

	defer func() error {
		if err := f.Close(); err != nil {
			return err
		}
		return nil
	}()

	if sheetName == "" {
		sheetName = "Sheet1"
	}
	headerNames := c.GetHeaderNames()
	f.SetSheetRow(sheetName, "A1", &headerNames)

	for i, record := range c.Records {
		row := fmt.Sprintf("A%d", i+2)
		f.SetSheetRow(sheetName, row, &record)
	}
	if err := f.SaveAs(filePath); err != nil {
		return err
	}
	return nil
}

// GetHeaderNames returns a Slice with the names of the columns in the CSV file.
func (c *CSV) GetHeaderNames() []string {
	columnNames := make([]string, len(c.Headers))
	for i, column := range c.Headers {
		columnNames[i] = column.Name
	}
	return columnNames
}

func Merge(files ...*CSV) (*CSV, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files to merge")
	}
	columnCount := len(files[0].Headers)

	for _, file := range files[1:] {
		if columnCount != len(file.Headers) {
			return nil, fmt.Errorf("inconsistent number of columns in either %s or %s", file.FilePath, files[0].FilePath)
		}
	}
	mergedFiles := make([][]interface{}, 0)
	for _, file := range files {
		mergedFiles = append(mergedFiles, file.Records...)
	}
	f := New(
		WithHeaders(files[0].Headers),
		WithDelimiter(files[0].Delimiter),
		WithRecords(mergedFiles),
	)
	return f, nil
}
