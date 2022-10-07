import sys
from typing import Sequence
from math import nan


def mean(nums: Sequence[float]) -> float:
    n = len(nums)
    if n < 1:
        return nan
    return sum(nums) / n


def variance(nums: Sequence[float], mean: float, bias_correct: bool=True) -> float:
    n = len(nums)
    if n < 2:
        return nan
    else:
        return sum((x - mean)**2 for x in nums) / (n - 1 if bias_correct else n)


# G1 from https://en.wikipedia.org/wiki/Skewness#Sample_skewness
def skewness(nums: Sequence[float], mean: float, variance: float) -> float:
    n = len(nums)
    if n < 3:
        return nan
    else:
        return sum((x - mean)**3 for x in nums) * n / ((n-1)*(n-2)*variance**1.5)


def main():
    nums = list(map(float, sys.argv[1:]))
    m = mean(nums)
    v = variance(nums, m)
    s = skewness(nums, m, v)
    print(f"mean: {m}")
    print(f"variance: {v}")
    print(f"skewness: {s}")


if __name__ == "__main__":
    main()

