// Package document renders a family's case record as a single,
// print-ready PDF — the same content already shown on the case result
// page (backend/internal/response), packaged as one downloadable
// document a family can hand to hospital billing staff, attach to a
// CGRMS complaint, or read from during a call to a NALSA Para Legal
// Volunteer. See this folder's README.md for the full design reasoning
// and docs/OPEN_QUESTIONS.md for how this relates to (and does not
// close) the "automated CGRMS complaint submission" gap this project
// has always been explicit about: this makes the draft a real,
// professional, portable document. It still does not submit anything
// anywhere — the family still does that themselves.
//
// Deliberately zero third-party dependencies, matching the rest of
// this backend (see ../../../ARCHITECTURE.md) — the PDF byte format
// (objects, cross-reference table, trailer) is written directly by
// pdf.go. That surface is real and non-trivial to get right, but it
// stayed tractable specifically because the actual generated text
// (internal/response/templates.go) uses exactly two non-ASCII
// characters in practice — an em dash and the rupee sign — verified
// directly, not assumed: see case_document_test.go's
// TestTemplateTextHasNoUnrenderableCharacters, which scans every
// template string this package ever has to render and fails loudly if
// a future edit introduces one this package can't. A full Unicode
// font-embedding pipeline (CID fonts, glyph subsetting, shaping for
// scripts like Malayalam) was never actually necessary as a result —
// see encodeWinAnsi's doc comment for exactly how the rupee sign is
// handled instead, and why.
package document

// Duplicate rune->byte mappings collapsed (kept the first/lowest byte):
//   U+2022 (bullet): byte 0x81 duplicates byte 0x7F, dropped from the encode map (both bytes still get correct widths below)
//   U+2022 (bullet): byte 0x8D duplicates byte 0x7F, dropped from the encode map (both bytes still get correct widths below)
//   U+2022 (bullet): byte 0x8F duplicates byte 0x7F, dropped from the encode map (both bytes still get correct widths below)
//   U+2022 (bullet): byte 0x90 duplicates byte 0x7F, dropped from the encode map (both bytes still get correct widths below)
//   U+2022 (bullet): byte 0x95 duplicates byte 0x7F, dropped from the encode map (both bytes still get correct widths below)
//   U+2022 (bullet): byte 0x9D duplicates byte 0x7F, dropped from the encode map (both bytes still get correct widths below)
//   U+0020 (space): byte 0xA0 duplicates byte 0x20, dropped from the encode map (both bytes still get correct widths below)
//   U+002D (hyphen): byte 0xAD duplicates byte 0x2D, dropped from the encode map (both bytes still get correct widths below)

