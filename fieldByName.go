package hashsqlite

func fieldByName(nm string, fldList []field) (int, bool) {
	var i int

	for i=0; i<len(fldList); i++ {
		if nm == fldList[i].name {
			return fldList[i].index, true
		}
	}

	return -1, false
}
