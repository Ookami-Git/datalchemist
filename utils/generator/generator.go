package generator

import (
	"crypto/rand"
	"encoding/base64"
)


func String(n int) (string) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return ""
    }
    return base64.URLEncoding.EncodeToString(b)[:n]
}