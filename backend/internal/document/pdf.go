package document

import (
	"fmt"
)

// A4 page geometry, in points (1/72 inch) — the standard PDF unit.
// India uses A4, not US Letter, for virtually all official and
// hospital paperwork, so this is not a default left un-considered; a
// document sized for US Letter would print with different margins on
// every printer this tool's actual users are likely to encounter.
const (
	pageWidth  = 595.0
	pageHeight = 842.0
)

// rgb is a color in the 0-1 range PDF's "rg"/"RG" operators expect.
type rgb struct{ r, g, b float64 }

// hexRGB converts a "#rrggbb" literal (as used throughout
// frontend/app/globals.css's design tokens) into an rgb — kept as a
// small helper specifically so the color constants in case_document.go
// can be copy-pasted directly from that CSS file instead of
// hand-converted, which is where a transcription mistake would
// otherwise be easy to make and easy to miss.
func hexRGB(hex string) rgb {
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return rgb{float64(r) / 255.0, float64(g) / 255.0, float64(b) / 255.0}
}

// textRun is one line of already-shaped text at a fixed position.
// layout.go is responsible for wrapping and positioning; by the time a
// textRun exists, no further decisions remain — canvas just draws it.
type textRun struct {
	x, y  float64
	bold  bool
	size  float64
	color rgb
	text  string // raw Go string; shaped (encoded+escaped) at draw time
}

// filledRect is one solid-color rectangle, used for the care-first
// banner, tier panels, and boxed text — mirroring the bordered/filled
// cards throughout frontend/app/components (CareFirstBanner,
// TierPanel, CopyableTextBox). PDF has no rounded-rectangle primitive
// without hand-rolling Bezier curves, so these are sharp-cornered by
// design — a plain rectangle reads as entirely normal in a print
// document (rounded cards are a screen-UI convention, not a print
// one), so this is not a fidelity compromise worth spending
// complexity on.
type filledRect struct {
	x, y, w, h float64
	color      rgb
}

// hLine is a thin horizontal rule, used as a section divider and above
// the per-page footer.
type hLine struct {
	x1, x2, y float64
	color     rgb
	widthPts  float64
}

// page accumulates every drawing primitive for one page, in the order
// they should be painted (rects before the text that sits on top of
// them — canvas's callers are responsible for that ordering; see
// layout.go).
type page struct {
	rects []filledRect
	lines []hLine
	text  []textRun
}

// canvas lays out content across as many A4 pages as needed and, once
// finished, serializes everything into real PDF bytes. Every method
// works in plain page-relative point coordinates with the origin at
// the bottom-left (PDF's native coordinate system) — nothing outside
// this file needs to know anything about PDF's binary object format.
type canvas struct {
	marginX, marginTop, marginBottom float64
	pages                            []*page
	y                                float64 // current write cursor on the last page, in points from the bottom
}

func newCanvas(marginX, marginTop, marginBottom float64) *canvas {
	c := &canvas{marginX: marginX, marginTop: marginTop, marginBottom: marginBottom}
	c.pages = append(c.pages, &page{})
	c.y = pageHeight - marginTop
	return c
}

// contentWidth is the usable horizontal space between the left and
// right margins (margins are symmetric).
func (c *canvas) contentWidth() float64 {
	return pageWidth - 2*c.marginX
}

func (c *canvas) currentPage() *page {
	return c.pages[len(c.pages)-1]
}

// newPage starts a fresh page and resets the write cursor to the top
// margin.
func (c *canvas) newPage() {
	c.pages = append(c.pages, &page{})
	c.y = pageHeight - c.marginTop
}

// ensureSpace starts a new page if fewer than h points remain above
// the bottom margin — called before drawing any block whose height is
// known ahead of time (a heading, a box, a line of text), so a block
// never gets visually split across a page boundary partway through.
func (c *canvas) ensureSpace(h float64) {
	if c.y-h < c.marginBottom {
		c.newPage()
	}
}

// advance moves the write cursor down by d points without drawing
// anything — used for inter-section spacing.
func (c *canvas) advance(d float64) {
	c.y -= d
}

// fillRect draws a filled rectangle, within the content column (from
// left margin to right margin), whose top-left corner is the canvas's
// CURRENT position and whose height is h. Deliberately the only
// rectangle primitive canvas exposes, and deliberately does not move
// the cursor itself — see layout.go's box, which is the one call site
// that should ever use this directly. box always calls ensureSpace(h)
// immediately beforehand and advance(h) immediately afterward, so the
// rect's footprint and the cursor's movement can never drift apart the
// way they did in this package's very first draft (a real bug, caught
// by actually rendering a sample and looking at it — see
// case_document_test.go's rendering-fidelity tests, which exist
// specifically so a regression like that fails a test instead of
// requiring another screenshot to notice).
func (c *canvas) fillRect(h float64, color rgb) {
	p := c.currentPage()
	p.rects = append(p.rects, filledRect{
		x: c.marginX, y: c.y - h, w: c.contentWidth(), h: h, color: color,
	})
}

// divider draws a thin horizontal rule at the current cursor and
// advances past it.
func (c *canvas) divider(color rgb, gapBefore, gapAfter float64) {
	c.ensureSpace(gapBefore + 1 + gapAfter)
	c.advance(gapBefore)
	p := c.currentPage()
	p.lines = append(p.lines, hLine{x1: c.marginX, x2: pageWidth - c.marginX, y: c.y, color: color, widthPts: 0.75})
	c.advance(gapAfter)
}

// line draws exactly one line of text at the given left offset from
// the margin (0 for flush-left) and font, then advances the cursor by
// leading. It does not wrap or measure — see drawParagraph for that;
// this is the primitive drawParagraph and the list/heading helpers in
// layout.go build on.
func (c *canvas) line(indent, size, leading float64, bold bool, color rgb, text string) {
	c.ensureSpace(leading)
	p := c.currentPage()
	p.text = append(p.text, textRun{
		x: c.marginX + indent, y: c.y - size*0.82, // baseline sits slightly below the cursor's top-of-line position
		bold: bold, size: size, color: color, text: text,
	})
	c.advance(leading)
}
