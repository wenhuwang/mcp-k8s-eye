package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func IsErrorReason(reason string) bool {
	failureReasons := []string{
		"CrashLoopBackOff", "ImagePullBackOff", "CreateContainerConfigError", "PreCreateHookError", "CreateContainerError",
		"PreStartHookError", "RunContainerError", "ImageInspectError", "ErrImagePull", "ErrImageNeverPull", "InvalidImageName",
	}

	for _, r := range failureReasons {
		if r == reason {
			return true
		}
	}
	return false
}

func IsEvtErrorReason(reason string) bool {
	failureReasons := []string{
		"FailedCreatePodSandBox", "FailedMount",
	}

	for _, r := range failureReasons {
		if r == reason {
			return true
		}
	}
	return false
}

var anonymizePattern = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func MaskString(input string) string {
	key := make([]byte, len(input))
	result := make([]rune, len(input))
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	for i := range result {
		result[i] = anonymizePattern[int(key[i])%len(anonymizePattern)]
	}
	return base64.StdEncoding.EncodeToString([]byte(string(result)))
}
