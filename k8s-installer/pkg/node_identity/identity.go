package node_identity

import (
	"encoding/base64"
	"github.com/google/uuid"
)

func GeneratorNodeID(identity, prefix, suffix string) string {
	encoded := base64.URLEncoding.EncodeToString([]byte(identity))
	if prefix != "" {
		encoded = prefix + "-" + encoded
	}
	if suffix != "" {
		encoded = encoded + "-" + suffix
	}
	return encoded
}

func GeneratorNodeIDWithUUID(prefix, suffix string) string {
	id := uuid.New().String()
	if prefix != "" {
		id = prefix + "-" + id
	}
	if suffix != "" {
		id = id + "-" + suffix
	}
	return id
}
