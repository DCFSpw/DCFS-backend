package util

// SliceContains - check whether a slice contains a value
//
// params:
//   - slice []comparable - slice of elements implementing the comparable interface
//   - element comparable - element to check for
//
// return type:
//   - bool - true if the slice contains the element, false otherwise
func SliceContains[K comparable](slice []K, element K) bool {
	for _, element2 := range slice {
		if element == element2 {
			return true
		}
	}

	return false
}
