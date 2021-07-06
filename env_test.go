// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package env

import (
	"log"
	"os"
	"testing"
)

var originalArguments = os.Args

// reset clears all the environment
func reset() {
	env.properties = make(map[string]*property)
	env.variables = make(map[string]*variable)
	os.Args = originalArguments
	os.Clearenv()
}

// TestAdHocUsage tests out-of-the-box usage
func TestAdHocUsage(t *testing.T) {
	t.Run("Test ad-hoc cml extraction", func(t *testing.T) {
		reset()
		originalArgs := os.Args
		os.Args = []string{"app", "-property1", "cmlValue1", "--property2", "longValue", "-property3=assignedValue"}
		Load()
		if v, isType := GetProperty("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "cmlValue1" {
			t.Error("value for property1 is not the expected one: ", v)
		}
		if v, isType := GetProperty("property2").(string); !isType {
			t.Error("property2 is not of the expected type")
		} else if v != "longValue" {
			t.Error("value for property2 is not the expected one: ", v)
		}
		if v, isType := GetProperty("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "assignedValue" {
			t.Error("value for property3 is not the expected one: ", v)
		}
		os.Args = originalArgs
	})

	t.Run("Test ad-hoc system environment variables extraction", func(t *testing.T) {
		reset()
		_ = os.Setenv("property1", "evValue1")
		Load()
		if v, isType := GetProperty("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "evValue1" {
			t.Error("value for property1 is not the expected one: ", v)
		}
		_ = os.Unsetenv("property1")
	})

	// Test json provided values
	t.Run("Test ad-hoc json configuration with overrides", func(t *testing.T) {
		reset()
		originalArgs := os.Args
		os.Args = []string{"app", "-Jsection.property2", "overrideSectionJsonValue2", "-j", "tests/config.json"}
		Load()

		if v, isType := GetProperty("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "jsonValue1" {
			t.Error("value for property1 is not the expected one: ", v)
		}
		if v, isType := GetProperty("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "jsonValue3" {
			t.Error("value for property3 is not the expected one: ", v)
		}
		if v, isType := GetProperty("section.property1").(string); !isType {
			t.Error("section.property1 is not of the expected type")
		} else if v != "sectionJsonValue1" {
			t.Error("value for section.property1 is not the expected one: ", v)
		}
		if v, isType := GetProperty("section.property2").(string); !isType {
			t.Error("section.property2 is not of the expected type")
		} else if v != "overrideSectionJsonValue2" {
			t.Error("value for section.property2 is not the expected one: ", v)
		}

		os.Args = originalArgs
	})

	// Test yml provided values
	t.Run("Test ad-hoc yml configuration with overrides", func(t *testing.T) {
		reset()
		originalArgs := os.Args
		os.Args = []string{"app", "-Ysection.property2", "overrideSectionYamlValue2", "-y", "tests/config.yml"}
		Load()

		if v, isType := GetProperty("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "yamlValue1" {
			t.Error("value for property1 is not the expected one: ", v)
		}
		if v, isType := GetProperty("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "yamlValue3" {
			t.Error("value for property3 is not the expected one: ", v)
		}
		if v, isType := GetProperty("section.property1").(string); !isType {
			t.Error("section.property1 is not of the expected type")
		} else if v != "sectionYamlValue1" {
			t.Error("value for section.property1 is not the expected one: ", v)
		}
		if v, isType := GetProperty("section.property2").(string); !isType {
			t.Error("section.property2 is not of the expected type")
		} else if v != "overrideSectionYamlValue2" {
			t.Error("value for section.property2 is not the expected one: ", v)
		}

		os.Args = originalArgs
	})
}

