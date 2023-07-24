package main

import (
	"os"

	"harrisonwaffel/ccg-gmsa-dll/pkg/codegen/generator"
	v1 "k8s.io/api/core/v1"
)

func main() {
	os.Unsetenv("GOPATH")

	generator.GenerateNativeTypes(v1.SchemeGroupVersion, []interface{}{
		v1.Secret{},
	}, nil)
}
