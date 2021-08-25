// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package env

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	err "github.com/gomatbase/go-error"
)

var originalArguments = os.Args

// reset clears all the environment
func reset() {
	env.variables = make(map[string]*variable)
	os.Args = originalArguments
	os.Clearenv()
	cmlLoaded = false
	copyFile("tests/config.original.json", "tests/config.json")
	copyFile("tests/config.original.yml", "tests/config.yml")
	jsonConfigurationProviderDefaultInstance.json = nil
	yamlConfigurationProviderDefaultInstance.yaml = nil
}

func deferredFileClose(file *os.File) {
	if e := file.Close(); e != nil {
		log.Println("Failed to close file:", e)
	}
}

func copyFile(src, dst string) {
	srcFile, _ := os.Open(src)
	defer deferredFileClose(srcFile)
	dstFile, _ := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer deferredFileClose(dstFile)
	if _, e := io.Copy(dstFile, srcFile); e != nil {
		log.Printf("Faile to copy %s to %s: %s", srcFile.Name(), dstFile.Name(), e)
	}
}

func updateJson() {
	copyFile("tests/config.updated.json", "tests/config.json")
}

func updateYaml() {
	copyFile("tests/config.updated.yml", "tests/config.yml")
}

func TestCmlArgumentsSource(t *testing.T) {
	t.Run("Test ad-hoc cml extraction", func(t *testing.T) {
		reset()
		os.Args = []string{"app", "-property1", "cmlValue1", "--property2", "longValue", "-property3=assignedValue"}
		Load()
		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "cmlValue1" {
			t.Error("value for property1 is not the expected one: ", v)
		}
		if v, isType := Get("property2").(string); !isType {
			t.Error("property2 is not of the expected type")
		} else if v != "longValue" {
			t.Error("value for property2 is not the expected one: ", v)
		}
		if v, isType := Get("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "assignedValue" {
			t.Error("value for property3 is not the expected one: ", v)
		}
	})

	t.Run("Test configured cml variable", func(t *testing.T) {
		reset()
		_ = Var("v1").
			Default("default1").
			From(CmlArgumentsSource()).
			Add()
		_ = Var("v2").
			Default("default2").
			From(CmlArgumentsSource()).
			Add()
		_ = Var("v3").
			From(CmlArgumentsSource()).
			Add()
		_ = Var("v4").
			From(CmlArgumentsSource().Name("something")).
			Add()

		os.Args = []string{"app", "-Vv1", "cmlValue1", "-Vsomething", "cmlValue2"}
		Load()

		if v, isType := Get("v1").(string); !isType {
			t.Error("v1 is not of the expected type")
		} else if v != "cmlValue1" {
			t.Error("value for v1 is not the expected one: ", v)
		}

		if v, isType := Get("v2").(string); !isType {
			t.Error("v2 is not of the expected type")
		} else if v != "default2" {
			t.Error("value for v2 is not the expected one: ", v)
		}

		if Get("v3") != nil {
			t.Error("v3 was found!")
		}

		if v, isType := Get("v4").(string); !isType {
			t.Error("v4 is not of the expected type")
		} else if v != "cmlValue2" {
			t.Error("value for v4 is not the expected one: ", v)
		}
	})
}

func TestEnvironmentVariablesSource(t *testing.T) {
	t.Run("Test ad-hoc environment variables extraction", func(t *testing.T) {
		reset()
		_ = os.Setenv("property1", "evValue1")
		Load()
		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "evValue1" {
			t.Error("value for property1 is not the expected one: ", v)
		}
	})

	t.Run("Test configured environment variable", func(t *testing.T) {
		reset()
		_ = Var("v1").
			Default("default1").
			From(EnvironmentVariablesSource()).
			Add()
		_ = Var("v2").
			Default("default2").
			From(EnvironmentVariablesSource()).
			Add()
		_ = Var("v3").
			From(EnvironmentVariablesSource()).
			Add()
		_ = Var("v4").
			From(EnvironmentVariablesSource().Name("something")).
			Add()

		_ = os.Setenv("v1", "envValue1")
		_ = os.Setenv("something", "envValue2")
		Load()

		if v, isType := Get("v1").(string); !isType {
			t.Error("v1 is not of the expected type")
		} else if v != "envValue1" {
			t.Error("value for v1 is not the expected one: ", v)
		}

		if v, isType := Get("v2").(string); !isType {
			t.Error("v2 is not of the expected type")
		} else if v != "default2" {
			t.Error("value for v2 is not the expected one: ", v)
		}

		if Get("v3") != nil {
			t.Error("v3 was found!")
		}

		if v, isType := Get("v4").(string); !isType {
			t.Error("v4 is not of the expected type")
		} else if v != "envValue2" {
			t.Error("value for v4 is not the expected one: ", v)
		}
	})
}

