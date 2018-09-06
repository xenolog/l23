package utils

type AppConfig struct {
	Debug  bool
	DryRun bool
	NsPath string
}

// Search first occurance of str into string slice
// return item index or -1 if not found
func IndexString(slice []string, str string) int {
	for i, a := range slice {
		if a == str {
			return i
		}
	}
	return -1
}
