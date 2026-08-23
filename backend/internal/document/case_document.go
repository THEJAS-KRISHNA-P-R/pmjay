package document

import (
	"fmt"
	"strings"

	"github.com/pmjay-advocate/backend/internal/store"
)

// Color palette. Every hex value here is copied verbatim from
// frontend/app/globals.css's design tokens (hexRGB takes the same
// "#rrggbb" literal that file uses) specifically so this palette can
// be diffed against that file directly rather than trusted to have
// been transcribed correctly by eye. colorGray has no direct
// equivalent there — the web achieves muted meta/footer text with
// Tailwind opacity utilities (e.g. "text-sand-900/60"), a concept PDF
// fill color doesn't have without introducing an ExtGState alpha
// dictionary; a fixed warm gray in the same family reads the same way
// without that additional format complexity.
var (
	colorTeal800  = hexRGB("#17453d")
	colorWhite    = hexRGB("#ffffff")
	colorSand900  = hexRGB("#2b2620")
	colorSand100  = hexRGB("#f4efe3")
	colorHairline = hexRGB("#e8dfc9")
	colorGray     = rgb{0.45, 0.42, 0.38}

	colorGreenBg     = hexRGB("#e6f2ea")
	colorGreenBorder = hexRGB("#a9d0b8")
	colorGreenText   = hexRGB("#245c3e")
	colorGreenStrong = hexRGB("#2f7d5a")

	colorAmberBg     = hexRGB("#faf0dd")
	colorAmberBorder = hexRGB("#e6c483")
	colorAmberText   = hexRGB("#7a4f0f")
	colorAmberStrong = hexRGB("#b06e1a")

	colorRedBg     = hexRGB("#f6e9e7")
	colorRedBorder = hexRGB("#d9aca5")
	colorRedText   = hexRGB("#7a3129")
	colorRedStrong = hexRGB("#a8433a")

	colorHandoffBg     = hexRGB("#e3edec")
	colorHandoffBorder = hexRGB("#a9c4c1")
	colorHandoffText   = hexRGB("#1d564c")
	colorHandoffStrong = hexRGB("#24685d")
)

// tierStyle mirrors frontend/app/components/TierBadge.tsx's
// TIER_STYLES map: one entry per outcome, giving the badge glyph,
// plain-language label, and the four-color family (background,
// border/rule, body text accent, and a "strong" accent for the badge
// itself) used throughout that outcome's panel.
type tierStyle struct {
	badge                    string
	label                    string
	bg, border, text, strong rgb
}

// tierStyles' badge glyphs are deliberately not the web's icon set
// (✓ ? i ½ →) verbatim — ✓ and → have no glyph in the standard 14
// PDF fonts under any encoding (see winansi.go's package doc comment
// on why this package doesn't embed a font to work around that), so
// "OK" and "->" stand in for them. "?" and "i" are already plain
// ASCII. "½" is the one icon that happens to have a real WinAnsi byte
// (0xBD) and is used as-is. Every badge is paired with the same
// plain-language label the web UI uses specifically so meaning never
// depends on the glyph rendering correctly — matching TierBadge.tsx's
// own accessibility principle of never relying on color or a single
// symbol alone.
var tierStyles = map[string]tierStyle{
	"green": {badge: "OK", label: "This looks covered",
		bg: colorGreenBg, border: colorGreenBorder, text: colorGreenText, strong: colorGreenStrong},
	"amber": {badge: "?", label: "Needs one more check",
		bg: colorAmberBg, border: colorAmberBorder, text: colorAmberText, strong: colorAmberStrong},
	"red": {badge: "i", label: "This is correctly not covered",
		bg: colorRedBg, border: colorRedBorder, text: colorRedText, strong: colorRedStrong},
	// Mixed uses the amber color family, not a distinct one — this
	// matches TIER_STYLES in TierBadge.tsx exactly (mixed's textClass
	// and strongClass are amber's; only its background is a distinct
	// green-to-amber gradient on the web, and this package draws flat
	// fills only — see pdf.go's filledRect doc comment).
	"mixed": {badge: "\u00BD", label: "Part covered, part not",
		bg: colorAmberBg, border: colorAmberBorder, text: colorAmberText, strong: colorAmberStrong},
	"handoff": {badge: "->", label: "Let's get you a person",
		bg: colorHandoffBg, border: colorHandoffBorder, text: colorHandoffText, strong: colorHandoffStrong},
}

// dateFormat matches the family-facing date style used elsewhere in
// this project's Go->text formatting (day, full month name, year) —
// unambiguous regardless of the reader's preferred date-order
// convention, which matters more here than usual since this document
// may be read by hospital staff, a PLV, or a CGRMS reviewer as well as
// the family themselves.
const dateFormat = "2 January 2006, 3:04 PM"