func TestJsonConfigurationSource(t *testing.T) {
	t.Run("Test ad-hoc json configuration with overrides", func(t *testing.T) {
		reset()
		os.Args = []string{"app", "-Jsection.property2", "overrideSectionJsonValue2", "-j", "tests/config.json"}
		Load()

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type", env)
		} else if v != "jsonValue1" {
			t.Error("value for property1 is not the expected one: ", v)
		}
		if v, isType := Get("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "jsonValue3" {
			t.Error("value for property3 is not the expected one: ", v)
		}
		if v, isType := Get("section.property1").(string); !isType {
			t.Error("section.property1 is not of the expected type")
		} else if v != "sectionJsonValue1" {
			t.Error("value for section.property1 is not the expected one: ", v)
		}
		if v, isType := Get("section.property2").(string); !isType {
			t.Error("section.property2 is not of the expected type")
		} else if v != "overrideSectionJsonValue2" {
			t.Error("value for section.property2 is not the expected one: ", v)
		}
	})

	t.Run("Test configured json variables", func(t *testing.T) {
		reset()
		_ = Var("property1").
			Default("default1").
			From(JsonConfigurationSource()).
			Add()
		_ = Var("property2").
			From(JsonConfigurationSource().Name("section.property2")).
			Add()
		_ = Var("property5").
			Default("default5").
			From(JsonConfigurationSource()).
			Add()

		os.Args = []string{"app", "-j", "tests/config.json"}
		Load()

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "jsonValue1" {
			t.Error("value for property1 is not the expected one: ", v)
		}

		if v, isType := Get("property2").(string); !isType {
			t.Error("property2 is not of the expected type")
		} else if v != "sectionJsonValue2" {
			t.Error("value for property2 is not the expected one: ", v)
		}

		if v, isType := Get("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "jsonValue3" {
			t.Error("value for property3 is not the expected one: ", v)
		}

		if Get("property4") != nil {
			t.Error("property4 was found!")
		}

		if v, isType := Get("property5").(string); !isType {
			t.Error("property5 is not of the expected type")
		} else if v != "default5" {
			t.Error("value for property5 is not the expected one: ", v)
		}

		if v, isType := Get("section.property1").(string); !isType {
			t.Error("section.property1 is not of the expected type")
		} else if v != "sectionJsonValue1" {
			t.Error("value for section.property1 is not the expected one: ", v)
		}
	})
}

