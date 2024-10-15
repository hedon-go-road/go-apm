package dogapm

import (
	"google.golang.org/grpc/metadata"
)

type metadataSupplier struct {
	metadata *metadata.MD
}

// Get returns the value associated with the passed key.
func (m *metadataSupplier) Get(key string) string {
	values := m.metadata.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Set stores the key-value pair.
func (m *metadataSupplier) Set(key, value string) {
	m.metadata.Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (m *metadataSupplier) Keys() []string {
	keys := make([]string, 0, len(*m.metadata))
	for key := range *m.metadata {
		keys = append(keys, key)
	}
	return keys
}
