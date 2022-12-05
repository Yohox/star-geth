package star

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
	"io/ioutil"
	"net/http"
)

type Backend struct {
	host string
}

func NewBackend(host string, blockchain *core.BlockChain) *Backend {
	backend := &Backend{host: host}
	chainHeadCh := make(chan core.ChainHeadEvent)
	blockchain.SubscribeChainHeadEvent(chainHeadCh)
	go func() {
		for {
			select {
			case chanHeadEvent := <-chainHeadCh:
				err := backend.NewGlobalModel(hex.EncodeToString(chanHeadEvent.Block.Extra()))
				if err != nil {
					log.Error(fmt.Sprintf("backend.NewGlobalModel err: %v", err))
				}
			}
		}
	}()
	return backend
}

func (b *Backend) request(path string, param map[string]interface{}) error {
	bs, err := json.Marshal(param)
	if err != nil {
		return fmt.Errorf("json.Marshal err: %v", err)
	}
	reader := bytes.NewReader(bs)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", b.host, path), reader)
	if err != nil {
		return fmt.Errorf("http.NewRequest err: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.DefaultClient.Do err: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http.statusCode == %d", resp.StatusCode)
	}

	return nil
}

func (b *Backend) Register(addr *common.Address, dataSize uint) error {
	s := map[string]interface{}{
		"address":   addr.Hex(),
		"data_size": dataSize,
	}
	err := b.request("register", s)
	if err != nil {
		return fmt.Errorf("b.request err: %v", err)
	}

	return nil
}

func (b *Backend) NewLocalModel(addr *common.Address, modelStateHex string) error {
	s := map[string]interface{}{
		"local_model_hex": modelStateHex,
	}

	err := b.request("newLocalModel/"+addr.Hex(), s)
	if err != nil {
		return fmt.Errorf("b.requestStarBackend err: %v", err)
	}

	return nil
}

func (b *Backend) NewGlobalModel(modelStateHex string) error {
	s := map[string]interface{}{
		"global_model_hex": modelStateHex,
	}

	err := b.request("newGlobalModel", s)
	if err != nil {
		return fmt.Errorf("b.requestStarBackend err: %v", err)
	}

	return nil
}

func (b *Backend) GetTrainInfo() (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", b.host, "getTrainInfo"), nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest err: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http.DefaultClient.Do err: %v", err)
	}
	bs, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll err: %v", err)
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(bs, &data)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %v", err)
	}

	return data, nil
}
