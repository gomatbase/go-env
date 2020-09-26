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
	t.Run("Test Unconfigured properties with default values", func(t *testing.T) {
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
}