var winAnsiEncode = map[rune]byte{
	0x0020: 0x20, // space
	0x0021: 0x21, // exclam
	0x0022: 0x22, // quotedbl
	0x0023: 0x23, // numbersign
	0x0024: 0x24, // dollar
	0x0025: 0x25, // percent
	0x0026: 0x26, // ampersand
	0x0027: 0x27, // quotesingle
	0x0028: 0x28, // parenleft
	0x0029: 0x29, // parenright
	0x002A: 0x2A, // asterisk
	0x002B: 0x2B, // plus
	0x002C: 0x2C, // comma
	0x002D: 0x2D, // hyphen
	0x002E: 0x2E, // period
	0x002F: 0x2F, // slash
	0x0030: 0x30, // zero
	0x0031: 0x31, // one
	0x0032: 0x32, // two
	0x0033: 0x33, // three
	0x0034: 0x34, // four
	0x0035: 0x35, // five
	0x0036: 0x36, // six
	0x0037: 0x37, // seven
	0x0038: 0x38, // eight
	0x0039: 0x39, // nine
	0x003A: 0x3A, // colon
	0x003B: 0x3B, // semicolon
	0x003C: 0x3C, // less
	0x003D: 0x3D, // equal
	0x003E: 0x3E, // greater
	0x003F: 0x3F, // question
	0x0040: 0x40, // at
	0x0041: 0x41, // A
	0x0042: 0x42, // B
	0x0043: 0x43, // C
	0x0044: 0x44, // D
	0x0045: 0x45, // E
	0x0046: 0x46, // F
	0x0047: 0x47, // G
	0x0048: 0x48, // H
	0x0049: 0x49, // I
	0x004A: 0x4A, // J
	0x004B: 0x4B, // K
	0x004C: 0x4C, // L
	0x004D: 0x4D, // M
	0x004E: 0x4E, // N
	0x004F: 0x4F, // O
	0x0050: 0x50, // P
	0x0051: 0x51, // Q
	0x0052: 0x52, // R
	0x0053: 0x53, // S
	0x0054: 0x54, // T
	0x0055: 0x55, // U
	0x0056: 0x56, // V
	0x0057: 0x57, // W
	0x0058: 0x58, // X
	0x0059: 0x59, // Y
	0x005A: 0x5A, // Z
	0x005B: 0x5B, // bracketleft
	0x005C: 0x5C, // backslash
	0x005D: 0x5D, // bracketright
	0x005E: 0x5E, // asciicircum
	0x005F: 0x5F, // underscore
	0x0060: 0x60, // grave
	0x0061: 0x61, // a
	0x0062: 0x62, // b
	0x0063: 0x63, // c
	0x0064: 0x64, // d
	0x0065: 0x65, // e
	0x0066: 0x66, // f
	0x0067: 0x67, // g
	0x0068: 0x68, // h
	0x0069: 0x69, // i
	0x006A: 0x6A, // j
	0x006B: 0x6B, // k
	0x006C: 0x6C, // l
	0x006D: 0x6D, // m
	0x006E: 0x6E, // n
	0x006F: 0x6F, // o
	0x0070: 0x70, // p
	0x0071: 0x71, // q
	0x0072: 0x72, // r
	0x0073: 0x73, // s
	0x0074: 0x74, // t
	0x0075: 0x75, // u
	0x0076: 0x76, // v
	0x0077: 0x77, // w
	0x0078: 0x78, // x
	0x0079: 0x79, // y
	0x007A: 0x7A, // z
	0x007B: 0x7B, // braceleft
	0x007C: 0x7C, // bar
	0x007D: 0x7D, // braceright
	0x007E: 0x7E, // asciitilde
	0x2022: 0x7F, // bullet
	0x20AC: 0x80, // Euro
	0x201A: 0x82, // quotesinglbase
	0x0192: 0x83, // florin
	0x201E: 0x84, // quotedblbase
	0x2026: 0x85, // ellipsis
	0x2020: 0x86, // dagger
	0x2021: 0x87, // daggerdbl
	0x02C6: 0x88, // circumflex
	0x2030: 0x89, // perthousand
	0x0160: 0x8A, // Scaron
	0x2039: 0x8B, // guilsinglleft
	0x0152: 0x8C, // OE
	0x017D: 0x8E, // Zcaron
	0x2018: 0x91, // quoteleft
	0x2019: 0x92, // quoteright
	0x201C: 0x93, // quotedblleft
	0x201D: 0x94, // quotedblright
	0x2013: 0x96, // endash
	0x2014: 0x97, // emdash
	0x02DC: 0x98, // tilde
	0x2122: 0x99, // trademark
	0x0161: 0x9A, // scaron
	0x203A: 0x9B, // guilsinglright
	0x0153: 0x9C, // oe
	0x017E: 0x9E, // zcaron
	0x0178: 0x9F, // Ydieresis
	0x00A1: 0xA1, // exclamdown
	0x00A2: 0xA2, // cent
	0x00A3: 0xA3, // sterling
	0x00A4: 0xA4, // currency
	0x00A5: 0xA5, // yen
	0x00A6: 0xA6, // brokenbar
	0x00A7: 0xA7, // section
	0x00A8: 0xA8, // dieresis
	0x00A9: 0xA9, // copyright
	0x00AA: 0xAA, // ordfeminine
	0x00AB: 0xAB, // guillemotleft
	0x00AC: 0xAC, // logicalnot
	0x00AE: 0xAE, // registered
	0x00AF: 0xAF, // macron
	0x00B0: 0xB0, // degree
	0x00B1: 0xB1, // plusminus
	0x00B2: 0xB2, // twosuperior
	0x00B3: 0xB3, // threesuperior
	0x00B4: 0xB4, // acute
	0x00B5: 0xB5, // mu
	0x00B6: 0xB6, // paragraph
	0x00B7: 0xB7, // periodcentered
	0x00B8: 0xB8, // cedilla
	0x00B9: 0xB9, // onesuperior
	0x00BA: 0xBA, // ordmasculine
	0x00BB: 0xBB, // guillemotright
	0x00BC: 0xBC, // onequarter
	0x00BD: 0xBD, // onehalf
	0x00BE: 0xBE, // threequarters
	0x00BF: 0xBF, // questiondown
	0x00C0: 0xC0, // Agrave
	0x00C1: 0xC1, // Aacute
	0x00C2: 0xC2, // Acircumflex
	0x00C3: 0xC3, // Atilde
	0x00C4: 0xC4, // Adieresis
	0x00C5: 0xC5, // Aring
	0x00C6: 0xC6, // AE
	0x00C7: 0xC7, // Ccedilla
	0x00C8: 0xC8, // Egrave
	0x00C9: 0xC9, // Eacute
	0x00CA: 0xCA, // Ecircumflex
	0x00CB: 0xCB, // Edieresis
	0x00CC: 0xCC, // Igrave
	0x00CD: 0xCD, // Iacute
	0x00CE: 0xCE, // Icircumflex
	0x00CF: 0xCF, // Idieresis
	0x00D0: 0xD0, // Eth
	0x00D1: 0xD1, // Ntilde
	0x00D2: 0xD2, // Ograve
	0x00D3: 0xD3, // Oacute
	0x00D4: 0xD4, // Ocircumflex
	0x00D5: 0xD5, // Otilde
	0x00D6: 0xD6, // Odieresis
	0x00D7: 0xD7, // multiply
	0x00D8: 0xD8, // Oslash
	0x00D9: 0xD9, // Ugrave
	0x00DA: 0xDA, // Uacute
	0x00DB: 0xDB, // Ucircumflex
	0x00DC: 0xDC, // Udieresis
	0x00DD: 0xDD, // Yacute
	0x00DE: 0xDE, // Thorn
	0x00DF: 0xDF, // germandbls
	0x00E0: 0xE0, // agrave
	0x00E1: 0xE1, // aacute
	0x00E2: 0xE2, // acircumflex
	0x00E3: 0xE3, // atilde
	0x00E4: 0xE4, // adieresis
	0x00E5: 0xE5, // aring
	0x00E6: 0xE6, // ae
	0x00E7: 0xE7, // ccedilla
	0x00E8: 0xE8, // egrave
	0x00E9: 0xE9, // eacute
	0x00EA: 0xEA, // ecircumflex
	0x00EB: 0xEB, // edieresis
	0x00EC: 0xEC, // igrave
	0x00ED: 0xED, // iacute
	0x00EE: 0xEE, // icircumflex
	0x00EF: 0xEF, // idieresis
	0x00F0: 0xF0, // eth
	0x00F1: 0xF1, // ntilde
	0x00F2: 0xF2, // ograve
	0x00F3: 0xF3, // oacute
	0x00F4: 0xF4, // ocircumflex
	0x00F5: 0xF5, // otilde
	0x00F6: 0xF6, // odieresis
	0x00F7: 0xF7, // divide
	0x00F8: 0xF8, // oslash
	0x00F9: 0xF9, // ugrave
	0x00FA: 0xFA, // uacute
	0x00FB: 0xFB, // ucircumflex
	0x00FC: 0xFC, // udieresis
	0x00FD: 0xFD, // yacute
	0x00FE: 0xFE, // thorn
	0x00FF: 0xFF, // ydieresis
}

