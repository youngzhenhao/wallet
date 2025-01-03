package api

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/lightninglabs/taproot-assets/proof"
	"github.com/lightninglabs/taproot-assets/taprpc/universerpc"
	"github.com/wallet/base"
	"github.com/wallet/service/apiConnect"
	"github.com/wallet/service/rpcclient"
)

func AddFederationServer() {}

func assetLeafKeys(id string, proofType universerpc.ProofType) (*universerpc.AssetLeafKeyResponse, error) {
	conn, clearUp, err := apiConnect.GetConnection("tapd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	client := universerpc.NewUniverseClient(conn)
	request := &universerpc.AssetLeafKeysRequest{
		Id: &universerpc.ID{
			Id: &universerpc.ID_AssetIdStr{
				AssetIdStr: id,
			},
			ProofType: proofType,
		},
		//Offset:    0,
		//Limit:     0,
		//Direction: 0,
	}
	response, err := client.AssetLeafKeys(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func AssetLeafKeysAndGetResponse(assetId string, proofType universerpc.ProofType) (*universerpc.AssetLeafKeyResponse, error) {
	return assetLeafKeys(assetId, proofType)
}

func AssetLeafKeys(id string, proofType string) string {
	var _proofType universerpc.ProofType
	if proofType == "issuance" || proofType == "ISSUANCE" || proofType == "PROOF_TYPE_ISSUANCE" {
		_proofType = universerpc.ProofType_PROOF_TYPE_ISSUANCE
	} else if proofType == "transfer" || proofType == "TRANSFER" || proofType == "PROOF_TYPE_TRANSFER" {
		_proofType = universerpc.ProofType_PROOF_TYPE_TRANSFER
	} else {
		_proofType = universerpc.ProofType_PROOF_TYPE_UNSPECIFIED
	}
	response, err := assetLeafKeys(id, _proofType)
	if err != nil {
		return MakeJsonErrorResult(assetLeafKeysErr, err.Error(), nil)
	}
	if len(response.AssetKeys) == 0 {
		return MakeJsonErrorResult(responseAssetKeysZeroErr, "Result length is zero.", nil)
	}
	return MakeJsonErrorResult(SUCCESS, "", processAssetKey(response))
}

type AssetKey struct {
	OpStr          string `json:"op_str"`
	ScriptKeyBytes string `json:"script_key_bytes"`
}

func processAssetKey(response *universerpc.AssetLeafKeyResponse) *[]AssetKey {
	var assetKey []AssetKey
	for _, keys := range response.AssetKeys {
		assetKey = append(assetKey, AssetKey{
			OpStr:          keys.GetOpStr(),
			ScriptKeyBytes: hex.EncodeToString(keys.GetScriptKeyBytes()),
		})
	}
	return &assetKey
}

func AssetLeaves(id string) string {
	response, err := assetLeaves(false, id, universerpc.ProofType_PROOF_TYPE_ISSUANCE)
	if err != nil {
		return MakeJsonErrorResult(assetLeavesErr, err.Error(), nil)
	}

	if response.Leaves == nil {
		return MakeJsonErrorResult(responseLeavesNullErr, "NOT_FOUND", nil)
	}

	return MakeJsonErrorResult(SUCCESS, "", response)
}

func GetAssetInfo(id string) string {
	root := rpcclient.QueryAssetRoots(id)
	if root == nil || root.IssuanceRoot.Id == nil {
		return MakeJsonErrorResult(QueryAssetRootsErr, "NOT_FOUND", nil)
	}
	queryId := id
	isGroup := false
	if groupKey, ok := root.IssuanceRoot.Id.Id.(*universerpc.ID_GroupKey); ok {
		isGroup = true
		queryId = hex.EncodeToString(groupKey.GroupKey)
	}
	response, err := assetLeaves(isGroup, queryId, universerpc.ProofType_PROOF_TYPE_ISSUANCE)
	if err != nil {
		return MakeJsonErrorResult(assetLeavesErr, err.Error(), nil)
	}
	if response.Leaves == nil {
		return MakeJsonErrorResult(responseLeavesNullErr, "NOT_FOUND", nil)
	}
	var blob proof.Blob
	for index, leaf := range response.Leaves {
		if hex.EncodeToString(leaf.Asset.AssetGenesis.GetAssetId()) == id {
			blob = response.Leaves[index].Proof
			break
		}
	}
	if len(blob) == 0 {
		return MakeJsonErrorResult(blobLenZeroErr, "NOT_FOUND", nil)
	}
	p, _ := blob.AsSingleProof()
	assetId := p.Asset.ID().String()
	assetName := p.Asset.Tag
	assetPoint := p.Asset.FirstPrevOut.String()
	assetType := p.Asset.Type.String()
	amount := p.Asset.Amount
	createHeight := p.BlockHeight
	createTime := p.BlockHeader.Timestamp
	var (
		newMeta Meta
		m       = ""
	)
	if p.MetaReveal != nil {
		m = string(p.MetaReveal.Data)
	}
	newMeta.GetMetaFromStr(m)
	var assetInfo = struct {
		AssetId      string  `json:"asset_Id"`
		Name         string  `json:"name"`
		Point        string  `json:"point"`
		AssetType    string  `json:"assetType"`
		GroupName    *string `json:"group_name"`
		GroupKey     *string `json:"group_key"`
		Amount       uint64  `json:"amount"`
		Meta         *string `json:"meta"`
		CreateHeight int64   `json:"create_height"`
		CreateTime   int64   `json:"create_time"`
		Universe     string  `json:"universe"`
	}{
		AssetId:      assetId,
		Name:         assetName,
		Point:        assetPoint,
		AssetType:    assetType,
		GroupName:    &newMeta.GroupName,
		Amount:       amount,
		Meta:         &newMeta.Description,
		CreateHeight: int64(createHeight),
		CreateTime:   createTime.Unix(),
		Universe:     "localhost",
	}
	if isGroup {
		assetInfo.GroupKey = &queryId
	}
	return MakeJsonErrorResult(SUCCESS, "", assetInfo)
}

func AssetRoots() {}

func DeleteAssetRoot() {}

func DeleteFederationServer() {}

// UniverseInfo
//
//	@Description: Info returns a set of information about the current state of the Universe.
//	@return string
func UniverseInfo() string {
	conn, clearUp, err := apiConnect.GetConnection("tapd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()

	client := universerpc.NewUniverseClient(conn)
	request := &universerpc.InfoRequest{}
	response, err := client.Info(context.Background(), request)
	if err != nil {
		return MakeJsonErrorResult(clientInfoErr, err.Error(), nil)
	}
	return MakeJsonErrorResult(SUCCESS, "", response)
}

func InsertProof() {}

// ListFederationServers
//
//	@Description: ListFederationServers lists the set of servers that make up the federation of the local Universe server.
//	This servers are used to push out new proofs, and also periodically call sync new proofs from the remote server.
//	@return string
func ListFederationServers() string {
	conn, clearUp, err := apiConnect.GetConnection("tapd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	client := universerpc.NewUniverseClient(conn)
	request := &universerpc.ListFederationServersRequest{}
	response, err := client.ListFederationServers(context.Background(), request)
	if err != nil {
		fmt.Printf("%s universerpc ListFederationServers Error: %v\n", GetTimeNow(), err)
		return ""
	}
	return response.String()
}

func MultiverseRoot() {}

func QueryAssetRoots(id string) string {
	response, err := queryAssetRoot(id)
	if err != nil {
		return MakeJsonErrorResult(queryAssetRootErr, err.Error(), nil)
	}
	return MakeJsonErrorResult(SUCCESS, "", response)
}

func QueryAssetStats(assetId string) string {
	response, err := queryAssetStats(assetId)
	if err != nil {
		return MakeJsonErrorResult(queryAssetStatsErr, err.Error(), "")
	}
	return MakeJsonErrorResult(SUCCESS, "", response)
}

func QueryEvents() {}

func QueryFederationSyncConfig() {}

func QueryProof() {}

func SetFederationSyncConfig() {}

func SyncUniverse(universeHost string, assetId string) string {
	var targets []*universerpc.SyncTarget
	universeID := &universerpc.ID{
		Id: &universerpc.ID_AssetIdStr{
			AssetIdStr: assetId,
		},
		ProofType: universerpc.ProofType_PROOF_TYPE_ISSUANCE,
	}
	if universeID != nil {
		targets = append(targets, &universerpc.SyncTarget{
			Id: universeID,
		})
	}
	var defaultHost string
	switch base.NetWork {
	case base.UseMainNet:
		defaultHost = "universe.lightning.finance:10029"
	case base.UseTestNet:
		defaultHost = "testnet.universe.lightning.finance:10029"
	}
	if universeHost == "" {
		universeHost = defaultHost
	}
	response, err := syncUniverse(universeHost, targets, universerpc.UniverseSyncMode_SYNC_FULL)
	if err != nil {
		return MakeJsonErrorResult(syncUniverseErr, err.Error(), "")
	}
	return MakeJsonErrorResult(SUCCESS, "", response)
}

func SyncUniverseByGroup(universeHost string, groupKey string) string {
	var targets []*universerpc.SyncTarget
	universeID := &universerpc.ID{
		Id: &universerpc.ID_GroupKeyStr{
			GroupKeyStr: groupKey,
		},
		ProofType: universerpc.ProofType_PROOF_TYPE_ISSUANCE,
	}
	if universeID != nil {
		targets = append(targets, &universerpc.SyncTarget{
			Id: universeID,
		})
	}
	var defaultHost string
	switch base.NetWork {
	case base.UseMainNet:
		defaultHost = "universe.lightning.finance:10029"
	case base.UseTestNet:
		defaultHost = "testnet.universe.lightning.finance:10029"
	}
	if universeHost == "" {
		universeHost = defaultHost
	}
	response, err := syncUniverse(universeHost, targets, universerpc.UniverseSyncMode_SYNC_ISSUANCE_ONLY)
	if err != nil {
		return MakeJsonErrorResult(syncUniverseErr, err.Error(), "")
	}
	return MakeJsonErrorResult(SUCCESS, "", response)
}

func UniverseStats() {}

func queryAssetRoot(id string) (*universerpc.QueryRootResponse, error) {
	conn, clearUp, err := apiConnect.GetConnection("tapd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()

	requst := &universerpc.AssetRootQuery{
		Id: &universerpc.ID{
			Id: &universerpc.ID_AssetIdStr{
				AssetIdStr: id,
			},
		},
	}
	client := universerpc.NewUniverseClient(conn)
	response, err := client.QueryAssetRoots(context.Background(), requst)
	return response, err
}

func assetLeaves(isGroup bool, id string, proofType universerpc.ProofType) (*universerpc.AssetLeafResponse, error) {
	conn, clearUp, err := apiConnect.GetConnection("tapd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	request := &universerpc.ID{
		ProofType: proofType,
	}

	if isGroup {
		groupKey := &universerpc.ID_GroupKeyStr{
			GroupKeyStr: id,
		}
		request.Id = groupKey
	} else {
		AssetId := &universerpc.ID_AssetIdStr{
			AssetIdStr: id,
		}
		request.Id = AssetId
	}

	client := universerpc.NewUniverseClient(conn)
	response, err := client.AssetLeaves(context.Background(), request)
	return response, err
}

func AssetLeavesAndGetResponse(isGroup bool, id string, proofType universerpc.ProofType) (*universerpc.AssetLeafResponse, error) {
	return assetLeaves(isGroup, id, proofType)
}

func queryAssetStats(assetId string) (*universerpc.UniverseAssetStats, error) {
	conn, clearUp, err := apiConnect.GetConnection("tapd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	id, err := hex.DecodeString(assetId)
	client := universerpc.NewUniverseClient(conn)
	request := &universerpc.AssetStatsQuery{
		AssetIdFilter: id,
	}
	response, err := client.QueryAssetStats(context.Background(), request)
	return response, err
}

func syncUniverse(universeHost string, syncTargets []*universerpc.SyncTarget, syncMode universerpc.UniverseSyncMode) (*universerpc.SyncResponse, error) {
	conn, clearUp, err := apiConnect.GetConnection("tapd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	request := &universerpc.SyncRequest{
		UniverseHost: universeHost,
		SyncMode:     syncMode,
		SyncTargets:  syncTargets,
	}
	client := universerpc.NewUniverseClient(conn)
	response, err := client.SyncUniverse(context.Background(), request)
	return response, err
}
