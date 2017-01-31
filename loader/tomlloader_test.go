package loader_test

import (
	"testing"

	"github.com/gophersgang/configor/loader"
)

func TestTomlAll(t *testing.T) {
	behavesLikeLoader(t, &loader.Tomlloader{}, "/tmp/toml_config.toml")
}
