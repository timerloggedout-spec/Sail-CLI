package domain_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/E-Timileyin/sail/internal/domain"
)

func TestServerAddress(t *testing.T) {
	tests := []struct {
		name string
		srv  domain.Server
		want string
	}{
		{"default port", domain.Server{Host: "example.com"}, "example.com:22"},
		{"explicit port", domain.Server{Host: "example.com", Port: 2222}, "example.com:2222"},
		{"ip host", domain.Server{Host: "10.0.0.8", Port: 2200}, "10.0.0.8:2200"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.srv.Address(); got != tt.want {
				t.Errorf("Address() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestServerPortString(t *testing.T) {
	if got := (&domain.Server{}).PortString(); got != "22" {
		t.Errorf("PortString() with zero port = %q, want %q", got, "22")
	}
	if got := (&domain.Server{Port: 8022}).PortString(); got != "8022" {
		t.Errorf("PortString() = %q, want %q", got, "8022")
	}
}

func TestServerValidate(t *testing.T) {
	valid := domain.Server{
		Name: "production", Host: "example.com", User: "deploy", KeyPath: "/home/deploy/.ssh/id_ed25519",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() on a complete server: %v", err)
	}

	tests := []struct {
		name    string
		mutate  func(*domain.Server)
		wantSub string
	}{
		{"missing name", func(s *domain.Server) { s.Name = "" }, "name is required"},
		{"missing host", func(s *domain.Server) { s.Host = "" }, "host is required"},
		{"missing user", func(s *domain.Server) { s.User = "" }, "user is required"},
		{"missing key path", func(s *domain.Server) { s.KeyPath = "" }, "key_path is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := valid
			tt.mutate(&srv)
			err := srv.Validate()
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tt.wantSub)
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("Validate() error = %q, want substring %q", err, tt.wantSub)
			}
		})
	}
}

func TestDeploymentValidate(t *testing.T) {
	valid := domain.Deployment{App: "web", Image: "ghcr.io/acme/web", Tag: "9f31c2e"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() on a complete deployment: %v", err)
	}

	tests := []struct {
		name    string
		mutate  func(*domain.Deployment)
		wantSub string
	}{
		{"missing app", func(d *domain.Deployment) { d.App = "" }, "app is required"},
		{"missing image", func(d *domain.Deployment) { d.Image = "" }, "image is required"},
		{"missing tag", func(d *domain.Deployment) { d.Tag = "" }, "tag is required"},
		{"latest tag rejected", func(d *domain.Deployment) { d.Tag = "latest" }, "\"latest\" is not allowed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := valid
			tt.mutate(&dep)
			err := dep.Validate()
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tt.wantSub)
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("Validate() error = %q, want substring %q", err, tt.wantSub)
			}
		})
	}
}

func TestDeploymentImageRef(t *testing.T) {
	d := domain.Deployment{Image: "ghcr.io/acme/web", Tag: "9f31c2e"}
	if got := d.ImageRef(); got != "ghcr.io/acme/web:9f31c2e" {
		t.Errorf("ImageRef() = %q, want %q", got, "ghcr.io/acme/web:9f31c2e")
	}
}

// TestDualStructTags guards the v0.1.0 regression: Viper reads mapstructure,
// not yaml, so every field of the config-facing structs must carry both tags
// or multi-word keys (key_path) silently unmarshal empty.
func TestDualStructTags(t *testing.T) {
	for _, typ := range []reflect.Type{
		reflect.TypeOf(domain.Server{}),
		reflect.TypeOf(domain.Deployment{}),
	} {
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if _, ok := f.Tag.Lookup("yaml"); !ok {
				t.Errorf("%s.%s: missing yaml tag", typ.Name(), f.Name)
			}
			if _, ok := f.Tag.Lookup("mapstructure"); !ok {
				t.Errorf("%s.%s: missing mapstructure tag", typ.Name(), f.Name)
			}
		}
	}
}

func TestTrustPolicyZeroValueIsStrict(t *testing.T) {
	var p domain.TrustPolicy
	if p != domain.Strict {
		t.Errorf("zero TrustPolicy = %v, want Strict (the documented zero value)", p)
	}
	if domain.AcceptNew == domain.Strict {
		t.Error("AcceptNew must differ from Strict")
	}
	srv := domain.Server{}
	if srv.TrustPolicy != domain.Strict {
		t.Errorf("zero Server.TrustPolicy = %v, want Strict", srv.TrustPolicy)
	}
}
