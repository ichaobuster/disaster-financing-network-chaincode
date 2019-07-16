package main

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)


// GetAllObjectsByDocType 查询全部项目
func GetAllObjectsByDocType(stub shim.ChaincodeStubInterface, args []string, docType string) ([]byte, error) {
	var result []byte
	var err error
	fmt.Println("starting GetAllObjectsByDocType")

	var queryBuffer bytes.Buffer
	queryBuffer.WriteString(`{"selector":{"docType":"`)
	queryBuffer.WriteString(docType)
	queryBuffer.WriteString(`"}}`)

	result, err = GetQueryResult(stub, queryBuffer.String())
	if err != nil {
		return nil, err
	}
	fmt.Println("end GetAllObjectsByDocType")

	return result, nil
}

// GetPagingObjectsByDocType 分页查询项目
func GetPagingObjectsByDocType(stub shim.ChaincodeStubInterface, args []string, docType string) ([]byte, error) {
	var result []byte
	var err error
	fmt.Println("starting GetPagingObjectsByDocType")

	limit, skip, err := SanitizePagingArgument(args[0:2])
	if err != nil {
		return nil, err
	}

	var queryBuffer bytes.Buffer
		queryBuffer.WriteString(`{"selector":{"docType":"`)
		queryBuffer.WriteString(docType)
		queryBuffer.WriteString(`"},"limit":`)
		queryBuffer.WriteString(limit)
		queryBuffer.WriteString(`,"skip":`)
		queryBuffer.WriteString(skip)
		queryBuffer.WriteString(`,"sort": [{"modifyTime": "desc"}]`)
		queryBuffer.WriteString(`}`)

	result, err = GetQueryResult(stub, queryBuffer.String())
	if err != nil {
		return nil, err
	}
	fmt.Println("end GetPagingObjectsByDocType")

	return result, nil
}

