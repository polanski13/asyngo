package gen

import (
	sigsyaml "sigs.k8s.io/yaml"
)

func jsonToYAML(data []byte) ([]byte, error) {
	return sigsyaml.JSONToYAML(data)
}
