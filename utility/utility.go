package utility

import (
	"encoding/base64"
	"crypto/rand"
	"os"
	"fmt"
	"io"
	"os/user"
	"path/filepath"
)

func CreateTempFile() io.Writer {
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	tmpDir := filepath.Join(usr.HomeDir,".nafue")
	err = os.MkdirAll(tmpDir, os.ModeDir)
	if err != nil {
		fmt.Println("Cannot create temp directory: ", err.Error())
		os.Exit(0)
	}

	// random file
	ran, err := GenerateRandomString(32)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	w, err := os.Create(filepath.Join(tmpDir, ran + ".enn"))
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	return w
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	if err != nil {
		return "", err
	}
	code := base64.URLEncoding.EncodeToString(b)
	return code[0:s], nil
}
