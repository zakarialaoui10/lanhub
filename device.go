package lanhub

import "time"

type Device struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Platform    string         `json:"platform"`
	Connected   bool           `json:"connected"`
	ConnectedAt time.Time      `json:"connectedAt"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

func (d *Device) Clone() *Device {
	if d == nil { return nil }
	out := *d
	if d.Metadata != nil {
		out.Metadata = make(map[string]any, len(d.Metadata))
		for k, v := range d.Metadata { out.Metadata[k] = v }
	}
	return &out
}
