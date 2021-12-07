package util

func Unique(ll ...[]string) []string {
	keys := make(map[string]bool)
	keys[""] = true // ignore empty string
	ret := []string{}
	for _, l := range ll {
		for _, item := range l {
			if _, value := keys[item]; !value {
				keys[item] = true
				ret = append(ret, item)
			}
		}
	}
	return ret
}
