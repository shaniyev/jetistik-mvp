package batch

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ParseResult holds the parsed columns and rows from a CSV/XLSX file.
type ParseResult struct {
	Columns []string
	Rows    []map[string]string
}

// ParseCSV reads a CSV file and returns columns and rows.
func ParseCSV(r io.Reader) (*ParseResult, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv headers: %w", err)
	}

	// Clean headers
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
		// Remove BOM if present
		headers[i] = strings.TrimPrefix(headers[i], "\xef\xbb\xbf")
	}

	var rows []map[string]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // skip malformed rows
		}
		row := make(map[string]string, len(headers))
		for i, h := range headers {
			if h == "" {
				continue
			}
			if i < len(record) {
				row[h] = strings.TrimSpace(record[i])
			} else {
				row[h] = ""
			}
		}
		rows = append(rows, row)
	}

	return &ParseResult{Columns: headers, Rows: rows}, nil
}

// ParseXLSX reads an XLSX file and returns columns and rows.
func ParseXLSX(r io.Reader) (*ParseResult, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()

	// Use the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in xlsx")
	}
	sheetName := sheets[0]

	xlsxRows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("read xlsx rows: %w", err)
	}

	if len(xlsxRows) == 0 {
		return &ParseResult{Columns: []string{}, Rows: []map[string]string{}}, nil
	}

	// First row is headers
	headers := make([]string, len(xlsxRows[0]))
	for i, h := range xlsxRows[0] {
		headers[i] = strings.TrimSpace(h)
	}

	var rows []map[string]string
	for _, xlsxRow := range xlsxRows[1:] {
		// Skip empty rows
		allEmpty := true
		for _, cell := range xlsxRow {
			if strings.TrimSpace(cell) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			continue
		}

		row := make(map[string]string, len(headers))
		for i, h := range headers {
			if h == "" {
				continue
			}
			if i < len(xlsxRow) {
				row[h] = strings.TrimSpace(xlsxRow[i])
			} else {
				row[h] = ""
			}
		}
		rows = append(rows, row)
	}

	return &ParseResult{Columns: headers, Rows: rows}, nil
}

// DefaultMapping creates a default column-to-token mapping based on column names.
// Ported from v1/core/utils.py default_mapping.
func DefaultMapping(columns []string, tokens []string) map[string]string {
	cols := make(map[string]string, len(columns))
	for _, c := range columns {
		if c != "" {
			cols[strings.ToLower(strings.TrimSpace(c))] = c
		}
	}

	pick := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := cols[k]; ok {
				return v
			}
		}
		return ""
	}

	mapping := make(map[string]string)
	mapping["name"] = pick("name", "fullname", "fio", "\u0444\u0438\u043e")
	mapping["id"] = pick("id", "code", "\u043d\u043e\u043c\u0435\u0440", "\u043d\u043e\u043c\u0435\u0440/\u043a\u043e\u0434 \u0434\u0438\u043f\u043b\u043e\u043c\u0430")
	mapping["school"] = pick("school", "school_name", "\u0448\u043a\u043e\u043b\u0430")
	mapping["class"] = pick("class", "grade", "\u043a\u043b\u0430\u0441\u0441")
	mapping["place"] = pick("place", "degree", "\u043c\u0435\u0441\u0442\u043e", "\u0441\u0442\u0435\u043f\u0435\u043d\u044c")
	mapping["teacher"] = pick("teacher", "teacher_name", "\u0443\u0447\u0438\u0442\u0435\u043b\u044c")
	mapping["nomination"] = pick("nomination", "category", "\u043d\u043e\u043c\u0438\u043d\u0430\u0446\u0438\u044f")
	mapping["text"] = pick("text", "subtitle", "description", "\u043e\u043f\u0438\u0441\u0430\u043d\u0438\u0435", "\u0442\u0435\u043a\u0441\u0442")

	// Legacy f* tokens mirror canonical ones
	mapping["fname"] = mapping["name"]
	mapping["fid"] = mapping["id"]
	mapping["fschool"] = mapping["school"]
	mapping["fclass"] = mapping["class"]
	mapping["fplace"] = mapping["place"]
	mapping["fteacher"] = mapping["teacher"]
	mapping["fnomination"] = mapping["nomination"]
	mapping["ftext"] = mapping["text"]

	// Only return tokens that were requested
	result := make(map[string]string, len(tokens))
	for _, tok := range tokens {
		if v, ok := mapping[tok]; ok {
			result[tok] = v
		} else {
			// Try direct match
			key := strings.TrimPrefix(tok, "f")
			if v, ok := cols[key]; ok {
				result[tok] = v
			} else {
				result[tok] = ""
			}
		}
	}
	return result
}
