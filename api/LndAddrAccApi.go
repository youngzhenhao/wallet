package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"github.com/wallet/base"
	"github.com/wallet/service/apiConnect"
	"github.com/wallet/service/rpcclient"
	"path/filepath"
	"strings"
	"time"
)

// GetNewAddress_P2TR
// @dev: Get a p2tr address
// @note: TAPROOT_PUBKEY
// @Description:NewAddress creates a new address under control of the local wallet.
// @return string
func GetNewAddress_P2TR() string {
	conn, clearUp, err := apiConnect.GetConnection("lnd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()

	client := lnrpc.NewLightningClient(conn)
	request := &lnrpc.NewAddressRequest{
		Type: lnrpc.AddressType_TAPROOT_PUBKEY,
	}
	response, err := client.NewAddress(context.Background(), request)
	if err != nil {
		fmt.Printf("%s lnrpc NewAddress err: %v\n", GetTimeNow(), err)
		return MakeJsonErrorResult(NewAddressP2trErr, "AddressType_TAPROOT_PUBKEY error", "")
	}
	return MakeJsonErrorResult(SUCCESS, "", Addr{
		Name:           "default",
		Address:        response.Address,
		Balance:        0,
		AddressType:    lnrpc.AddressType_TAPROOT_PUBKEY.String(),
		DerivationPath: AddressTypeToDerivationPath(lnrpc.AddressType_TAPROOT_PUBKEY.String()),
		IsInternal:     false,
	})
}

// GetNewAddress_P2WKH
// @dev: Get a p2wkh address
// @note: WITNESS_PUBKEY_HASH
// @Description:NewAddress creates a new address under control of the local wallet.
// @return string
func GetNewAddress_P2WKH() string {
	conn, clearUp, err := apiConnect.GetConnection("lnd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	client := lnrpc.NewLightningClient(conn)
	request := &lnrpc.NewAddressRequest{
		Type: lnrpc.AddressType_WITNESS_PUBKEY_HASH,
	}
	response, err := client.NewAddress(context.Background(), request)
	if err != nil {
		fmt.Printf("%s lnrpc NewAddress err: %v\n", GetTimeNow(), err)
		return MakeJsonErrorResult(NewAddressP2wkhErr, "AddressType_WITNESS_PUBKEY_HASH error", "")
	}
	return MakeJsonErrorResult(SUCCESS, "", Addr{
		Name:           "default",
		Address:        response.Address,
		Balance:        0,
		AddressType:    lnrpc.AddressType_WITNESS_PUBKEY_HASH.String(),
		DerivationPath: AddressTypeToDerivationPath(lnrpc.AddressType_WITNESS_PUBKEY_HASH.String()),
		IsInternal:     false,
	})
}

// GetNewAddress_NP2WKH
// @dev: Get a np2wkh address
// @note: HYBRID_NESTED_WITNESS_PUBKEY_HASH
// @Description: NewAddress creates a new address under control of the local wallet.
// @return string
func GetNewAddress_NP2WKH() string {
	conn, clearUp, err := apiConnect.GetConnection("lnd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	client := lnrpc.NewLightningClient(conn)
	request := &lnrpc.NewAddressRequest{
		Type: lnrpc.AddressType_NESTED_PUBKEY_HASH,
	}
	response, err := client.NewAddress(context.Background(), request)
	if err != nil {
		fmt.Printf("%s lnrpc NewAddress err: %v\n", GetTimeNow(), err)
		return MakeJsonErrorResult(NewAddressNp2wkhErr, "AddressType_NESTED_PUBKEY_HASH error", "")
	}
	return MakeJsonErrorResult(SUCCESS, "", Addr{
		Name:           "default",
		Address:        response.Address,
		Balance:        0,
		AddressType:    lnrpc.AddressType_NESTED_PUBKEY_HASH.String(),
		DerivationPath: AddressTypeToDerivationPath(lnrpc.AddressType_NESTED_PUBKEY_HASH.String()),
		IsInternal:     false,
	})
}

func GetNewAddress_P2TR_Example() string {
	address := "bc1pxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	return MakeJsonErrorResult(SUCCESS, SuccessError, Addr{
		Name:           "default",
		Address:        address,
		Balance:        0,
		AddressType:    lnrpc.AddressType_TAPROOT_PUBKEY.String(),
		DerivationPath: AddressTypeToDerivationPath(lnrpc.AddressType_TAPROOT_PUBKEY.String()),
		IsInternal:     false,
	})
}

func GetNewAddress_P2WKH_Example() string {
	address := "bc1qxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	return MakeJsonErrorResult(SUCCESS, SuccessError, Addr{
		Name:           "default",
		Address:        address,
		Balance:        0,
		AddressType:    lnrpc.AddressType_WITNESS_PUBKEY_HASH.String(),
		DerivationPath: AddressTypeToDerivationPath(lnrpc.AddressType_WITNESS_PUBKEY_HASH.String()),
		IsInternal:     false,
	})
}

