package loader_test

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/gophersgang/configor/loader"
)

type Config struct {
	APPName string `default:"configor"`

	DB struct {
		Name     string
		User     string `default:"root"`
		Password string `required:"true" env:"DBPassword"`
		Port     uint   `default:"3306"`
	}

	Contacts []struct {
		Name  string
		Email string `required:"true"`
	}

	Anonymous `anonymous:"true"`
}

type Anonymous struct {
	Description string
}

func generateDefaultConfig() Config {
	config := Config{
		APPName: "configor",
		DB: struct {
			Name     string
			User     string `default:"root"`
			Password string `required:"true" env:"DBPassword"`
			Port     uint   `default:"3306"`
		}{
			Name:     "configor",
			User:     "configor",
			Password: "configor",
			Port:     3306,
		},
		Contacts: []struct {
			Name  string
			Email string `required:"true"`
		}{
			{
				Name:  "Jinzhu",
				Email: "wosmvp@gmail.com",
			},
		},
		Anonymous: Anonymous{
			Description: "This is an anonymous embedded struct whose environment variables should NOT include 'ANONYMOUS'",
		},
	}
	return config
}

func behavesLikeLoader(t *testing.T, loader loader.ConfigAll, configFile string) {
	fmt.Printf("*** FOR %T\n\n", loader)
	behavesLikeLoaderLoad(t, loader, configFile)
	behavesLikeLoaderDump(t, loader, configFile)
	behavesLikeLoaderErrors(t, loader, configFile)

}

func behavesLikeLoaderLoad(t *testing.T, loader loader.ConfigAll, configFile string) {
	config := generateDefaultConfig()
	loader.Dump(config, configFile)

	dat, err := ioutil.ReadFile(configFile)
	if err != nil {
		t.Error(err)
	}
	_ = dat
}

func behavesLikeLoaderDump(t *testing.T, loader loader.ConfigAll, configFile string) {
	// Dumping / Loading / Overriding
	config := generateDefaultConfig()
	little := Config{
		APPName: "little config",
	}

	loader.Dump(config, configFile)
	loader.Load(&little, configFile)

	if little.APPName != "configor" {
		t.Errorf("expected AppName to be configor, was %s", little.APPName)
	}
	fmt.Println(little)
}

func behavesLikeLoaderErrors(t *testing.T, loader loader.ConfigAll, configFile string) {
	dumbPath := "/laskdnf/lasdjflkasj/lasdjflkajsd/alsdjflkasd"
	// non-createable file
	config := generateDefaultConfig()
	err := loader.Dump(config, dumbPath)
	if err == nil {
		t.Errorf("Expected error on writing to a non-createable file")
	}

	// reading from non-existing file
	little := Config{
		APPName: "little config",
	}
	err = loader.Load(&little, dumbPath)
	if err == nil {
		t.Errorf("Expected error on reading from non-existing file")
	}

	err = loader.PlainLoad(&little, dumbPath)
	if err == nil {
		t.Errorf("Expected error on reading from non-existing file")
	}

	myType := reflect.TypeOf(loader)

	// YAML parses all kinda of junk... leave it out of this test...
	if myType.String() != "*loader.Yamlloader" && myType.String() != "*loader.ChainedLoader" {
		// reading junk from existing file
		ioutil.WriteFile("/tmp/junkfile.txt", []byte("1111 1111 Here is a string...."), 0644)
		err = loader.PlainLoad(&little, "/tmp/junkfile.txt")
		fmt.Println(err)
		if err == nil {
			t.Errorf("Expected error on reading junk")
		}
	}
}
