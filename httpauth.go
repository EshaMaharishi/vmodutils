package vmodutils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"go.viam.com/rdk/utils"
)

// nameType is BoatAuth or the like
func HTTPAuthHeader(nameType, machine, apiKeyId, apiKey string) string {
	hasher := sha256.New()
	hasher.Write([]byte(apiKey))
	hash := hasher.Sum(nil)
	return fmt.Sprintf("%s robot_id=\"%s\", api_key_id=\"%s\", api_key_hash=\"%s\"",
		nameType, machine, apiKeyId, hex.EncodeToString(hash))
}

func HTTPAuthHeaderFromEnv(nameType string) (string, error) {
	machine := os.Getenv(utils.MachineIDEnvVar)
	if machine == "" {
		return "", fmt.Errorf("need a machine")
	}

	apiKeyId := os.Getenv(utils.APIKeyIDEnvVar)
	if apiKeyId == "" {
		return "", fmt.Errorf("need an api key id")
	}

	apiKey := os.Getenv(utils.APIKeyEnvVar)
	if apiKey == "" {
		return "", fmt.Errorf("need an api key")
	}

	return HTTPAuthHeader(nameType, machine, apiKeyId, apiKey), nil
}
