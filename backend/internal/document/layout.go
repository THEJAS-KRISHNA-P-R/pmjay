package document

import (
	"strconv"
	"strings"
)

// wrapWords greedily packs words from s onto lines no wider than
// maxWidth at the given font/size, using this package's real Adobe
// Core 14 metrics (see winansi.go) rather than a fixed
// characters-per-line guess — the guessing approach is what produces
// the ragged, sometimes-overflowing wrapping common to quick PDF
// scripts. A single word wider than maxWidth on its own (not expected
// for this document's actual vocabulary, but not impossible for a
// long package name) is placed on its own line rather than dropped or
// truncated — see TestWrapWords_SingleWordWiderThanMaxWidth_KeptWholeNotDropped
// in layout_test.go.
func wrapWords(s string, size float64, bold bool, maxWidth float64) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	current := words[0]
	for _, w := range words[1:] {
		candidate := current + " " + w
		if textWidth(candidate, size, bold) <= maxWidth {
			current = candidate
			continue
		}
		lines = append(lines, current)
		current = w
	}
	lines = append(lines, current)
	return lines
}

// wrapParagraphs splits s on literal newlines first — internal/response's
// templates use "\n\n" for a paragraph break and a single "\n" for a
// tight line break within the same paragraph (e.g. the two bulleted
// lines in templates.go's pending-preauth message) — then word-wraps
// each non-empty segment independently. An empty segment (from a "\n\n"
// double break) becomes a "" entry in the result, which draw
// interprets as a blank-paragraph gap rather than a zero-width line of
// text; a single "\n" produces no such entry, so consecutive lines
// stay tightly spaced exactly where the source text put them next to
// each other with no blank line between.
func wrapParagraphs(s string, size float64, bold bool, maxWidth float64) []string {
	if strings.TrimSpace(s) == "" {
		// strings.Split("", "\n") returns [""], not an empty slice —
		// without this guard, a genuinely empty input would produce one
		// spurious blank-paragraph gap marker rather than "nothing to
		// draw". Every real call site in case_document.go already
		// checks for an empty field before calling into this (directly
		// or via textBlock), so this never currently changes actual
		// document output — it's here so this function is correct on
		// its own terms for whoever calls it next, not just for today's
		// call sites. See TestWrapParagraphs_EmptyOrWhitespaceInput.
		return nil
	}
	var out []string
	for _, segment := range strings.Split(s, "\n") {
		if strings.TrimSpace(segment) == "" {
			out = append(out, "")
			continue
		}
		out = append(out, wrapWords(segment, size, bold, maxWidth)...)
	}
	return out
}

// styledLine is one drawable unit inside a box or a free-flowing
// section: either a real line of already-wrapped text, or a blank-line
// gap (text == "" && !rule), or a thin divider rule.
type styledLine struct {
	text  string
	size  float64
	bold  bool
	color rgb
	rule  bool
}

// textBlock wraps s and returns one styledLine per resulting line,
// tagging every line with the given style — the single function every
// content section in case_document.go goes through, so a paragraph's
// wrapped lines and its visual style can never be assembled
// inconsistently between two call sites.
func textBlock(s string, size float64, bold bool, color rgb, maxWidth float64) []styledLine {
	var out []styledLine
	for _, l := range wrapParagraphs(s, size, bold, maxWidth) {
		out = append(out, styledLine{text: l, size: size, bold: bold, color: color})
	}
	return out
}

// blockHeight sums the vertical space lines will occupy when drawn
// with the given normal line leading and gap height for blank lines —
// used up front, before anything is drawn, so a box can reserve
// exactly the right amount of space with a single ensureSpace call
// rather than drawing speculatively (see this package's first-draft
// bug, fixed in pdf.go's fillRect doc comment).
func blockHeight(lines []styledLine, leading, gapHeight float64) float64 {
	var h float64
	for _, l := range lines {
		switch {
		case l.rule:
			h += leading // a rule takes one line-height's worth of room (small margin above/below is folded into leading here for simplicity)
		case l.text == "":
			h += gapHeight
		default:
			h += leading
		}
	}
	return h
}

// drawBlock draws lines starting at the canvas's current cursor,
// advancing past each one exactly as blockHeight predicted for it —
// the two functions are kept side by side in this file specifically so
// a change to one's per-line accounting is hard to make without
// noticing the other needs the same change.
func (c *canvas) drawBlock(lines []styledLine, leading, gapHeight, indent float64) {
	for _, l := range lines {
		switch {
		case l.rule:
			c.divider(l.color, 0, 0)
			c.advance(leading) // divider() itself moves the cursor by 0; account for the leading blockHeight reserved for it
		case l.text == "":
			c.advance(gapHeight)
		default:
			c.line(indent, l.size, leading, l.bold, l.color, l.text)
		}
	}
}