func TestConfigureVariables(t *testing.T) {

	t.Run("Test unprovided variables with default values", func(t *testing.T) {
		reset()
		Var("v1").Default("default1").Add()
		Var("v2").Add()

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

	t.Run("Test configured environment variable", func(t *testing.T) {
		reset()
		Var("v1").Default("default1").From(EnvironmentVariablesSource()).Add()

		os.Setenv("v1", "envValue1")
		Load()

		if v, isType := Get("v1").(string); !isType {
			t.Error("v1 is not of the expected type")
		} else if v != "envValue1" {
			t.Error("value for v1 is not the expected one: ", v)
		}
	})

	t.Run("Test configured cml variable", func(t *testing.T) {
		reset()
		Var("v1").Default("default1").From(CmlArgumentsSource()).Add()

		os.Args = []string{"app", "-Vv1", "cmlValue1"}
		Load()

		if v, isType := Get("v1").(string); !isType {
			t.Error("v1 is not of the expected type")
		} else if v != "cmlValue1" {
			t.Error("value for v1 is not the expected one: ", v)
		}
	})

	t.Run("Test configured json variable", func(t *testing.T) {
		reset()
		Var("property1").Default("default1").From(JsonConfigurationSource()).Add()

		os.Args = []string{"app", "-j", "tests/config.json"}
		Load()

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "jsonValue1" {
			t.Error("value for property1 is not the expected one: ", v)
		}
	})

	t.Run("Test configured yaml variable", func(t *testing.T) {
		reset()
		Var("property1").Default("default1").From(YamlConfigurationSource()).Add()

		os.Args = []string{"app", "-y", "tests/config.yml"}
		Load()

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "yamlValue1" {
			t.Error("value for property1 is not the expected one: ", v)
		}
	})

	// Test defined environment variables with non-provided values and with/without default values
	t.Run("Test unprovided properties with default values", func(t *testing.T) {
		reset()
		AddProperty("property1").WithDefaultValue("default1")
		AddProperty("property2")
		Load()
		if v, isType := GetProperty("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "default1" {
			t.Error("value for property1 is not the expected one: ", v)
		}
		if GetProperty("property2") != nil {
			t.Error("property2 was expected to be nil")
		}
		if GetProperty("property3") != nil {
			t.Error("property3 was expected to be nil")
		}
	})

	// Test chain of extraction
	t.Run("Test chain of extraction", func(t *testing.T) {
		reset()
		os.Args = []string{"app", "-property1", "cmlValue1", "-property2", "cmlValue2"}
		_ = os.Setenv("property1", "evValue1")
		_ = os.Setenv("property3", "evValue3")
		AddProperty("property1").WithDefaultValue("default1")
		AddProperty("property2").WithDefaultValue("default2")
		AddProperty("property3").WithDefaultValue("default3")
		AddProperty("property4").WithDefaultValue("default4")
		Load()

		if v, isType := GetProperty("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "cmlValue1" {
			t.Errorf("value for property1 is not the expected one: %v", v)
		}
		if v, isType := GetProperty("property2").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "cmlValue2" {
			t.Errorf("value for property2 is not the expected one: %v", v)
		}
		if v, isType := GetProperty("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "evValue3" {
			t.Errorf("value for property3 is not the expected one: %v", v)
		}
		if v, isType := GetProperty("property4").(string); !isType {
			t.Error("property4 is not of the expected type")
		} else if v != "default4" {
			t.Errorf("value for property4 is not the expected one: %v", v)
		}
		if GetProperty("property5") != nil {
			t.Error("property5 was expected to be nil")
		}
	})

	t.Run("Test Fully configured property", func(t *testing.T) {
		reset()
		AddProperty("property1").
			Required().
			From(JsonConfigurationProvider()).
			From(EnvironmentVariablesProvider())

		Load()
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Error("Validate did not trigger panic")
				}
				log.Print("RECOVERED")
			}()
			Validate()
		}()

		_ = os.Setenv("property1", "envValue1")
		Load()
		Validate()

		if v, isType := GetProperty("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "envValue1" {
			t.Errorf("value for property1 is not the expected one: %v", v)
		}

		os.Args = []string{"app", "-property1", "cmlValue1", "-j", "tests/config.json"}
		Load()
		Validate()

		if v, isType := GetProperty("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "jsonValue1" {
			t.Errorf("value for property1 is not the expected one: %v", v)
		}
	})
}
