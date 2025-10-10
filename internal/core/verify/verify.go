package verify

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func VerifyLevel1(path, upstreamHash string) (bool, *string, error) {
	hash, err := GetSha256(path)
	if err != nil {
		return false, nil, err
	}

	hash = fmt.Sprintf("sha256:%s", hash)
	return hash == upstreamHash, &hash, nil
}

func GetSha256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
