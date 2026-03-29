package worker

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

// makeQRPNGBytes generates a QR code PNG for the given URL.
func makeQRPNGBytes(url string) ([]byte, error) {
	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("generate qr: %w", err)
	}
	return png, nil
}

// replaceTokensInPPTX opens a PPTX file, replaces tokens in slide text, inserts a QR code,
// and writes the modified PPTX to a new temp file. Returns path to the new PPTX.
func replaceTokensInPPTX(pptxData []byte, tokenValues map[string]string, qrPNG []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(pptxData), int64(len(pptxData)))
	if err != nil {
		return nil, fmt.Errorf("open pptx zip: %w", err)
	}

	// Build expanded token map: include both "token" and "{token}" forms
	expanded := make(map[string]string, len(tokenValues)*2)
	for k, v := range tokenValues {
		expanded[k] = v
		if k != "" && !strings.HasPrefix(k, "{") {
			expanded["{"+k+"}"] = v
		}
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Track QR insertion state
	qrInserted := false
	var qrShapeInfo *shapeInfo

	// First pass: find QR shape info in slide XMLs
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		if !strings.HasPrefix(name, "ppt/slides/slide") || !strings.HasSuffix(name, ".xml") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}
		if info := findQRShape(data); info != nil {
			qrShapeInfo = info
			qrShapeInfo.slidePath = f.Name
			break
		}
	}

	// Add QR image to pptx if we found a shape or have QR data
	qrImagePath := ""
	qrRID := ""
	if qrPNG != nil && len(qrPNG) > 0 {
		qrImagePath = "ppt/media/qr_generated.png"
	}

	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open zip entry %s: %w", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("read zip entry %s: %w", f.Name, err)
		}

		name := strings.ToLower(f.Name)
		isSlideXML := strings.HasPrefix(name, "ppt/slides/slide") && strings.HasSuffix(name, ".xml")

		if isSlideXML {
			// Replace tokens in slide XML
			data = replaceTokensInXML(data, expanded)

			// Handle QR insertion
			if qrPNG != nil && !qrInserted {
				if qrShapeInfo != nil && strings.EqualFold(f.Name, qrShapeInfo.slidePath) {
					// Replace the shape with a picture reference
					data, qrRID = replaceQRShapeWithPicture(data, qrShapeInfo, qrImagePath)
					qrInserted = true
				} else if qrShapeInfo == nil {
					// Try text marker fallback
					newData, rid, ok := replaceQRTextMarker(data, qrImagePath)
					if ok {
						data = newData
						qrRID = rid
						qrInserted = true
						qrShapeInfo = &shapeInfo{slidePath: f.Name}
					}
				}
			}
		}

		// Update rels file for the slide where QR was inserted
		if qrInserted && qrRID != "" && qrShapeInfo != nil {
			slideFilename := filepath.Base(qrShapeInfo.slidePath)
			relsPath := "ppt/slides/_rels/" + slideFilename + ".rels"
			if strings.EqualFold(f.Name, relsPath) {
				data = addImageRelationship(data, qrRID, "../media/qr_generated.png")
				qrRID = "" // Only add once
			}
		}

		// Update [Content_Types].xml to include PNG content type
		if strings.EqualFold(f.Name, "[Content_Types].xml") && qrImagePath != "" {
			data = ensurePNGContentType(data)
		}

		w, err := zw.Create(f.Name)
		if err != nil {
			return nil, fmt.Errorf("create zip entry %s: %w", f.Name, err)
		}
		if _, err := w.Write(data); err != nil {
			return nil, fmt.Errorf("write zip entry %s: %w", f.Name, err)
		}
	}

	// Add the QR image file
	if qrInserted && qrPNG != nil {
		w, err := zw.Create(qrImagePath)
		if err != nil {
			return nil, fmt.Errorf("create qr image entry: %w", err)
		}
		if _, err := w.Write(qrPNG); err != nil {
			return nil, fmt.Errorf("write qr image: %w", err)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close zip: %w", err)
	}

	return buf.Bytes(), nil
}

type shapeInfo struct {
	slidePath string
	name      string
	// Position attributes from the XML
	x, y, cx, cy string
}

