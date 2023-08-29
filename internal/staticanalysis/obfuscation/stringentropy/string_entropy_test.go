package stringentropy

import (
	"math"
	"testing"

	"github.com/ossf/package-analysis/internal/utils"
)

type entropyTestCase struct {
	s        string
	expected float64
}

func TestStringEntropy(t *testing.T) {
	tolerance := 1e-6
	testCases := []entropyTestCase{
		{"", 0},
		{"a", 0},
		{"abc", math.Log(3)},
		{"abcdefghij", math.Log(10)},
		{"aaa", 0},
		{"aA", math.Log(2)},
		{"aaA", -2*2.0/3.0*math.Log(2.0/3.0) - math.Log(1.0/3.0)/3.0},
	}
	for index, test := range testCases {
		actual := Calculate(test.s, nil)
		if !utils.FloatEquals(test.expected, actual, tolerance) {
			t.Errorf("Test case %d failed (str: %s, expected: %f, actual: %f\n",
				index+1, test.s, test.expected, actual)
		}
	}
}

func TestStringEntropyWithFixedProbs(t *testing.T) {
	tolerance := 1e-6
	a := 0.125
	b := 0.25
	c := 0.125
	d := 0.5
	probs := map[rune]float64{
		'a': 0.125,
		'b': 0.25,
		'c': 0.125,
		'd': 0.5,
	}
	testCases := []entropyTestCase{
		{"", 0},
		{"a", -a * math.Log(a)},
		{"abc", -a*math.Log(a) - b*math.Log(b) - c*math.Log(c)},
		{"abcdefg", -a*math.Log(a) - b*math.Log(b) - c*math.Log(c) - d*math.Log(d)},
		{"aaaAAA", -3 * a * math.Log(a)},
		{" a \n", -a * math.Log(a)},
	}
	for index, test := range testCases {
		actual := Calculate(test.s, probs)
		if !utils.FloatEquals(test.expected, actual, tolerance) {
			t.Errorf("Test case %d failed (str: %s, expected: %f, actual: %f\n",
				index+1, test.s, test.expected, actual)
		}
	}
}

func TestStringEntropyRatio(t *testing.T) {
	tolerance := 1e-9
	testCases := []entropyTestCase{
		{"", 0},
		{"a", 1},
		{"abc", 1},
		{"abcdefghij", 1},
		{"aaa", 0},
		{"aA", 1},
		{"aaA", (-2*2.0*math.Log(2.0/3.0) - math.Log(1.0/3.0)) / (3.0 * math.Log(3))},
	}
	for index, test := range testCases {
		actual := CalculateNormalised(test.s, nil)
		if !utils.FloatEquals(test.expected, actual, tolerance) {
			t.Errorf("Test case %d failed (str: %s, expected: %f, actual: %f\n",
				index+1, test.s, test.expected, actual)
		}
	}
}

func runeOf(char string) rune {
	return []rune(char)[0]
}

func TestCharacterProbabilities(t *testing.T) {
	tolerance := 1e-6
	str := "hello there"
	str2 := "lady emma"
	countsExpected := map[rune]int{
		runeOf("h"): 2,
		runeOf("e"): 4,
		runeOf("l"): 3,
		runeOf("o"): 1,
		runeOf(" "): 2,
		runeOf("t"): 1,
		runeOf("r"): 1,
		runeOf("a"): 2,
		runeOf("d"): 1,
		runeOf("m"): 2,
	}
	countsActual, _ := CharacterCounts([]string{str, str2})
	for char, expectedCount := range countsExpected {
		actualCount := countsActual[char]
		if expectedCount != actualCount {
			t.Errorf("Incorrect count for character '%s' (%d): expected %d, actual %d",
				string(char), char, expectedCount, actualCount)
		}
	}

	probsExpected := map[rune]float64{
		runeOf("h"): 2.0 / 20.0,
		runeOf("e"): 4.0 / 20.0,
		runeOf("l"): 3.0 / 20.0,
		runeOf("o"): 1.0 / 20.0,
		runeOf(" "): 2.0 / 20.0,
		runeOf("t"): 1.0 / 20.0,
		runeOf("r"): 1.0 / 20.0,
		runeOf("a"): 2.0 / 20.0,
		runeOf("d"): 1.0 / 20.0,
		runeOf("m"): 2.0 / 20.0,
	}

	probsActual := CharacterProbabilities([]string{str, str2})

	for char, expectedProb := range probsExpected {
		actualProb := probsActual[char]
		if !utils.FloatEquals(expectedProb, actualProb, tolerance) {
			t.Errorf("Incorrect prob for character '%s' (%d): expected %.2f, actual %.2f",
				string(char), char, expectedProb, actualProb)
		}
	}
}
