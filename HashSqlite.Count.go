package hashsqlite

func (hs *HashSqlite) Count(table interface{}) (int64, error) {
	var count int64
	var tabName string
//	var refRow reflect.Value
	var err error
	var ok bool

	switch tab := table.(type) {
	case string:
		if tabName, ok = hs.tableType[tab]; !ok {
			for _, tabName = range hs.tableType {
				if tabName == tab {
					ok = true
					break
				}
			}
			if !ok {
				Goose.Query.Logf(1, "Parameter type error: %s", ErrNoTablesFound)
				return 0, ErrNoTablesFound
			}
		}

	default:
		tabName, _, err = hs.getSingleRefs(tab)
		if err != nil {
			Goose.Query.Logf(1, "Parameter type error: %s", ErrNoTablesFound)
			return 0, ErrNoTablesFound
		}
	}

	_, err = hs.count[tabName].SelectOneRow(&count)
	if err != nil {
		Goose.Query.Logf(1, "Count error on %s: %s", tabName, err)
		return 0, err
	}

	return count, nil
}
