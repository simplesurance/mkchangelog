package parser

func dedupStringSlice(slice []string) []string {
	var result []string
	dedupMap := map[string]struct{}{}

	for _, elem := range slice {
		if _, exist := dedupMap[elem]; exist {
			continue
		}

		dedupMap[elem] = struct{}{}
		result = append(result, elem)
	}

	return result
}