var helvW, helvBoldW [256]int

func init() {
	for i := range helvW {
		helvW[i] = 556
		helvBoldW[i] = 556
	}
	helvW[0x20], helvBoldW[0x20] = 278, 278   // space
	helvW[0x21], helvBoldW[0x21] = 278, 333   // exclam
	helvW[0x22], helvBoldW[0x22] = 355, 474   // quotedbl
	helvW[0x23], helvBoldW[0x23] = 556, 556   // numbersign
	helvW[0x24], helvBoldW[0x24] = 556, 556   // dollar
	helvW[0x25], helvBoldW[0x25] = 889, 889   // percent
	helvW[0x26], helvBoldW[0x26] = 667, 722   // ampersand
	helvW[0x27], helvBoldW[0x27] = 191, 238   // quotesingle
	helvW[0x28], helvBoldW[0x28] = 333, 333   // parenleft
	helvW[0x29], helvBoldW[0x29] = 333, 333   // parenright
	helvW[0x2A], helvBoldW[0x2A] = 389, 389   // asterisk
	helvW[0x2B], helvBoldW[0x2B] = 584, 584   // plus
	helvW[0x2C], helvBoldW[0x2C] = 278, 278   // comma
	helvW[0x2D], helvBoldW[0x2D] = 333, 333   // hyphen
	helvW[0x2E], helvBoldW[0x2E] = 278, 278   // period
	helvW[0x2F], helvBoldW[0x2F] = 278, 278   // slash
	helvW[0x30], helvBoldW[0x30] = 556, 556   // zero
	helvW[0x31], helvBoldW[0x31] = 556, 556   // one
	helvW[0x32], helvBoldW[0x32] = 556, 556   // two
	helvW[0x33], helvBoldW[0x33] = 556, 556   // three
	helvW[0x34], helvBoldW[0x34] = 556, 556   // four
	helvW[0x35], helvBoldW[0x35] = 556, 556   // five
	helvW[0x36], helvBoldW[0x36] = 556, 556   // six
	helvW[0x37], helvBoldW[0x37] = 556, 556   // seven
	helvW[0x38], helvBoldW[0x38] = 556, 556   // eight
	helvW[0x39], helvBoldW[0x39] = 556, 556   // nine
	helvW[0x3A], helvBoldW[0x3A] = 278, 333   // colon
	helvW[0x3B], helvBoldW[0x3B] = 278, 333   // semicolon
	helvW[0x3C], helvBoldW[0x3C] = 584, 584   // less
	helvW[0x3D], helvBoldW[0x3D] = 584, 584   // equal
	helvW[0x3E], helvBoldW[0x3E] = 584, 584   // greater
	helvW[0x3F], helvBoldW[0x3F] = 556, 611   // question
	helvW[0x40], helvBoldW[0x40] = 1015, 975  // at
	helvW[0x41], helvBoldW[0x41] = 667, 722   // A
	helvW[0x42], helvBoldW[0x42] = 667, 722   // B
	helvW[0x43], helvBoldW[0x43] = 722, 722   // C
	helvW[0x44], helvBoldW[0x44] = 722, 722   // D
	helvW[0x45], helvBoldW[0x45] = 667, 667   // E
	helvW[0x46], helvBoldW[0x46] = 611, 611   // F
	helvW[0x47], helvBoldW[0x47] = 778, 778   // G
	helvW[0x48], helvBoldW[0x48] = 722, 722   // H
	helvW[0x49], helvBoldW[0x49] = 278, 278   // I
	helvW[0x4A], helvBoldW[0x4A] = 500, 556   // J
	helvW[0x4B], helvBoldW[0x4B] = 667, 722   // K
	helvW[0x4C], helvBoldW[0x4C] = 556, 611   // L
	helvW[0x4D], helvBoldW[0x4D] = 833, 833   // M
	helvW[0x4E], helvBoldW[0x4E] = 722, 722   // N
	helvW[0x4F], helvBoldW[0x4F] = 778, 778   // O
	helvW[0x50], helvBoldW[0x50] = 667, 667   // P
	helvW[0x51], helvBoldW[0x51] = 778, 778   // Q
	helvW[0x52], helvBoldW[0x52] = 722, 722   // R
	helvW[0x53], helvBoldW[0x53] = 667, 667   // S
	helvW[0x54], helvBoldW[0x54] = 611, 611   // T
	helvW[0x55], helvBoldW[0x55] = 722, 722   // U
	helvW[0x56], helvBoldW[0x56] = 667, 667   // V
	helvW[0x57], helvBoldW[0x57] = 944, 944   // W
	helvW[0x58], helvBoldW[0x58] = 667, 667   // X
	helvW[0x59], helvBoldW[0x59] = 667, 667   // Y
	helvW[0x5A], helvBoldW[0x5A] = 611, 611   // Z
	helvW[0x5B], helvBoldW[0x5B] = 278, 333   // bracketleft
	helvW[0x5C], helvBoldW[0x5C] = 278, 278   // backslash
	helvW[0x5D], helvBoldW[0x5D] = 278, 333   // bracketright
	helvW[0x5E], helvBoldW[0x5E] = 469, 584   // asciicircum
	helvW[0x5F], helvBoldW[0x5F] = 556, 556   // underscore
	helvW[0x60], helvBoldW[0x60] = 333, 333   // grave
	helvW[0x61], helvBoldW[0x61] = 556, 556   // a
	helvW[0x62], helvBoldW[0x62] = 556, 611   // b
	helvW[0x63], helvBoldW[0x63] = 500, 556   // c
	helvW[0x64], helvBoldW[0x64] = 556, 611   // d
	helvW[0x65], helvBoldW[0x65] = 556, 556   // e
	helvW[0x66], helvBoldW[0x66] = 278, 333   // f
	helvW[0x67], helvBoldW[0x67] = 556, 611   // g
	helvW[0x68], helvBoldW[0x68] = 556, 611   // h
	helvW[0x69], helvBoldW[0x69] = 222, 278   // i
	helvW[0x6A], helvBoldW[0x6A] = 222, 278   // j
	helvW[0x6B], helvBoldW[0x6B] = 500, 556   // k
	helvW[0x6C], helvBoldW[0x6C] = 222, 278   // l
	helvW[0x6D], helvBoldW[0x6D] = 833, 889   // m
	helvW[0x6E], helvBoldW[0x6E] = 556, 611   // n
	helvW[0x6F], helvBoldW[0x6F] = 556, 611   // o
	helvW[0x70], helvBoldW[0x70] = 556, 611   // p
	helvW[0x71], helvBoldW[0x71] = 556, 611   // q
	helvW[0x72], helvBoldW[0x72] = 333, 389   // r
	helvW[0x73], helvBoldW[0x73] = 500, 556   // s
	helvW[0x74], helvBoldW[0x74] = 278, 333   // t
	helvW[0x75], helvBoldW[0x75] = 556, 611   // u
	helvW[0x76], helvBoldW[0x76] = 500, 556   // v
	helvW[0x77], helvBoldW[0x77] = 722, 778   // w
	helvW[0x78], helvBoldW[0x78] = 500, 556   // x
	helvW[0x79], helvBoldW[0x79] = 500, 556   // y
	helvW[0x7A], helvBoldW[0x7A] = 500, 500   // z
	helvW[0x7B], helvBoldW[0x7B] = 334, 389   // braceleft
	helvW[0x7C], helvBoldW[0x7C] = 260, 280   // bar
	helvW[0x7D], helvBoldW[0x7D] = 334, 389   // braceright
	helvW[0x7E], helvBoldW[0x7E] = 584, 584   // asciitilde
	helvW[0x7F], helvBoldW[0x7F] = 350, 350   // bullet
	helvW[0x80], helvBoldW[0x80] = 556, 556   // Euro
	helvW[0x81], helvBoldW[0x81] = 350, 350   // bullet
	helvW[0x82], helvBoldW[0x82] = 222, 278   // quotesinglbase
	helvW[0x83], helvBoldW[0x83] = 556, 556   // florin
	helvW[0x84], helvBoldW[0x84] = 333, 500   // quotedblbase
	helvW[0x85], helvBoldW[0x85] = 1000, 1000 // ellipsis
	helvW[0x86], helvBoldW[0x86] = 556, 556   // dagger
	helvW[0x87], helvBoldW[0x87] = 556, 556   // daggerdbl
	helvW[0x88], helvBoldW[0x88] = 333, 333   // circumflex
	helvW[0x89], helvBoldW[0x89] = 1000, 1000 // perthousand
	helvW[0x8A], helvBoldW[0x8A] = 667, 667   // Scaron
	helvW[0x8B], helvBoldW[0x8B] = 333, 333   // guilsinglleft
	helvW[0x8C], helvBoldW[0x8C] = 1000, 1000 // OE
	helvW[0x8D], helvBoldW[0x8D] = 350, 350   // bullet
	helvW[0x8E], helvBoldW[0x8E] = 611, 611   // Zcaron
	helvW[0x8F], helvBoldW[0x8F] = 350, 350   // bullet
	helvW[0x90], helvBoldW[0x90] = 350, 350   // bullet
	helvW[0x91], helvBoldW[0x91] = 222, 278   // quoteleft
	helvW[0x92], helvBoldW[0x92] = 222, 278   // quoteright
	helvW[0x93], helvBoldW[0x93] = 333, 500   // quotedblleft
	helvW[0x94], helvBoldW[0x94] = 333, 500   // quotedblright
	helvW[0x95], helvBoldW[0x95] = 350, 350   // bullet
	helvW[0x96], helvBoldW[0x96] = 556, 556   // endash
	helvW[0x97], helvBoldW[0x97] = 1000, 1000 // emdash
	helvW[0x98], helvBoldW[0x98] = 333, 333   // tilde
	helvW[0x99], helvBoldW[0x99] = 1000, 1000 // trademark
	helvW[0x9A], helvBoldW[0x9A] = 500, 556   // scaron
	helvW[0x9B], helvBoldW[0x9B] = 333, 333   // guilsinglright
	helvW[0x9C], helvBoldW[0x9C] = 944, 944   // oe
	helvW[0x9D], helvBoldW[0x9D] = 350, 350   // bullet
	helvW[0x9E], helvBoldW[0x9E] = 500, 500   // zcaron
	helvW[0x9F], helvBoldW[0x9F] = 667, 667   // Ydieresis
	helvW[0xA0], helvBoldW[0xA0] = 278, 278   // space
	helvW[0xA1], helvBoldW[0xA1] = 333, 333   // exclamdown
	helvW[0xA2], helvBoldW[0xA2] = 556, 556   // cent
	helvW[0xA3], helvBoldW[0xA3] = 556, 556   // sterling
	helvW[0xA4], helvBoldW[0xA4] = 556, 556   // currency
	helvW[0xA5], helvBoldW[0xA5] = 556, 556   // yen
	helvW[0xA6], helvBoldW[0xA6] = 260, 280   // brokenbar
	helvW[0xA7], helvBoldW[0xA7] = 556, 556   // section
	helvW[0xA8], helvBoldW[0xA8] = 333, 333   // dieresis
	helvW[0xA9], helvBoldW[0xA9] = 737, 737   // copyright
	helvW[0xAA], helvBoldW[0xAA] = 370, 370   // ordfeminine
	helvW[0xAB], helvBoldW[0xAB] = 556, 556   // guillemotleft
	helvW[0xAC], helvBoldW[0xAC] = 584, 584   // logicalnot
	helvW[0xAD], helvBoldW[0xAD] = 333, 333   // hyphen
	helvW[0xAE], helvBoldW[0xAE] = 737, 737   // registered
	helvW[0xAF], helvBoldW[0xAF] = 333, 333   // macron
	helvW[0xB0], helvBoldW[0xB0] = 400, 400   // degree
	helvW[0xB1], helvBoldW[0xB1] = 584, 584   // plusminus
	helvW[0xB2], helvBoldW[0xB2] = 333, 333   // twosuperior
	helvW[0xB3], helvBoldW[0xB3] = 333, 333   // threesuperior
	helvW[0xB4], helvBoldW[0xB4] = 333, 333   // acute
	helvW[0xB5], helvBoldW[0xB5] = 556, 611   // mu
	helvW[0xB6], helvBoldW[0xB6] = 537, 556   // paragraph
	helvW[0xB7], helvBoldW[0xB7] = 278, 278   // periodcentered
	helvW[0xB8], helvBoldW[0xB8] = 333, 333   // cedilla
	helvW[0xB9], helvBoldW[0xB9] = 333, 333   // onesuperior
	helvW[0xBA], helvBoldW[0xBA] = 365, 365   // ordmasculine
	helvW[0xBB], helvBoldW[0xBB] = 556, 556   // guillemotright
	helvW[0xBC], helvBoldW[0xBC] = 834, 834   // onequarter
	helvW[0xBD], helvBoldW[0xBD] = 834, 834   // onehalf
	helvW[0xBE], helvBoldW[0xBE] = 834, 834   // threequarters
	helvW[0xBF], helvBoldW[0xBF] = 611, 611   // questiondown
	helvW[0xC0], helvBoldW[0xC0] = 667, 722   // Agrave
	helvW[0xC1], helvBoldW[0xC1] = 667, 722   // Aacute
	helvW[0xC2], helvBoldW[0xC2] = 667, 722   // Acircumflex
	helvW[0xC3], helvBoldW[0xC3] = 667, 722   // Atilde
	helvW[0xC4], helvBoldW[0xC4] = 667, 722   // Adieresis
	helvW[0xC5], helvBoldW[0xC5] = 667, 722   // Aring
	helvW[0xC6], helvBoldW[0xC6] = 1000, 1000 // AE
	helvW[0xC7], helvBoldW[0xC7] = 722, 722   // Ccedilla
	helvW[0xC8], helvBoldW[0xC8] = 667, 667   // Egrave
	helvW[0xC9], helvBoldW[0xC9] = 667, 667   // Eacute
	helvW[0xCA], helvBoldW[0xCA] = 667, 667   // Ecircumflex
	helvW[0xCB], helvBoldW[0xCB] = 667, 667   // Edieresis
	helvW[0xCC], helvBoldW[0xCC] = 278, 278   // Igrave
	helvW[0xCD], helvBoldW[0xCD] = 278, 278   // Iacute
	helvW[0xCE], helvBoldW[0xCE] = 278, 278   // Icircumflex
	helvW[0xCF], helvBoldW[0xCF] = 278, 278   // Idieresis
	helvW[0xD0], helvBoldW[0xD0] = 722, 722   // Eth
	helvW[0xD1], helvBoldW[0xD1] = 722, 722   // Ntilde
	helvW[0xD2], helvBoldW[0xD2] = 778, 778   // Ograve
	helvW[0xD3], helvBoldW[0xD3] = 778, 778   // Oacute
	helvW[0xD4], helvBoldW[0xD4] = 778, 778   // Ocircumflex
	helvW[0xD5], helvBoldW[0xD5] = 778, 778   // Otilde
	helvW[0xD6], helvBoldW[0xD6] = 778, 778   // Odieresis
	helvW[0xD7], helvBoldW[0xD7] = 584, 584   // multiply
	helvW[0xD8], helvBoldW[0xD8] = 778, 778   // Oslash
	helvW[0xD9], helvBoldW[0xD9] = 722, 722   // Ugrave
	helvW[0xDA], helvBoldW[0xDA] = 722, 722   // Uacute
	helvW[0xDB], helvBoldW[0xDB] = 722, 722   // Ucircumflex
	helvW[0xDC], helvBoldW[0xDC] = 722, 722   // Udieresis
	helvW[0xDD], helvBoldW[0xDD] = 667, 667   // Yacute
	helvW[0xDE], helvBoldW[0xDE] = 667, 667   // Thorn
	helvW[0xDF], helvBoldW[0xDF] = 611, 611   // germandbls
	helvW[0xE0], helvBoldW[0xE0] = 556, 556   // agrave
	helvW[0xE1], helvBoldW[0xE1] = 556, 556   // aacute
	helvW[0xE2], helvBoldW[0xE2] = 556, 556   // acircumflex
	helvW[0xE3], helvBoldW[0xE3] = 556, 556   // atilde
	helvW[0xE4], helvBoldW[0xE4] = 556, 556   // adieresis
	helvW[0xE5], helvBoldW[0xE5] = 556, 556   // aring
	helvW[0xE6], helvBoldW[0xE6] = 889, 889   // ae
	helvW[0xE7], helvBoldW[0xE7] = 500, 556   // ccedilla
	helvW[0xE8], helvBoldW[0xE8] = 556, 556   // egrave
	helvW[0xE9], helvBoldW[0xE9] = 556, 556   // eacute
	helvW[0xEA], helvBoldW[0xEA] = 556, 556   // ecircumflex
	helvW[0xEB], helvBoldW[0xEB] = 556, 556   // edieresis
	helvW[0xEC], helvBoldW[0xEC] = 278, 278   // igrave
	helvW[0xED], helvBoldW[0xED] = 278, 278   // iacute
	helvW[0xEE], helvBoldW[0xEE] = 278, 278   // icircumflex
	helvW[0xEF], helvBoldW[0xEF] = 278, 278   // idieresis
	helvW[0xF0], helvBoldW[0xF0] = 556, 611   // eth
	helvW[0xF1], helvBoldW[0xF1] = 556, 611   // ntilde
	helvW[0xF2], helvBoldW[0xF2] = 556, 611   // ograve
	helvW[0xF3], helvBoldW[0xF3] = 556, 611   // oacute
	helvW[0xF4], helvBoldW[0xF4] = 556, 611   // ocircumflex
	helvW[0xF5], helvBoldW[0xF5] = 556, 611   // otilde
	helvW[0xF6], helvBoldW[0xF6] = 556, 611   // odieresis
	helvW[0xF7], helvBoldW[0xF7] = 584, 584   // divide
	helvW[0xF8], helvBoldW[0xF8] = 611, 611   // oslash
	helvW[0xF9], helvBoldW[0xF9] = 556, 611   // ugrave
	helvW[0xFA], helvBoldW[0xFA] = 556, 611   // uacute
	helvW[0xFB], helvBoldW[0xFB] = 556, 611   // ucircumflex
	helvW[0xFC], helvBoldW[0xFC] = 556, 611   // udieresis
	helvW[0xFD], helvBoldW[0xFD] = 500, 556   // yacute
	helvW[0xFE], helvBoldW[0xFE] = 556, 611   // thorn
	helvW[0xFF], helvBoldW[0xFF] = 500, 556   // ydieresis
}

