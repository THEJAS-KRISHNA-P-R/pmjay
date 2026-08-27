// Package retrieval narrows ~hundreds of HBP packages down to a short,
// plausible candidate list before any LLM call happens.
//
// Why this layer exists at all (cost, not just latency): the spec's own
// implementation notes (Appendix Z) describe this as retrieval-then-
// disambiguation, not a single end-to-end model call. Sending the full
// package list as context on every single family query would mean every
// query pays to re-read the entire dataset — with the real ~1,900-code
// HBP list that is not a rounding error. Narrowing to ~15-20 plausible
// candidates first, in plain Go with no external calls, means the LLM
// only ever has to read and reason about a handful of real options,
// which is the single largest lever this system has on its own running
// cost (see ../../../ARCHITECTURE.md, "reducing bills").
//
// This layer is deliberately dumb: English-lexicon keyword overlap
// against package names, specialties, and curated description keywords.
// It does not need to understand Malayalam or code-mixed phrasing — the
// LLM extraction step downstream receives the family's full original
// text, in whatever language mix they used, and does the actual
// understanding (spec Section 58: "AI where genuine language ambiguity
// exists, deterministic logic everywhere else"). This layer only needs
// to be recall-biased enough not to accidentally exclude the right
// answer from the shortlist. If it returns too broad a shortlist, that's
// a cost problem to tune later; if it returns too narrow a shortlist,
// that's a correctness problem — so ties and near-misses are kept in,
// not filtered out.
package retrieval

import (
	"sort"
	"strings"

	"github.com/pmjay-advocate/backend/internal/hbp"
)

// MaxCandidates bounds how many packages get forwarded to the LLM step
// when keyword scoring found real signal (at least one package scored
// above zero). Chosen generously relative to the seed dataset size so
// the shortlist essentially never excludes a real match; on the real
// ~1,900-code dataset this is the number that actually controls
// per-query cost.
const MaxCandidates = 20

// MaxCandidatesWhenBlind bounds the fallback slice used when keyword
// scoring found ZERO signal — every package scored 0, so "top N by
// score" is not a ranking at all, just the dataset's on-disk order.
// This is a materially different situation from the normal MaxCandidates
// case and deliberately gets a larger allowance: with real signal, a
// tight shortlist is a considered trade of recall for cost; with no
// signal, a tight shortlist is pure luck-of-file-order, and the
// consequence is silently excluding the right answer for exactly the
// families this system most needs to work for (see the doc comment on
// Retrieve below for how this is currently reached and confirmed).
//
// This is a partial mitigation, not a fix: even 4x MaxCandidates is a
// small fraction of the eventual ~1,900-code dataset, so a genuinely
// zero-signal description can still miss the real answer — it just
// misses less often than at MaxCandidates. The actual fix is one of:
// language-specific keyword lists added to the HBP data by a qualified
// translator (not fabricated here — wrong medical terminology in this
// specific field is actively harmful, not just imprecise), or a cheap
// translation/pre-classification pass ahead of this keyword layer. That
// is a product and cost decision, not one this constant can make on its
// own — revisit this number, and whether it's enough, once real query
// logs show how often the zero-signal case actually happens.
const MaxCandidatesWhenBlind = 4 * MaxCandidates

// Candidate is a package paired with the retrieval score that surfaced it,
// kept only for debugging/observability — internal/tiering and
// internal/extract only care about the ordered package list.
type Candidate struct {
	Package hbp.Package
	Score   int
}

// Retrieve scores every package in the dataset against the family's raw
// description text and returns the top matches, most plausible first,
// always including the Unspecified Procedure catch-all so the downstream
// LLM step can consider it explicitly rather than never seeing it.
//
// Retrieve never returns an empty slice for a non-empty dataset: if
// nothing scores above zero, it falls back to returning a larger slice
// of the dataset (bounded by MaxCandidatesWhenBlind, not MaxCandidates —
// see that constant's doc comment for why) so a genuinely novel
// description still reaches the LLM rather than being silently dropped
// before it gets a chance at real reasoning. Silently returning nothing
// here would turn a retrieval miss into a false "no coverage" answer,
// which is exactly the overclaiming failure mode Section 10 exists to
// prevent — so this layer is built to fail open toward "let the LLM see
// more", never toward "decide nothing matches" on its own authority.
//
// The zero-signal case is not hypothetical: this package's own keyword
// matching is ASCII-letters-and-digits only (see tokenize), so a
// description with no Latin-alphabet or digit substrings anywhere — a
// short sentence in Malayalam, Tamil, Hindi, etc. native script with no
// English loanword, proper noun, or number mixed in — tokenizes to
// nothing, every package scores 0, and "top N by score" stops being a
// ranking. This is expected to be rarer than the code-mixed case (Indian
// speech commonly mixes in English medical/technical terms even
// mid-sentence, which tokenizes and scores normally), but it is
// reachable, not a contrived edge case — a short, simple description can
// plausibly contain zero such tokens.
func Retrieve(ds *hbp.Dataset, description string) []Candidate {
	tokens := tokenize(description)
	tokenSet := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		tokenSet[t] = true
	}

	scored := make([]Candidate, 0, len(ds.Packages))
	for _, p := range ds.Packages {
		score := scorePackage(p, tokenSet)
		scored = append(scored, Candidate{Package: p, Score: score})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	hasPositive := len(scored) > 0 && scored[0].Score > 0

	var result []Candidate
	if hasPositive {
		for _, c := range scored {
			if c.Score <= 0 {
				break
			}
			result = append(result, c)
			if len(result) >= MaxCandidates {
				break
			}
		}
	} else {
		// Fail open: nothing scored, so hand the LLM a bounded slice of
		// everything rather than nothing. See doc comments above.
		limit := len(scored)
		if limit > MaxCandidatesWhenBlind {
			limit = MaxCandidatesWhenBlind
		}
		result = scored[:limit]
	}

	// Always guarantee the discretionary catch-all is visible to the LLM,
	// since by definition it won't score well on keyword overlap — that's
	// exactly the case it exists for (spec Section 8, point 2).
	result = ensureUnspecifiedIncluded(result, ds)

	return result
}

