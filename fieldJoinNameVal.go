package hashsqlite

func fieldJoinNameVal(fld []field) string {
	var f field
	var s string

	for _, f = range fld {
		if f.joinList {
			continue
		}
		if len(s)>0 {
			s += ","
		}
		s += "`" + f.name + "`=?"
	}

	return s
}