// rupeeSubstitute is what the rupee sign (U+20B9) becomes in generated
// PDF text. The standard 14 PDF fonts predate the symbol (it was only
// adopted in 2010, decades after Helvetica's metrics were fixed) and
// have no glyph for it under any standard encoding, WinAnsi included —
// this isn't a gap in this package's encoding table, the glyph simply
// does not exist in an unembedded standard font. "Rs." is the
// established fallback used throughout formal/typewritten Indian
// financial and legal documents specifically because of this exact
// limitation, and reads unambiguously next to a number either way. The
// alternative (embedding a Unicode font) was deliberately not done —
// see the package doc comment.
const rupeeSubstitute = "Rs."

// unmappableSubstitute stands in for any rune this package encounters
// that isn't the rupee sign and still isn't in winAnsiEncode — not
// expected to ever fire on this package's own template text (see
// case_document_test.go), but a live server must never panic or emit
// invalid PDF bytes over a single unanticipated character, so this is
// the deliberate, safe degradation path rather than no path at all.
const unmappableSubstitute = '?'

// shapedText is already-encoded, ready-to-embed PDF content: exactly
// the bytes to place inside a literal-string PDF operator (already
// escaped), plus the total width that same content will occupy when
// drawn, in points. Producing both together from one pass over the
// input is deliberate — layout.go's line-wrapping decisions and pdf.go's
// actual drawing must agree on width down to the same substitutions
// (e.g. "₹" becoming "Rs."), and computing them separately would risk
// the two silently drifting apart, which is exactly the kind of "looks
// fine until a real family sees ₹49,300 rendered as a broken glyph"
// bug this package exists to avoid.
type shapedText struct {
	pdfBytes   []byte // escaped, ready for "(" + pdfBytes + ")" in a content stream
	widthPts   float64
	unmappable []rune // non-empty only if a rune outside {rupee, WinAnsi} was hit
}

