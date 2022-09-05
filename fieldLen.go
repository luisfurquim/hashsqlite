package hashsqlite

func fieldLen(fld []field) int {
	var f field
	var n int

	for _, f = range fld {
		if f.joinList {
			continue
		}
		n++
	}

	return n
}

