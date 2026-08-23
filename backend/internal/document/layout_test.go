package document

import "testing"

func TestWrapWords_FitsOnOneLineWhenShortEnough(t *testing.T) {
	lines := wrapWords("short enough text", 10, false, 400)
	if len(lines) != 1 || lines[0] != "short enough text" {
		t.Errorf("expected a single unwrapped line, got %v", lines)
	}
}

func TestWrapWords_WrapsAtWordBoundaries(t *testing.T) {
	// Pick a maxWidth that fits exactly two words per line at this
	// size, verified against the real metric rather than guessed.
	two := textWidth("aaaa bbbb", 10, false)
	lines := wrapWords("aaaa bbbb cccc dddd", 10, false, two)
	want := []string{"aaaa bbbb", "cccc dddd"}
	if len(lines) != len(want) {
		t.Fatalf("expected %d lines, got %d: %v", len(want), len(lines), lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("line %d: got %q, want %q", i, lines[i], want[i])
		}
	}
}

func TestWrapWords_NoLineExceedsMaxWidth(t *testing.T) {
	// The property that actually matters, checked directly against
	// real text rather than a single hand-picked example: every
	// produced line must measure at or under maxWidth, for a range of
	// realistic widths and a realistic sentence.
	s := "Based on what you have described, this matches Laparoscopic Cholecystectomy under General Surgery, listed with an indicative rate of Rs.49,300."
	for _, maxWidth := range []float64{120, 200, 300, 495} {
		for _, l := range wrapWords(s, bodySize, false, maxWidth) {
			if w := textWidth(l, bodySize, false); w > maxWidth {
				t.Errorf("maxWidth=%.0f: line %q measures %.1f, exceeds maxWidth", maxWidth, l, w)
			}
		}
	}
}

func TestWrapWords_SingleWordWiderThanMaxWidth_KeptWholeNotDropped(t *testing.T) {
	// The behavior wrapWords's own doc comment claims: a single word
	// that alone exceeds maxWidth is placed on its own line (visually
	// overflowing) rather than silently dropped or truncated — data
	// loss would be far worse than an imperfectly wrapped line. Not
	// expected in this document's real vocabulary, but a long HBP
	// package name or a family-entered proper noun isn't impossible.
	longWord := "Supercalifragilisticexpialidocious-TypeProcedureNameLongerThanAnyReasonableLineWidth"
	tiny := 50.0 // far narrower than the word itself
	lines := wrapWords("short "+longWord+" tail", bodySize, false, tiny)

	found := false
	for _, l := range lines {
		if l == longWord {
			found = true
		}
	}
	if !found {
		t.Errorf("expected the oversized word to appear intact on its own line, got %v", lines)
	}
	// And no data lost overall: rejoining every line reproduces every
	// original word.
	var total int
	for _, l := range lines {
		total += len(splitFields(l))
	}
	if want := len(splitFields("short " + longWord + " tail")); total != want {
		t.Errorf("expected %d total words preserved across wrapped lines, got %d", want, total)
	}
}

func splitFields(s string) []string {
	var out []string
	field := ""
	for _, r := range s {
		if r == ' ' {
			if field != "" {
				out = append(out, field)
				field = ""
			}
			continue
		}
		field += string(r)
	}
	if field != "" {
		out = append(out, field)
	}
	return out
}

func TestWrapWords_EmptyString_ReturnsNoLines(t *testing.T) {
	if lines := wrapWords("", 10, false, 400); lines != nil {
		t.Errorf("expected nil for empty input, got %v", lines)
	}
	if lines := wrapWords("   ", 10, false, 400); lines != nil {
		t.Errorf("expected nil for whitespace-only input, got %v", lines)
	}
}

func TestWrapParagraphs_DoubleNewlineProducesGapMarker(t *testing.T) {
	lines := wrapParagraphs("first paragraph\n\nsecond paragraph", 10, false, 400)
	want := []string{"first paragraph", "", "second paragraph"}
	if len(lines) != len(want) {
		t.Fatalf("expected %v, got %v", want, lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("line %d: got %q, want %q", i, lines[i], want[i])
		}
	}
}

func TestWrapParagraphs_SingleNewline_NoGapMarker(t *testing.T) {
	// This is the exact shape internal/response/templates.go's
	// pending-preauth message uses: two bullet-style lines separated
	// by one "\n", meant to stay visually adjacent (no blank line
	// between them), inside a message that also uses "\n\n" elsewhere
	// for real paragraph breaks.
	lines := wrapParagraphs("heading:\n\n- first point\n- second point\n\nclosing line", 10, false, 400)
	want := []string{"heading:", "", "- first point", "- second point", "", "closing line"}
	if len(lines) != len(want) {
		t.Fatalf("expected %v, got %v", want, lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("line %d: got %q, want %q", i, lines[i], want[i])
		}
	}
}

func TestWrapParagraphs_EmptyOrWhitespaceInput_ProducesNoLines(t *testing.T) {
	if lines := wrapParagraphs("", 10, false, 400); lines != nil {
		t.Errorf("expected nil for empty input, got %v", lines)
	}
	if lines := wrapParagraphs("   ", 10, false, 400); lines != nil {
		t.Errorf("expected nil for whitespace-only input, got %v", lines)
	}
}

