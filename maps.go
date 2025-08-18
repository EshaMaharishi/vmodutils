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
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	}

	return 0, false
}

func GetInt64FromMap(m map[string]interface{}, n string) (int64, bool) {
	if m == nil {
		return 0, false
	}

	i, ok := m[n]
	if !ok {
		return 0, false
	}

	switch x := i.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case float64:
		return int64(x), true
	}

	return 0, false
}

func GetFloat64FromMap(m map[string]interface{}, n string) (float64, bool) {
	if m == nil {
		return 0, false
	}

	i, ok := m[n]
	if !ok {
		return 0, false
	}

	switch x := i.(type) {
	case int64:
		return float64(x), true
	case int:
		return float64(x), true
	case float64:
		return x, true
	}

	return 0, false
}
