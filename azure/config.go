package azure

import (
	"fmt"
	"os"
)

type Config struct {
	TenantID     string
	AppID        string
	SecretID     string
	Subscription string
	ResourceName string
	AKSName      string
}

func NewConfig() (*Config, error) {
	required := map[string]string{
		"AZURE_TENANTID":     "",
		"AZURE_APPID":        "",
		"AZURE_SECRETID":     "",
		"AZURE_SUBSCRIPTION": "",
		"AKS_RESOURCE_NAME":  "",
		"AKS_NAME":           "",
	}

	for key := range required {
		val := os.Getenv(key)
		if val == "" {
			return nil, fmt.Errorf("missing required environment variable: %s", key)
		}
		required[key] = val
	}

	return &Config{
		TenantID:     required["AZURE_TENANTID"],
		AppID:        required["AZURE_APPID"],
		SecretID:     required["AZURE_SECRETID"],
		Subscription: required["AZURE_SUBSCRIPTION"],
		ResourceName: required["AKS_RESOURCE_NAME"],
		AKSName:      required["AKS_NAME"],
	}, nil
}
