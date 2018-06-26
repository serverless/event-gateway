package metadata

// Metadata stores additional information about resource.
type Metadata map[string]string

// Check checks if Metadata is compliant with all of the provided filters.
func (m Metadata) Check(filters ...Filter) bool {
	for _, filter := range filters {
		value, ok := m[filter.Key]
		if !ok || value != filter.Value {
			return false
		}
	}

	return true
}
