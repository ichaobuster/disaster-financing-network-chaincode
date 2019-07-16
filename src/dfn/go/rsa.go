package main

import (
	"fmt"
	"bytes"
	"crypto/x509"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"encoding/pem"
	"encoding/base64"
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type EncryptedData struct {
	DocType        string   `json:"docType"`
	Id             string   `json:"id"`
	StateId        string   `json:"stateId"`
	TargetOrg      string   `json:"targetOrg"`
	EncryptedData  string   `json:"encryptedData"`
	Creator        string   `json:"creator"`      // 创建人
	LastModifier   string   `json:"lastModifier"` // 最后修改人
	CreateTime     string   `json:"createTime"`   // 创建时间
	ModifyTime     string   `json:"modifyTime"`   // 修改时间
}

type OrgPublicKey struct {
	DocType        string   `json:"docType"`
	Id             string   `json:"id"`
	Organization   string   `json:"organization"`
	PublicKey      string   `json:"publicKey"`
	Version        int      `json:"version"`
	CreateTime     string   `json:"createTime"`   // 创建时间
	ModifyTime     string   `json:"modifyTime"`   // 修改时间
}

// 加密数据
func encrypt_data(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting encrypt_data")
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	var organizations []string
	stateID := args[0]
	modifyTime := args[2]

	err := json.Unmarshal([]byte(args[1]), &organizations)
	if err != nil {
		fmt.Println(err.Error())
		return shim.Error(err.Error())
	}

	tMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("Could not retrieve transient")
	}
	// 从transientMap获取需要加密的数据
	originData := tMap["dataString"]

	// 获取当前机构
	submitterOrg, err := GetOrgFromCert(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	// 将当前机构添加到加密清单
	if !ContainsString(organizations, submitterOrg) {
		organizations = append(organizations, submitterOrg)
	}

	// 更新数据的情况下，需要添加所有既有数据的公钥
	storedOrgs, err := GetEncryptTargetOrgs(stub, stateID)
	if storedOrgs != nil{
		// 合并
		organizations = Append2Slices(organizations, storedOrgs)
		// 去重
		organizations = RemoveRepStringByMap(organizations)
	}

	// 加密并存储
	err = EncryptAndStoreByOrgs(stub, stateID, originData, organizations, modifyTime)
	if err != nil {
		return shim.Error(err.Error())
	}
	returnDataAsBytes, _ := json.Marshal(stateID)

	fmt.Println("- end encrypt_data")
	return shim.Success(returnDataAsBytes)
}


// 解密数据
func decrypt_data(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting decrypt_data")
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	tMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("Could not retrieve transient")
	}
	// 从transientMap获取密钥清单
	privateKey := tMap["privateKey"]
	
	stateID := args[0]

	returnData, err := GetEncryptDataAndDecrypt(stub, stateID, privateKey)

	fmt.Println("- end decrypt_data")
	return shim.Success(returnData)
}

// 存储机构公钥
func save_org_public_key(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting save_org_public_key")
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	publicKey := args[0]
	modifyTime := args[1]
	organization, err := GetOrgFromCert(stub)
	if err != nil {
		shim.Error(err.Error())
	}

	SaveOrgPublicKey(stub, organization, publicKey, modifyTime)

	return shim.Success(nil)
}

// 使用公钥加密某机构数据
func EncryptAndStore(stub shim.ChaincodeStubInterface, stateID string, originData []byte, organization string, modifyTime string) error {
	orgPubKey, err := GetOrgPublicKey(stub, organization)
	if err != nil {
		return err
	}

	base64EncStr, err := Encrypt64(originData, []byte(orgPubKey.PublicKey))
	if err != nil {
		return err
	}

	creator, err := GetSubmitterName(stub)
	if err != nil {
		return err
	}

	storeID := GetStoreDataID(stateID, organization)
	dataStoredAsBytes, err := stub.GetState(storeID) 
	if err != nil {
		return err
	}
	updateState := false
	if dataStoredAsBytes != nil {
		updateState = true
	}

	encryptedData := EncryptedData{}
	encryptedData.DocType = "encryptedData"
	encryptedData.Id = storeID
	encryptedData.StateId = stateID
	encryptedData.TargetOrg = organization
	encryptedData.EncryptedData = base64EncStr
	if !updateState {
		encryptedData.Creator = creator
		encryptedData.CreateTime = modifyTime
	}
	encryptedData.LastModifier = creator
	encryptedData.ModifyTime = modifyTime

	encryptedDataAsBytes, _ := json.Marshal(encryptedData)
	err = stub.PutState(storeID, encryptedDataAsBytes)
	return err
}

// 对列出的机构进行数据加密
func EncryptAndStoreByOrgs(stub shim.ChaincodeStubInterface, stateID string, originData []byte, organizations []string, modifyTime string) error {
	for i := 0; i < len(organizations); i++ {
		err := EncryptAndStore(stub, stateID, originData, organizations[i], modifyTime)
		if err != nil {
			return err
		}
	}
	return nil
}

