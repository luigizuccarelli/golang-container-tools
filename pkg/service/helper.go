package service

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/luigizuccarelli/golang-container-tools/pkg/schema"
)

// GetBasicAuthCredentials - simple basic auth helper function
func GetBasicAuthCredentials() (*schema.BasicAuth, error) {
	creds := os.Getenv("BASIC_AUTH_CREDENTIALS")
	base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(creds)))
	base64.StdEncoding.Decode(base64Text, []byte(creds))
	hld := strings.Split(string(base64Text), ":")
	user := strings.ReplaceAll(hld[0], "\n", "")
	pwd := strings.ReplaceAll(hld[1], "\n", "")
	ba := &schema.BasicAuth{User: user, Password: pwd}
	return ba, nil
}
