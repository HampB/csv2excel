// Package file provides functionality for reading, processing, and saving CSV files.

// It includes features for inferring column data types and saving the data to Excel files.

package file

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func Test_readCSV(t *testing.T) {

	type args struct {
		reader    io.Reader
		delimiter rune
	}
	tests := []struct {
		name    string
		args    args
		want    [][]string
		wantErr bool
	}{
		{
			name: "Read CSV with comma delimiter",
			args: args{
				reader:    bytes.NewBufferString("a,b,c\nd,e,f"),
				delimiter: ',',
			},
			want: [][]string{
				{"a", "b", "c"},
				{"d", "e", "f"},
			},
			wantErr: false,
		},
		{
			name: "Read CSV with semicolon delimiter",
			args: args{
				reader:    bytes.NewBufferString("a;b;c\nd;e;f"),
				delimiter: ';',
			},
			want: [][]string{
				{"a", "b", "c"},
				{"d", "e", "f"},
			},
			wantErr: false,
		},
		{
			name: "Read CSV with tab delimiter",
			args: args{
				reader:    bytes.NewBufferString("a\tb\tc\nd\te\tf"),
				delimiter: '\t',
			},
			want: [][]string{
				{"a", "b", "c"},
				{"d", "e", "f"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readCSV(tt.args.reader, tt.args.delimiter)
			if (err != nil) != tt.wantErr {
				t.Errorf("readCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_New(t *testing.T) {
	tests := []struct {
		name     string
		options  []func(*CSV)
		expected *CSV
	}{
		{
			name: "Create CSV with file path",
			options: []func(*CSV){
				WithFilePath("test.csv"),
			},
			expected: &CSV{
				FilePath: "test.csv",
			},
		},
		{
			name: "Create CSV with delimiter",
			options: []func(*CSV){
				WithDelimiter(';'),
			},
			expected: &CSV{
				Delimiter: ';',
			},
		},
		{
			name: "Create CSV with headers",
			options: []func(*CSV){
				WithHeaders([]Column{
					{Name: "Column1", Type: StringType},
					{Name: "Column2", Type: IntegerType},
				}),
			},
			expected: &CSV{
				Headers: []Column{
					{Name: "Column1", Type: StringType},
					{Name: "Column2", Type: IntegerType},
				},
			},
		},
		{
			name: "Create CSV with records",
			options: []func(*CSV){
				WithRecords([][]interface{}{
					{"a", 1},
					{"b", 2},
				}),
			},
			expected: &CSV{
				Records: [][]interface{}{
					{"a", 1},
					{"b", 2},
				},
			},
		},
		{
			name: "Create CSV with multiple options",
			options: []func(*CSV){
				WithFilePath("test.csv"),
				WithDelimiter(';'),
				WithHeaders([]Column{
					{Name: "Column1", Type: StringType},
					{Name: "Column2", Type: IntegerType},
				}),
				WithRecords([][]interface{}{
					{"a", 1},
					{"b", 2},
				}),
			},
			expected: &CSV{
				FilePath:  "test.csv",
				Delimiter: ';',
				Headers: []Column{
					{Name: "Column1", Type: StringType},
					{Name: "Column2", Type: IntegerType},
				},
				Records: [][]interface{}{
					{"a", 1},
					{"b", 2},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.options...)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("New() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
func Test_CSV_ConvertColumnTypes(t *testing.T) {
	tests := []struct {
		name     string
		csv      *CSV
		expected [][]interface{}
	}{
		{
			name: "Convert string to float and integer",
			csv: &CSV{
				Headers: []Column{
					{Name: "Column1", Type: FloatType},
					{Name: "Column2", Type: IntegerType},
					{Name: "Column3", Type: StringType},
				},
				Records: [][]interface{}{
					{"1.23", "456", "text"},
					{"4.56", "789", "more text"},
				},
			},
			expected: [][]interface{}{
				{float64(1.23), int64(456), "text"},
				{float64(4.56), int64(789), "more text"},
			},
		},
		{
			name: "No conversion needed",
			csv: &CSV{
				Headers: []Column{
					{Name: "Column1", Type: StringType},
					{Name: "Column2", Type: StringType},
				},
				Records: [][]interface{}{
					{"text1", "text2"},
					{"text3", "text4"},
				},
			},
			expected: [][]interface{}{
				{"text1", "text2"},
				{"text3", "text4"},
			},
		},
		{
			name: "Mixed conversion",
			csv: &CSV{
				Headers: []Column{
					{Name: "Column1", Type: FloatType},
					{Name: "Column2", Type: IntegerType},
					{Name: "Column3", Type: StringType},
				},
				Records: [][]interface{}{
					{"1.23", "invalid", "text"},
					{"invalid", "789", "more text"},
				},
			},
			expected: [][]interface{}{
				{float64(1.23), "invalid", "text"},
				{"invalid", int64(789), "more text"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.csv.ConvertColumnTypes()
			for i, row := range tt.csv.Records {
				for j, val := range row {
					fmt.Printf("Actual[%d][%d]: value=%v, type=%T\n", i, j, val, val)
				}
			}
			for i, row := range tt.expected {
				for j, val := range row {
					fmt.Printf("Expected[%d][%d]: value=%v, type=%T\n", i, j, val, val)
				}
			}
			if !reflect.DeepEqual(tt.csv.Records, tt.expected) {
				t.Errorf("ConvertColumnTypes() = %v, expected %v", tt.csv.Records, tt.expected)
			}
		})
	}
}
func Test_CSV_InferColumnTypes(t *testing.T) {
	tests := []struct {
		name     string
		csv      *CSV
		expected []Column
	}{
		{
			name: "Infer types for mixed data",
			csv: &CSV{
				Headers: []Column{
					{Name: "Column1", Type: StringType},
					{Name: "Column2", Type: StringType},
					{Name: "Column3", Type: StringType},
				},
				Records: [][]interface{}{
					{"1.23", "456", "text"},
					{"4.56", "789", "more text"},
				},
			},
			expected: []Column{
				{Name: "Column1", Type: FloatType},
				{Name: "Column2", Type: IntegerType},
				{Name: "Column3", Type: StringType},
			},
		},
		{
			name: "Infer types for all strings",
			csv: &CSV{
				Headers: []Column{
					{Name: "Column1", Type: StringType},
					{Name: "Column2", Type: StringType},
				},
				Records: [][]interface{}{
					{"text1", "text2"},
					{"text3", "text4"},
				},
			},
			expected: []Column{
				{Name: "Column1", Type: StringType},
				{Name: "Column2", Type: StringType},
			},
		},
		{
			name: "Infer types for all integers",
			csv: &CSV{
				Headers: []Column{
					{Name: "Column1", Type: StringType},
					{Name: "Column2", Type: StringType},
				},
				Records: [][]interface{}{
					{"123", "456"},
					{"789", "101112"},
				},
			},
			expected: []Column{
				{Name: "Column1", Type: IntegerType},
				{Name: "Column2", Type: IntegerType},
			},
		},
		{
			name: "Infer types for all floats",
			csv: &CSV{
				Headers: []Column{
					{Name: "Column1", Type: StringType},
					{Name: "Column2", Type: StringType},
				},
				Records: [][]interface{}{
					{"1.23", "4.56"},
					{"7.89", "10.11"},
				},
			},
			expected: []Column{
				{Name: "Column1", Type: FloatType},
				{Name: "Column2", Type: FloatType},
			},
		},
		{
			name: "Infer types for mixed valid and invalid data",
			csv: &CSV{
				Headers: []Column{
					{Name: "Column1", Type: StringType},
					{Name: "Column2", Type: StringType},
				},
				Records: [][]interface{}{
					{"1.23", "invalid"},
					{"invalid", "789"},
				},
			},
			expected: []Column{
				{Name: "Column1", Type: StringType},
				{Name: "Column2", Type: StringType},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.csv.InferColumnTypes()
			if !reflect.DeepEqual(tt.csv.Headers, tt.expected) {
				t.Errorf("InferColumnTypes() = %v, expected %v", tt.csv.Headers, tt.expected)
			}
		})
	}
}
func Test_CSV_GetHeaderNames(t *testing.T) {
	tests := []struct {
		name     string
		csv      *CSV
		expected []string
	}{
		{
			name: "Get header names from CSV with headers",
			csv: &CSV{
				Headers: []Column{
					{Name: "Column1", Type: StringType},
					{Name: "Column2", Type: IntegerType},
					{Name: "Column3", Type: FloatType},
				},
			},
			expected: []string{"Column1", "Column2", "Column3"},
		},
		{
			name:     "Get header names from CSV with no headers",
			csv:      &CSV{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.csv.GetHeaderNames()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetHeaderNames() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
