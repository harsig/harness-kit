package yamlconfig

import (
	"os"
	"path/filepath"
	"testing"
)

type config struct {
	Name  string `yaml:"name"`
	Count int    `yaml:"count"`
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    config
		wantErr bool
	}{
		{name: "valid", data: "name: a\ncount: 2\n", want: config{Name: "a", Count: 2}},
		{name: "empty", data: "", want: config{}},
		{name: "comment only", data: "# nothing\n", want: config{}},
		{name: "unknown key", data: "nmae: a\n", wantErr: true},
		{name: "wrong type", data: "count: x\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got config
			err := Decode([]byte(tt.data), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.yml")
	if err := os.WriteFile(path, []byte("name: a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var c config
	got, err := Load(path, &c)
	if err != nil {
		t.Fatal(err)
	}
	if got != dir || c.Name != "a" {
		t.Errorf("Load = %q, %+v; want %q, name a", got, c, dir)
	}
	if _, err := Load(filepath.Join(dir, "missing.yml"), &c); err == nil {
		t.Error("missing file: no error")
	}
}

func TestResolvePath(t *testing.T) {
	t.Setenv("HOME", "/home/u")
	tests := []struct{ p, want string }{
		{p: "a/b", want: "/base/a/b"},
		{p: "./a/../b", want: "/base/b"},
		{p: "/abs", want: "/abs"},
		{p: "~", want: "/home/u"},
		{p: "~/x", want: "/home/u/x"},
		{p: "~x", want: "/base/~x"},
	}
	for _, tt := range tests {
		if got := ResolvePath("/base", tt.p); got != tt.want {
			t.Errorf("ResolvePath(%q) = %q, want %q", tt.p, got, tt.want)
		}
	}
}
