package document

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Fixed object numbers for the four objects that exist exactly once
// regardless of page count. Page and content-stream objects follow
// sequentially after these with no gaps (see render): page i (0-indexed)
// is object firstPageObj+2*i, and its content stream is the very next
// object. Keeping the numbering gap-free is what lets the xref table
// below be a single unbroken subsection instead of needing PDF's
// (rarely-implemented-correctly) multi-subsection xref format.
const (
	catalogObj   = 1
	pagesObj     = 2
	fontRegObj   = 3
	fontBoldObj  = 4
	firstPageObj = 5
)

// objWriter accumulates the full PDF byte stream while recording the
// exact byte offset each indirect object starts at — the one piece of
// bookkeeping the xref table absolutely cannot get wrong, since a
// wrong offset there is what actually makes a PDF fail to open (as
// opposed to a content-stream mistake, which at worst mis-renders one
// piece of content).
type objWriter struct {
	buf     bytes.Buffer
	offsets map[int]int
}

func newObjWriter() *objWriter {
	return &objWriter{offsets: make(map[int]int)}
}

func (w *objWriter) beginObj(n int) {
	w.offsets[n] = w.buf.Len()
	fmt.Fprintf(&w.buf, "%d 0 obj\n", n)
}

func (w *objWriter) endObj() {
	w.buf.WriteString("endobj\n")
}

// fmtNum formats a coordinate/dimension for embedding directly in a
// PDF content stream or object dictionary — fixed two-decimal
// notation, never scientific notation (which "%g"/strconv's shortest
// form can silently produce for very small or large floats, and which
// is not valid PDF number syntax).
func fmtNum(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// render serializes every page's accumulated drawing primitives into a
// complete, spec-valid PDF 1.4 byte stream: header, every indirect
// object (catalog, page tree, two standard-14 font references, and one
// Page+content-stream pair per page), the cross-reference table, and
// the trailer.
//
// Deliberately uncompressed content streams (no /Filter
// /FlateDecode) — this document is a handful of KB of text at most, so
// compression would save bytes nobody will notice at the cost of every
// test in this package (and any future debugging) needing a zlib
// decoder just to read what the file actually says. Real-world size is
// not the constraint this format choice is optimizing for.
func (c *canvas) render() []byte {
	n := len(c.pages)
	w := newObjWriter()

	w.buf.WriteString("%PDF-1.4\n%\xE2\xE3\xCF\xD3\n")

	w.beginObj(catalogObj)
	fmt.Fprintf(&w.buf, "<< /Type /Catalog /Pages %d 0 R >>\n", pagesObj)
	w.endObj()

	kids := make([]string, n)
	for i := 0; i < n; i++ {
		kids[i] = fmt.Sprintf("%d 0 R", firstPageObj+2*i)
	}
	w.beginObj(pagesObj)
	fmt.Fprintf(&w.buf, "<< /Type /Pages /Kids [%s] /Count %d >>\n", strings.Join(kids, " "), n)
	w.endObj()

	w.beginObj(fontRegObj)
	w.buf.WriteString("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>\n")
	w.endObj()

	w.beginObj(fontBoldObj)
	w.buf.WriteString("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>\n")
	w.endObj()

	for i, pg := range c.pages {
		pageObj := firstPageObj + 2*i
		contentObj := pageObj + 1
		content := renderContentStream(pg)

		w.beginObj(pageObj)
		fmt.Fprintf(&w.buf, "<< /Type /Page /Parent %d 0 R /MediaBox [0 0 %s %s] /Resources << /Font << /F1 %d 0 R /F2 %d 0 R >> >> /Contents %d 0 R >>\n",
			pagesObj, fmtNum(pageWidth), fmtNum(pageHeight), fontRegObj, fontBoldObj, contentObj)
		w.endObj()

		w.beginObj(contentObj)
		fmt.Fprintf(&w.buf, "<< /Length %d >>\nstream\n", len(content))
		w.buf.Write(content)
		w.buf.WriteString("endstream\n")
		w.endObj()
	}

	highestObj := fontBoldObj + 2*n // == firstPageObj + 2*n - 1, i.e. 4 + 2n
	xrefOffset := w.buf.Len()
	fmt.Fprintf(&w.buf, "xref\n0 %d\n", highestObj+1)
	w.buf.WriteString("0000000000 65535 f \n")
	for objNum := 1; objNum <= highestObj; objNum++ {
		offset, ok := w.offsets[objNum]
		if !ok {
			// Would only happen if the numbering scheme above had a gap
			// — defensive, not expected to ever fire (see the constant
			// block's comment), but an unresolvable reference is worse
			// than a loud failure at generation time.
			panic(fmt.Sprintf("document: internal numbering gap — no offset recorded for object %d", objNum))
		}
		fmt.Fprintf(&w.buf, "%010d 00000 n \n", offset)
	}

	fmt.Fprintf(&w.buf, "trailer\n<< /Size %d /Root %d 0 R >>\nstartxref\n%d\n%%%%EOF", highestObj+1, catalogObj, xrefOffset)

	return w.buf.Bytes()
}

// renderContentStream turns one page's accumulated primitives into raw
// PDF content-stream operators. Paint order is rects, then rules, then
// text — each primitive is independently wrapped in q/Q (save/restore
// graphics state) specifically so no drawing operation's color or line
// width can ever leak into the next one; that redundancy costs a
// little file size and buys not having to reason about shared mutable
// graphics state across dozens of independently-generated primitives.
func renderContentStream(pg *page) []byte {
	var buf bytes.Buffer

	for _, r := range pg.rects {
		fmt.Fprintf(&buf, "q\n%s %s %s rg\n%s %s %s %s re\nf\nQ\n",
			fmtNum(r.color.r), fmtNum(r.color.g), fmtNum(r.color.b),
			fmtNum(r.x), fmtNum(r.y), fmtNum(r.w), fmtNum(r.h))
	}

	for _, l := range pg.lines {
		fmt.Fprintf(&buf, "q\n%s %s %s RG\n%s w\n%s %s m\n%s %s l\nS\nQ\n",
			fmtNum(l.color.r), fmtNum(l.color.g), fmtNum(l.color.b),
			fmtNum(l.widthPts),
			fmtNum(l.x1), fmtNum(l.y), fmtNum(l.x2), fmtNum(l.y))
	}

	for _, t := range pg.text {
		shaped := shapeText(t.text, t.size, t.bold)
		font := "/F1"
		if t.bold {
			font = "/F2"
		}
		fmt.Fprintf(&buf, "q\n%s %s %s rg\nBT\n%s %s Tf\n%s %s Td\n(",
			fmtNum(t.color.r), fmtNum(t.color.g), fmtNum(t.color.b),
			font, fmtNum(t.size), fmtNum(t.x), fmtNum(t.y))
		buf.Write(shaped.pdfBytes)
		buf.WriteString(") Tj\nET\nQ\n")
	}

	return buf.Bytes()
}
