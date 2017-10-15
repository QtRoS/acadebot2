package shared

// Min of two values.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max of two values.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
