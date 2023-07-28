package pkg

import (
	"fmt"

	v1 "harrisonwaffel/ccg-gmsa-dll/pkg/generated/norman/core/v1"
	"k8s.io/client-go/rest"
)

type CredentialClient struct {
	Secrets v1.SecretInterface
}

func NewClient(ns string) (*CredentialClient, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("this application must be run as a Kubernetes workload : %v", err)
	}

	secretCnfg, err := v1.NewForConfig(*cfg)
	if err != nil {
		return nil, err
	}

	return &CredentialClient{
		Secrets: secretCnfg.Secrets(ns),
	}, nil
}

type Response struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	DomainName string `json:"domainName"`
}
