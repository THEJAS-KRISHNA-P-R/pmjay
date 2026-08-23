package document

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func sampleCanvas() *canvas {
	c := newCanvas(50, 56, 60)
	c.fillRect(40, hexRGB("#17453d"))
	c.line(0, 14, 20, true, hexRGB("#ffffff"), "Care first message")
	c.advance(20)
	c.divider(hexRGB("#a9d0b8"), 10, 10)
	c.line(0, 11, 16, false, hexRGB("#2b2620"), "Body text with an em—dash and a rupee amount ₹49,300.")
	return c
}

func TestRender_ProducesValidPDFHeaderAndTrailer(t *testing.T) {
	out := sampleCanvas().render()

	if !bytes.HasPrefix(out, []byte("%PDF-1.4\n")) {
		t.Errorf("expected output to start with %%PDF-1.4 header, got: %q", out[:min(40, len(out))])
	}
	trimmed := bytes.TrimRight(out, "\n")
	if !bytes.HasSuffix(trimmed, []byte("%%EOF")) {
		t.Errorf("expected output to end with %%%%EOF, got: %q", trimmed[max(0, len(trimmed)-40):])
	}
}

func TestRender_BalancedObjMarkers(t *testing.T) {
	out := sampleCanvas().render()
	s := string(out)

	// "endobj" itself contains "obj" as a substring, so count "N 0 obj"
	// openers specifically rather than naively counting "obj".
	opens := strings.Count(s, " 0 obj\n")
	closes := strings.Count(s, "endobj\n")
	if opens == 0 {
		t.Fatal("expected at least one object in output")
	}
	if opens != closes {
		t.Errorf("expected balanced obj/endobj: %d opens, %d closes", opens, closes)
	}
}

func TestRender_BalancedStreamMarkers(t *testing.T) {
	out := sampleCanvas().render()
	s := string(out)
	// "\nstream\n" (leading newline required) rather than bare
	// "stream\n" — the latter also matches inside "endstream\n" (which
	// ends in exactly that suffix), silently over-counting opens.
	opens := strings.Count(s, "\nstream\n")
	closes := strings.Count(s, "endstream\n")
	if opens == 0 || opens != closes {
		t.Errorf("expected balanced stream/endstream: %d opens, %d closes", opens, closes)
	}
}

func TestRender_SinglePageDocument_HasExactlyOnePageObject(t *testing.T) {
	out := sampleCanvas().render()
	s := string(out)
	if got := strings.Count(s, "/Type /Page "); got != 1 {
		// Note the trailing space distinguishes "/Type /Page " from
		// "/Type /Pages" (the parent tree object), which also contains
		// "/Type /Page" as a prefix.
		t.Errorf("expected exactly 1 /Type /Page object for single-page content, got %d", got)
	}
	if !strings.Contains(s, "/Count 1") {
		t.Error("expected page tree /Count 1")
	}
}

func TestRender_XrefTableSizeMatchesDeclaredCountAndEntryWidth(t *testing.T) {
	out := sampleCanvas().render()
	s := string(out)

	xrefStart := strings.Index(s, "\nxref\n")
	trailerStart := strings.Index(s, "trailer\n")
	if xrefStart == -1 || trailerStart == -1 || trailerStart < xrefStart {
		t.Fatal("could not locate both xref table and trailer, in order")
	}
	// xrefStart+1 skips the leading '\n' that precedes the "xref" line
	// itself (that separator belongs to the previous object, not to
	// this section).
	lines := strings.Split(s[xrefStart+1:trailerStart], "\n")
	if len(lines) < 2 || lines[0] != "xref" {
		t.Fatalf("malformed xref section, first lines: %v", lines[:min(3, len(lines))])
	}

	declaredCount, err := strconv.Atoi(strings.TrimPrefix(lines[1], "0 "))
	if err != nil {
		t.Fatalf("could not parse declared xref entry count from %q: %v", lines[1], err)
	}

	// Everything from lines[2] onward is entries, except the final
	// element, which is "" from the section's trailing newline before
	// "trailer".
	entryLines := lines[2 : len(lines)-1]

	if len(entryLines) != declaredCount {
		t.Errorf("xref header declares %d entries but %d entry lines follow", declaredCount, len(entryLines))
	}
	for i, line := range entryLines {
		// strings.Split consumes the '\n' delimiter, so each entry's
		// on-disk length is len(line)+1; the PDF spec's 20-byte-per-entry
		// requirement is about that on-disk length, so the in-memory
		// (post-split) string is expected to be 19.
		if len(line) != 19 {
			t.Errorf("xref entry %d is %d bytes on disk, want exactly 20 (PDF spec requirement): %q", i, len(line)+1, line)
		}
	}

	// /Size in the trailer must equal the entry count exactly (Size =
	// highest object number + 1 = total entries including the free head).
	if !strings.Contains(s, "/Size "+strconv.Itoa(declaredCount)+" ") {
		t.Errorf("expected trailer /Size %d matching xref entry count", declaredCount)
	}
}

