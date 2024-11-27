package raft_db

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func GetMostCommonValue(arr []int) int {
	values := make(map[int]int, 0)
	max_freq := 0

	for _, a := range arr {
		values[a] += 1
		max_freq = max(max_freq, values[a])
	}

	for key, freq := range values {
		if freq == max_freq {
			return key
		}
	}
	return 0
}

