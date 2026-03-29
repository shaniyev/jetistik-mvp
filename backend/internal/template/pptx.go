package template

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

// TokenList defines the canonical set of tokens from v1.
var TokenList = []string{
	"name", "school", "class", "place", "teacher", "nomination", "id", "text",
	"fqr",
	"fname", "fschool", "fclass", "fplace", "fteacher", "fnomination", "fid", "ftext",
}

var (
	patternF      = regexp.MustCompile(`(?i)\bf[a-z0-9_]+\b`)
	patternBraces = regexp.MustCompile(`(?i)\{([a-z0-9_]+)\}`)
)

// ExtractTokensFromPPTX reads a PPTX file (as io.ReaderAt + size) and extracts tokens.
// Ported from v1/core/utils.py extract_tokens_from_pptx.
func ExtractTokensFromPPTX(r io.ReaderAt, size int64) ([]string, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("open pptx zip: %w", err)
	}

	tokens := make(map[string]bool)

	// Scan slide XML files for text content
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		// Check slides, slide layouts, and slide masters
		if !strings.HasPrefix(name, "ppt/slides/") &&
			!strings.HasPrefix(name, "ppt/slidelayouts/") &&
			!strings.HasPrefix(name, "ppt/slidemasters/") {
			continue
		}
		if !strings.HasSuffix(name, ".xml") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}
		text := extractTextFromXML(rc)
		rc.Close()

		for _, m := range patternF.FindAllString(text, -1) {
			tokens[strings.ToLower(m)] = true
		}
		for _, m := range patternBraces.FindAllStringSubmatch(text, -1) {
			if len(m) > 1 {
				tokens[strings.ToLower(m[1])] = true
			}
		}

		// Check for shape names containing "qr" (simplified: check if any element has name="QR")
		if strings.Contains(text, "qr") || strings.Contains(text, "QR") {
			// Check if "qr" appears as a standalone word or shape name
			if strings.Contains(strings.ToLower(text), "fqr") || strings.Contains(text, "{qr}") {
				tokens["fqr"] = true
			}
		}
	}

	// Also scan for shape names "QR" in slide XML (check nvSpPr/nvPr name attributes)
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		if !strings.HasPrefix(name, "ppt/slides/slide") || !strings.HasSuffix(name, ".xml") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		if hasQRShapeName(rc) {
			tokens["fqr"] = true
		}
		rc.Close()
	}

	// Normalize: if "qr" is in tokens, replace with "fqr"
	if tokens["qr"] {
		delete(tokens, "qr")
		tokens["fqr"] = true
	}

	// Order: canonical tokens first, then any extras sorted
	ordered := make([]string, 0, len(tokens))
	for _, t := range TokenList {
		if tokens[t] {
			ordered = append(ordered, t)
			delete(tokens, t)
		}
	}
	extras := make([]string, 0, len(tokens))
	for t := range tokens {
		extras = append(extras, t)
	}
	sort.Strings(extras)
	ordered = append(ordered, extras...)

	return ordered, nil
}

// extractTextFromXML reads all text content from a PPTX XML file.
func extractTextFromXML(r io.Reader) string {
	decoder := xml.NewDecoder(r)
	var texts []string
	var inText bool

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			// a:t elements contain text in PPTX
			if t.Name.Local == "t" {
				inText = true
			}
			// Check nvSpPr name attribute for shape names
			for _, attr := range t.Attr {
				if attr.Name.Local == "name" {
					texts = append(texts, attr.Value)
				}
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inText = false
			}
		case xml.CharData:
			if inText {
				texts = append(texts, string(t))
			}
		}
	}

	return strings.Join(texts, " ")
}

// hasQRShapeName checks if any shape in the slide XML has name="QR" (case-insensitive).
func hasQRShapeName(r io.Reader) bool {
	decoder := xml.NewDecoder(r)
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		if start, ok := tok.(xml.StartElement); ok {
			for _, attr := range start.Attr {
				if attr.Name.Local == "name" && strings.EqualFold(attr.Value, "qr") {
					return true
				}
			}
		}
	}
	return false
}
