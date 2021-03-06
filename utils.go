package configor

import (
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/gophersgang/configor/loader"

	yaml "gopkg.in/yaml.v1"
)

func (configor *Configor) getENVPrefix(config interface{}) string {
	if configor.Config.ENVPrefix == "" {
		if prefix := os.Getenv("CONFIGOR_ENV_PREFIX"); prefix != "" {
			return prefix
		}
		return "Configor"
	}
	return configor.Config.ENVPrefix
}

func getConfigurationFileWithENVPrefix(file, env string) (string, error) {
	var (
		envFile string
		extname = path.Ext(file)
	)

	if extname == "" {
		envFile = fmt.Sprintf("%v.%v", file, env)
	} else {
		envFile = fmt.Sprintf("%v.%v%v", strings.TrimSuffix(file, extname), env, extname)
	}

	if fileInfo, err := os.Stat(envFile); err == nil && fileInfo.Mode().IsRegular() {
		return envFile, nil
	}
	return "", fmt.Errorf("failed to find file %v", file)
}

func (configor *Configor) getConfigurationFiles(files ...string) []string {
	var results []string

	for i := len(files) - 1; i >= 0; i-- {
		foundFile := false
		file := files[i]

		// check configuration
		if fileInfo, err := os.Stat(file); err == nil && fileInfo.Mode().IsRegular() {
			foundFile = true
			results = append(results, file)
		}

		// check configuration with env
		if file, err := getConfigurationFileWithENVPrefix(file, configor.GetEnvironment()); err == nil {
			foundFile = true
			results = append(results, file)
		}

		// check example configuration
		if !foundFile {
			if example, err := getConfigurationFileWithENVPrefix(file, "example"); err == nil {
				fmt.Printf("Failed to find configuration %v, using example file %v\n", file, example)
				results = append(results, example)
			} else {
				fmt.Printf("Failed to find configuration %v\n", file)
			}
		}
	}
	return results
}

func processFile(config interface{}, file string) error {
	loader := &loader.ChainedLoader{}
	err := loader.Load(config, file)
	if err == nil {
		return nil
	}
	return loader.PlainLoad(config, file)
}

func getPrefixForStruct(prefixes *[]string, fieldStruct *reflect.StructField) []string {
	if fieldStruct.Anonymous && fieldStruct.Tag.Get("anonymous") == "true" {
		return *prefixes
	}
	return append(*prefixes, fieldStruct.Name)
}

func processTags(config interface{}, prefixes ...string) error {
	configValue := reflect.Indirect(reflect.ValueOf(config))
	if configValue.Kind() != reflect.Struct {
		return errors.New("invalid config, should be struct")
	}

	configType := configValue.Type()

	for i := 0; i < configType.NumField(); i++ {
		var (
			fieldStruct = configType.Field(i)
			field       = configValue.Field(i)
		)
		field = reflect.Indirect(field) // so we dont deal with pointers
		err := processTag(&fieldStruct, &field, &prefixes)
		if err != nil {
			return err
		}
	}
	return nil
}

func processTag(fieldStruct *reflect.StructField, field *reflect.Value, prefixes *[]string) error {
	var err error

	err = loadFromShellEnv(fieldStruct, field, prefixes)
	if err != nil {
		return err
	}

	err = setDefaultFromTag(fieldStruct, field)
	if err != nil {
		return err
	}

	err = processTagIfStruct(fieldStruct, field, prefixes)
	if err != nil {
		return err
	}

	err = processTagIfSlice(fieldStruct, field, prefixes)
	if err != nil {
		return err
	}
	return nil
}

func loadFromShellEnv(fieldStruct *reflect.StructField, field *reflect.Value, prefixes *[]string) error {
	envNames := envNames(fieldStruct, prefixes)

	// Load From Shell ENV
	for _, env := range envNames {
		if value := os.Getenv(env); value != "" {
			err := yaml.Unmarshal([]byte(value), field.Addr().Interface())
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func setDefaultFromTag(fieldStruct *reflect.StructField, field *reflect.Value) error {
	isBlank := reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface())
	if !isBlank {
		return nil
	}

	// Set default configuration
	defValue := fieldStruct.Tag.Get("default")
	if defValue != "" {
		err := yaml.Unmarshal([]byte(defValue), field.Addr().Interface())
		if err != nil {
			return err
		}
	} else if fieldStruct.Tag.Get("required") == "true" {
		// return error if it is required but blank
		return errors.New(fieldStruct.Name + " is required, but blank")
	}
	return nil
}

func processTagIfStruct(fieldStruct *reflect.StructField, field *reflect.Value, prefixes *[]string) error {
	if field.Kind() != reflect.Struct {
		return nil
	}

	err := processTags(field.Addr().Interface(), getPrefixForStruct(prefixes, fieldStruct)...)
	if err != nil {
		return err
	}
	return nil
}

func processTagIfSlice(fieldStruct *reflect.StructField, field *reflect.Value, prefixes *[]string) error {
	if field.Kind() != reflect.Slice {
		return nil
	}
	for i := 0; i < field.Len(); i++ {
		if reflect.Indirect(field.Index(i)).Kind() == reflect.Struct {
			newPrefixes := append(getPrefixForStruct(prefixes, fieldStruct), fmt.Sprint(i))
			err := processTags(field.Index(i).Addr().Interface(), newPrefixes...)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func envNames(fieldStruct *reflect.StructField, prefixes *[]string) []string {
	var res []string
	envName := fieldStruct.Tag.Get("env") // read configuration from shell env
	if envName == "" {
		res = append(res, strings.Join(append(*prefixes, fieldStruct.Name), "_"))                  // Configor_DB_Name
		res = append(res, strings.ToUpper(strings.Join(append(*prefixes, fieldStruct.Name), "_"))) // CONFIGOR_DB_NAME
	} else {
		res = []string{envName}
	}
	return res
}