func TestRender_EveryPageObjectOffsetIsAccurate(t *testing.T) {
	// The single property that actually determines whether a PDF opens:
	// each xref offset must point exactly at the start of "N 0 obj" for
	// that object number. Parse the xref table ourselves (independently
	// of render's own bookkeeping) and check every offset directly
	// against the byte content at that position.
	out := sampleCanvas().render()

	xrefIdx := bytes.Index(out, []byte("\nxref\n"))
	trailerIdx := bytes.Index(out, []byte("trailer\n"))
	if xrefIdx == -1 || trailerIdx == -1 {
		t.Fatal("missing xref or trailer")
	}
	body := string(out[xrefIdx+len("\nxref\n") : trailerIdx])
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatal("xref table too short")
	}
	// First line is "0 N" (start object number and count).
	entries := lines[1:]

	for i, line := range entries {
		if len(line) != 19 { // 20 on disk; Split strips the trailing '\n' — see the sibling width test
			t.Errorf("xref entry %d has length %d, want 19 post-split (20 on disk): %q", i, len(line), line)
			continue
		}
		if i == 0 {
			continue // object 0 is the free-list head, not a real object
		}
		offset, err := strconv.Atoi(line[:10])
		if err != nil {
			t.Errorf("xref entry %d: could not parse offset %q: %v", i, line[:10], err)
			continue
		}
		expectedHeader := strconv.Itoa(i) + " 0 obj"
		if offset+len(expectedHeader) > len(out) || string(out[offset:offset+len(expectedHeader)]) != expectedHeader {
			got := ""
			if offset < len(out) {
				end := offset + 20
				if end > len(out) {
					end = len(out)
				}
				got = string(out[offset:end])
			}
			t.Errorf("xref entry for object %d points to offset %d, expected %q there, found %q", i, offset, expectedHeader, got)
		}
	}
}

func TestRender_MultiPage_WhenContentExceedsOnePage(t *testing.T) {
	c := newCanvas(50, 56, 60)
	// Force enough content to overflow one A4 page.
	for i := 0; i < 80; i++ {
		c.line(0, 11, 16, false, hexRGB("#2b2620"), "A line of body text to fill up the page vertically.")
	}
	out := c.render()
	s := string(out)

	pageCount := strings.Count(s, "/Type /Page ")
	if pageCount < 2 {
		t.Errorf("expected multi-page output for overflowing content, got %d page(s)", pageCount)
	}
	if !strings.Contains(s, "/Count "+strconv.Itoa(pageCount)) {
		t.Errorf("expected /Count %d in page tree", pageCount)
	}
}

func TestRender_RupeeAndEmDashDoNotBreakStructure(t *testing.T) {
	// A regression guard specifically for the two non-ASCII characters
	// this package's real input actually contains (see winansi.go's
	// package doc comment) — confirms they don't produce unbalanced
	// parens/escapes in the content stream's literal strings.
	out := sampleCanvas().render()
	s := string(out)
	if strings.Count(s, "(") < strings.Count(s, ")") {
		// Escaped parens inside Tj strings are "\(" / "\)", which still
		// contain a literal '(' or ')' byte, so a raw imbalance here
		// would indicate a genuinely malformed literal string, not just
		// an escaping artifact.
		t.Error("unbalanced literal-string parentheses in content stream")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
