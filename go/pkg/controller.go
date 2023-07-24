package pkg

import (
	v1 "harrisonwaffel/ccg-gmsa-dll/pkg/generated/norman/core/v1"
	"k8s.io/client-go/rest"
)

type CredentialController struct {
	Secrets v1.SecretInterface
}

func NewController(ns string) *CredentialController {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	secretCnfg, err := v1.NewForConfig(*cfg)
	if err != nil {
		panic(err)
	}

	s := secretCnfg.Secrets(ns)

	controller := &CredentialController{
		Secrets: s,
	}

	return controller
}
