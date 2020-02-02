package cryptoengine

// Storage provides the ability to persist keys to custom locations.
type Storage interface {
	// Read
	Read(name string) ([]byte, error)
	// Write
	Write(name string, dat []byte) error

	// Delete
	Delete(name string) error
}

// StorageMemory...
type StorageMemory struct {
	keys map[string][]byte
}

// Read...
func (s *StorageMemory) Read(name string) ([]byte, error) {
	if s == nil {
		return nil, nil
	}

	if s.keys == nil {
		s.keys = make(map[string][]byte)
	}

	dat, _ := s.keys[name]

	return dat, nil
}

// Write...
func (s *StorageMemory) Write(name string, dat []byte) error {
	if s == nil {
		return nil
	}

	if s.keys == nil {
		s.keys = make(map[string][]byte)
	}

	s.keys[name] = dat

	return nil
}

// Delete...
func (s *StorageMemory) Delete(name string) error {
	if s == nil || s.keys == nil {
		return nil
	}

	delete(s.keys, name)

	return nil
}

// NewStorageMemory implements the interface Storage to store a single key in memory.
func NewStorageMemory() (*StorageMemory, error) {
	storage := &StorageMemory{
		keys: make(map[string][]byte),
	}

	return storage, nil
}
