package util

func SliceContains[K comparable](slice []K, element K) bool {
	for _, element2 := range slice {
		if element == element2 {
			return true
		}
	}

	return false
}
