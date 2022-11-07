package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

type ModelPool struct {
	pendingMux   sync.RWMutex
	pending      map[string]*modelList
	lastModelHex string
}

func NewModelPool() *ModelPool {
	return &ModelPool{
		pendingMux: sync.RWMutex{},
		pending:    make(map[string]*modelList),
	}
}

type modelList struct {
	hex string
	cnt int
}

func (m *ModelPool) Clear() {
	m.pending = make(map[string]*modelList)
}

func (m *ModelPool) GetModelId(h string) (string, error) {
	md := sha256.New()
	decodeString, err := hex.DecodeString(h)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString err: %v", err)
	}

	md.Write(decodeString)

	expectedCryptogram := md.Sum(nil)

	return string(expectedCryptogram), nil
}

func (m *ModelPool) AddModel(h string) error {
	k, err := m.GetModelId(h)
	if err != nil {
		return fmt.Errorf("m.getMapKey err: %v", err)
	}

	m.pendingMux.Lock()
	defer m.pendingMux.Unlock()

	l, ok := m.pending[k]
	if !ok {
		l = &modelList{
			hex: h,
			cnt: 0,
		}
		m.pending[k] = l
	}
	l.cnt++
	return nil
}

func (m *ModelPool) GetGreatestModel() string {
	m.pendingMux.RLock()
	defer m.pendingMux.RUnlock()
	var result *modelList
	for _, l := range m.pending {
		if result == nil || result.cnt < l.cnt {
			result = l
		}
	}

	return result.hex
}

func (m *ModelPool) SetLastModelHex(lastModelHex string) {
	m.lastModelHex = lastModelHex
}

func (m *ModelPool) GetLastModelHex() string {
	return m.lastModelHex
}
