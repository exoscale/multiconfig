package multiconfig

import (
	"strings"
	"testing"

	"github.com/fatih/structs"
	"github.com/stretchr/testify/require"
)

func TestENV(t *testing.T) {
	m := EnvironmentLoader{}
	s := &Server{}
	structName := structs.Name(s)

	// set env variables
	setEnvVars(t, structName, "")

	if err := m.Load(s); err != nil {
		t.Error(err)
	}

	testStruct(t, s, getDefaultServer())
}

func TestCamelCaseEnv(t *testing.T) {
	m := EnvironmentLoader{
		CamelCase: true,
	}
	s := &CamelCaseServer{}
	structName := structs.Name(s)

	// set env variables
	setEnvVars(t, structName, "")

	if err := m.Load(s); err != nil {
		t.Error(err)
	}

	testCamelcaseStruct(t, s, getDefaultCamelCaseServer())
}

func TestENVWithPrefix(t *testing.T) {
	const prefix = "Prefix"

	m := EnvironmentLoader{Prefix: prefix}
	s := &Server{}
	structName := structs.New(s).Name()

	// set env variables
	setEnvVars(t, structName, prefix)

	if err := m.Load(s); err != nil {
		t.Error(err)
	}

	testStruct(t, s, getDefaultServer())
}

func TestENVFlattenStructPrefix(t *testing.T) {
	const prefix = "Prefix"

	m := EnvironmentLoader{Prefix: prefix}
	s := &TaggedServer{}
	structName := structs.New(s).Name()

	// set env variables
	setEnvVars(t, structName, prefix)

	if err := m.Load(s); err != nil {
		t.Error(err)
	}

	testPostgres(t, s.Postgres, getDefaultServer().Postgres)
}

func setEnvVars(t *testing.T, structName, prefix string) {
	t.Helper()
	if structName == "" {
		t.Fatal("struct name can not be empty")
	}

	var env map[string]string
	switch structName {
	case "Server":
		env = map[string]string{
			"NAME":                       "koding",
			"PORT":                       "6060",
			"ENABLED":                    "true",
			"USERS":                      "ankara,istanbul",
			"INTERVAL":                   "10s",
			"ID":                         "1234567890",
			"LABELS":                     "123,456",
			"POSTGRES_ENABLED":           "true",
			"POSTGRES_PORT":              "5432",
			"POSTGRES_HOSTS":             "192.168.2.1,192.168.2.2,192.168.2.3",
			"POSTGRES_DBNAME":            "configdb",
			"POSTGRES_AVAILABILITYRATIO": "8.23",
			"POSTGRES_FOO":               "8.23,9.12,11,90",
			"EPOCH":                      "1638551008",
			"EPOCH32":                    "1638551009",
			"EPOCH64":                    "1638551010",
		}
	case "CamelCaseServer":
		env = map[string]string{
			"ACCESS_KEY":         "123456",
			"NORMAL":             "normal",
			"DB_NAME":            "configdb",
			"AVAILABILITY_RATIO": "8.23",
		}
	case "TaggedServer":
		env = map[string]string{
			"NAME":              "koding",
			"ENABLED":           "true",
			"PORT":              "5432",
			"HOSTS":             "192.168.2.1,192.168.2.2,192.168.2.3",
			"DBNAME":            "configdb",
			"AVAILABILITYRATIO": "8.23",
			"FOO":               "8.23,9.12,11,90",
		}
	}

	if prefix == "" {
		prefix = structName
	}

	prefix = strings.ToUpper(prefix)

	for key, val := range env {
		env := prefix + "_" + key
		t.Setenv(env, val)
	}
}

func TestENVgetPrefix(t *testing.T) {
	e := &EnvironmentLoader{}
	s := &Server{}

	st := structs.New(s)

	prefix := st.Name()

	if p := e.getPrefix(st); p != prefix {
		t.Errorf("Prefix is wrong: %s, want: %s", p, prefix)
	}

	prefix = "Test"
	e = &EnvironmentLoader{Prefix: prefix}
	if p := e.getPrefix(st); p != prefix {
		t.Errorf("Prefix is wrong: %s, want: %s", p, prefix)
	}
}

func TestMapEnvSupport(t *testing.T) {
	m := &EnvironmentLoader{CamelCase: true}

	env := map[string]string{
		"_MAP_STRING_INT":    "key1=1234,key2=456",
		"_MAP_STRING_STRING": "key1=val1,key2=val2",
	}
	for key, val := range env {
		t.Setenv(key, val)
	}

	var e struct {
		MapStringInt    map[string]int
		MapStringString map[string]string
	}

	require.NoError(t, m.Load(&e))
	require.EqualValues(t, map[string]int{"key1": 1234, "key2": 456}, e.MapStringInt)
	require.EqualValues(t, map[string]string{"key1": "val1", "key2": "val2"}, e.MapStringString)
}
