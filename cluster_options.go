package proxmox

import (
	"context"
	"encoding/json"
	"reflect"
)

// Wrappers for /cluster/options — the datacenter.cfg surface. The PVE schema
// here is wide and growing (50+ keys across HA, CRS, migration, replication,
// notifications, MAC prefix, tag styles, U2F/WebAuthn, console default, …).
// We expose:
//
//   - a typed struct for the most common scalars (console, keyboard,
//     language, mac_prefix, max_workers, http_proxy, fencing, migration,
//     description, email_from, …) — round-trippable cleanly;
//
//   - an Extra map[string]any for the long tail (HA, CRS, replication,
//     notify, location, tag-style, u2f, webauthn, next-id, etc.) so callers
//     can read/write keys we haven't gone through the trouble of typing.
//     PVE's `additionalProperties: 0` means *unknown* keys are rejected on
//     PUT, but Extra is populated from the GET response so callers can read
//     whatever the server returned and pass it back as a map.
//
// The Update path accepts a typed *ClusterOptionsUpdate to keep the typed
// scalars discoverable without forcing callers to learn the Extra-map
// convention for fields they DO want to type.

// ClusterOptions reads the cluster-wide datacenter.cfg.
//
// GET /cluster/options
func (cl *Cluster) ClusterOptions(ctx context.Context) (opts *ClusterOptionsResponse, err error) {
	opts = &ClusterOptionsResponse{}
	err = cl.client.Get(ctx, "/cluster/options", opts)
	return
}

// UpdateClusterOptions mutates the datacenter.cfg. opts.Delete is a
// comma-separated list of keys to reset to their PVE defaults.
//
// PUT /cluster/options
func (cl *Cluster) UpdateClusterOptions(ctx context.Context, opts *ClusterOptionsUpdate) error {
	if opts == nil {
		opts = &ClusterOptionsUpdate{}
	}
	return cl.client.Put(ctx, "/cluster/options", opts, nil)
}

// UnmarshalJSON splits the wire payload into typed scalars and the Extra
// catch-all map so callers can read every key PVE returned without us
// enumerating the entire wide datacenter.cfg surface.
func (r *ClusterOptionsResponse) UnmarshalJSON(data []byte) error {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Decode the typed scalars by round-tripping the raw map through the
	// alias to avoid recursion into this method.
	type alias ClusterOptionsResponse
	scalar, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	tmp := alias{}
	if err := json.Unmarshal(scalar, &tmp); err != nil {
		return err
	}
	*r = ClusterOptionsResponse(tmp)

	// Extra captures every key not consumed by a typed field.
	typed := map[string]struct{}{}
	t := reflect.TypeOf(*r)
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := tag
		if c := indexComma(tag); c >= 0 {
			name = tag[:c]
		}
		typed[name] = struct{}{}
	}
	extra := map[string]any{}
	for k, v := range raw {
		if _, ok := typed[k]; ok {
			continue
		}
		var any any
		if err := json.Unmarshal(v, &any); err != nil {
			return err
		}
		extra[k] = any
	}
	if len(extra) > 0 {
		r.Extra = extra
	}
	return nil
}

// MarshalJSON merges Extra into the wire payload alongside the typed scalars.
// Typed keys take precedence over Extra (a caller mistake — typed scalar
// should be the source of truth).
func (u ClusterOptionsUpdate) MarshalJSON() ([]byte, error) {
	type alias ClusterOptionsUpdate
	scalar, err := json.Marshal(alias(u))
	if err != nil {
		return nil, err
	}
	if len(u.Extra) == 0 {
		return scalar, nil
	}
	m := map[string]json.RawMessage{}
	if err := json.Unmarshal(scalar, &m); err != nil {
		return nil, err
	}
	for k, v := range u.Extra {
		if _, exists := m[k]; exists {
			continue
		}
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		m[k] = raw
	}
	return json.Marshal(m)
}

func indexComma(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			return i
		}
	}
	return -1
}
