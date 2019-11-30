package win32

// stringToCharPtr converts a Go string into pointer to a null-terminated cstring.
// This assumes the go string is already ANSI encoded.
// https://medium.com/jettech/breaking-all-the-rules-using-go-to-call-windows-api-2cbfd8c79724
func stringToCharPtr(str string) *uint8 {
	chars := append([]byte(str), 0) // null terminated
	return &chars[0]
}
