package hashsqlite

func fieldJoin(fld []field) string {
	var i int
	var f field
	var s string

	for i, f = range fld {
		if i>0 {
			s += ","
		}
		s += f.name
	}

	return s
}

