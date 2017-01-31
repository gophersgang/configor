package loader_test

import (
	"testing"

	"github.com/gophersgang/configor/loader"
)

func TestChainedLoaderAll(t *testing.T) {
	behavesLikeLoader(t, &loader.ChainedLoader{}, "/tmp/toml_config.toml")
}
