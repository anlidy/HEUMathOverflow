package utils

func SliceFilter[T comparable](src []T, filter []T) []T {
	filterSet := make(map[T]struct{}, len(filter))
	for _, v := range filter {
		filterSet[v] = struct{}{}
	}
	res := make([]T, len(src))
	for _, item := range src {
		if _, ok := filterSet[item]; !ok {
			res = append(res, item)
		}
	}
	return res
}
