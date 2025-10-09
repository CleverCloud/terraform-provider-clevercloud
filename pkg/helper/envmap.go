package helper

// EnvMap is a wrapper around map[string]string that provides
// a Pop method to retrieve and remove keys in one operation.
// This is useful for parsing environment variables where you want
// to extract known keys and keep the rest.
type EnvMap struct {
	All map[string]string
}

// NewEnvMap creates a new EnvMap from a map[string]string.
// The input map is copied to avoid modifying the original.
func NewEnvMap(m map[string]string) *EnvMap {
	copied := make(map[string]string, len(m))
	for k, v := range m {
		copied[k] = v
	}
	return &EnvMap{All: copied}
}

// Pop retrieves the value for the given key and removes it from the map.
// Returns an empty string if the key doesn't exist.
func (em *EnvMap) Pop(key string) string {
	value := em.All[key]
	delete(em.All, key)
	return value
}

// Size returns the number of remaining entries in the map.
func (em *EnvMap) Size() int {
	return len(em.All)
}
