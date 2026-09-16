package store

import (
	"loadtest/internal/model"
)

// CreateAssertion 创建断言。
func (s *MemoryStore) CreateAssertion(a *model.Assertion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.assertions[a.ID] = a
	return nil
}

// GetAssertion 按 ID 查询断言。
func (s *MemoryStore) GetAssertion(id string) (*model.Assertion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.assertions[id]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

// ListAssertions 返回全部断言。
func (s *MemoryStore) ListAssertions() []*model.Assertion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Assertion, 0, len(s.assertions))
	for _, a := range s.assertions {
		list = append(list, a)
	}
	return list
}

// DeleteAssertion 删除断言。
func (s *MemoryStore) DeleteAssertion(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.assertions[id]; !ok {
		return ErrNotFound
	}
	delete(s.assertions, id)
	return nil
}