// findQRShape looks for a shape named "QR" (case-insensitive) or containing fqr/{qr} text.
func findQRShape(slideXML []byte) *shapeInfo {
	decoder := xml.NewDecoder(bytes.NewReader(slideXML))
	var inSp bool
	var currentName string
	var hasQRText bool
	var x, y, cx, cy string

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "sp" {
				inSp = true
				currentName = ""
				hasQRText = false
				x, y, cx, cy = "", "", "", ""
			}
			if inSp {
				// Check for shape name in nvSpPr > cNvPr
				if t.Name.Local == "cNvPr" {
					for _, attr := range t.Attr {
						if attr.Name.Local == "name" && strings.EqualFold(attr.Value, "qr") {
							currentName = attr.Value
						}
					}
				}
				// Capture position from spPr > xfrm > off and ext
				if t.Name.Local == "off" {
					for _, attr := range t.Attr {
						if attr.Name.Local == "x" {
							x = attr.Value
						}
						if attr.Name.Local == "y" {
							y = attr.Value
						}
					}
				}
				if t.Name.Local == "ext" {
					for _, attr := range t.Attr {
						if attr.Name.Local == "cx" {
							cx = attr.Value
						}
						if attr.Name.Local == "cy" {
							cy = attr.Value
						}
					}
				}
			}
		case xml.CharData:
			if inSp {
				text := strings.TrimSpace(string(t))
				textLower := strings.ToLower(text)
				if strings.Contains(textLower, "fqr") || strings.Contains(textLower, "{qr}") || textLower == "qr" {
					hasQRText = true
				}
			}
		case xml.EndElement:
			if t.Name.Local == "sp" {
				if inSp && (strings.EqualFold(currentName, "qr") || hasQRText) {
					return &shapeInfo{
						name: currentName,
						x:    x, y: y,
						cx: cx, cy: cy,
					}
				}
				inSp = false
			}
		}
	}
	return nil
}

// replaceTokensInXML replaces token strings in the text content of slide XML.
// Handles tokens split across multiple <a:r>/<a:t> runs by working on the raw XML.
func replaceTokensInXML(xmlData []byte, tokenValues map[string]string) []byte {
	// Sort tokens by length descending so longer matches replace first
	tokens := make([]string, 0, len(tokenValues))
	for k := range tokenValues {
		if k != "" {
			tokens = append(tokens, k)
		}
	}
	sort.Slice(tokens, func(i, j int) bool {
		return len(tokens[i]) > len(tokens[j])
	})

	result := xmlData
	for _, token := range tokens {
		value := tokenValues[token]
		// Escape XML special characters in the value
		value = xmlEscape(value)
		// Direct replacement in the XML text content
		// This handles the common case where tokens appear within a single <a:t> element
		result = bytes.ReplaceAll(result, []byte(token), []byte(value))
	}

	return result
}

// xmlEscape escapes special XML characters.
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// replaceQRShapeWithPicture replaces a shape element with a picture element referencing the QR image.
// Returns the modified XML and the relationship ID used.
func replaceQRShapeWithPicture(xmlData []byte, info *shapeInfo, imagePath string) ([]byte, string) {
	rid := "rIdQR1"

	// Make QR square: use cy (height) for both dimensions
	width := info.cy
	if width == "" {
		width = info.cx
	}
	height := info.cy
	if height == "" {
		height = "1000000"
	}
	if info.x == "" {
		info.x = "0"
	}
	if info.y == "" {
		info.y = "0"
	}

	// Build the picture XML element
	picXML := fmt.Sprintf(`<p:pic xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <p:nvPicPr>
    <p:cNvPr id="9999" name="QR Code"/>
    <p:cNvPicPr><a:picLocks noChangeAspect="1"/></p:cNvPicPr>
    <p:nvPr/>
  </p:nvPicPr>
  <p:blipFill>
    <a:blip r:embed="%s"/>
    <a:stretch><a:fillRect/></a:stretch>
  </p:blipFill>
  <p:spPr>
    <a:xfrm>
      <a:off x="%s" y="%s"/>
      <a:ext cx="%s" cy="%s"/>
    </a:xfrm>
    <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
  </p:spPr>
</p:pic>`, rid, info.x, info.y, width, height)

	// Find and replace the QR shape element
	// We need to find the <p:sp> element that contains the QR shape
	namePattern := info.name
	if namePattern == "" {
		namePattern = "qr"
	}

	// Use regex to find the sp element containing the QR shape name or text
	// Match the entire <p:sp>...</p:sp> block
	patternStr := `(?s)<p:sp\b[^>]*>.*?` + regexp.QuoteMeta(namePattern) + `.*?</p:sp>`
	re, err := regexp.Compile(patternStr)
	if err != nil {
		// Fallback: try case-insensitive
		re = regexp.MustCompile(`(?si)<p:sp\b[^>]*>.*?(?:fqr|\{qr\}|name="[Qq][Rr]").*?</p:sp>`)
	}

	loc := re.FindIndex(xmlData)
	if loc != nil {
		result := make([]byte, 0, len(xmlData)+len(picXML))
		result = append(result, xmlData[:loc[0]]...)
		result = append(result, []byte(picXML)...)
		result = append(result, xmlData[loc[1]:]...)
		return result, rid
	}

	return xmlData, ""
}