func TestYamlConfigurationSource(t *testing.T) {
	t.Run("Test ad-hoc yml configuration with overrides", func(t *testing.T) {
		reset()
		originalArgs := os.Args
		os.Args = []string{"app", "-Ysection.property2", "overrideSectionYamlValue2", "-y", "tests/config.yml"}
		Load()

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "yamlValue1" {
			t.Error("value for property1 is not the expected one: ", v)
		}
		if v, isType := Get("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "yamlValue3" {
			t.Error("value for property3 is not the expected one: ", v)
		}
		if v, isType := Get("section.property1").(string); !isType {
			t.Error("section.property1 is not of the expected type")
		} else if v != "sectionYamlValue1" {
			t.Error("value for section.property1 is not the expected one: ", v)
		}
		if v, isType := Get("section.property2").(string); !isType {
			t.Error("section.property2 is not of the expected type")
		} else if v != "overrideSectionYamlValue2" {
			t.Error("value for section.property2 is not the expected one: ", v)
		}

		os.Args = originalArgs
	})

	t.Run("Test configured yaml variable", func(t *testing.T) {
		reset()
		_ = Var("property1").
			Default("default1").
			From(YamlConfigurationSource()).
			Add()
		_ = Var("property2").
			From(YamlConfigurationSource().Name("section.property2")).
			Add()
		_ = Var("property5").
			Default("default5").
			From(YamlConfigurationSource()).
			Add()

		os.Args = []string{"app", "-y", "tests/config.yml"}
		Load()

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "yamlValue1" {
			t.Error("value for property1 is not the expected one: ", v)
		}

		if v, isType := Get("property2").(string); !isType {
			t.Error("property2 is not of the expected type")
		} else if v != "sectionYamlValue2" {
			t.Error("value for property2 is not the expected one: ", v)
		}

		if v, isType := Get("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "yamlValue3" {
			t.Error("value for property3 is not the expected one: ", v)
		}

		if Get("property6") != nil {
			t.Error("property5 was found!")
		}

		if v, isType := Get("property5").(string); !isType {
			t.Error("property5 is not of the expected type")
		} else if v != "default5" {
			t.Error("value for property5 is not the expected one: ", v)
		}

		if v, isType := Get("section.property1").(string); !isType {
			t.Error("section.property1 is not of the expected type")
		} else if v != "sectionYamlValue1" {
			t.Error("value for section.property1 is not the expected one: ", v)
		}
	})
}

