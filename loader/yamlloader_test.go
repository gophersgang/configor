package loader_test

import (
	"testing"

	"github.com/gophersgang/configor/loader"
)

func TestYamlAll(t *testing.T) {
	behavesLikeLoader(t, &loader.Yamlloader{}, "/tmp/yaml_config.yaml")
}
