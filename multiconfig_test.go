package multiconfig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type (
	Server struct {
		Name       string `required:"true"`
		Port       int    `default:"6060"`
		ID         int64
		Labels     []int
		Enabled    bool
		Users      []string
		Postgres   Postgres
		unexported string
		Interval   time.Duration
		Epoch      uint   `default:"1638551008"`
		Epoch32    uint32 `default:"1638551009"`
		Epoch64    uint64 `default:"1638551010"`
	}

	// Postgres holds Postgresql database related configuration
	Postgres struct {
		Enabled           bool
		Port              uint16   `required:"true" customRequired:"yes"`
		Hosts             []string `required:"true"`
		DBName            string   `default:"configdb"`
		AvailabilityRatio float64
		unexported        string
	}

	TaggedServer struct {
		Name     string `required:"true"`
		Postgres `structs:",flatten"`
	}

	Database struct {
		Postgres Postgres
	}

	NestedServer struct {
		Name            string `required:"true"`
		DatabaseOptions Database
	}
)

type FlattenedServer struct {
	Postgres Postgres
}

type CamelCaseServer struct {
	AccessKey         string
	Normal            string
	DBName            string `default:"configdb"`
	AvailabilityRatio float64
}

var (
	testTOML = "testdata/config.toml"
	testJSON = "testdata/config.json"
	testYAML = "testdata/config.yaml"
)

func getDefaultServer() *Server {
	return &Server{
		Name:     "koding",
		Port:     6060,
		Enabled:  true,
		ID:       1234567890,
		Labels:   []int{123, 456},
		Users:    []string{"ankara", "istanbul"},
		Interval: 10 * time.Second,
		Postgres: Postgres{
			Enabled:           true,
			Port:              5432,
			Hosts:             []string{"192.168.2.1", "192.168.2.2", "192.168.2.3"},
			DBName:            "configdb",
			AvailabilityRatio: 8.23,
			unexported:        "unexported",
		},
		Epoch:      1638551008,
		Epoch32:    1638551009,
		Epoch64:    1638551010,
		unexported: "unexported",
	}
}

func getDefaultCamelCaseServer() *CamelCaseServer {
	return &CamelCaseServer{
		AccessKey:         "123456",
		Normal:            "normal",
		DBName:            "configdb",
		AvailabilityRatio: 8.23,
	}
}

func getDefaultNestedServer() *NestedServer {
	return &NestedServer{
		Name: "koding",
		DatabaseOptions: Database{
			Postgres: Postgres{
				Enabled:           true,
				Port:              5432,
				Hosts:             []string{"192.168.2.1", "192.168.2.2", "192.168.2.3"},
				DBName:            "configdb",
				AvailabilityRatio: 8.23,
			},
		},
	}
}

func TestNewWithPath(t *testing.T) {
	var _ Loader = NewWithPath(testTOML)
}

func TestLoad(t *testing.T) {
	m := NewWithPath(testTOML)

	s := new(Server)
	if err := m.Load(s); err != nil {
		t.Error(err)
	}

	testStruct(t, s, getDefaultServer())
}

func TestDefaultLoader(t *testing.T) {
	m := New()
	setEnvVars(t, "Server", "")

	s := new(Server)
	if err := m.Load(s); err != nil {
		t.Error(err)
	}

	if err := m.Validate(s); err != nil {
		t.Error(err)
	}
	testStruct(t, s, getDefaultServer())

	s.Name = ""
	if err := m.Validate(s); err == nil {
		t.Error("Name should be required")
	}
}

