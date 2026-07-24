package crypto

import "fmt"

type Key struct {
	ID       string
	Material []byte
}

type KeyRing struct {
	ActiveID string
	Keys     map[string][]byte
}

func (r KeyRing) Validate() error {
	if r.ActiveID == "" || len(r.Keys) == 0 {
		return fmt.Errorf("%w: empty keyring", ErrInvalidConfig)
	}
	if _, ok := r.Keys[r.ActiveID]; !ok {
		return fmt.Errorf("%w: active key %q missing", ErrInvalidConfig, r.ActiveID)
	}
	for id, mat := range r.Keys {
		if id == "" || len(mat) == 0 {
			return fmt.Errorf("%w: empty key entry", ErrInvalidConfig)
		}
	}
	return nil
}

func (r KeyRing) Active() (Key, error) {
	return r.Get(r.ActiveID)
}

func (r KeyRing) Get(id string) (Key, error) {
	mat, ok := r.Keys[id]
	if !ok {
		return Key{}, fmt.Errorf("%w: %s", ErrUnknownKey, id)
	}
	return Key{ID: id, Material: mat}, nil
}