// replaceQRTextMarker finds a shape with fqr/{qr}/qr text and replaces it with a picture.
func replaceQRTextMarker(xmlData []byte, imagePath string) ([]byte, string, bool) {
	re := regexp.MustCompile(`(?si)<p:sp\b[^>]*>.*?(?:fqr|\{qr\}).*?</p:sp>`)
	loc := re.FindIndex(xmlData)
	if loc == nil {
		return xmlData, "", false
	}

	// Extract position from the matched shape
	matched := xmlData[loc[0]:loc[1]]
	info := extractPositionFromShape(matched)

	rid := "rIdQR1"
	width := info.cy
	if width == "" {
		width = info.cx
	}
	if width == "" {
		width = "1000000"
	}
	height := info.cy
	if height == "" {
		height = "1000000"
	}
	if info.x == "" {
		info.x = "0"
	}
	if info.y == "" {
		info.y = "0"
	}

	picXML := fmt.Sprintf(`<p:pic xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <p:nvPicPr>
    <p:cNvPr id="9999" name="QR Code"/>
    <p:cNvPicPr><a:picLocks noChangeAspect="1"/></p:cNvPicPr>
    <p:nvPr/>
  </p:nvPicPr>
  <p:blipFill>
    <a:blip r:embed="%s"/>
    <a:stretch><a:fillRect/></a:stretch>
  </p:blipFill>
  <p:spPr>
    <a:xfrm>
      <a:off x="%s" y="%s"/>
      <a:ext cx="%s" cy="%s"/>
    </a:xfrm>
    <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
  </p:spPr>
</p:pic>`, rid, info.x, info.y, width, height)

	result := make([]byte, 0, len(xmlData)+len(picXML))
	result = append(result, xmlData[:loc[0]]...)
	result = append(result, []byte(picXML)...)
	result = append(result, xmlData[loc[1]:]...)
	return result, rid, true
}

func extractPositionFromShape(shapeXML []byte) *shapeInfo {
	info := &shapeInfo{}
	// Extract offset
	offRe := regexp.MustCompile(`<a:off[^/]*x="(\d+)"[^/]*y="(\d+)"`)
	if m := offRe.FindSubmatch(shapeXML); len(m) > 2 {
		info.x = string(m[1])
		info.y = string(m[2])
	}
	// Extract extent
	extRe := regexp.MustCompile(`<a:ext[^/]*cx="(\d+)"[^/]*cy="(\d+)"`)
	if m := extRe.FindSubmatch(shapeXML); len(m) > 2 {
		info.cx = string(m[1])
		info.cy = string(m[2])
	}
	return info
}

// addImageRelationship adds an image relationship to the rels XML.
func addImageRelationship(relsXML []byte, rid, target string) []byte {
	relElement := fmt.Sprintf(
		`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="%s"/>`,
		rid, target,
	)
	// Insert before closing </Relationships>
	closing := []byte("</Relationships>")
	idx := bytes.LastIndex(relsXML, closing)
	if idx < 0 {
		return relsXML
	}
	result := make([]byte, 0, len(relsXML)+len(relElement)+1)
	result = append(result, relsXML[:idx]...)
	result = append(result, []byte(relElement)...)
	result = append(result, relsXML[idx:]...)
	return result
}

// ensurePNGContentType adds PNG content type to [Content_Types].xml if not already present.
func ensurePNGContentType(contentTypesXML []byte) []byte {
	if bytes.Contains(contentTypesXML, []byte(`Extension="png"`)) {
		return contentTypesXML
	}
	pngType := `<Default Extension="png" ContentType="image/png"/>`
	closing := []byte("</Types>")
	idx := bytes.LastIndex(contentTypesXML, closing)
	if idx < 0 {
		return contentTypesXML
	}
	result := make([]byte, 0, len(contentTypesXML)+len(pngType)+1)
	result = append(result, contentTypesXML[:idx]...)
	result = append(result, []byte(pngType)...)
	result = append(result, contentTypesXML[idx:]...)
	return result
}

// convertPPTXToPDF sends a PPTX file to Gotenberg for PDF conversion.
func convertPPTXToPDF(pptxData []byte, gotenbergURL string) ([]byte, error) {
	// Write PPTX to temp file for multipart upload
	tmpFile, err := os.CreateTemp("", "cert-*.pptx")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(pptxData); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	return convertFileWithGotenberg(tmpPath, gotenbergURL)
}

// convertFileWithGotenberg sends a file to Gotenberg's LibreOffice endpoint.
func convertFileWithGotenberg(filePath, gotenbergURL string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var body bytes.Buffer
	writer := newMultipartWriter(&body)
	part, err := writer.CreateFormFile("files", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("copy file: %w", err)
	}
	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer: %w", err)
	}

	url := strings.TrimRight(gotenbergURL, "/") + "/forms/libreoffice/convert"

	resp, err := httpPost(url, contentType, &body)
	if err != nil {
		return nil, fmt.Errorf("gotenberg request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gotenberg returned %d: %s", resp.StatusCode, string(errBody))
	}

	pdfData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read pdf response: %w", err)
	}
	return pdfData, nil
}
