package utils

type AppConfig struct {
	Debug  bool
	DryRun bool
	NsPath string
}

// IndexString -- Search first occurance of str into string slice
// returns item index or -1 if not found
func IndexString(slice []string, str string) int {
	for i, a := range slice {
		if a == str {
			return i
		}
	}
	return -1
}

// PrependString -- Prepend string `str` to a given string slice
// returns new string slice
func PrependString(slice []string, str string) []string {
	return append([]string{str}, slice...)
}

// ReverseString -- Return reversed string slice
func ReverseString(a []string) []string {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
	return a
}
