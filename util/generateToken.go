package util

import (
	"os/exec"
)

func GenerateToken() string {
	// gcloud auth print-access-token
	cmdStruct := exec.Command("gcloud", "auth", "print-access-token")
	cmdOut, err := cmdStruct.Output()
	if err != nil {
		panic(err)
	}

	return string(cmdOut)
}
