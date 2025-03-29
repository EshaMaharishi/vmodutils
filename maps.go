package vmodutils

func GetIntFromMap(m map[string]interface{}, n string) (int, bool) {
	if m == nil {
		return 0, false
	}

	i, ok := m[n]
	if !ok {
		return 0, false
	}

	switch x := i.(type) {
	case int:
		return x, true
	case float64:
		return int(x), true
	}

	return 0, false
}
