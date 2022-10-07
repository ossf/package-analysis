package stringentropy

import (
	"math"
)

/*
CalculateEntropy

Calculates entropy of a string S of characters over an alphabet A, which is defined as

	E(S) = - sum(i in A) { (p(i)) * log(p(i)) },

where p(i) is the probability of observing character i, and the summation is performed over all characters in A.
If S is the empty string, we define E(S) to be 0.

The probabilities p(i) can be given a priori, or simply calculated by counting characters within the string S.
In the latter case, we have p(i) = c(i) / |S|, where c(i) counts the number of times character i appears in S,
and |S| is the length of S. Then,

	E(S) = - sum(i in A) { (c(i) / |S|) * log(c(i) / |S|) }.

In this case, the maximum value for E is log(|S|). When the number of distinct characters in S is small,
the entropy approaches 0.

Reference: https://link.springer.com/chapter/10.1007/978-3-642-10509-8_19
*/
func CalculateEntropy(s string, prob *map[rune]float64) float64 {
	if len(s) == 0 {
		return 0
	}

	counts := CharacterCounts([]string{s})

	if prob == nil {
		prob = CharacterProbabilitiesFromCounts(counts)
	}

	entropy := 0.0

	// can't iterate over string as we need to iterate over distinct chars
	for char := range *counts {
		p := (*prob)[char]
		entropy -= p * math.Log(p)
	}

	return entropy
}

/*
CalculateNormalisedEntropy
Returns string entropy normalised by the log of the length of the string. This quantity is used
because for log(N) is the maximum possible entropy out of all strings with length N, where N > 0.
If the string has one or fewer characters, the ratio is defined to be 1.
As a formula:

	E_n(S) := {
	    0,               if |S| = 0
	    1,               if |S| = 1
	    E(S) / log(|S|), otherwise
	}
*/
// TODO does this make sense when a general probability structure is used?
// TODO calculate max string entropy for a given set of character counts
func CalculateNormalisedEntropy(s string, prob *map[rune]float64) float64 {
	switch len(s) {
	case 0:
		return 0
	case 1:
		return 1
	default:
		return CalculateEntropy(s, prob) / math.Log(float64(len(s)))
	}
}

func CharacterCounts(strs []string) *map[rune]int {
	counts := make(map[rune]int)

	for _, s := range strs {
		for _, b := range s {
			// if b is not in map, the read value will be zero, which is what we want
			counts[b] = counts[b] + 1
		}
	}
	return &counts
}

func CharacterProbabilitiesFromCounts(counts *map[rune]int) *map[rune]float64 {
	total := 0
	for _, count := range *counts {
		total += count
	}

	prob := make(map[rune]float64, len(*counts))
	for char, count := range *counts {
		prob[char] = float64(count) / float64(total)
	}
	return &prob
}

// CharacterProbabilities
// Just a convenience function that chains together the other fwo function calls.
func CharacterProbabilities(strs []string) *map[rune]float64 {
	counts := CharacterCounts(strs)
	return CharacterProbabilitiesFromCounts(counts)
}