func TestConfiguredVariables(t *testing.T) {
	t.Run("Test unprovided variables with default values", func(t *testing.T) {
		reset()
		_ = Var("v1").Default("default1").Add()
		_ = Var("v2").Add()

		if v, isType := Get("v1").(string); !isType {
			t.Error("v1 is not of the expected type")
		} else if v != "default1" {
			t.Error("value for v1 is not the expected one: ", v)
		}
		if Get("v2") != nil {
			t.Error("v2 was expected to be nil")
		}
		if Get("v3") != nil {
			t.Error("v3 was expected to be nil")
		}
	})

	t.Run("Test default chain of extraction", func(t *testing.T) {
		reset()
		os.Args = []string{"app", "-property1", "cmlValue1", "-j", "tests/config.json", "-y", "tests/config.yml"}
		_ = os.Setenv("property1", "envValue1")
		_ = os.Setenv("property3", "envValue3")
		_ = os.Setenv("property5", "envValue5")
		_ = Var("property1").Default("default1").Add()
		_ = Var("property2").Default("default2").Add()
		_ = Var("property3").Default("default3").Add()
		_ = Var("property4").Default("default4").Add()
		_ = Var("property5").Default("default5").Add()
		Load()

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "cmlValue1" {
			t.Errorf("value for property1 is not the expected one: %v", v)
		}
		if v, isType := Get("property2").(string); !isType {
			t.Error("property2 is not of the expected type")
		} else if v != "default2" {
			t.Errorf("value for property2 is not the expected one: %v", v)
		}
		if v, isType := Get("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "jsonValue3" {
			fmt.Println(jsonConfigurationProviderDefaultInstance.json)
			t.Errorf("value for property3 is not the expected one: %v", v)
		}
		if v, isType := Get("property4").(string); !isType {
			t.Error("property4 is not of the expected type")
		} else if v != "yamlValue4" {
			t.Errorf("value for property4 is not the expected one: %v", v)
		}
		if v, isType := Get("property5").(string); !isType {
			t.Error("property5 is not of the expected type")
		} else if v != "envValue5" {
			t.Errorf("value for property5 is not the expected one: %v", v)
		}
		if Get("property6") != nil {
			t.Error("property6 was expected to be nil")
		}
	})

	t.Run("Test custom chain of extraction", func(t *testing.T) {
		reset()
		os.Args = []string{"app", "-property1", "cmlValue1", "-property4", "cmlValue4", "-j", "tests/config.json", "-y", "tests/config.yml"}
		_ = os.Setenv("property1", "envValue1")
		_ = os.Setenv("property2", "envValue2")
		_ = os.Setenv("property3", "envValue3")
		_ = Var("property1").
			Default("default1").
			From(JsonConfigurationSource()).
			From(YamlConfigurationSource()).
			From(EnvironmentVariablesSource()).
			From(CmlArgumentsSource()).
			Add()
		_ = Var("property2").
			Default("default2").
			From(JsonConfigurationSource()).
			From(YamlConfigurationSource()).
			From(EnvironmentVariablesSource()).
			From(CmlArgumentsSource()).
			Add()
		_ = Var("property3").
			Default("default3").
			From(YamlConfigurationSource()).
			From(JsonConfigurationSource()).
			Add()
		_ = Var("property4").
			From(JsonConfigurationSource()).
			Add()
		Load()

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "jsonValue1" {
			t.Errorf("value for property1 is not the expected one: %v", v)
		}
		if v, isType := Get("property2").(string); !isType {
			t.Error("property2 is not of the expected type")
		} else if v != "envValue2" {
			t.Errorf("value for property2 is not the expected one: %v", v)
		}
		if v, isType := Get("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "yamlValue3" {
			t.Errorf("value for property3 is not the expected one: %v", v)
		}
		if Get("property4") != nil {
			t.Error("property4 was found")
		}
		if v := Get("property5"); v != nil {
			t.Error("property5 was expected to be nil:", v)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("Test valid environment", func(t *testing.T) {
		reset()
		os.Args = []string{"app", "tests/config.json"}
		_ = Var("property1").
			Default("default1").
			Required().
			Add()
		_ = Var("property2").
			Default("default2").
			Required().
			Add()
		_ = Var("property3").
			Default("default3").
			Required().
			Add()
		_ = Var("property4").
			Add()
		Load()

		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Error("Validate should not have raised panic")
				}
			}()
			if e := Validate(); e != nil {
				t.Error("Validate should have succeeded")
			}
		}()

	})

	t.Run("Test missing variable", func(t *testing.T) {
		reset()
		os.Args = []string{"app", "tests/config.json"}
		_ = Var("property1").
			Default("default1").
			Required().
			Add()
		_ = Var("property2").
			Required().
			Add()
		_ = Var("property3").
			Default("default3").
			Required().
			Add()
		_ = Var("property4").
			Required().
			Add()
		Load()

		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Error("Validate should not have raised panic")
				}
			}()
			if e := Validate(); e == nil {
				t.Error("Validate should have failed")
			} else if err.Count(e) != 2 {
				t.Error("Validate should have failed with 2 errors")
			}
		}()

	})

	t.Run("Test panic on failed validation", func(t *testing.T) {
		reset()
		FailOnMissingVariables(true)
		os.Args = []string{"app", "tests/config.json"}
		_ = Var("property1").
			Default("default1").
			Required().
			Add()
		_ = Var("property2").
			Required().
			Add()
		_ = Var("property3").
			Default("default3").
			Required().
			Add()
		_ = Var("property4").
			Required().
			Add()
		Load()

		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Error("Validate should have raised panic")
				}
			}()
			_ = Validate()
			t.Error("Validate should have panicked")
		}()

	})
}

