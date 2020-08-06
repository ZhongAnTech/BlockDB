package og

import (
	"fmt"
	"sync"

	"github.com/annchain/OG/common"
	"github.com/annchain/OG/common/crypto"
	"go.uber.org/atomic"
)

// SampleAccount 账户
type SampleAccount struct {
	PrivateKey  crypto.PrivateKey
	PublicKey   crypto.PublicKey
	Address     common.Address
	nonce       atomic.Uint64
	nonceInited bool
	mutex       sync.RWMutex
}

// NewAccount 新建账户
func NewAccount(privateKeyHex string) *SampleAccount {
	s := &SampleAccount{}
	pr, err := crypto.PrivateKeyFromString(privateKeyHex)
	if err != nil {
		fmt.Println(err)
	}
	signer := crypto.NewSigner(pr.Type)
	s.PrivateKey = pr
	s.PublicKey = signer.PubKey(pr)
	s.Address = signer.Address(s.PublicKey)
	return s
}

// ConsumeNonce 消费nonce
func (s *SampleAccount) ConsumeNonce() (uint64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !s.nonceInited {
		return 0, fmt.Errorf("nonce is not initialized. Query first")
	}
	s.nonce.Inc() /* nonce++，包装并返回 */
	return s.nonce.Load(), nil
}

// SetNonce 人为初始化nonce
func (s *SampleAccount) SetNonce(lastUsedNonce uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.nonce.Store(lastUsedNonce)
	s.nonceInited = true
}
