package strutil

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"text/template"
)

var (
	// escape string.
	EscapeJS   = template.JSEscapeString
	EscapeHTML = template.HTMLEscapeString
)

// Md5 Generate a 32-bit md5 string
func Md5(src interface{}) string {
	return GenMd5(src)
}

// GenMd5 Generate a 32-bit md5 string
func GenMd5(src interface{}) string {
	h := md5.New()

	if s, ok := src.(string); ok {
		h.Write([]byte(s))
	} else {
		h.Write([]byte(fmt.Sprint(src)))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// Base64Encode base64 encode
func Base64Encode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}