func GetNewAddress_NP2WKH_Example() string {
	address := "3xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	return MakeJsonErrorResult(SUCCESS, SuccessError, Addr{
		Name:           "default",
		Address:        address,
		Balance:        0,
		AddressType:    lnrpc.AddressType_NESTED_PUBKEY_HASH.String(),
		DerivationPath: AddressTypeToDerivationPath(lnrpc.AddressType_NESTED_PUBKEY_HASH.String()),
		IsInternal:     false,
	})
}

// StoreAddr
// @Description: Store Addr after being chosen.
// @param address
// @param balance
// @param _type
// @return string
func StoreAddr(name string, address string, balance int, addressType string, derivationPath string, isInternal bool) string {
	_ = InitAddrDB()
	path := filepath.Join(base.QueryConfigByKey("dirpath"), "phone.db")
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Printf("%s bolt.Open :%v\n", GetTimeNow(), err)
	}
	defer func(db *bolt.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("%s db.Close :%v\n", GetTimeNow(), err)
		}
	}(db)
	s := &AddrStore{DB: db}
	err = s.CreateOrUpdateAddr("addresses", &Addr{
		Name:           name,
		Address:        address,
		Balance:        balance,
		AddressType:    addressType,
		DerivationPath: derivationPath,
		IsInternal:     isInternal,
	})
	if err != nil {
		return MakeJsonErrorResult(CreateOrUpdateAddrErr, "Store address fail", "")
	}
	return MakeJsonErrorResult(SUCCESS, "", address)
}

func StoreAddrAndGetResponse(name string, address string, balance int, addressType string, derivationPath string, isInternal bool) (string, error) {
	_ = InitAddrDB()
	path := filepath.Join(base.QueryConfigByKey("dirpath"), "phone.db")
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Printf("%s bolt.Open :%v\n", GetTimeNow(), err)
		return "", err
	}
	defer func(db *bolt.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("%s db.Close :%v\n", GetTimeNow(), err)
		}
	}(db)
	s := &AddrStore{DB: db}
	err = s.CreateOrUpdateAddr("addresses", &Addr{
		Name:           name,
		Address:        address,
		Balance:        balance,
		AddressType:    addressType,
		DerivationPath: derivationPath,
		IsInternal:     isInternal,
	})
	if err != nil {
		return "", err
	}
	return address, nil
}

func StoreAddrAndGetResponseByAddr(addr Addr) (string, error) {
	return StoreAddrAndGetResponse(addr.Name, addr.Address, addr.Balance, addr.AddressType, addr.DerivationPath, addr.IsInternal)
}

// RemoveAddr
// @Description: Remove a addr in all addresses
// @param address
// @return string
func RemoveAddr(address string) string {
	_ = InitAddrDB()
	path := filepath.Join(base.QueryConfigByKey("dirpath"), "phone.db")
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Printf("%s bolt.Open :%v\n", GetTimeNow(), err)
	}
	defer func(db *bolt.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("%s db.Close :%v\n", GetTimeNow(), err)
		}
	}(db)
	s := &AddrStore{DB: db}
	_, err = s.ReadAddr("addresses", address)
	if err != nil {
		return MakeJsonErrorResult(ReadAddrErr, "No such address available for deletion. Read Addr fail.", "")
	}
	err = s.DeleteAddr("addresses", address)
	if err != nil {
		return MakeJsonErrorResult(DeleteAddrErr, "Delete Addr fail. "+err.Error(), "")
	}
	return MakeJsonErrorResult(SUCCESS, "", address)
}