// GetQueryResult 使用 stub.GetQueryResult 接口查询并转换结果为 bytes
func GetQueryResult(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {
	fmt.Println("queryString is :" + queryString)

	var err error
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	return ConvQueryResult(resultsIterator)
}

// ========================================================
// 转换查询结果 shim.StateQueryIteratorInterface 为 bytes
// ========================================================
func ConvQueryResult(resultsIterator shim.StateQueryIteratorInterface) ([]byte, error) {
	// buffer is a JSON array containing QueryResults
	fmt.Println("starting convert query result.")

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		aKeyValue, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// queryResultKey := aKeyValue.Key
		queryResultValue := aKeyValue.Value

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		buffer.WriteString(string(queryResultValue))

		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

// ========================================================
// 转换查询结果 shim.HistoryQueryIteratorInterface 为 bytes
// ========================================================
func ConvHistoryResult(resultsIterator shim.HistoryQueryIteratorInterface) ([]byte, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		aKeyValue, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// queryResultKey := aKeyValue.Key
		queryResultValue := aKeyValue.Value

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		buffer.WriteString(string(queryResultValue))

		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

// UpdateStruct 更新struct的field值
func UpdateStruct(o interface{}, key string, value string) error {
	protectedKeys := []string{"id", "docType", "enabled", "companyDomain", "creator", "lastModifier", "createTime", "modifyTime"}
	if ContainsString(protectedKeys, key) {
		return errors.New("You are not allowed to update field '" + key + "'!")
	}

	v := reflect.ValueOf(o).Elem()
	fieldKey := strings.Title(key)

	field := v.FieldByName(fieldKey)
	if !field.IsValid() || !field.CanSet() {
		return errors.New("Cannot find field " + key + " or the field cannot set a value")
	}

	if field.Kind() == reflect.String {
		field.SetString(value)
	} else {
		return errors.New("Field type is not string")
	}

	return nil
}

// ========================================================
// 检查正整数数字
// ========================================================
func SanitizePositiveIntArgument(arg string) error {
	intArg, err := strconv.Atoi(arg)
	if err != nil {
		return errors.New("argument must be a numeric string")
	}
	if intArg < 0 {
		return errors.New("argument must be a positive integer or a zero")
	}
	return nil
}

// ========================================================
// 检查分页参数
// ========================================================
func SanitizePagingArgument(args []string) (string, string, error) {
	var err error
	var limit string
	var skip string

	if len(args) != 2 {
		return limit, skip, errors.New("Incorrect number of paging arguments. Expecting 2")
	}

	limit = args[0]
	skip = args[1]

	err = SanitizePositiveIntArgument(limit)
	if err != nil {
		return limit, skip, err
	}
	err = SanitizePositiveIntArgument(skip)
	if err != nil {
		return limit, skip, err
	}

	return limit, skip, nil
}

// ========================================================
// 根据整数解析用户信息
// ========================================================
func GetSubmitterName(stub shim.ChaincodeStubInterface) (string, error) {
	var err error
	var commonName string

	_, isMock := stub.(*shim.MockStub) 
	if isMock {
		// MOCK测试情况
		// TODO 是否会被用于攻击或造假？
		return "Test@org1.example.com", nil
	}

	creator, err := stub.GetCreator()
	if err != nil {
		fmt.Println(err.Error())
		return commonName, errors.New(err.Error())
	}
	certStart := bytes.Index(creator, []byte("-----BEGIN CERTIFICATE-----"))
	if certStart == -1 {
		fmt.Println("No certificate found")
		return "", errors.New("No certificate found")
	}
	certText := creator[certStart:]
	block, _ := pem.Decode(certText)
	if block == nil {
		fmt.Println("Error received on pem.Decode of certificate")
		return "", errors.New("Error received on pem.Decode of certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		fmt.Println("Error received on ParseCertificate")
		return "", errors.New("Error received on ParseCertificate")
	}
	// rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)

	return cert.Subject.CommonName, nil
}

// ========================================================
// 从用户信息获取用户公司信息
// ========================================================
func GetOrgFromCertCommonName(certCommonName string) (string, error) {
	var orgName string
	// 目前仅通过"@"分割来获取公司信息
	if !strings.Contains(certCommonName, "@") {
		return orgName, errors.New("CertCommonName does not contains @ mark")
	}
	orgName = "@" + strings.SplitN(certCommonName, "@", 2)[1]
	return orgName, nil
}

// ========================================================
// 直接从stub的证书获取公司信息
// ========================================================
func GetOrgFromCert(stub shim.ChaincodeStubInterface) (string, error) {
	submitterName, err := GetSubmitterName(stub)
	if err != nil {
		return "", err
	}
	orgName, err := GetOrgFromCertCommonName(submitterName)
	if err != nil {
		return "", err
	}
	return orgName, nil
}

// ========================================================
// 获取字符串是否在某个slice里
// ========================================================
func ContainsString(sli []string, str string) bool {
	if sli == nil {
		return false
	}
	for i := 0; i < len(sli); i++ {
		if sli[i] == str {
			return true
		}
	}
	return false
}

// 通过map主键唯一的特性过滤slice重复string元素
func RemoveRepStringByMap(slc []string) []string {
    result := []string{}
    tempMap := map[string]byte{}  // 存放不重复主键
    for _, e := range slc{
        l := len(tempMap)
        tempMap[e] = 0
        if len(tempMap) != l{  // 加入map后，map长度变化，则元素不重复
            result = append(result, e)
        }
    }
    return result
}

// 合并两个slice
func Append2Slices(slc1 []string, slc2 []string) []string {
	result := make([]string, len(slc1) + len(slc2))
  	copy(result, slc1)
	copy(result[len(slc1):], slc2)
	return result
}

// ========================================================
// 获取文档名称
// ========================================================
func GetDocNameByDocTypeAndId(stub shim.ChaincodeStubInterface, docType string, docId string) (string, error) {
	var attachDocName string
	var err error
	//check if attachDoc exists and docType is correct
	// TODO 检查文档是否存在并且类型正确
	switch docType {
	case "project":
		project, err := GetProjectById(stub, docId)
		if err == nil {
			attachDocName = project.ProjectName
		}
	default:
		// error out
		fmt.Println("Received unknown DocType - " + docType)
		return "", errors.New("Received unknown DocType - " + docType)
	}

	return attachDocName, err
}

// 包装Event内容
func SendEvent(stub shim.ChaincodeStubInterface, eventName string, eventBytes []byte) {
	var buffer bytes.Buffer
	buffer.WriteString(`{"eventName":"`)
	buffer.WriteString(eventName)
	buffer.WriteString(`", "txId":"`)
	buffer.WriteString(stub.GetTxID())
	buffer.WriteString(`", "payload":`)
	buffer.Write(eventBytes)
	buffer.WriteString("}")

	stub.SetEvent("NewEvent", buffer.Bytes())
}

// PutState的包装
func PutState(stub shim.ChaincodeStubInterface, stateID string, stateBytes []byte) error{
	err := stub.PutState(stateID, stateBytes) //store with id as key
	if err != nil {
		return err
	}
	SendEvent(stub, "PutState", stateBytes)
	return nil
}

// DelState的包装
func DelState(stub shim.ChaincodeStubInterface, stateID string, docType string) error{
	err := stub.DelState(stateID)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	buffer.WriteString(`{"id":"`)
	buffer.WriteString(stateID)
	buffer.WriteString(`", "docType":"`)
	buffer.WriteString(docType)
	buffer.WriteString(`"}`)
	SendEvent(stub, "DelState", buffer.Bytes())
	return nil
}
