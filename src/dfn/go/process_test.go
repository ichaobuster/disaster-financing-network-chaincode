package main

import (
	"fmt"
	"testing"
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// mock 启动流程
func MockStartProcess1(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("start_process"),
		[]byte(`{"id":"test_process_002:test_linear_workflow-001","workflowId":"test_linear_workflow-001","attachDocType":"project","attachDocId":"project-bankcomm-000002","createTime":"2018-3-19 09:43:02"}`),
	})
	return response
}

// mock 启动流程
func MockStartProcess2(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("start_process"),
		[]byte(`{"id":"test_process_003:test_linear_workflow-002","workflowId":"test_linear_workflow-002","attachDocType":"project","attachDocId":"project-bankcomm-000002","createTime":"2018-3-19 09:43:02"}`),
	})
	return response
}

// mock 根据id获取流程实例
func MockGetProcessByID(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("get_process_by_id"),
		[]byte("test_process_002:test_linear_workflow-001"),
	})
	return response
}

// mock 提交流程
func MockTransferProcess(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("transfer_process"),
		[]byte("test_process_002:test_linear_workflow-001"),
		[]byte("test_linear_workflow-001:node-2"),
		[]byte("@org1.example.com"),
		[]byte("2018-03-16 15:54:00"),
	})
	return response
}

// mock 退回流程
func MockReturnProcess(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("return_process"),
		[]byte("test_process_002:test_linear_workflow-001"),
		[]byte("2018-03-16 15:54:00"),
	})
	return response
}

// mock 撤回流程
func MockWithdrawProcess(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("withdraw_process"),
		[]byte("test_process_002:test_linear_workflow-001"),
		[]byte("2018-03-16 15:54:00"),
	})
	return response
}

// mock 取消流程
func MockCancelProcess(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("cancel_process"),
		[]byte("test_process_002:test_linear_workflow-001"),
		[]byte("2018-03-16 15:54:00"),
	})
	return response
}

// mock 查询待办流程
func MockQueryTodoProcess(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{[]byte("query_todo_process")})
	return response
}

// mock 查询已办流程
func MockQueryDoneProcess(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{[]byte("query_done_process")})
	return response
}

// 测试启动流程
func Test_StartProcess(t *testing.T) {
	stub := GetMockStub()
	MockInit(t, stub)
	// 启动将失败，因为缺少workflow和attachdoc
	response := MockStartProcess2(t, stub)
	if response.Status != shim.ERROR {
		fmt.Println("不存在工作流和文档，应该失败。")
		t.FailNow()
	}
	// 添加工作流
	MockCreateLinearWorkflow2(t, stub)
	// 启动将失败，因为缺少attachdoc
	response = MockStartProcess2(t, stub)
	if response.Status != shim.ERROR {
		fmt.Println("不存在文档，应该失败。")
		t.FailNow()
	}
	// 添加文档
	MockCreateProject2(t, stub)
	// 启动将失败，因为首节点里不包含测试机构
	response = MockStartProcess2(t, stub)
	if response.Status != shim.ERROR {
		fmt.Println("不存在文档，应该失败。")
		t.FailNow()
	}
	MockCreateLinearWorkflow1(t, stub)
	// 由于mock引擎还没有实现GetQueryResult，测试将直接失败
	// 启动将可以成功
	response = MockStartProcess1(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		// t.FailNow()
	}
}

func Test_GetProcessById(t *testing.T) {
	// 由于mock引擎还没有实现GetQueryResult，测试将直接失败
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateLinearWorkflow1(t, stub)
	MockCreateProject2(t, stub)
	MockStartProcess1(t, stub)
	response := MockGetProcessByID(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		// t.FailNow()
	}
	if len(response.Payload) <= 0 {
		fmt.Println("response is incorrect")
		// t.FailNow()
	}
}

func Test_TransferProcess(t *testing.T) {
	// 由于mock引擎还没有实现GetQueryResult，测试将直接失败
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateLinearWorkflow1(t, stub)
	MockCreateProject2(t, stub)
	MockStartProcess1(t, stub)
	MockGetProcessByID(t, stub)
	response := MockTransferProcess(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		// t.FailNow()
	}
}

func Test_ReturnProcess(t *testing.T) {
	// 由于mock引擎还没有实现GetQueryResult，测试将直接失败
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateLinearWorkflow1(t, stub)
	MockCreateProject2(t, stub)
	MockStartProcess1(t, stub)
	MockGetProcessByID(t, stub)
	MockTransferProcess(t, stub)
	response := MockReturnProcess(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		// t.FailNow()
	}
}

func Test_WithdrawProcess(t *testing.T) {
	// 由于mock引擎还没有实现GetQueryResult，测试将直接失败
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateLinearWorkflow1(t, stub)
	MockCreateProject2(t, stub)
	MockStartProcess1(t, stub)
	MockGetProcessByID(t, stub)
	MockTransferProcess(t, stub)
	response := MockWithdrawProcess(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		// t.FailNow()
	}
}

func Test_CancelProcess(t *testing.T) {
	// 由于mock引擎还没有实现GetQueryResult，测试将直接失败
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateLinearWorkflow1(t, stub)
	MockCreateProject2(t, stub)
	MockStartProcess1(t, stub)
	MockGetProcessByID(t, stub)
	MockTransferProcess(t, stub)
	response := MockCancelProcess(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		// t.FailNow()
	}
	// 校验目前只能先跳过
	var result Process
	state := stub.State["test_process_002:test_linear_workflow-001"]
	json.Unmarshal(state, &result)
	if !result.Canceled {
		fmt.Println("应为取消状态")
		// t.FailNow()
	}
}

func Test_QueryTodoProcess(t *testing.T) {
	// 由于mock引擎还没有实现GetQueryResult，测试将直接失败
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateLinearWorkflow1(t, stub)
	MockCreateProject2(t, stub)
	MockStartProcess1(t, stub)
	MockGetProcessByID(t, stub)
	MockTransferProcess(t, stub)
	response := MockQueryTodoProcess(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		// t.FailNow()
	}
    var result []interface{}
	json.Unmarshal(response.Payload, &result)
	if len(result) != 1 {
		fmt.Println("应有1条")
		// t.FailNow()
	}
}

func Test_QueryDoneProcess(t *testing.T) {
	// 由于mock引擎还没有实现GetQueryResult，测试将直接失败
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateLinearWorkflow1(t, stub)
	MockCreateProject2(t, stub)
	MockStartProcess1(t, stub)
	MockGetProcessByID(t, stub)
	MockTransferProcess(t, stub)
	response := MockQueryDoneProcess(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		// t.FailNow()
	}
    var result []interface{}
	json.Unmarshal(response.Payload, &result)
	if len(result) != 1 {
		fmt.Println("应有1条")
		// t.FailNow()
	}
}