// ExclusionCandidate mirrors Candidate for the exclusion reference list.
type ExclusionCandidate struct {
	Exclusion hbp.Exclusion
	Score     int
}

// RetrieveExclusions scores every confirmed exclusion category against the
// family's description. Unlike Retrieve, this deliberately does NOT fail
// open to "return everything" when nothing scores — the exclusion list is
// short enough (a handful of categories) that the LLM step is always given
// the complete list regardless of pre-filter score, so a real exclusion
// can never be filtered out before the model gets a chance to reason about
// it. The score is kept only to sort the list, most-plausible-first, as a
// small readability aid to whoever reviews the model's reasoning later.
func RetrieveExclusions(ds *hbp.Dataset, description string) []ExclusionCandidate {
	tokens := tokenize(description)
	tokenSet := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		tokenSet[t] = true
	}

	scored := make([]ExclusionCandidate, 0, len(ds.Exclusions))
	for _, e := range ds.Exclusions {
		score := scoreExclusion(e, tokenSet)
		scored = append(scored, ExclusionCandidate{Exclusion: e, Score: score})
	}
	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})
	return scored
}

func scoreExclusion(e hbp.Exclusion, tokenSet map[string]bool) int {
	score := 0
	for _, kw := range e.Keywords {
		kwTokens := tokenize(kw)
		if len(kwTokens) == 0 {
			continue
		}
		allPresent := true
		for _, kt := range kwTokens {
			if !tokenSet[kt] {
				allPresent = false
				break
			}
		}
		if allPresent {
			score += 3 * len(kwTokens)
		}
	}
	for _, t := range tokenize(e.DisplayName) {
		if tokenSet[t] && !isStopword(t) {
			score++
		}
	}
	return score
}

func ensureUnspecifiedIncluded(candidates []Candidate, ds *hbp.Dataset) []Candidate {
	for _, c := range candidates {
		if c.Package.PackageCode == "UNSPECIFIED" {
			return candidates
		}
	}
	for _, p := range ds.Packages {
		if p.PackageCode == "UNSPECIFIED" {
			return append(candidates, Candidate{Package: p, Score: 0})
		}
	}
	return candidates
}

// scorePackage counts token overlap between the description and the
// package's name, specialty, and curated keywords. Keyword hits are
// weighted higher than incidental name/specialty word overlap because
// they were curated specifically to represent how a family actually
// talks (spec Section 15.1).
func scorePackage(p hbp.Package, tokenSet map[string]bool) int {
	score := 0

	for _, kw := range p.CommonDescriptionKeywords {
		kwTokens := tokenize(kw)
		if len(kwTokens) == 0 {
			continue
		}
		allPresent := true
		for _, kt := range kwTokens {
			if !tokenSet[kt] {
				allPresent = false
				break
			}
		}
		if allPresent {
			score += 3 * len(kwTokens) // reward multi-word phrase matches more
		}
	}

	for _, t := range tokenize(p.PackageName) {
		if tokenSet[t] && !isStopword(t) {
			score++
		}
	}
	for _, t := range tokenize(p.Specialty) {
		if tokenSet[t] && !isStopword(t) {
			score++
		}
	}

	return score
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	out := fields[:0]
	for _, f := range fields {
		if !isStopword(f) {
			out = append(out, f)
		}
	}
	return out
}

var stopwords = map[string]bool{
	"the": true, "a": true, "an": true, "is": true, "are": true, "was": true,
	"for": true, "to": true, "of": true, "and": true, "or": true, "in": true,
	"on": true, "with": true, "at": true, "it": true, "he": true, "she": true,
	"his": true, "her": true, "my": true, "our": true, "we": true, "us": true,
	"has": true, "have": true, "had": true, "this": true, "that": true,
}

func isStopword(s string) bool { return stopwords[s] }
