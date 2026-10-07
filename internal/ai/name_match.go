package ai

import (
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Matching a name the way a person remembers it.
//
// Exact, prefix and contains matches find a name typed right. A name
// remembered is often not: its vowels dropped ("hyto" for "hayato"), letters
// left out ("ordrs"), or a letter wrong. These find those, and rank below
// anything spelled right, so they never push a real match down.

// soundsLikeMinimum is the shortest search these apply to: two letters would
// sound like half the schema.
const soundsLikeMinimum = 3

// Why a name matched, for saying so.
const (
	matchExact      = "exact"
	matchPrefix     = "starts with it"
	matchContains   = "contains it"
	matchSoundsLike = "sounds like it"
	matchLetters    = "has its letters, in order"
	matchSpelling   = "spelled close to it"
)

// nameMatch scores how well a name matches a search, 0 for not at all, and
// says why. The scores follow the table search's: 100 exact, 90 prefix, 70
// contains, and below those the near-misses.
func nameMatch(query, name string) (float64, string) {
	q, n := strings.ToLower(strings.TrimSpace(query)), strings.ToLower(name)
	switch {
	case q == "":
		return 0, ""
	case n == q:
		return 100, matchExact
	case strings.HasPrefix(n, q):
		return 90, matchPrefix
	case strings.Contains(n, q):
		return 70, matchContains
	}
	return soundsLike(q, n)
}

// soundsLike is the near-miss part of nameMatch, for a search and a name
// already lower-cased that did not match as spelled.
func soundsLike(q, n string) (float64, string) {
	if len([]rune(q)) < soundsLikeMinimum {
		return 0, ""
	}
	qs, ns := skeleton(q), skeleton(n)
	switch {
	case qs != "" && qs == ns:
		return 60, matchSoundsLike
	case len(qs) >= soundsLikeMinimum && strings.HasPrefix(ns, qs):
		return 50, matchSoundsLike
	case fuzzy.Match(q, n) && len([]rune(n)) <= 2*len([]rune(q))+2:
		// Letters in order, in a name not much longer than the search: in a
		// long name, a short search's letters are nearly always somewhere.
		return 40, matchLetters
	}
	if d := fuzzy.LevenshteinDistance(q, n); d <= allowedTypos(q) {
		return 35 - 5*float64(d), matchSpelling
	}
	return 0, ""
}

// skeleton is a name's first letter and the consonants after it, in order,
// with no letter twice running and no digits or underscores: what is left of a
// name remembered by its sound. "hayato", "hayato2" and "hyto" are all "hyt";
// "orders" and "ordrs" are both "ordrs". Y counts as a consonant: as a vowel,
// "hyto" would be "ht", and sound like "hat", "hot" and "hit" too.
func skeleton(s string) string {
	var b strings.Builder
	var last rune
	for _, r := range s {
		if r < 'a' || r > 'z' || (strings.ContainsRune("aeiou", r) && b.Len() > 0) {
			continue
		}
		if r == last {
			continue
		}
		b.WriteRune(r)
		last = r
	}
	return b.String()
}

// allowedTypos is how many letters may be wrong for a name of this length to
// still match: one in a short search, two in a long one.
func allowedTypos(q string) int {
	if len([]rune(q)) >= 8 {
		return 2
	}
	return 1
}