func TestRefresh(t *testing.T) {
	t.Run("Test refresh with ad-hoc variables", func(t *testing.T) {
		reset()

		os.Args = []string{"app", "-property1", "cmlValue1", "-j", "tests/config.json", "-y", "tests/config.yml"}
		_ = os.Setenv("property1", "envValue1")
		_ = os.Setenv("property3", "envValue3")
		_ = os.Setenv("property5", "envValue5")
		Load()

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "cmlValue1" {
			t.Errorf("value for property1 is not the expected one: %v", v)
		}
		if Get("property2") != nil {
			t.Error("property2 was expected to be nil")
		}
		if v, isType := Get("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "jsonValue3" {
			t.Errorf("value for property3 is not the expected one: %v", v)
		}
		if v, isType := Get("property4").(string); !isType {
			t.Error("property4 is not of the expected type")
		} else if v != "yamlValue4" {
			t.Errorf("value for property4 is not the expected one: %v", v)
		}
		if v, isType := Get("property5").(string); !isType {
			t.Error("property5 is not of the expected type")
		} else if v != "envValue5" {
			t.Errorf("value for property5 is not the expected one: %v", v)
		}

		updateJson()
		updateYaml()
		os.Args = originalArguments
		_ = os.Setenv("property5", "envNewValue5")
		if e := Refresh(); e != nil {
			t.Error("Unexpected refresh errors :\n", e.Error())
		}

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "cmlValue1" {
			t.Errorf("value for property1 is not the expected one: %v", v)
		}
		if v, isType := Get("property2").(string); !isType {
			t.Error("property2 is not of the expected type")
		} else if v != "jsonValue2" {
			t.Errorf("value for property2 is not the expected one: %v", v)
		}
		if v, isType := Get("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "jsonValue3" {
			t.Errorf("value for property3 is not the expected one: %v", v)
		}
		if v, isType := Get("property4").(string); !isType {
			t.Error("property4 is not of the expected type")
		} else if v != "yamlNewValue4" {
			t.Errorf("value for property4 is not the expected one: %v", v)
		}

		// ad-hoc variables will only get refreshed if the first value in the default chain is refreshed. CML doesn't
		// refresh as that doesn't happen in a real-life scenario. Environment variables also don't change in real life
		// scenarios but the provider always takes the value directly from the os environment variable.
		if v, isType := Get("property5").(string); !isType {
			t.Error("property5 is not of the expected type")
		} else if v != "envNewValue5" {
			t.Errorf("value for property5 is not the expected one: %v", v)
		}
	})

	t.Run("Test refresh with configured variables", func(t *testing.T) {
		reset()

		_ = Var("property1").
			Default("default1").
			Required().
			Add()
		_ = Var("property2").
			Add()
		_ = Var("property3").
			Default("default3").
			Required().
			Add()
		_ = Var("property4").
			Default("default4").
			Required().
			Add()
		_ = Var("property5").
			Default("default5").
			Required().
			Add()

		os.Args = []string{"app", "-property1", "cmlValue1", "-j", "tests/config.json", "-y", "tests/config.yml"}
		_ = os.Setenv("property1", "envValue1")
		_ = os.Setenv("property3", "envValue3")
		_ = os.Setenv("property5", "envValue5")
		Load()

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "cmlValue1" {
			t.Errorf("value for property1 is not the expected one: %v", v)
		}
		if Get("property2") != nil {
			t.Error("property2 was expected to be nil")
		}
		if v, isType := Get("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "jsonValue3" {
			t.Errorf("value for property3 is not the expected one: %v", v)
		}
		if v, isType := Get("property4").(string); !isType {
			t.Error("property4 is not of the expected type")
		} else if v != "yamlValue4" {
			t.Errorf("value for property4 is not the expected one: %v", v)
		}
		if v, isType := Get("property5").(string); !isType {
			t.Error("property5 is not of the expected type")
		} else if v != "envValue5" {
			t.Errorf("value for property5 is not the expected one: %v", v)
		}

		updateJson()
		updateYaml()
		os.Args = originalArguments
		_ = os.Setenv("property5", "envNewValue5")
		if e := Refresh(); e != nil {
			t.Error("Unexpected refresh errors :\n", e.Error())
		}

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "jsonNewValue1" {
			t.Errorf("value for property1 is not the expected one: %v", v)
		}
		if v, isType := Get("property2").(string); !isType {
			t.Error("property2 is not of the expected type")
		} else if v != "jsonValue2" {
			t.Errorf("value for property2 is not the expected one: %v", v)
		}
		if v, isType := Get("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "yamlNewValue3" {
			t.Errorf("value for property3 is not the expected one: %v", v)
		}
		if v, isType := Get("property4").(string); !isType {
			t.Error("property4 is not of the expected type")
		} else if v != "yamlNewValue4" {
			t.Errorf("value for property4 is not the expected one: %v", v)
		}
		// For defined variables, values are cached to do comparison. The environment variable provider should never
		// report being dirty and the value for the property should not change.
		if v, isType := Get("property5").(string); !isType {
			t.Error("property5 is not of the expected type")
		} else if v != "envValue5" {
			t.Errorf("value for property5 is not the expected one: %v", v)
		}
	})

	t.Run("Test refresh with configured variable listeners", func(t *testing.T) {
		reset()
	})
}
