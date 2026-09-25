package model

type Data = Map[string, any]

func (m Data) GetSlice(key string) (arr []any) {
	if !m.Has(key) {
		return
	}
	aa, ok := m.Get(key).([]any)
	if !ok {
		return
	}

	return aa
}

func (m Data) GetStrs(key string) (arr []string) {
	if !m.Has(key) {
		return
	}

	aa, ok := m.Get(key).([]any)
	if !ok {
		return
	}

	for _, a := range aa {
		if s, k := a.(string); k {
			arr = append(arr, s)
		}
	}
	return
}
