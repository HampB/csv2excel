# csv2excel

`csv2excel` is a CLI tool written in Go that allows you to convert CSV files to Excel format. It supports specifying the input CSV file, output Excel file, and the delimiter used in the CSV file. Additionally, it can infer and convert column types from strings to integers or floats.

**Note:** This is my first Go project, and I appreciate any feedback or contributions to improve it.

## Features

- Convert CSV files to Excel format
- Specify custom delimiters for CSV files
- Infer and convert column types (string to integer or float)
- Specify output file name or path

## Installation

To install `csv2excel`, you need to have Go installed on your machine. Then, you can use the following command to install the tool:

```sh
go install github.com/HampB/csv2excel@latest
```

## Usage

You can use the `csv2excel` command to convert a CSV file to an Excel file. Below are the available options:

```sh
csv2excel -i <input-file> -o <output-file> -d <delimiter> -c
```

### Options

- `-i, --input`: Path to the input CSV file (required)
- `-o, --output`: Path to the output Excel file (optional)
- `-n, --name`: Name of the output Excel file (optional)
- `-d, --delimiter`: Delimiter for CSV file (default is `,`)
- `-c, --convert`: Convert column types to inferred types (optional)

### Examples

Convert a CSV file to an Excel file with default settings:

```sh
csv2excel -i data.csv
```

Convert a CSV file with a custom delimiter and infer column types:

```sh
csv2excel -i data.csv -d ";" -c
```

Specify the output file name:

```sh
csv2excel -i data.csv -n output
```

Specify the full path for the output file:

```sh
csv2excel -i data.csv -o /path/to/output.xlsx
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
