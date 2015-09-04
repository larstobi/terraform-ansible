package tory

import (
	"encoding/json"
	"regexp"
	"strings"
	"sync"
)

var (
	groupNameUnsafe = regexp.MustCompile("[^-A-Za-z0-9]")
)

// Inventory is a struct containing host information.
type Inventory struct {
	Meta       *meta `json:"_meta"`
	groups     map[string][]string
	groupMutex *sync.Mutex
}

// NewInventory returns an inventory struct including a meta struct.
func NewInventory() *Inventory {
	return &Inventory{
		Meta:       newMeta(),
		groups:     map[string][]string{},
		groupMutex: &sync.Mutex{},
	}
}

// GetGroup gets a group by name.
func (inv *Inventory) GetGroup(group string) []string {
	inv.groupMutex.Lock()
	defer inv.groupMutex.Unlock()

	if g, ok := inv.groups[group]; ok {
		return g
	}

	return nil
}

// AddHostnameToGroup adds a hostname to a group and sanitizes it.
func (inv *Inventory) AddHostnameToGroup(group, hostname string) {
	sanitizedGroup := groupNameUnsafe.ReplaceAllString(strings.ToLower(group), "_")
	sanitizedGroup = strings.Replace(sanitizedGroup, ".", "_", -1)
	inv.AddHostnameToGroupUnsanitized(sanitizedGroup, hostname)
}

// AddHostnameToGroupUnsanitized adds a hostname to a group without touching it.
func (inv *Inventory) AddHostnameToGroupUnsanitized(group, hostname string) {
	inv.groupMutex.Lock()
	defer inv.groupMutex.Unlock()

	if _, ok := inv.groups[group]; !ok {
		inv.groups[group] = []string{}
	}
	inv.groups[group] = append(inv.groups[group], hostname)
}

// MarshalJSON does exactly that.
func (inv *Inventory) MarshalJSON() ([]byte, error) {
	serialized := map[string]interface{}{}
	serialized["_meta"] = inv.Meta
	for key, value := range inv.groups {
		serialized[key] = value
	}

	return json.MarshalIndent(serialized, "", "    ")
}

// UnmarshalJSON does exacltly that.
func (inv *Inventory) UnmarshalJSON(b []byte) error {
	raw := map[string]json.RawMessage{}
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	for key, value := range raw {
		if key == "_meta" {
			m := &meta{}
			err := json.Unmarshal(value, m)
			if err != nil {
				return err
			}
			inv.Meta = m
		} else {
			group := []string{}
			err := json.Unmarshal(value, &group)
			if err == nil {
				for _, hostname := range group {
					inv.AddHostnameToGroupUnsanitized(key, hostname)
				}
			}
		}
	}

	return nil
}