// QueryAddr
// @Description: Query Addr in all addresses
// @param address
// @return string
func QueryAddr(address string) string {
	_ = InitAddrDB()
	path := filepath.Join(base.QueryConfigByKey("dirpath"), "phone.db")
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Printf("%s bolt.Open :%v\n", GetTimeNow(), err)
	}
	defer func(db *bolt.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("%s db.Close :%v\n", GetTimeNow(), err)
		}
	}(db)
	s := &AddrStore{DB: db}
	addr, err := s.ReadAddr("addresses", address)
	if err != nil {
		return MakeJsonErrorResult(ReadAddrErr, "No such address, read Addr fail.", "")
	}
	return MakeJsonErrorResult(SUCCESS, "", addr)
}

// QueryAllAddr
// @Description: get a json list of addresses
// @return string
func QueryAllAddr() string {
	_ = InitAddrDB()
	path := filepath.Join(base.QueryConfigByKey("dirpath"), "phone.db")
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Printf("%s bolt.Open :%v\n", GetTimeNow(), err)
	}
	defer func(db *bolt.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("%s db.Close :%v\n", GetTimeNow(), err)
		}
	}(db)
	s := &AddrStore{DB: db}
	addresses, err := s.AllAddresses("addresses")
	if err != nil || len(addresses) == 0 {
		return MakeJsonErrorResult(AllAddressesErr, "Addresses is NULL or read fail.", "")
	}
	return MakeJsonErrorResult(SUCCESS, "", addresses)
}

func QueryAllAddrAndGetResponse() (*[]Addr, error) {
	_ = InitAddrDB()
	path := filepath.Join(base.QueryConfigByKey("dirpath"), "phone.db")
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Printf("%s bolt.Open :%v\n", GetTimeNow(), err)
	}
	defer func(db *bolt.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("%s db.Close :%v\n", GetTimeNow(), err)
		}
	}(db)
	s := &AddrStore{DB: db}
	addresses, err := s.AllAddresses("addresses")
	if err != nil || len(addresses) == 0 {
		return nil, errors.New("Addresses is NULL or read fail.")
	}
	return &addresses, nil
}

//	 QueryAddresses
//	 @Description:  Use listAddresses to query the non-zero balance address, exported.
//					List of non-zero balance addresses constitutes the Total balance.
//	 @return string
func GetNonZeroBalanceAddresses() string {
	listAddrResp, err := rpcclient.ListAddresses()
	if err != nil {
		return MakeJsonErrorResult(ListAddressesErr, "Query addresses fail. "+err.Error(), "")
	}
	var addrs []Addr
	listAddrs := listAddrResp.GetAccountWithAddresses()
	if len(listAddrs) == 0 {
		return MakeJsonErrorResult(GetAccountWithAddressesErr, "Queried non-zero balance addresses NULL.", "")
	}
	for _, accWithAddr := range listAddrs {
		addresses := accWithAddr.Addresses
		for _, address := range addresses {
			if address.Balance != 0 {
				addrs = append(addrs, Addr{
					Name:           accWithAddr.Name,
					Address:        address.Address,
					Balance:        int(address.Balance),
					AddressType:    accWithAddr.AddressType.String(),
					DerivationPath: accWithAddr.DerivationPath,
					IsInternal:     address.IsInternal,
				})
			}
		}
	}
	return MakeJsonErrorResult(SUCCESS, "", addrs)
}

