package utils

import "os"

func GetConfigPath() string {
	return os.Getenv("GOPATH") + "/src/github.com/matba/slyde-server/configs/"
}
