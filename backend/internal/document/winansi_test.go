package document

import (
	"bytes"
	"strings"
	"testing"
)

func TestShapeText_PlainASCIIRoundTrips(t *testing.T) {
	got := shapeText("Hello, world.", 10, false)
	if string(got.pdfBytes) != "Hello, world." {
		t.Errorf("expected plain ASCII to pass through unchanged, got %q", got.pdfBytes)
	}
	if len(got.unmappable) != 0 {
		t.Errorf("expected no unmappable runes, got %v", got.unmappable)
	}
	if got.widthPts <= 0 {
		t.Errorf("expected positive width, got %f", got.widthPts)
	}
}

func TestShapeText_EscapesPDFSpecialCharacters(t *testing.T) {
	got := shapeText(`say (hello) or \slash\`, 10, false)
	want := `say \(hello\) or \\slash\\`
	if string(got.pdfBytes) != want {
		t.Errorf("expected %q, got %q", want, got.pdfBytes)
	}
}

func TestShapeText_RupeeSignBecomesRs(t *testing.T) {
	got := shapeText("₹49,300", 10, false)
	if string(got.pdfBytes) != "Rs.49,300" {
		t.Errorf(`expected "Rs.49,300", got %q`, got.pdfBytes)
	}
	if len(got.unmappable) != 0 {
		t.Errorf("rupee sign should be a deliberate substitution, not an unmappable rune: got %v", got.unmappable)
	}
}

func TestShapeText_EmDashIsDirectlyMappable(t *testing.T) {
	// The em dash is the other non-ASCII character templates.go actually
	// uses (126 times) — unlike the rupee sign, WinAnsiEncoding has a
	// real glyph for it (byte 0x97), so this must NOT go through the
	// unmappable-substitution path.
	got := shapeText("before—after", 10, false)
	if len(got.unmappable) != 0 {
		t.Errorf("em dash should be directly mappable via WinAnsiEncoding, got unmappable: %v", got.unmappable)
	}
	// Deliberately checking for the raw WinAnsi byte 0x97, not
	// string(rune(0x97)) — that would UTF-8-encode the codepoint
	// U+0097 (a 2-byte sequence) rather than testing for the single
	// raw byte this package actually emits.
	if !bytes.Contains(got.pdfBytes, []byte{0x97}) {
		t.Errorf("expected raw WinAnsi byte 0x97 (emdash) in output, got %q", got.pdfBytes)
	}
}

func TestShapeText_TrulyUnmappableRuneDegradesSafely(t *testing.T) {
	// A rune with no WinAnsi glyph and no deliberate substitution (e.g.
	// a checkmark, or Malayalam script) must never panic and must never
	// produce invalid PDF bytes — it degrades to '?' and is reported.
	got := shapeText("ok ✓ done", 10, false)
	if len(got.unmappable) != 1 || got.unmappable[0] != '✓' {
		t.Errorf("expected the checkmark reported as unmappable, got %v", got.unmappable)
	}
	if !strings.Contains(string(got.pdfBytes), "?") {
		t.Errorf("expected '?' substitute in output, got %q", got.pdfBytes)
	}
}

func TestShapeText_NewlinesNormalizedToSpace(t *testing.T) {
	got := shapeText("a\nb\rc\td", 10, false)
	if string(got.pdfBytes) != "a b c d" {
		t.Errorf("expected control characters normalized to spaces, got %q", got.pdfBytes)
	}
}

func TestTextWidth_AgreesWithShapeText(t *testing.T) {
	// textWidth is a thin wrapper specifically so it can never drift
	// from what shapeText actually draws (see winansi.go's doc comment
	// on shapedText) — assert that equivalence directly rather than
	// trusting the implementation comment.
	for _, s := range []string{"plain text", "₹49,300 with rupee", "em—dash", "unmapped ✓ rune"} {
		w1 := textWidth(s, 11, false)
		w2 := shapeText(s, 11, false).widthPts
		if w1 != w2 {
			t.Errorf("textWidth(%q) = %f, shapeText(...).widthPts = %f; must be identical", s, w1, w2)
		}
	}
}

func TestTextWidth_BoldIsWiderThanRegularForKnownCase(t *testing.T) {
	s := "PMJAY Advocate Case Summary"
	regular := textWidth(s, 12, false)
	bold := textWidth(s, 12, true)
	if bold <= regular {
		t.Errorf("expected bold width > regular width for %q, got bold=%f regular=%f", s, bold, regular)
	}
}

func TestTextWidth_LongerStringsAreWider(t *testing.T) {
	short := textWidth("short", 10, false)
	long := textWidth("a much longer string of text", 10, false)
	if long <= short {
		t.Errorf("expected longer string to measure wider: short=%f long=%f", short, long)
	}
}

func TestWinAnsiEncode_NoDuplicateByteAmbiguityForCommonPunctuation(t *testing.T) {
	// Sanity check on the generated table itself: the handful of runes
	// with more than one valid WinAnsi byte (space, hyphen, bullet) must
	// still resolve to *some* valid, correctly-widthed byte after
	// deduplication — this test would catch a broken code-generation
	// step even though the map only exposes one byte per rune.
	for _, r := range []rune{' ', '-', '\u2022'} {
		b, ok := winAnsiEncode[r]
		if !ok {
			t.Errorf("expected rune %q (U+%04X) to be mapped", r, r)
			continue
		}
		if helvW[b] <= 0 || helvBoldW[b] <= 0 {
			t.Errorf("rune %q mapped to byte 0x%02X with non-positive width: helv=%d helvBold=%d", r, b, helvW[b], helvBoldW[b])
		}
	}
}

func TestWinAnsiWidthTables_FullyPopulated(t *testing.T) {
	// Every byte that winAnsiEncode can ever produce must have a real
	// (non-default-556-by-coincidence) width, and no width should be
	// zero or negative — either would make wrapping silently wrong.
	for r, b := range winAnsiEncode {
		if helvW[b] <= 0 {
			t.Errorf("rune %q (byte 0x%02X) has non-positive Helvetica width %d", r, b, helvW[b])
		}
		if helvBoldW[b] <= 0 {
			t.Errorf("rune %q (byte 0x%02X) has non-positive Helvetica-Bold width %d", r, b, helvBoldW[b])
		}
	}
}
