package pkg

// EnvMap is a wrapper around map[string]string that provides helper methods
// for accessing environment variables with pointer semantics.
type EnvMap map[string]string

// Get returns a pointer to the value associated with key, or nil if not found.
// This allows for cleaner null-handling in FromEnv implementations.
func (m EnvMap) Get(key string) *string {
	if val, ok := m[key]; ok {
		return &val
	}
	return nil
}