// 加密
func Encrypt(originData []byte, publicKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, originData)
}

// 加密后取得base64编码字符串
func Encrypt64(originData []byte, publicKey []byte) (string, error) {
	encData, err := Encrypt(originData, publicKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encData), nil
}

// 解密
func Decrypt(cipherData []byte, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, cipherData)
}

// base64解码字符串后解密
func Decrypt64(ciperData64 string, privateKey []byte) ([]byte, error) {
	encBytes, err := base64.StdEncoding.DecodeString(ciperData64)
	if err != nil {
		return nil, err
	}
	oriData, err := Decrypt(encBytes, privateKey)
	if err != nil {
		return nil, err
	}
	return oriData, nil
}

// 获取加密数据
func GetEncryptData(stub shim.ChaincodeStubInterface, stateID string) (EncryptedData, error) {	
	var data EncryptedData

	orgName, err := GetOrgFromCert(stub)
	if err != nil {
		return data, err
	}

	storeID := GetStoreDataID(stateID, orgName)
	dataAsBytes, err := stub.GetState(storeID) 
	if err != nil {
		return data, err
	}
	if dataAsBytes == nil {
		return data, errors.New("This storeID is not existed - " + storeID)
	}
	json.Unmarshal(dataAsBytes, &data) //un stringify it aka JSON.parse()
	return data, nil
}

// 获取加密数据机构清单
func GetEncryptTargetOrgs(stub shim.ChaincodeStubInterface, stateID string) ([]string, error) {	
	var queryBuffer bytes.Buffer
		queryBuffer.WriteString(`{"selector":{"docType":"encryptedData","stateId":"`)
		queryBuffer.WriteString(stateID)
		queryBuffer.WriteString(`"}}`)

	result, err := GetQueryResult(stub, queryBuffer.String())
	var encryptedDatas []EncryptedData
	err = json.Unmarshal(result, &encryptedDatas)
	if err != nil {
		return nil, err
	}

	organizations := []string{}
	for i := 0; i < len(encryptedDatas); i++ {
		organizations = append(organizations, encryptedDatas[i].TargetOrg)
	}
	return organizations, nil
}

// 获取加密数据并解密
func GetEncryptDataAndDecrypt(stub shim.ChaincodeStubInterface, stateID string, privateKey []byte) ([]byte, error) {
	encData, err := GetEncryptData(stub, stateID)
	if err != nil {
		return nil, err
	}

	return Decrypt64(encData.EncryptedData, privateKey)
}

// 生成存储用的ID
func GetStoreDataID(stateID string, orgName string) string{
	return stateID + ":rsa:org:" + orgName
}

// 生成机构存储用的ID
func GetOrgPublicKeyId(organization string) string{
	return "org_public_key:" + organization
}

// 获取机构公钥
func GetOrgPublicKey(stub shim.ChaincodeStubInterface, organization string) (OrgPublicKey, error) {
	var data OrgPublicKey
	storeID := GetOrgPublicKeyId(organization)
	dataAsBytes, err := stub.GetState(storeID) 
	if err != nil {
		return data, err
	}
	if dataAsBytes == nil {
		return data, errors.New("No public key founded - " + organization)
	}
	json.Unmarshal(dataAsBytes, &data) //un stringify it aka JSON.parse()
	return data, nil
}

// 存储机构公钥
func SaveOrgPublicKey(stub shim.ChaincodeStubInterface, organization string, publicKey string, modifyTime string) {
	storeID := GetOrgPublicKeyId(organization)
	orgPublicKey, err :=  GetOrgPublicKey(stub, organization)
	if err == nil && orgPublicKey.Id == storeID {
		// 数据存在，更新
		fmt.Println("rsa public key exists")
		if publicKey == orgPublicKey.PublicKey {
			// 公钥未更新的情况下直接结束
			return
		} 
		orgPublicKey.PublicKey = publicKey
		orgPublicKey.ModifyTime = modifyTime
		orgPublicKey.Version = orgPublicKey.Version + 1
	} else {
		// 数据不存在，新增
		fmt.Println("rsa public key does not exist")
		orgPublicKey = OrgPublicKey{}
		orgPublicKey.DocType = "orgPublicKey"
		orgPublicKey.Id = storeID
		orgPublicKey.Organization = organization
		orgPublicKey.PublicKey = publicKey
		orgPublicKey.Version = 1
		orgPublicKey.CreateTime = modifyTime
		orgPublicKey.ModifyTime = modifyTime
	}
	fmt.Println("saving public key")
	orgPublicKeyAsBytes, _ := json.Marshal(orgPublicKey)
	PutState(stub, storeID, orgPublicKeyAsBytes)
}
