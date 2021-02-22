package dripfeed

import (
	"crypto/sha1"
	"fmt"
)

func UUID(userUUID string, alistUUID string) string {
	b := []byte(fmt.Sprintf("%s/%s", userUUID, alistUUID))
	hash := fmt.Sprintf("%x", sha1.Sum(b))
	return hash
}
