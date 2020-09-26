// Copyright 2020 GOM. All rights reserved.
// Since 24/09/2020 By GOM
// Licensed under MIT License

package env

import (
	"os"
	"testing"
)

func TestVariableExctraction(t *testing.T) {

	t.Run("Test ad-hoc cml extraction", func(t *testing.T) {
		originalArgs := os.Args
		os.Args = []string{"app", "-property1", "cmlValue1", "--property2", "longValue", "-property3=assignedValue"}
		Refresh()
		os.Args = originalArgs
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

	t.Run("Test ad-hoc system environment variables extraction", func(t *testing.T) {
		_ = os.Setenv("property1", "evValue1")
		Refresh()
		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "evValue1" {
			t.Error("value for property1 is not the expected one: ", v)
		}
		_ = os.Unsetenv("property1")
	})

	//Test defined environment variables with non-provided values and with/without default values
	t.Run("Test json configuration with defaults and overrides", func(t *testing.T) {
		originalArgs := os.Args
		os.Args = []string{"app", "-Jsection.property2", "overrideSectionJsonValue2", "-j", "tests/config.json"}

		AddProperty("property1").WithDefaultValue("default1")
		Build()
		os.Args = originalArgs

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
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

	//Test json configuration provider with non-provided values and with/without default values
	t.Run("Test unconfigured properties with default values", func(t *testing.T) {
		AddProperty("property1").WithDefaultValue("default1")
		AddProperty("property2")
		Build()
		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "default1" {
			t.Error("value for property1 is not the expected one: ", v)
		}
		if Get("property2") != nil {
			t.Error("property2 was expected to be nil")
		}
		if Get("property3") != nil {
			t.Error("property3 was expected to be nil")
		}
	})

	// Test chain of extraction
	t.Run("Test chain of extraction", func(t *testing.T) {
		originalArgs := os.Args
		os.Args = []string{"app", "-property1", "cmlValue1", "-property2", "cmlValue2"}
		_ = os.Setenv("property1", "evValue1")
		_ = os.Setenv("property3", "evValue3")
		AddProperty("property1").WithDefaultValue("default1")
		AddProperty("property2").WithDefaultValue("default2")
		AddProperty("property3").WithDefaultValue("default3")
		AddProperty("property4").WithDefaultValue("default4")
		Build()
		os.Args = originalArgs

		if v, isType := Get("property1").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "cmlValue1" {
			t.Errorf("value for property1 is not the expected one: %v", v)
		}
		if v, isType := Get("property2").(string); !isType {
			t.Error("property1 is not of the expected type")
		} else if v != "cmlValue2" {
			t.Errorf("value for property2 is not the expected one: %v", v)
		}
		if v, isType := Get("property3").(string); !isType {
			t.Error("property3 is not of the expected type")
		} else if v != "evValue3" {
			t.Errorf("value for property3 is not the expected one: %v", v)
		}
		if v, isType := Get("property4").(string); !isType {
			t.Error("property4 is not of the expected type")
		} else if v != "default4" {
			t.Errorf("value for property4 is not the expected one: %v", v)
		}
		if Get("property5") != nil {
			t.Error("property5 was expected to be nil")
		}
	})
}
