package joker

type comboCache struct {
	l      int
	r      int
	result [][]int
}

var (
	cache = []comboCache{}
)

func combinations(l int, r int) [][]int {
	for _, c := range cache {
		if c.l == l && c.r == r {
			return c.result
		}
	}

	results := [][]int{}

	pool := []int{}
	for i := 0; i < l; i++ {
		pool = append(pool, i)
	}
	n := len(pool)

	if r > n {
		return results
	}

	indices := make([]int, r)
	for i := range indices {
		indices[i] = i
	}

	result := make([]int, r)
	for i, el := range indices {
		result[i] = pool[el]
	}

	resultCopy := []int{}
	for _, el := range result {
		resultCopy = append(resultCopy, el)
	}
	results = append(results, resultCopy)

	for {
		i := r - 1
		for ; i >= 0 && indices[i] == i+n-r; i-- {
		}

		if i < 0 {
			break
		}

		indices[i]++
		for j := i + 1; j < r; j++ {
			indices[j] = indices[j-1] + 1
		}

		for ; i < len(indices); i++ {
			result[i] = pool[indices[i]]
		}

		resultCopy2 := []int{}
		for _, el := range result {
			resultCopy2 = append(resultCopy2, el)
		}
		results = append(results, resultCopy2)
	}

	cache = append(cache, comboCache{l, r, results})
	return results
}
