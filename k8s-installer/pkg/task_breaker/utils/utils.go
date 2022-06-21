package utils

import (
	"github.com/google/uuid"
)

func GenNodeStepID() string {
	return "NodeStep-" + uuid.New().String()
}
