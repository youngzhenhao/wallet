package api

import (
	"encoding/hex"
	"github.com/lightninglabs/taproot-assets/taprpc"
)

// Tapdroot Addr
type JsonResultAddr struct {
	Encoded          string `json:"encoded"`
	AssetId          string `json:"asset_id"`
	AssetType        int    `json:"asset_type"`
	Amount           int    `json:"amount"`
	GroupKey         string `json:"group_key"`
	ScriptKey        string `json:"script_key"`
	InternalKey      string `json:"internal_key"`
	TapscriptSibling string `json:"tapscript_sibling"`
	TaprootOutputKey string `json:"taproot_output_key"`
	ProofCourierAddr string `json:"proof_courier_addr"`
	AssetVersion     int    `json:"asset_version"`
	AddressVersion   int    `json:"address_version"`
	ReceiveNum       int    `json:"receive_num"`
}

func (r *JsonResultAddr) GetData(response *taprpc.Addr) {
	r.Encoded = response.Encoded
	r.AssetId = hex.EncodeToString(response.AssetId)
	r.AssetType = int(response.AssetType)
	r.Amount = int(response.Amount)
	r.GroupKey = hex.EncodeToString(response.GroupKey)
	r.ScriptKey = hex.EncodeToString(response.ScriptKey)
	r.InternalKey = hex.EncodeToString(response.InternalKey)
	r.TaprootOutputKey = hex.EncodeToString(response.TaprootOutputKey)
	r.ProofCourierAddr = response.ProofCourierAddr
	r.AssetVersion = int(response.AssetVersion)
	r.AddressVersion = int(response.AddressVersion)
}

// Tapdroot AssetTransfer
type Inputs struct {
	AnchorPoint string `json:"anchor_point"`
	Address     string `json:"address"`
	AssetID     string `json:"asset_id"`
	ScriptKey   string `json:"script_key"`
	Amount      int64  `json:"amount"`
}

type Anchor struct {
	Outpoint         string `json:"outpoint"`
	Address          string `json:"address"`
	Value            int64  `json:"value"`
	InternalKey      string `json:"internal_key"`
	TaprootAssetRoot string `json:"taproot_asset_root"`
	MerkleRoot       string `json:"merkle_root"`
	TapscriptSibling string `json:"tapscript_sibling"`
	NumPassiveAssets int    `json:"num_passive_assets"`
}

type Outputs struct {
	Anchor              Anchor `json:"anchor"`
	ScriptKey           string `json:"script_key"`
	ScriptKeyIsLocal    bool   `json:"script_key_is_local"`
	Amount              int64  `json:"amount"`
	NewProofBlob        string `json:"new_proof_blob"`
	SplitCommitRootHash string `json:"split_commit_root_hash"`
	OutputType          string `json:"output_type"`
	AssetVersion        string `json:"asset_version"`
}

type Transfer struct {
	Txid               string     `json:"txid"`
	TransferTimestamp  int64      `json:"transfer_timestamp"`
	AnchorTxHash       string     `json:"anchor_tx_hash"`
	AnchorTxHeightHint int        `json:"anchor_tx_height_hint"`
	AnchorTxChainFees  int64      `json:"anchor_tx_chain_fees"`
	Inputs             []*Inputs  `json:"inputs"`
	Outputs            []*Outputs `json:"outputs"`
}

func (r *Transfer) GetData(response *taprpc.AssetTransfer) {
	r.TransferTimestamp = response.TransferTimestamp
	r.AnchorTxHash = hex.EncodeToString(response.AnchorTxHash)
	r.AnchorTxHeightHint = int(response.AnchorTxHeightHint)
	r.AnchorTxChainFees = response.AnchorTxChainFees
	r.Txid, _ = outpointToTransactionAndIndex(response.Outputs[0].Anchor.Outpoint)
	for _, input := range response.Inputs {
		newInput := &Inputs{}
		newInput.AnchorPoint = input.AnchorPoint
		newInput.AssetID = hex.EncodeToString(input.AssetId)
		newInput.ScriptKey = hex.EncodeToString(input.ScriptKey)
		newInput.Amount = int64(input.Amount)
		r.Inputs = append(r.Inputs, newInput)
	}
	for _, output := range response.Outputs {
		newOutput := &Outputs{}
		newOutput.Anchor.Outpoint = output.Anchor.Outpoint
		newOutput.Anchor.Value = output.Anchor.Value
		newOutput.Anchor.InternalKey = hex.EncodeToString(output.Anchor.InternalKey)
		newOutput.Anchor.TaprootAssetRoot = hex.EncodeToString(output.Anchor.TaprootAssetRoot)
		newOutput.Anchor.MerkleRoot = hex.EncodeToString(output.Anchor.MerkleRoot)
		newOutput.Anchor.TapscriptSibling = hex.EncodeToString(output.Anchor.TapscriptSibling)
		newOutput.Anchor.NumPassiveAssets = int(output.Anchor.NumPassiveAssets)
		newOutput.ScriptKey = hex.EncodeToString(output.ScriptKey)
		newOutput.ScriptKeyIsLocal = output.ScriptKeyIsLocal
		newOutput.Amount = int64(output.Amount)
		newOutput.NewProofBlob = ""
		newOutput.SplitCommitRootHash = hex.EncodeToString(output.SplitCommitRootHash)
		newOutput.OutputType = output.OutputType.String()
		newOutput.AssetVersion = output.AssetVersion.String()
		r.Outputs = append(r.Outputs, newOutput)
	}
}
