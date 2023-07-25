package pkg

import (
	v1 "harrisonwaffel/ccg-gmsa-dll/pkg/generated/norman/core/v1"
	"k8s.io/client-go/rest"
)

type CredentialController struct {
	Secrets v1.SecretInterface
}

func NewController(ns string) (*CredentialController, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	secretCnfg, err := v1.NewForConfig(*cfg)
	if err != nil {
		return nil, err
	}

	return &CredentialController{
		Secrets: secretCnfg.Secrets(ns),
	}, nil
}
