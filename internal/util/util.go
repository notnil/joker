package util

// Combinations returns the combinations of n and k, explained
// in http://en.wikipedia.org/wiki/Combination, as a two dimensional
// slice of indexes.  If n or k are negative or k > n the return value
// will be empty.
func Combinations(n, k int) [][]int {
	results := [][]int{}

	if n <= 0 || k <= 0 || k > n {
		return results
	}

	pool := indexRange(n)
	indices := indexRange(k)
	result := indexRange(k)
	results = append(results, indexRange(k))

	for {
		i := k - 1
		for ; i >= 0 && indices[i] == i+len(pool)-k; i-- {
		}

		if i < 0 {
			break
		}

		indices[i]++
		for j := i + 1; j < k; j++ {
			indices[j] = indices[j-1] + 1
		}

		for ; i < len(indices); i++ {
			result[i] = pool[indices[i]]
		}

		resultCopy := make([]int, len(result))
		copy(resultCopy, result)
		results = append(results, resultCopy)
	}

	return results
}

func indexRange(n int) []int {
	r := []int{}
	for i := 0; i < n; i++ {
		r = append(r, i)
	}
	return r
}