// UpdateAllAddressesByGNZBA
// @Description: Update all addresses by query non zero balance addresses
// @return string
func UpdateAllAddressesByGNZBA() string {
	listAddrResp, err := rpcclient.ListAddresses()
	if err != nil {
		return MakeJsonErrorResult(ListAddressesErr, "Query addresses fail. "+err.Error(), nil)
	}
	var addresses []string
	listAddrs := listAddrResp.GetAccountWithAddresses()
	if len(listAddrs) == 0 {
		return MakeJsonErrorResult(GetAccountWithAddressesErr, "Queried non-zero balance addresses NULL.", nil)
	}
	allAddr, err := QueryAllAddrAndGetResponse()
	if err != nil {
		return MakeJsonErrorResult(QueryAllAddrAndGetResponseErr, err.Error(), nil)
	}
	// @dev: Update allAddr balance
	err = UpdateAllAddrByAccountWithAddresses(allAddr, &listAddrs)
	if err != nil {
		return MakeJsonErrorResult(UpdateAllAddrByAccountWithAddressesErr, err.Error(), nil)
	}
	// @dev: UpdateNoneZeroAddress
	for _, accWithAddr := range listAddrs {
		if accWithAddr.Name != "default" {
			continue
		}
		_addresses := accWithAddr.Addresses
		for _, _address := range _addresses {
			// @dev: remove is_internal check
			if _address.Balance != 0 {
				var result JsonResult
				// @dev: Store
				_re := StoreAddr(accWithAddr.Name, _address.Address, int(_address.Balance), accWithAddr.AddressType.String(), accWithAddr.DerivationPath, _address.IsInternal)
				err = json.Unmarshal([]byte(_re), &result)
				if err != nil {
					return MakeJsonErrorResult(UnmarshalErr, "Store address Unmarshal fail. "+err.Error(), nil)
				}
				if !result.Success {
					return MakeJsonErrorResult(resultIsNotSuccessErr, "Store address result false", nil)
				}
				addresses = append(addresses, _address.Address)
			}
		}
	}
	return MakeJsonErrorResult(SUCCESS, SuccessError, nil)
}

func UpdateAllAddrByAccountWithAddresses(addrs *[]Addr, accountWithAddresses *[]*walletrpc.AccountWithAddresses) error {
	addressToAddr := AddrToAddressMapAddr(addrs)
	for _, accWithAddr := range *accountWithAddresses {
		if accWithAddr.Name != "default" {
			continue
		}
		_addresses := accWithAddr.Addresses
		for _, _address := range _addresses {
			addr, ok := (*addressToAddr)[_address.Address]
			if !ok {
				continue
			}
			(*addr).Balance = int(_address.Balance)
			_, err := StoreAddrAndGetResponseByAddr(*addr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func AddrToAddressMapAddr(addrs *[]Addr) *map[string]*Addr {
	addressToAddr := make(map[string]*Addr)
	for _, addr := range *addrs {
		addressToAddr[addr.Address] = &addr
	}
	return &addressToAddr
}

// @dev: Acc

type Account struct {
	Name              string `json:"name"`
	AddressType       string `json:"address_type"`
	ExtendedPublicKey string `json:"extended_public_key"`
	DerivationPath    string `json:"derivation_path"`
}

func GetAllAccountsString() string {
	accs := GetAllAccounts()
	if accs == nil {
		return MakeJsonErrorResult(GetAllAccountsErr, "get all accounts fail.", "")
	}
	return MakeJsonErrorResult(SUCCESS, "", accs)
}

func GetAllAccounts() []Account {
	var accs []Account
	response, err := listAccounts()
	if err != nil {
		fmt.Printf("%s listAccounts fail. %v\n", GetTimeNow(), err)
		return nil
	}
	for _, v := range response.Accounts {
		accs = append(accs, Account{
			v.Name,
			v.AddressType.String(),
			v.ExtendedPublicKey,
			v.DerivationPath,
		})
	}
	return accs
}

// AddressTypeToDerivationPath
// @dev: NOT STANDARD RESULT RETURN
func AddressTypeToDerivationPath(addressType string) string {
	accs := GetAllAccounts()
	addressType = strings.ToUpper(addressType)
	if addressType == "NESTED_PUBKEY_HASH" {
		addressType = "HYBRID_NESTED_WITNESS_PUBKEY_HASH"
	}
	for _, acc := range accs {
		if acc.AddressType == addressType {
			return acc.DerivationPath
		}
	}
	fmt.Printf("%s %v is not a valid address type.\n", GetTimeNow(), addressType)
	return ""
}

func GetPathByAddressType(addressType string) string {
	accs := GetAllAccounts()
	if addressType == "NESTED_PUBKEY_HASH" {
		addressType = "HYBRID_NESTED_WITNESS_PUBKEY_HASH"
	}
	for _, acc := range accs {
		if acc.AddressType == addressType {
			return MakeJsonErrorResult(SUCCESS, "", acc.DerivationPath)
		}
	}
	fmt.Printf("%s %v is not a valid address type.\n", GetTimeNow(), addressType)
	return MakeJsonErrorResult(InvalidAddressTypeErr, "can't find path by given address type.", "")
}
