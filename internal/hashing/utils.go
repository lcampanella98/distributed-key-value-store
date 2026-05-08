package hashing

func IsInRange(start, end, k uint64) bool {
	if start < end {
		return k >= start && k <= end
	}
	// end <= start
	return k >= start || k <= end
}