// Standard vertical rhythm used throughout the document. Kept as named
// constants rather than repeated magic numbers so every section's
// spacing stays visually consistent without each call site having to
// remember the same numbers independently.
const (
	bodySize     = 10.5
	bodyLeading  = 15.0
	bodyGap      = 7.0
	headingSize  = 13.0
	headingLead  = 18.0
	smallSize    = 8.5
	smallLeading = 12.0
)

// box draws a padded, filled-background block containing lines,
// guaranteed to render on a single page (space is reserved with
// ensureSpace before the fill rect is drawn, so a page break can never
// land in the middle of a box — the exact failure mode this package's
// first draft had). Returns the total height consumed, including
// padding, in case a caller wants it (none currently do, but it costs
// nothing to expose and avoids a future caller needing to
// recompute it).
func (c *canvas) box(bg rgb, padX, padTop, padBottom, leading, gapHeight, indent float64, lines []styledLine) float64 {
	contentH := blockHeight(lines, leading, gapHeight)
	total := padTop + padBottom + contentH
	c.ensureSpace(total)
	c.fillRect(total, bg)
	c.advance(padTop)
	c.drawBlock(lines, leading, gapHeight, indent+padX)
	c.advance(padBottom)
	return total
}

// heading draws a section heading in the document's standard heading
// style (bold, teal, with breathing room above and below) — the
// print-document equivalent of the bold "text-teal-800" section
// headers used throughout frontend/app/components (ActionSteps,
// CopyableTextBox, etc.), so a reader who has already seen the web
// result page recognizes the same visual hierarchy in the PDF.
func (c *canvas) heading(text string, color rgb) {
	c.advance(6)
	c.line(0, headingSize, headingLead, true, color, text)
	c.advance(2)
}

// paragraph draws body text at the document's standard body style,
// wrapped to the full content width.
func (c *canvas) paragraph(text string, color rgb) {
	lines := textBlock(text, bodySize, false, color, c.contentWidth())
	c.drawBlock(lines, bodyLeading, bodyGap, 0)
}

// smallNote draws text at footer-scale size — for content that must be
// present and legible but is deliberately de-emphasized relative to the
// document's main guidance (e.g. the not-legal-advice disclaimer right
// under the loud care-first banner: both must be read, but they are not
// the same kind of urgent).
func (c *canvas) smallNote(text string, color rgb) {
	const smallLeading = 12.0
	lines := textBlock(text, smallSize, false, color, c.contentWidth())
	c.drawBlock(lines, smallLeading, bodyGap, 0)
}

// numberedList draws each item with a hanging "N." prefix and wraps
// each item's own text to account for that prefix's width, matching
// frontend/app/components/ActionSteps.tsx's numbered layout (a plain
// numeral prefix here rather than that component's circular badge — see
// this package's README for why that's a deliberate simplification,
// not an oversight).
func (c *canvas) numberedList(items []string, color rgb) {
	const prefixWidth = 22.0
	for i, item := range items {
		prefix := strconv.Itoa(i+1) + "."
		lines := wrapParagraphs(item, bodySize, false, c.contentWidth()-prefixWidth)
		prefixDrawn := false
		for _, l := range lines {
			if l == "" {
				c.advance(bodyGap)
				continue
			}
			if !prefixDrawn {
				// The "N." prefix sits in the indent gutter to the left
				// of the item's own wrapped text, drawn once per item —
				// on its first *non-blank* wrapped line specifically
				// (not simply lines[0]): an item whose text happens to
				// start with a literal newline would otherwise wrap to
				// a blank marker at index 0, and a plain "index == 0"
				// check would skip the prefix entirely for that item —
				// a real bug this package's own tests caught (see
				// layout_test.go's TestNumberedList_PrefixSurvivesLeadingBlankLine).
				c.ensureSpace(bodyLeading)
				p := c.currentPage()
				p.text = append(p.text, textRun{
					x: c.marginX, y: c.y - bodySize*0.82,
					bold: true, size: bodySize, color: color, text: prefix,
				})
				prefixDrawn = true
			}
			c.line(prefixWidth, bodySize, bodyLeading, false, color, l)
		}
	}
}