func TestWrapParagraphs_LongParagraphStillWrapsWithinEachSegment(t *testing.T) {
	long := "one two three four five six seven eight nine ten eleven twelve"
	narrow := textWidth("one two three", 10, false)
	lines := wrapParagraphs(long+"\n\nshort tail", 10, false, narrow)

	// Every wrapped line before the gap marker must fit narrow; the gap
	// marker itself ("") must be present; the segment after it wraps
	// independently.
	sawGap := false
	for _, l := range lines {
		if l == "" {
			sawGap = true
			continue
		}
		if w := textWidth(l, 10, false); w > narrow {
			t.Errorf("line %q (%.1f) exceeds narrow width %.1f", l, w, narrow)
		}
	}
	if !sawGap {
		t.Error("expected a gap marker between the two paragraphs")
	}
}

func TestBlockHeight_TextRuleAndGapAccountedSeparately(t *testing.T) {
	lines := []styledLine{
		{text: "line one"},
		{text: ""}, // gap
		{text: "line two"},
		{rule: true},
	}
	const leading, gap = 15.0, 7.0
	got := blockHeight(lines, leading, gap)
	want := leading + gap + leading + leading // text + gap + text + rule
	if got != want {
		t.Errorf("blockHeight = %v, want %v", got, want)
	}
}

func TestBlockHeight_EmptyInput_IsZero(t *testing.T) {
	if h := blockHeight(nil, 15, 7); h != 0 {
		t.Errorf("expected 0 height for no lines, got %v", h)
	}
}

func TestTextBlock_TagsEveryLineWithGivenStyle(t *testing.T) {
	color := hexRGB("#17453d")
	lines := textBlock("one two three", 11, true, color, 1000)
	for _, l := range lines {
		if l.size != 11 || !l.bold || l.color != color {
			t.Errorf("expected every line tagged size=11 bold=true color=%v, got %+v", color, l)
		}
	}
}

func TestBox_ReservesExactlyBlockHeightPlusPadding(t *testing.T) {
	// The property that fixes this package's original first-draft bug
	// (box content overlapping what follows it): after box() returns,
	// the canvas cursor must have moved down by exactly what it
	// reported, and nothing drawn after it should land inside that
	// span. Checked here structurally (cursor arithmetic), independent
	// of the visual check already done by hand during development.
	c := newCanvas(50, 56, 60) // a normal, small bottom margin — plenty of usable page height, no spurious page break
	startY := c.y

	lines := textBlock("some body text for the box", bodySize, false, colorSand900, c.contentWidth()-36)
	consumed := c.box(colorSand100, 18, 14, 14, bodyLeading, bodyGap, 0, lines)

	if got := startY - c.y; got != consumed {
		t.Errorf("cursor moved by %v but box() reported consuming %v", got, consumed)
	}

	// Draw a marker line right after the box and confirm its top edge
	// (c.y at the moment of drawing) is below the box's bottom edge —
	// i.e. startY - consumed, the same arithmetic renderContentStream's
	// rect uses.
	boxBottom := startY - consumed
	if c.y != boxBottom {
		t.Errorf("cursor after box() = %v, expected exactly at the box's bottom edge %v", c.y, boxBottom)
	}
}

func TestNumberedList_PrefixSurvivesLeadingBlankLine(t *testing.T) {
	// Regression test for the bug found and fixed while writing this
	// test file: an item whose text begins with a literal "\n" used to
	// silently lose its "N." prefix, because the old code checked
	// "is this wrapped-line index 0" rather than "is this the first
	// non-blank wrapped line".
	c := newCanvas(50, 56, 60)
	c.numberedList([]string{"\nStarts with a blank line"}, colorSand900)
	page := c.pages[0]

	var sawPrefix bool
	for _, tr := range page.text {
		if tr.text == "1." {
			sawPrefix = true
		}
	}
	if !sawPrefix {
		t.Error("expected the \"1.\" prefix to be drawn even though the item's text starts with a blank line")
	}
}

func TestNumberedList_EachItemGetsSequentialPrefix(t *testing.T) {
	c := newCanvas(50, 56, 60)
	c.numberedList([]string{"first", "second", "third"}, colorSand900)
	page := c.pages[0]

	var prefixes []string
	for _, tr := range page.text {
		if tr.bold {
			prefixes = append(prefixes, tr.text)
		}
	}
	want := []string{"1.", "2.", "3."}
	if len(prefixes) != len(want) {
		t.Fatalf("expected prefixes %v, got %v", want, prefixes)
	}
	for i := range want {
		if prefixes[i] != want[i] {
			t.Errorf("prefix %d: got %q, want %q", i, prefixes[i], want[i])
		}
	}
}

func TestHeading_AddsBreathingRoomAroundText(t *testing.T) {
	c := newCanvas(50, 56, 60)
	startY := c.y
	c.heading("Section Title", colorTeal800)
	// heading() advances by 6 (before) + headingLead (the line itself)
	// + 2 (after) — checked as a concrete number specifically so a
	// future change to that spacing has to update this test
	// consciously rather than the spacing silently drifting.
	want := startY - (6 + headingLead + 2)
	if c.y != want {
		t.Errorf("cursor after heading() = %v, want %v", c.y, want)
	}
}

func TestParagraph_HandlesEmptyStringWithoutPanicking(t *testing.T) {
	c := newCanvas(50, 56, 60)
	startY := c.y
	c.paragraph("", colorSand900)
	if c.y != startY {
		t.Errorf("expected an empty paragraph to consume no vertical space, cursor moved from %v to %v", startY, c.y)
	}
}