// BuildCase renders c as a complete, print-ready PDF: the care-first
// message, the outcome explanation, handoff routing if applicable,
// action steps, hospital script, draft complaint, and an evidence log
// — the same content already shown on the case result page
// (frontend/app/case/[id]/page.tsx) — with case ID and both helpline
// numbers repeated in a footer on every page. See this package's
// README.md for the full design reasoning.
//
// The error return exists for interface consistency with the rest of
// this codebase's handler-facing functions (see internal/api/handlers.go);
// nothing in this function's actual control flow can currently fail,
// since it only ever reads already-validated, already-serialized
// CaseRecord fields and does no I/O.
func BuildCase(c store.CaseRecord) ([]byte, error) {
	canvas := newCanvas(50, 60, 78)

	drawDocHeader(canvas, c)
	canvas.advance(6)

	canvas.box(colorTeal800, 18, 14, 14, bodyLeading+1.5, bodyGap, 0,
		textBlock(c.CareFirstMessage, 11.5, true, colorWhite, canvas.contentWidth()-36))
	canvas.advance(8)

	if strings.TrimSpace(c.Disclaimer) != "" {
		canvas.smallNote(c.Disclaimer, colorGray)
		canvas.advance(8)
	}

	style, known := tierStyles[c.Outcome]
	if !known {
		// Mirrors internal/response/builder.go's own stated philosophy
		// (see its default-case comment): an outcome value this package
		// doesn't recognize is a bug elsewhere, and the safe direction
		// to fail is toward more caution and a person, not toward
		// silently dropping the section. Handoff's styling (not red's)
		// is used for that same reason.
		style = tierStyle{badge: "?", label: "Outcome: " + c.Outcome,
			bg: colorHandoffBg, border: colorHandoffBorder, text: colorHandoffText, strong: colorHandoffStrong}
	}
	drawTierPanel(canvas, style, c)
	canvas.advance(16)

	if c.Outcome == "handoff" && strings.TrimSpace(c.HandoffSummary) != "" {
		drawHandoffPanel(canvas, c.HandoffSummary)
		canvas.advance(16)
	}

	if len(c.ActionSteps) > 0 {
		canvas.heading("What to do right now", colorTeal800)
		canvas.numberedList(c.ActionSteps, colorSand900)
		canvas.advance(10)
	}

	if strings.TrimSpace(c.HospitalScript) != "" {
		drawBoxedText(canvas, "Exact words to use at the desk", "", c.HospitalScript)
		canvas.advance(10)
	}

	if strings.TrimSpace(c.ComplaintText) != "" {
		drawBoxedText(canvas, "Draft complaint, ready to review",
			"Submit this yourself through the official Ayushman App (or your chosen CGRMS channel) when you're ready — this document does not submit anything on its own.",
			c.ComplaintText)
		canvas.advance(10)
	}

	if len(c.Evidence) > 0 {
		drawEvidenceLog(canvas, c.Evidence)
	}

	drawFooter(canvas, c)

	return canvas.render(), nil
}

func drawDocHeader(c *canvas, rec store.CaseRecord) {
	c.line(0, 18, 23, true, colorTeal800, "PMJAY Point-of-Denial Advocate")
	c.line(0, 10, 14, false, colorGray, "Case summary, prepared automatically. Not an official government document.")
	meta := fmt.Sprintf("Case ID: %s   |   Prepared: %s", rec.ID, formatWhenOrFallback(rec))
	c.line(0, 9, 15, false, colorGray, meta)
	c.divider(colorHairline, 6, 0)
}

// formatWhenOrFallback guards against an unset CreatedAt (a bare
// store.CaseRecord{} in a test, for instance) rendering as the
// misleading "1 January 0001" zero-value date rather than something
// that honestly signals "not set".
func formatWhenOrFallback(rec store.CaseRecord) string {
	if rec.CreatedAt.IsZero() {
		return "(not recorded)"
	}
	return rec.CreatedAt.Format(dateFormat)
}

func drawTierPanel(c *canvas, style tierStyle, rec store.CaseRecord) {
	innerWidth := c.contentWidth() - 36

	var lines []styledLine
	lines = append(lines, styledLine{
		text: style.badge + "   " + style.label, size: headingSize, bold: true, color: style.strong,
	})
	lines = append(lines, textBlock(rec.TierMessage, bodySize, false, colorSand900, innerWidth)...)
	if strings.TrimSpace(rec.Citation) != "" {
		lines = append(lines, styledLine{rule: true, color: style.border})
		lines = append(lines, textBlock("Based on: "+rec.Citation, smallSize, false, style.text, innerWidth)...)
	}
	c.box(style.bg, 18, 14, 14, bodyLeading, bodyGap, 0, lines)
}