// shapeText converts s (arbitrary Go/UTF-8 text, as produced by
// internal/response's templates) into PDF-ready bytes at the given
// font/size, applying the rupee substitution and escaping PDF's three
// special literal-string characters. Raw CR/LF/TAB in s are normalized
// to a single space — this function always produces single-line
// output; splitting text across lines is layout.go's job, done before
// this is ever called, so a stray newline reaching here would
// otherwise silently corrupt the content stream's line structure.
func shapeText(s string, size float64, bold bool) shapedText {
	widths := &helvW
	if bold {
		widths = &helvBoldW
	}

	var out []byte
	var width float64
	var bad []rune

	appendByte := func(b byte, w int) {
		switch b {
		case '(', ')', '\\':
			out = append(out, '\\', b)
		default:
			out = append(out, b)
		}
		width += float64(w) * size / 1000.0
	}

	for _, r := range s {
		switch r {
		case '\r', '\n', '\t':
			appendByte(' ', widths[0x20])
			continue
		case '₹':
			for _, sub := range rupeeSubstitute {
				b := byte(sub) // rupeeSubstitute is pure ASCII by construction
				appendByte(b, widths[b])
			}
			continue
		}

		b, ok := winAnsiEncode[r]
		if !ok {
			bad = append(bad, r)
			b = unmappableSubstitute
		}
		appendByte(b, widths[b])
	}

	return shapedText{pdfBytes: out, widthPts: width, unmappable: bad}
}

// textWidth is shapeText's width computation without the byte-encoding
// work, for call sites (line-wrapping) that only need to know whether a
// candidate line fits before committing to build it. Deliberately
// implemented as a thin wrapper around shapeText itself, rather than a
// second hand-written pass, specifically so it is structurally
// impossible for this function's idea of a string's width to drift from
// what shapeText actually draws for that same string — see shapedText's
// doc comment on why that agreement matters.
func textWidth(s string, size float64, bold bool) float64 {
	return shapeText(s, size, bold).widthPts
}
