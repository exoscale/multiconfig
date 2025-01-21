package multiconfig

import (
	"flag"
	"net/url"
	"strings"
	"testing"

	"github.com/fatih/structs"
	"github.com/stretchr/testify/require"
)

func TestFlag(t *testing.T) {
	m := &FlagLoader{}
	s := &Server{}
	structName := structs.Name(s)

	// get flags
	args := getFlags(t, structName, "")

	m.Args = args[1:]

	if err := m.Load(s); err != nil {
		t.Error(err)
	}

	testStruct(t, s, getDefaultServer())
}

func TestFlagWithPrefix(t *testing.T) {
	const prefix = "Prefix"

	m := FlagLoader{Prefix: prefix}
	s := &Server{}
	structName := structs.Name(s)

	// get flags
	args := getFlags(t, structName, prefix)

	m.Args = args[1:]

	if err := m.Load(s); err != nil {
		t.Error(err)
	}

	testStruct(t, s, getDefaultServer())
}

func TestNestedFlags(t *testing.T) {
	m := FlagLoader{}
	s := &NestedServer{}
	structName := structs.Name(s)

	args := getFlags(t, structName, "")
	m.Args = args[1:]

	if err := m.Load(s); err != nil {
		t.Error(err)
	}

	testNestedStruct(t, s, getDefaultNestedServer())
}

func TestStructSeparatorFlags(t *testing.T) {
	m := FlagLoader{StructSeparator: "."}
	s := &NestedServer{}

	m.Args = []string{
		"--name", "koding",
		"--databaseoptions.postgres.enabled",
		"--databaseoptions.postgres.port", "5432",
		"--databaseoptions.postgres.hosts", "192.168.2.1,192.168.2.2,192.168.2.3",
		"--databaseoptions.postgres.dbname", "configdb",
		"--databaseoptions.postgres.availabilityratio", "8.23",
	}

	require.NoError(t, m.Load(s))

	testNestedStruct(t, s, getDefaultNestedServer())
}

func TestFlattenFlags(t *testing.T) {
	m := FlagLoader{
		Flatten: true,
	}
	s := &FlattenedServer{}
	structName := structs.Name(s)

	// get flags
	args := getFlags(t, structName, "")

	m.Args = args[1:]

	if err := m.Load(s); err != nil {
		t.Error(err)
	}

	testFlattenedStruct(t, s, getDefaultServer())
}

func TestCamelCaseFlags(t *testing.T) {
	m := FlagLoader{
		CamelCase: true,
	}
	s := &CamelCaseServer{}
	structName := structs.Name(s)

	// get flags
	args := getFlags(t, structName, "")

	m.Args = args[1:]

	if err := m.Load(s); err != nil {
		t.Error(err)
	}

	testCamelcaseStruct(t, s, getDefaultCamelCaseServer())
}

func TestCamelCaseAndStructSeparator(t *testing.T) {
	m := FlagLoader{
		CamelCase:       true,
		StructSeparator: ".",
	}
	s := &NestedServer{}

	m.Args = []string{
		"--name", "koding",
		"--database-options.postgres.enabled",
		"--database-options.postgres.port", "5432",
		"--database-options.postgres.hosts", "192.168.2.1,192.168.2.2,192.168.2.3",
		"--database-options.postgres.db-name", "configdb",
		"--database-options.postgres.availability-ratio", "8.23",
	}

	require.NoError(t, m.Load(s))

	testNestedStruct(t, s, getDefaultNestedServer())
}

func TestFlattenAndCamelCaseFlags(t *testing.T) {
	m := FlagLoader{
		Flatten:   true,
		CamelCase: true,
	}
	s := &FlattenedServer{}

	// get flags
	args := getFlags(t, "FlattenedCamelCaseServer", "")

	m.Args = args[1:]

	if err := m.Load(s); err != nil {
		t.Error(err)
	}
}

func TestCustomUsageFunc(t *testing.T) {
	const usageMsg = "foobar help"
	strt := struct {
		Foobar string
	}{}
	m := FlagLoader{
		FlagUsageFunc: (func(s string) string { return usageMsg }),
	}
	err := m.Load(&strt)

	if err != nil {
		t.Fatalf("Unable to load struct: %s", err)
	}
	f := m.flagSet.Lookup("foobar")
	if f == nil {
		t.Fatalf("Flag foobar is not set")
	}
	if f.Usage != usageMsg {
		t.Fatalf("usage message was %q, expected %q", f.Usage, usageMsg)
	}
}

type URL struct {
	*url.URL
}

var _ flag.Value = (*URL)(nil)

