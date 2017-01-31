package loader_test

import (
	"testing"

	"github.com/gophersgang/configor/loader"
)

func TestJsonAll(t *testing.T) {
	behavesLikeLoader(t, &loader.Jsonloader{}, "/tmp/json_config.json")
}
