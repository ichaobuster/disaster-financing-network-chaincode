package main

import (
	"fmt"
	"testing"
	"time"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func GetTestTxID() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func MockInit(t *testing.T, stub *shim.MockStub) {
	response := stub.MockInit(GetTestTxID(), [][]byte{[]byte("1")})
	if response.Status != shim.OK {
		fmt.Println("Init failed", string(response.Message))
		t.FailNow()
	}
}

func GetMockStub() *shim.MockStub{
	chaincode := new(SimpleChaincode)
	stub := shim.NewMockStub("abs_chaincode", chaincode)
	return stub
}

// 测试初始化
func Test_Init(t *testing.T) {
	stub := GetMockStub()
	MockInit(t, stub)
}