func testStruct(t *testing.T, s *Server, d *Server) {
	t.Helper()

	if s.Name != d.Name {
		t.Errorf("Name value is wrong: %s, want: %s", s.Name, d.Name)
	}

	if s.Port != d.Port {
		t.Errorf("Port value is wrong: %d, want: %d", s.Port, d.Port)
	}

	if s.Enabled != d.Enabled {
		t.Errorf("Enabled value is wrong: %t, want: %t", s.Enabled, d.Enabled)
	}

	if s.Interval != d.Interval {
		t.Errorf("Interval value is wrong: %v, want: %v", s.Interval, d.Interval)
	}

	if s.ID != d.ID {
		t.Errorf("ID value is wrong: %v, want: %v", s.ID, d.ID)
	}

	if len(s.Labels) != len(d.Labels) {
		t.Errorf("Labels value is wrong: %d, want: %d", len(s.Labels), len(d.Labels))
	} else {
		for i, label := range d.Labels {
			if s.Labels[i] != label {
				t.Errorf("Label is wrong for index: %d, label: %d, want: %d", i, s.Labels[i], label)
			}
		}
	}

	if len(s.Users) != len(d.Users) {
		t.Errorf("Users value is wrong: %d, want: %d", len(s.Users), len(d.Users))
	} else {
		for i, user := range d.Users {
			if s.Users[i] != user {
				t.Errorf("User is wrong for index: %d, user: %s, want: %s", i, s.Users[i], user)
			}
		}
	}

	testPostgres(t, s.Postgres, d.Postgres)

	if s.Epoch != d.Epoch {
		t.Errorf("Epoch value is wrong: %v, want: %v", s.Epoch, d.Epoch)
	}

	if s.Epoch32 != d.Epoch32 {
		t.Errorf("Epoch32 value is wrong: %v, want: %v", s.Epoch32, d.Epoch32)
	}

	if s.Epoch64 != d.Epoch64 {
		t.Errorf("Epoch64 value is wrong: %v, want: %v", s.Epoch64, d.Epoch64)
	}
}

func testFlattenedStruct(t *testing.T, s *FlattenedServer, d *Server) {
	t.Helper()

	// Explicitly state that Enabled should be true, no need to check
	// `x == true` infact.
	testPostgres(t, s.Postgres, d.Postgres)
}

func testPostgres(t *testing.T, s Postgres, d Postgres) {
	t.Helper()

	if s.Enabled != d.Enabled {
		t.Errorf("Postgres enabled is wrong %t, want: %t", s.Enabled, d.Enabled)
	}

	if s.Port != d.Port {
		t.Errorf("Postgres Port value is wrong: %d, want: %d", s.Port, d.Port)
	}

	if s.DBName != d.DBName {
		t.Errorf("DBName is wrong: %s, want: %s", s.DBName, d.DBName)
	}

	if s.AvailabilityRatio != d.AvailabilityRatio {
		t.Errorf("AvailabilityRatio is wrong: %f, want: %f", s.AvailabilityRatio, d.AvailabilityRatio)
	}

	if len(s.Hosts) != len(d.Hosts) {
		// do not continue testing if this fails, because others is depending on this test
		t.Fatalf("Hosts len is wrong: %v, want: %v", s.Hosts, d.Hosts)
	}

	for i, host := range d.Hosts {
		if s.Hosts[i] != host {
			t.Fatalf("Hosts number %d is wrong: %v, want: %v", i, s.Hosts[i], host)
		}
	}
}

func testCamelcaseStruct(t *testing.T, s *CamelCaseServer, d *CamelCaseServer) {
	t.Helper()

	if s.AccessKey != d.AccessKey {
		t.Errorf("AccessKey is wrong: %s, want: %s", s.AccessKey, d.AccessKey)
	}

	if s.Normal != d.Normal {
		t.Errorf("Normal is wrong: %s, want: %s", s.Normal, d.Normal)
	}

	if s.DBName != d.DBName {
		t.Errorf("DBName is wrong: %s, want: %s", s.DBName, d.DBName)
	}

	if s.AvailabilityRatio != d.AvailabilityRatio {
		t.Errorf("AvailabilityRatio is wrong: %f, want: %f", s.AvailabilityRatio, d.AvailabilityRatio)
	}

}

func testNestedStruct(t *testing.T, s *NestedServer, d *NestedServer) {
	t.Helper()

	require.Equal(t, s.Name, d.Name)
	require.NotNil(t, d.DatabaseOptions)
	require.NotNil(t, s.DatabaseOptions)

	testPostgres(t, s.DatabaseOptions.Postgres, d.DatabaseOptions.Postgres)
}
