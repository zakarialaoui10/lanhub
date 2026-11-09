package lanhub

import "testing"

func TestDefaultOptions(t *testing.T) {
	o := DefaultOptions()
	if o.Host != "0.0.0.0" { t.Fatalf("unexpected host: %s", o.Host) }
	if o.Port != 8080 { t.Fatalf("unexpected port: %d", o.Port) }
	if o.Path != "/socket.io/" { t.Fatalf("unexpected path: %s", o.Path) }
}

func TestOptionsValidation(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		wantErr bool
	}{
		{"valid", DefaultOptions(), false},
		{"invalid port", Options{Host: "0.0.0.0", Port: 70000, Path: "/socket.io/"}, true},
		{"invalid path", Options{Host: "0.0.0.0", Port: 8080, Path: "socket.io"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr { t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr) }
		})
	}
}

func TestNewAppliesDefaults(t *testing.T) {
	s := New(Options{Port: 9090})
	if s.opts.Host != "0.0.0.0" || s.opts.Port != 9090 || s.opts.Path != "/socket.io/" {
		t.Fatalf("defaults not applied: %#v", s.opts)
	}
}
