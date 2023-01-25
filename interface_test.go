package multiconfig

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInterfaceLoader(t *testing.T) {
	t.Run("Should be a noop if nothing implements the interface", func(t *testing.T) {
		loader := InterfaceLoader{}
		conf := &TestConfig{}

		require.NoError(t, loader.Load(conf))
		require.Equal(t, &TestConfig{}, conf)
	})

	t.Run("Should invoke ApplyDefaults if top struct implements interface", func(t *testing.T) {
		loader := InterfaceLoader{}
		conf := &Config1{}
		expected := &Config1{
			Input:  "default string",
			Result: 42,
			Flag:   true,
		}

		require.NoError(t, loader.Load(conf))
		require.Equal(t, expected, conf)
	})

	t.Run("Should return an error when trying to load into a non-pointer", func(t *testing.T) {
		loader := InterfaceLoader{}
		conf := Config1{}

		require.EqualError(t, loader.Load(conf), "cannot load into a value: target must be a pointer")
	})

	t.Run("Should return an error if the input is nil", func(t *testing.T) {
		loader := InterfaceLoader{}
		var conf *Config1

		require.EqualError(t, loader.Load(conf), "cannot load into a nil pointer")
	})

	t.Run("Should invoke ApplyDefaults on nested struct", func(t *testing.T) {
		type NestedConfig struct {
			Output string
			Config1
		}
		loader := &InterfaceLoader{}
		conf := &NestedConfig{}
		expected := &NestedConfig{
			Output: "",
			Config1: Config1{
				Input:  "default string",
				Result: 42,
				Flag:   true,
			},
		}

		require.NoError(t, loader.Load(conf))
		require.Equal(t, expected, conf)
	})

	t.Run("Should invoke ApplyDefaults on nested struct-pointer", func(t *testing.T) {
		type NestedConfig struct {
			Output  string
			Config1 *Config1
		}
		loader := &InterfaceLoader{}
		conf := &NestedConfig{}
		expected := &NestedConfig{
			Output: "",
			Config1: &Config1{
				Input:  "default string",
				Result: 42,
				Flag:   true,
			},
		}

		require.NoError(t, loader.Load(conf))
		require.Equal(t, expected, conf)
	})

	t.Run("Should invoke ApplyDefaults recursively depth first", func(t *testing.T) {
		loader := &InterfaceLoader{}
		conf := &Config2{}
		expected := &Config2{
			Output: "my output",
			Config1: Config1{
				Input:  "my input",
				Result: 42,
				Flag:   true,
			},
		}

		require.NoError(t, loader.Load(conf))
		require.Equal(t, expected, conf)
	})

	// Not sure why anyone would do this, but someone will try so let's try to support it anyway
	t.Run("Should work when ApplyDefaults is implemented on a non-struct type", func(t *testing.T) {
		loader := &InterfaceLoader{}
		conf := new(IntWithDefault)
		expected := IntWithDefault(42)

		require.NoError(t, loader.Load(conf))
		require.Equal(t, &expected, conf)
	})

	t.Run("Should work when ApplyDefaults is implemented on a nested non-struct type", func(t *testing.T) {
		loader := &InterfaceLoader{}
		conf := &Config3{}
		expected := &Config3{
			Config1: Config1{
				Input:  "default string",
				Result: 42,
				Flag:   true,
			},
			IntWithDefault: 42,
		}

		require.NoError(t, loader.Load(conf))
		require.Equal(t, expected, conf)
	})
}

type TestConfig struct {
	pField string //nolint:unused
	Input  string
	Result int
	Flag   bool
}

type Config1 TestConfig

func (c *Config1) ApplyDefaults() {
	c.Input = "default string"
	c.Result = 42
	c.Flag = true
}

type Config2 struct {
	Output string
	Config1
}

func (c *Config2) ApplyDefaults() {
	c.Output = "my output"
	c.Config1.Input = "my input"
}

type IntWithDefault int

func (i *IntWithDefault) ApplyDefaults() {
	*i = IntWithDefault(42)
}

type Config3 struct {
	Config1
	IntWithDefault
}
