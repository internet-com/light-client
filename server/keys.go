package server

import (
	"net/http"

	"github.com/tendermint/light-client/keystore"
)

type KeyServer struct {
	store keystore.Store
}

func NewKeyStore(keydir string) *KeyServer {
	return &KeyServer{
		store: keystore.New(keydir),
	}
}

func (k *KeyServer) GenerateKey(w http.ResponseWriter, r *http.Request) {
	req := CreateKeyRequest{}
	err := readRequest(r, &req)
	if err != nil {
		writeError(w, err)
		return
	}

	err = k.store.CreateKey(req.Name, req.Passphrase)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := CreateKeyResponse{}
	writeSuccess(w, &resp)
}

func (k *KeyServer) SignMessage(w http.ResponseWriter, r *http.Request) {
	req := SignMessageRequest{}
	err := readRequest(r, &req)
	if err != nil {
		writeError(w, err)
		return
	}

	sig, pubkey, err := k.store.SignMessage(req.Data, req.KeyName, req.Passphrase)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := SignMessageResponse{
		Signature: sig,
		PubKey:    pubkey,
	}
	writeSuccess(w, &resp)
}

type CreateKeyRequest struct {
	Name       string `json:"name"`
	Passphrase string `json:"password"`
}

type CreateKeyResponse struct{}

type SignMessageRequest struct {
	KeyName    string `json:"name"`
	Passphrase string `json:"password"`
	Data       []byte `json:"data"`
}

type SignMessageResponse struct {
	Signature []byte `json:"signature"`
	PubKey    []byte `json:"pubkey"`
}
