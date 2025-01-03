package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lightninglabs/taproot-assets/taprpc"
	"github.com/lightninglabs/taproot-assets/taprpc/mintrpc"
	"github.com/vincent-petithory/dataurl"
	"github.com/wallet/service/apiConnect"
	"os"
	"path/filepath"
	"strings"
)

// CancelBatch
//
//	@Description: CancelBatch will attempt to cancel the current pending batch.
//	@return bool
func CancelBatch() bool {
	conn, clearUp, err := apiConnect.GetConnection("tapd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	client := mintrpc.NewMintClient(conn)
	request := &mintrpc.CancelBatchRequest{}
	response, err := client.CancelBatch(context.Background(), request)
	if err != nil {
		fmt.Printf("%s mintrpc CancelBatch Error: %v\n", GetTimeNow(), err)
		return false
	}
	fmt.Printf("%s %v\n", GetTimeNow(), response)
	return true
}

// FinalizeBatch
//
//	@Description: Wraps the finalizeBatch. FinalizeBatch will attempt to finalize the current pending batch.
//	@param shortResponse
//	@param feeRate
//	@return bool
func FinalizeBatch(feeRate int, token string, deviceId string) string {
	if feeRate > FeeRateSatPerBToSatPerKw(500) {
		err := errors.New("fee rate exceeds max(500)")
		return MakeJsonErrorResult(FeeRateExceedMaxErr, err.Error(), nil)
	}
	return finalizeBatch(false, feeRate, token, deviceId)
}

// ListBatches
//
//	@Description: ListBatches lists the set of batches submitted to the daemon, including pending and cancelled batches.
//	@return string
func ListBatches() string {
	conn, clearUp, err := apiConnect.GetConnection("tapd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	client := mintrpc.NewMintClient(conn)
	request := &mintrpc.ListBatchRequest{}
	response, err := client.ListBatches(context.Background(), request)
	if err != nil {
		fmt.Printf("%s mintrpc ListBatches Error: %v\n", GetTimeNow(), err)
		return ""
	}
	return response.String()
}

// MintAsset
//
//	@Description: Wraps the mintAsset, omitting most of the parameters and making them default values.
//	MintAsset will attempt to mint the set of assets (async by default to ensure proper batching) specified in the request.
//	The pending batch is returned that shows the other pending assets that are part of the next batch.
//	This call will block until the operation succeeds (asset is staged in the batch) or fails.
func MintAsset(name string, assetTypeIsCollectible bool, assetMetaData *Meta, amount int, newGroupedAsset bool) string {
	Metastr := assetMetaData.ToJsonStr()
	return mintAsset(false, assetTypeIsCollectible, name, Metastr, false, amount, newGroupedAsset, false, "", "", false)
}

func AddGroupAsset(name string, assetTypeIsCollectible bool, assetMetaData *Meta, amount int, groupKey string) string {
	Metastr := assetMetaData.ToJsonStr()
	return mintAsset(false, assetTypeIsCollectible, name, Metastr, false, amount, false, true, groupKey, "", false)

}

type Meta struct {
	Acronym     string `json:"acronym,omitempty"`
	Description string `json:"description,omitempty"`
	ImageData   string `json:"image_data,omitempty"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	GroupName   string `json:"groupName,omitempty"`
}

func NewMeta(description string) *Meta {
	meta := Meta{
		Description: description,
	}
	return &meta
}

func (m *Meta) LoadImageByByte(image []byte) (bool, error) {
	if len(image) == 0 {
		fmt.Println("image data is nil")
		return false, fmt.Errorf("image data is nil")
	}
	imageStr := dataurl.EncodeBytes(image)
	m.ImageData = imageStr
	return true, nil
}

func (m *Meta) LoadImage(file string) (bool, error) {
	if file != "" {
		image, err := os.ReadFile(file)
		if err != nil {
			fmt.Println("open image file is error:", err)
			return false, err
		}
		imageStr := dataurl.EncodeBytes(image)
		m.ImageData = imageStr
	}
	return true, nil
}

func (m *Meta) ToJsonStr() string {
	metastr, _ := json.Marshal(m)
	return string(metastr)
}

func (m *Meta) GetMetaFromStr(metaStr string) {
	if metaStr == "" {
		m.Description = "This asset has no meta."
		return
	}

	first := metaStr[:1]
	end := metaStr[len(metaStr)-1:]
	var s string
	if first == "\"" && end == "\"" {
		s = metaStr[1 : len(metaStr)-1]
	} else {
		s = metaStr
	}
	err := json.Unmarshal([]byte(s), m)
	if err != nil {
		m.Description = s
	}
}

func (m *Meta) SaveImage(dir string, name string) bool {
	if m.ImageData == "" {
		return false
	}
	dataUrl, err := dataurl.DecodeString(m.ImageData)
	if err != nil {
		return false
	}
	ContentType := dataUrl.MediaType.ContentType()
	datatype := strings.Split(ContentType, "/")
	if datatype[0] != "image" {
		fmt.Println("is not image dataurl")
		return false
	}
	formatName := strings.Split(name, ".")
	file := filepath.Join(dir, formatName[0]+"."+datatype[1])
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("create new image error:", err)
		return false
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(f)
	_, err = f.Write(dataUrl.Data)
	if err != nil {
		fmt.Println("Write data fail:", err)
		return false
	}
	return true
}

func (m *Meta) GetImage() []byte {
	if m.ImageData == "" {
		return nil
	}
	dataUrl, err := dataurl.DecodeString(m.ImageData)
	if err != nil {
		return nil
	}
	ContentType := dataUrl.MediaType.ContentType()
	datatype := strings.Split(ContentType, "/")
	if datatype[0] != "image" {
		fmt.Println("is not image dataurl")
		return nil
	}
	return dataUrl.Data
}

func (m *Meta) FetchAssetMeta(isHash bool, data string) string {
	response, err := fetchAssetMeta(isHash, data)
	if err != nil {
		return MakeJsonErrorResult(fetchAssetMetaErr, err.Error(), nil)
	}
	m.GetMetaFromStr(string(response.Data))
	return MakeJsonErrorResult(SUCCESS, "", nil)
}

// finalizeBatch
//
//	@Description: FinalizeBatch will attempt to finalize the current pending batch.
//	@param shortResponse
//	@param feeRate
//	@return string
func finalizeBatch(shortResponse bool, feeRate int, token string, deviceId string) string {
	conn, clearUp, err := apiConnect.GetConnection("tapd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	client := mintrpc.NewMintClient(conn)
	request := &mintrpc.FinalizeBatchRequest{
		ShortResponse: shortResponse,
		FeeRate:       uint32(feeRate),
	}
	response, err := client.FinalizeBatch(context.Background(), request)
	if err != nil {
		fmt.Printf("%s mintrpc FinalizeBatch Error: %v\n", GetTimeNow(), err)
		return MakeJsonErrorResult(FinalizeBatchErr, err.Error(), nil)
	}
	// @dev: Upload asset local mint
	err = UploadAssetLocalMints(token, deviceId, response)
	if err != nil {
		LogError("", err)
		// @dev: Do not return
	}
	// @Description: Set group first asset meta
	err = SetGroupFirstAssetMeta(token, deviceId, response)
	if err != nil {
		LogError("", err)
		// @dev: Do not return
	}
	return MakeJsonErrorResult(SUCCESS, "", FinalizeBatchResponseToPendingBatch(response))
}

// mintAsset
//
//	@Description:MintAsset will attempt to mint the set of assets (async by default to ensure proper batching) specified in the request.
//	The pending batch is returned that shows the other pending assets that are part of the next batch. This call will block until the operation succeeds (asset is staged in the batch) or fails.
//	@param assetVersionIsV1
//	@param assetTypeIsCollectible
//	@param name
//	@param assetMetaData
//	@param AssetMetaTypeIsJsonNotOpaque
//	@param amount
//	@param newGroupedAsset
//	@param groupedAsset
//	@param groupKey
//	@param groupAnchor
//	@param shortResponse
//	@return bool
func mintAsset(assetVersionIsV1 bool, assetTypeIsCollectible bool, name string, assetMetaData string, AssetMetaTypeIsJsonNotOpaque bool, amount int, newGroupedAsset bool, groupedAsset bool, groupKey string, groupAnchor string, shortResponse bool) string {
	conn, clearUp, err := apiConnect.GetConnection("tapd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	client := mintrpc.NewMintClient(conn)
	var _assetVersion taprpc.AssetVersion
	if assetVersionIsV1 {
		_assetVersion = taprpc.AssetVersion_ASSET_VERSION_V1
	} else {
		_assetVersion = taprpc.AssetVersion_ASSET_VERSION_V0
	}
	var _assetType taprpc.AssetType
	if assetTypeIsCollectible {
		_assetType = taprpc.AssetType_COLLECTIBLE
	} else {
		_assetType = taprpc.AssetType_NORMAL
	}
	_assetMetaDataByteSlice := []byte(assetMetaData)
	var _assetMetaType taprpc.AssetMetaType
	if AssetMetaTypeIsJsonNotOpaque {
		//_assetMetaType = taprpc.AssetMetaType_META_TYPE_JSON
	} else {
		_assetMetaType = taprpc.AssetMetaType_META_TYPE_OPAQUE
	}
	_groupKeyByteSlices, err := hex.DecodeString(groupKey)
	if err != nil {
		return MakeJsonErrorResult(DecodeStringErr, err.Error(), nil)
	}
	request := &mintrpc.MintAssetRequest{
		Asset: &mintrpc.MintAsset{
			AssetVersion: _assetVersion,
			AssetType:    _assetType,
			Name:         name,
			AssetMeta: &taprpc.AssetMeta{
				Data: _assetMetaDataByteSlice,
				Type: _assetMetaType,
			},
			Amount:          uint64(amount),
			NewGroupedAsset: newGroupedAsset,
			GroupedAsset:    groupedAsset,
			GroupKey:        _groupKeyByteSlices,
			GroupAnchor:     groupAnchor,
		},
		ShortResponse: shortResponse,
	}
	response, err := client.MintAsset(context.Background(), request)
	if err != nil {
		fmt.Printf("%s mintrpc MintAsset Error: %v\n", GetTimeNow(), err)
		return MakeJsonErrorResult(MintAssetErr, err.Error(), nil)
	}
	return MakeJsonErrorResult(SUCCESS, "", MintAssetResponseToPendingBatch(response))
}
