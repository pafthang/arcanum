package routing

// MergeRoutes merges multiple route lists.
// On ID conflict the last entry wins.
func MergeRoutes(lists ...[]*Route) []*Route {
	byID := make(map[string]*Route)
	var order []string

	for _, list := range lists {
		for _, r := range list {
			if r == nil || r.ID == "" {
				continue
			}
			if _, exists := byID[r.ID]; !exists {
				order = append(order, r.ID)
			}
			byID[r.ID] = r
		}
	}

	out := make([]*Route, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

// MergeTables merges multiple routing tables.
func MergeTables(tables ...*Table) *Table {
	t := NewTable()
	for _, src := range tables {
		if src == nil {
			continue
		}
		for _, r := range src.List() {
			t.Add(r)
		}
	}
	return t
}