func drawHandoffPanel(c *canvas, summary string) {
	innerWidth := c.contentWidth() - 36
	strong := tierStyles["handoff"].strong

	var lines []styledLine
	lines = append(lines, styledLine{text: "Free legal help, right now", size: headingSize, bold: true, color: strong})
	lines = append(lines, textBlock(
		"NALSA (the National Legal Services Authority) gives free legal help to families who can't afford a lawyer, including for exactly this kind of situation. A Para Legal Volunteer can help in person or by phone, at no cost.",
		bodySize, false, colorSand900, innerWidth)...)
	lines = append(lines, styledLine{text: ""})
	lines = append(lines, styledLine{text: "Call 15100 — free, toll-free", size: 12.5, bold: true, color: strong})
	lines = append(lines, styledLine{text: ""})
	lines = append(lines, styledLine{text: "What we'll have ready to share, so nothing has to be re-explained:",
		size: smallSize + 0.5, bold: true, color: colorHandoffText})
	lines = append(lines, textBlock(summary, bodySize, false, colorSand900, innerWidth)...)

	c.box(colorHandoffBg, 18, 14, 14, bodyLeading, bodyGap, 0, lines)
}

// drawBoxedText renders a titled section whose main content sits in a
// light, bordered-feeling panel — the print equivalent of
// frontend/app/components/CopyableTextBox.tsx's monospace-ish boxed
// text (hospital script, draft complaint). helper, if non-empty, is a
// small explanatory line drawn above the box, outside it.
func drawBoxedText(c *canvas, title, helper, text string) {
	c.heading(title, colorTeal800)
	if strings.TrimSpace(helper) != "" {
		c.paragraph(helper, colorGray)
		c.advance(6)
	}
	innerWidth := c.contentWidth() - 28
	lines := textBlock(text, bodySize, false, colorSand900, innerWidth)
	c.box(colorSand100, 14, 12, 12, bodyLeading, bodyGap, 0, lines)
}

func drawEvidenceLog(c *canvas, evidence []store.EvidenceEntry) {
	c.heading("Evidence recorded", colorTeal800)
	for i, e := range evidence {
		header := fmt.Sprintf("%d.  Recorded %s", i+1, formatEvidenceTime(e))
		c.line(0, bodySize, bodyLeading, true, colorSand900, header)

		var parts []string
		if strings.TrimSpace(e.StaffName) != "" {
			parts = append(parts, "Staff: "+e.StaffName)
		}
		if strings.TrimSpace(e.ApproxTime) != "" {
			parts = append(parts, "Approx. time: "+e.ApproxTime)
		}
		if strings.TrimSpace(e.Note) != "" {
			parts = append(parts, "Note: "+e.Note)
		}
		if len(parts) > 0 {
			c.paragraph(strings.Join(parts, "   \u2022   "), colorSand900)
		} else {
			c.paragraph("(no further details recorded)", colorGray)
		}
		c.advance(6)
	}
}

func formatEvidenceTime(e store.EvidenceEntry) string {
	if e.CapturedAt.IsZero() {
		return "(time not recorded)"
	}
	return e.CapturedAt.Format(dateFormat)
}

// footerY is the fixed baseline every page's footer text sits on,
// measured from the page bottom — independent of the cursor-based
// canvas.line/divider helpers, which is deliberate: the footer belongs
// at the same physical spot on every page regardless of how much body
// content that page holds, whereas the cursor's position when
// drawFooter runs is wherever the last content section happened to
// stop.
const footerY = 30.0

// drawFooter adds a two-line footer — case ID and page number on one
// line, both helpline numbers on the next — to every page, once all
// pages are known (hence "Page X of Y" is only possible as a pass
// after the rest of the document is fully laid out, not while it's
// being written). Two lines rather than one is a legibility and
// headroom choice, not a strict width necessity: the combined single
// line measures comfortably under the content width today (see
// case_document_test.go's TestFooterFitsPageWidth, which measures this
// directly), but at footer-sized type, cramming case ID + both
// helpline numbers + page count onto one ~8.5pt line the full width of
// the page reads as dense rather than calm, and leaves little margin
// before a future wording tweak (a longer helpline label, say) would
// silently start overflowing. Two lines costs nothing and removes that
// whole failure mode.
func drawFooter(c *canvas, rec store.CaseRecord) {
	total := len(c.pages)
	for i, pg := range c.pages {
		pg.lines = append(pg.lines, hLine{
			x1: c.marginX, x2: pageWidth - c.marginX, y: footerY + 26, color: colorHairline, widthPts: 0.75,
		})

		left := fmt.Sprintf("Case %s", rec.ID)
		right := fmt.Sprintf("Page %d of %d", i+1, total)
		pg.text = append(pg.text, textRun{x: c.marginX, y: footerY + 12, size: smallSize, color: colorGray, text: left})
		rightW := textWidth(right, smallSize, false)
		pg.text = append(pg.text, textRun{x: pageWidth - c.marginX - rightW, y: footerY + 12, size: smallSize, color: colorGray, text: right})

		helplines := "PMJAY helpline: 14555   |   NALSA free legal aid: 15100"
		pg.text = append(pg.text, textRun{x: c.marginX, y: footerY, size: smallSize, color: colorGray, text: helplines})
	}
}
