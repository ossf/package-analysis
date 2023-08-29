package stringentropy

import (
	"math"
	"unicode/utf8"
)

/*
Calculate finds the entropy of a string S of characters over an alphabet A, which is defined as

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
func Calculate(s string, prob map[rune]float64) float64 {
	if len(s) == 0 {
		return 0
	}

	if prob == nil {
		counts, sumCounts := CharacterCounts([]string{s})
		prob = characterProbabilitiesFromCounts(counts, sumCounts)
	}

	entropy := 0.0
	for _, char := range s {
		p := prob[char]
		if p > 0 {
			entropy -= p * math.Log(p)
		}
	}

	return entropy
}

/*
CalculateNormalised returns the string entropy normalised by the log of the length of the string.
This quantity is used because for log(N) is the maximum possible entropy out of all strings with length N,
where N > 0. Special cases are empty strings (0) and single character strings (1).
As a formula:

	E_n(S) := {
	    0,               if |S| = 0
	    1,               if |S| = 1
	    E(S) / log(|S|), otherwise
	}
*/
// TODO does this make sense when a general probability structure is used?
// TODO calculate max string entropy for a given set of character counts.
func CalculateNormalised(s string, prob map[rune]float64) float64 {
	length := utf8.RuneCountInString(s)
	switch length {
	case 0:
		return 0
	case 1:
		return 1
	default:
		return Calculate(s, prob) / math.Log(float64(length))
	}
}

// CharacterCounts computes a map of character (rune) to number of occurrences
// in the input strings
func CharacterCounts(strs []string) (map[rune]int, int64) {
	counts := make(map[rune]int)
	var sumCounts int64 = 0
	for _, s := range strs {
		for _, b := range s {
			counts[b] += 1
			sumCounts += 1
		}
	}
	return counts, sumCounts
}

// CharacterProbabilities computes a map of character (rune) to
// frequency/probability of occurrence in the input strings
func CharacterProbabilities(strs []string) map[rune]float64 {
	counts, sumCounts := CharacterCounts(strs)
	return characterProbabilitiesFromCounts(counts, sumCounts)
}

func characterProbabilitiesFromCounts(counts map[rune]int, sumCounts int64) map[rune]float64 {
	prob := make(map[rune]float64, len(counts))
	for char, count := range counts {
		prob[char] = float64(count) / float64(sumCounts)
	}
	return prob
}
