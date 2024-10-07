package proxmox

import (
	"reflect"
	"strconv"
	"strings"
)

// mergeIndexedString uses reflection to merge the ordinal/indexed fields
// returned by the Proxmox API. Such as Mp0..9, and Net0..9
func (cc *ContainerConfig) mergeIndexedFields(prefix string) map[string]string {
	stringMap := make(map[string]string)
	t := reflect.TypeOf(*cc)
	v := reflect.ValueOf(*cc)
	count := v.NumField()

	for i := 0; i < count; i++ {
		fn := t.Field(i).Name
		fv := v.Field(i).String()
		if fv == "" {
			continue
		}
		if strings.HasPrefix(fn, prefix) {
			// Ignore non-numeric suffixes like SCSIHW
			suffix := strings.TrimPrefix(fn, prefix)
			if _, err := strconv.Atoi(suffix); err != nil {
				continue
			}
			stringMap[strings.ToLower(fn)] = fv
		}
	}

	return stringMap
}

// MergeDevs merges and assigns the indexed Dev0..9 fields to a string map
func (cc *ContainerConfig) MergeDevs() map[string]string {
	if cc.Devs == nil {
		cc.Devs = cc.mergeIndexedFields("Dev")
	}
	return cc.Devs
}

// MergeMps merges and assigns the indexed Mp0..9 fields to a string map
func (cc *ContainerConfig) MergeMps() map[string]string {
	if cc.Mps == nil {
		cc.Mps = cc.mergeIndexedFields("Mp")
	}
	return cc.Mps
}

// MergeNets merges and assigns the indexed Net0..9 fields to a string map
func (cc *ContainerConfig) MergeNets() map[string]string {
	if cc.Nets == nil {
		cc.Nets = cc.mergeIndexedFields("Net")
	}
	return cc.Nets
}

// MergeUnuseds merges and assigns the indexed Unused0..9 fields to a string map
func (cc *ContainerConfig) MergeUnuseds() map[string]string {
	if cc.Unuseds == nil {
		cc.Unuseds = cc.mergeIndexedFields("Unused")
	}
	return cc.Unuseds
}
