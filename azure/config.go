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
	// Variables obligatoires
	required := map[string]string{
		"AKS_RESOURCE_NAME": "",
		"AKS_NAME":          "",
	}

	// Variables optionnelles pour SPN
	optional := map[string]string{
		"AZURE_TENANTID":     "",
		"AZURE_APPID":        "",
		"AZURE_SECRETID":     "",
		"AZURE_SUBSCRIPTION": "",
	}

	// Vérifier les variables obligatoires
	for key := range required {
		val := os.Getenv(key)
		if val == "" {
			return nil, fmt.Errorf("missing required environment variable: %s", key)
		}
		required[key] = val
	}

	// Récupérer les variables optionnelles
	for key := range optional {
		optional[key] = os.Getenv(key)
	}

	return &Config{
		TenantID:     optional["AZURE_TENANTID"],
		AppID:        optional["AZURE_APPID"],
		SecretID:     optional["AZURE_SECRETID"],
		Subscription: optional["AZURE_SUBSCRIPTION"],
		ResourceName: required["AKS_RESOURCE_NAME"],
		AKSName:      required["AKS_NAME"],
	}, nil
}

// HasSPNConfig vérifie si la configuration SPN est complète
func (c *Config) HasSPNConfig() bool {
	return c.TenantID != "" && c.AppID != "" && c.SecretID != ""
}