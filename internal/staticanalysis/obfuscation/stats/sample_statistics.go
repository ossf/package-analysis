package stats

import (
	"fmt"
	"math"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"

	"github.com/ossf/package-analysis/internal/utils"
)

type RealNumber interface {
	constraints.Integer | constraints.Float
}

type SampleStatistics struct {
	Size      int
	Mean      float64
	Variance  float64
	Skewness  float64
	Quartiles [5]float64
}

func NoData() SampleStatistics {
	nan := math.NaN()
	return SampleStatistics{
		Size:      0,
		Mean:      nan,
		Variance:  nan,
		Skewness:  nan,
		Quartiles: [5]float64{nan, nan, nan, nan, nan},
	}
}

func (s SampleStatistics) Min() float64    { return s.Quartiles[0] }
func (s SampleStatistics) Q1() float64     { return s.Quartiles[1] }
func (s SampleStatistics) Median() float64 { return s.Quartiles[2] }
func (s SampleStatistics) Q3() float64     { return s.Quartiles[3] }
func (s SampleStatistics) Max() float64    { return s.Quartiles[4] }

func (s SampleStatistics) toFloatData() (data [8]float64) {
	data[0] = s.Mean
	data[1] = s.Variance
	data[2] = s.Skewness
	for i := 0; i < 5; i++ {
		data[i+3] = s.Quartiles[i]
	}
	return
}

func fromFloatData(size int, data [8]float64) SampleStatistics {
	return SampleStatistics{
		Size:      size,
		Mean:      data[0],
		Variance:  data[1],
		Skewness:  data[2],
		Quartiles: [5]float64{data[3], data[4], data[5], data[6], data[7]},
	}
}

func (s SampleStatistics) String() string {
	q := s.Quartiles
	return fmt.Sprintf("size: %d, mean: %f, variance: %f, skewness: %f, quartiles: [%f, %f, %f, %f, %f]",
		s.Size, s.Mean, s.Variance, s.Skewness, q[0], q[1], q[2], q[3], q[4])
}

func (s SampleStatistics) Equals(other SampleStatistics, absTol float64) bool {
	if s.Size != other.Size {
		return false
	}
	thisData := s.toFloatData()
	otherData := other.toFloatData()
	for i := 0; i < len(thisData); i++ {
		if !utils.FloatEquals(thisData[i], otherData[i], absTol) {
			return false
		}
	}
	return true
}

// ReplaceNaNs returns a copy of this SampleStatistics object with all NaN values
// replaced by the given value r.
func (s SampleStatistics) ReplaceNaNs(r float64) SampleStatistics {
	data := s.toFloatData()
	for i, x := range data {
		if math.IsNaN(x) {
			data[i] = r
		}
	}
	return fromFloatData(s.Size, data)
}

// mean computes the sample mean.
func mean[T RealNumber](sample []T) float64 {
	if len(sample) < 1 {
		return math.NaN()
	} else {
		sum := 0.0
		for _, x := range sample {
			sum += float64(x)
		}
		return sum / float64(len(sample))
	}
}

// variance calculates sample variance with bias correction
// mean is the value returned by the mean() function.
func variance[T RealNumber](sample []T, mean float64) float64 {
	if len(sample) < 2 {
		return math.NaN()
	} else {
		n := float64(len(sample))
		sumSquares := 0.0
		for _, x := range sample {
			d := float64(x) - mean
			sumSquares += d * d
		}
		return sumSquares / (n - 1)
	}
}

func squareRootCubed(x float64) float64 {
	y := math.Sqrt(x)
	return y * y * y
}

// skewness calculates sample skewness using the G1 estimator from
// https://en.wikipedia.org/wiki/Skewness#Sample_skewness
// mean and variance respectively are the values returned by the
// mean() and (bias-corrected) variance() functions.
func skewness[T RealNumber](sample []T, mean, variance float64) float64 {
	if len(sample) < 3 {
		return math.NaN()
	} else {
		// G1 = n^2/((n-1)(n-2) b1,
		// where b1 = 1/n * sum of cubed deviations / variance^3/2
		n := float64(len(sample))
		sumCubes := 0.0
		for _, x := range sample {
			d := float64(x) - mean
			sumCubes += d * d * d
		}
		return sumCubes * n / (n - 1) / (n - 2) / squareRootCubed(variance)
	}
}

// quartile calculates sample quartiles of a dataset, which is assumed to be sorted.
// quartile zero is defined as the minimum and quartile 4 is defined as the maximum.
//
// Implements 'type 2' calculation from https://en.wikipedia.org/wiki/Quartile,
// which is also the type=2 version of R's quantile function:
// https://www.rdocumentation.org/packages/stats/versions/3.6.1/topics/quantile.
//
// This method was chosen as it is commonly used for discrete data
// (e.g. string lengths), and also because (I think) it is also
// equivalent to method 2 of https://en.wikipedia.org/wiki/Quartile.
// It has the nice property that the five quartiles of a sample [1, 2, 3, 4, 5]
// (including min and max) are exactly, 1, 2, 3, 4, 5.
func quartile[T RealNumber](sortedSample []T, whichQuartile int) float64 {
	if whichQuartile < 0 || whichQuartile > 4 {
		panic(fmt.Errorf("invalid quartile %d", whichQuartile))
	}
	n := len(sortedSample)
	if n == 0 {
		return math.NaN()
	}
	if whichQuartile == 0 {
		return float64(sortedSample[0])
	}
	if whichQuartile == 4 {
		return float64(sortedSample[n-1])
	}
	// here n >= 1; whichQuartile = 1, 2, 3

	// calculations from https://www.rdocumentation.org/packages/stats/versions/3.6.1/topics/quantile.
	j := n * whichQuartile / 4
	if n*whichQuartile%4 == 0 && j > 0 {
		// empirical CDF discontinuous at this point; average with prev. value if we can
		// (though it may be the same sample value)
		return float64(sortedSample[j-1]+sortedSample[j]) / 2.0
	}
	return float64(sortedSample[j])
}

func quartiles[T RealNumber](sample []T) [5]float64 {
	nan := math.NaN()
	result := [5]float64{nan, nan, nan, nan, nan}

	if len(sample) > 0 {
		sortedSample := make([]T, len(sample))
		copy(sortedSample, sample)
		slices.Sort(sortedSample)

		for i := 0; i <= 4; i++ {
			result[i] = quartile(sortedSample, i)
		}
	}

	return result
}

func Summarise[T RealNumber](sample []T) SampleStatistics {
	l := len(sample)
	m := mean(sample)
	v := variance(sample, m)
	s := skewness(sample, m, v)
	q := quartiles(sample)
	return SampleStatistics{Size: l, Mean: m, Variance: v, Skewness: s, Quartiles: q}
}
