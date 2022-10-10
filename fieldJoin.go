package hashsqlite

func fieldJoin(fld []field) (string, []int) {
	var f field
	var s string
	var i []int

	for _, f = range fld {
		if f.joinList {
			continue
		}
		if len(s)>0 {
			s += ","
		}
		s += "`" + f.name + "`"
		i = append(i, f.index)
	}

	return s, i
}

