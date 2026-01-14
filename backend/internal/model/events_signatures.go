package model

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var EventsSignatures = []*EventSignature{
	{Name: "CollateralDeposited", Signature: "CollateralDeposited(address,address,uint256)"},
	{Name: "CollateralRedeemed", Signature: "CollateralRedeemed(address,address,address,uint256)"},
	{Name: "AUSDMinted", Signature: "AUSDMinted(address,uint256)"},
	{Name: "AUSDBurned", Signature: "AUSDBurned(address,uint256)"},
}

type CollateralDeposited struct {
	From  			common.Address
	TokenAddr 		common.Address
	Amount 			*big.Int
}

type CollateralRedeemed struct {
	From 		common.Address
	To   		common.Address
	TokenAddr 	common.Address
	Amount 		*big.Int
}

type AUSDMinted struct {
	To    	common.Address
	Amount 	*big.Int
}

type AUSDBurned struct {
	From  	common.Address
	Amount 	*big.Int
}

type EventSignature struct {
	Name, Signature string
}

func (es *EventSignature) GetName() string {
	return es.Name
}

func (es *EventSignature) GetStringSignature() string {
	return es.Signature
}

func (es *EventSignature) GetBytesSignature() []byte {
	return []byte(es.Signature)
}

func (es *EventSignature) GetHexSignature() string {
	hash := crypto.Keccak256Hash(es.GetBytesSignature())

	return hash.Hex()
}

func (es *EventSignature) MatchesHexSignature(hexSig string) bool {
	return es.GetHexSignature() == hexSig
}
