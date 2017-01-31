package loader

import "fmt"

type ChainedLoader struct{}

var (
	myLoaders []ConfigAll
)

func init() {
	fmt.Println("INIT FOR CHAINED")
	// default: toml -> json -> yaml
	myLoaders = append(myLoaders, &Tomlloader{})
	myLoaders = append(myLoaders, &Jsonloader{})
	myLoaders = append(myLoaders, &Yamlloader{})
}

// Load will read the file and unmarshal
func (l *ChainedLoader) Load(config interface{}, file string) error {
	for _, l := range myLoaders {
		err := l.Load(config, file)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("Could not load from file %s", file)

}

// PlainLoad just does the unmarshalling
func (l *ChainedLoader) PlainLoad(config interface{}, file string) error {
	for _, l := range myLoaders {
		err := l.PlainLoad(config, file)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("Could not PlainLoad from file %s", file)
}

// Dump will marshal config to a file
func (l *ChainedLoader) Dump(config interface{}, file string) error {
	return myLoaders[0].Dump(config, file)
}