func (u *URL) Set(s string) error {
	ur, err := url.Parse(s)
	if err != nil {
		return err
	}
	u.URL = ur
	return nil
}

type Endpoint struct {
	Private *URL `required:"true"`
	Public  *URL `required:"true"`
}

func TestFlagValueSupport(t *testing.T) {
	m := &FlagLoader{}

	m.Args = []string{
		"-private", "http://127.0.0.1/kloud/kite",
		"-public", "http://127.0.0.1/kloud/kite",
	}

	var e Endpoint

	if err := m.Load(&e); err != nil {
		t.Fatalf("Load()=%s", err)
	}

	if e.Private.String() != m.Args[1] {
		t.Fatalf("got %q, want %q", e.Private, m.Args[3])
	}

	if e.Public.String() != m.Args[3] {
		t.Fatalf("got %q, want %q", e.Public, m.Args[3])
	}
}

func TestCustomUsageTag(t *testing.T) {
	const usageMsg = "foobar help"
	strt := struct {
		Foobar string `flagUsage:"foobar help"`
	}{}
	m := FlagLoader{}
	err := m.Load(&strt)

	if err != nil {
		t.Fatalf("Unable to load struct: %s", err)
	}
	f := m.flagSet.Lookup("foobar")
	if f == nil {
		t.Fatalf("Flag foobar is not set")
	}
	if f.Usage != usageMsg {
		t.Fatalf("usage message was %q, expected %q", f.Usage, usageMsg)
	}
}

func TestMapFlagSupport(t *testing.T) {
	m := &FlagLoader{CamelCase: true}

	m.Args = []string{
		"-map-string-int", "key1=1234,key2=456",
		"-map-string-string", "key1=val1,key2=val2",
	}

	var e struct {
		MapStringInt    map[string]int
		MapStringString map[string]string
	}

	require.NoError(t, m.Load(&e))
	require.EqualValues(t, map[string]int{"key1": 1234, "key2": 456}, e.MapStringInt)
	require.EqualValues(t, map[string]string{"key1": "val1", "key2": "val2"}, e.MapStringString)
}

// getFlags returns a slice of arguments that can be passed to flag.Parse()
func getFlags(t *testing.T, structName, prefix string) []string {
	t.Helper()
	if structName == "" {
		t.Fatal("struct name can not be empty")
	}

	var flags map[string]string
	switch structName {
	case "Server":
		flags = map[string]string{
			"-name":                       "koding",
			"-port":                       "6060",
			"-enabled":                    "",
			"-users":                      "ankara,istanbul",
			"-interval":                   "10s",
			"-id":                         "1234567890",
			"-labels":                     "123,456",
			"-postgres-enabled":           "",
			"-postgres-port":              "5432",
			"-postgres-hosts":             "192.168.2.1,192.168.2.2,192.168.2.3",
			"-postgres-dbname":            "configdb",
			"-postgres-availabilityratio": "8.23",
			"-epoch":                      "1638551008",
			"-epoch32":                    "1638551009",
			"-epoch64":                    "1638551010",
		}
	case "FlattenedServer":
		flags = map[string]string{
			"--enabled":           "",
			"--port":              "5432",
			"--hosts":             "192.168.2.1,192.168.2.2,192.168.2.3",
			"--dbname":            "configdb",
			"--availabilityratio": "8.23",
		}
	case "FlattenedCamelCaseServer":
		flags = map[string]string{
			"--enabled":            "",
			"--port":               "5432",
			"--hosts":              "192.168.2.1,192.168.2.2,192.168.2.3",
			"--db-name":            "configdb",
			"--availability-ratio": "8.23",
		}
	case "CamelCaseServer":
		flags = map[string]string{
			"--access-key":         "123456",
			"--normal":             "normal",
			"--db-name":            "configdb",
			"--availability-ratio": "8.23",
		}
	case "NestedServer":
		flags = map[string]string{
			"--name":                                       "koding",
			"--databaseoptions-postgres-enabled":           "",
			"--databaseoptions-postgres-port":              "5432",
			"--databaseoptions-postgres-hosts":             "192.168.2.1,192.168.2.2,192.168.2.3",
			"--databaseoptions-postgres-dbname":            "configdb",
			"--databaseoptions-postgres-availabilityratio": "8.23",
		}
	}

	prefix = strings.ToLower(prefix)

	args := []string{"multiconfig-test"}
	for key, val := range flags {
		flag := key
		if prefix != "" {
			flag = "-" + prefix + key
		}

		if val == "" {
			args = append(args, flag)
		} else {
			args = append(args, flag, val)
		}
	}

	return args
}
