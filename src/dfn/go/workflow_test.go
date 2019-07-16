package main

import (
	"fmt"
	"testing"
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// mock 创建一个线性工作流
func MockCreateLinearWorkflow1(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("create_linear_workflow"),
		[]byte(`{"id":"test_linear_workflow-001","subDocType":"linearWorkflow","workflowName":"测试线性流程001","createTime":"2018-3-16 16:08:51"}`),
		[]byte(`{"nodeName":"发起行","accessOrgs":["@bankcomm.com","@icbc.com.cn","@org1.example.com"]}`),
		[]byte(`{"nodeName":"尽调机构","accessOrgs":["@pwccn.com","@org1.example.com"]}`),
		[]byte(`{"nodeName":"发行机构","accessOrgs":["@bocommtrust.com","@org1.example.com"]}`),
	})
	return response
}

// mock 创建一个线性工作流
func MockCreateLinearWorkflow2(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("create_linear_workflow"),
		[]byte(`{"id":"test_linear_workflow-002","subDocType":"linearWorkflow","workflowName":"测试线性流程002","createTime":"2018-3-16 16:08:51"}`),
		[]byte(`{"nodeName":"发起行","accessOrgs":["@bankcomm.com","@icbc.com.cn"]}`),
		[]byte(`{"nodeName":"尽调机构","accessOrgs":["@pwccn.com","@org1.example.com"]}`),
		[]byte(`{"nodeName":"发行机构","accessOrgs":["@bocommtrust.com","@org1.example.com"]}`),
	})
	return response
}

// mock 根据id获取工作流
func MockGetWorkflowByID(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("get_workflow_by_id"), 
		[]byte("test_linear_workflow-001"),
	})
	return response
}

// mock 启用工作流
func MockEnableWorkflow(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("enable_or_disable_workflow"), 
		[]byte("test_linear_workflow-001"),
		[]byte("true"),
		[]byte("2018-03-16 15:54:00"),
	})
	return response
}

// mock 禁用工作流
func MockDisableWorkflow(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("enable_or_disable_workflow"), 
		[]byte("test_linear_workflow-001"),
		[]byte("false"),
		[]byte("2018-03-16 15:54:00"),
	})
	return response
}

// mock 获取全部工作流
func MockQueryAllWorkflow(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{[]byte("query_all_workflows")})
	return response
}

// mock 获取全部工作流
func MockQueryAccessableWorkflow(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{[]byte("query_accessable_workflows")})
	return response
}

// mock 修改一个工作流
func MockModifyWorkflowDef(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("modify_workflow_def"), 
		[]byte("test_linear_workflow-001"),
		[]byte("2018-03-16 15:54:00"),
		[]byte("workflowName"),
		[]byte("测试线性流程001_修改"),
	})
	return response
}

// 测试创建线性工作流
func Test_CreateLinearWorkflow(t *testing.T) {
	stub := GetMockStub()
	MockInit(t, stub)
	response := MockCreateLinearWorkflow1(t, stub)
	if response.Status != shim.OK {
		fmt.Println(string(response.Message))
		t.FailNow()
	}
	// 重复id将报错
	response = MockCreateLinearWorkflow1(t, stub)
	if response.Status != shim.ERROR {
		fmt.Println("应该不可重复添加ID")
		t.FailNow()
	}
}

func Test_GetWorkflowById(t *testing.T) {
	// 由于mock引擎还没有实现GetQueryResult，测试将直接失败
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateLinearWorkflow1(t, stub)
	response := MockGetWorkflowByID(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		// t.FailNow()
	}
	var result Workflow
	json.Unmarshal(response.Payload, &result)
	if result.WorkflowDef.WorkflowName != "测试线性流程001" {
		fmt.Println("WorkflowName is incorrect")
		// t.FailNow()
	}
	if len(result.WorkflowNodes) != 3 {
		fmt.Println("WorkflowNodes is incorrect")
		// t.FailNow()
	}
}

func Test_QueryAllWorkflow(t *testing.T) {
	// 由于mock引擎还没有实现GetQueryResult，测试将直接失败
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateLinearWorkflow1(t, stub)
	response := MockQueryAllWorkflow(t, stub)
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

func Test_QueryAccessableWorkflow(t *testing.T) {
	// 由于mock引擎还没有实现GetQueryResult，测试将直接失败
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateLinearWorkflow1(t, stub)
	response := MockQueryAccessableWorkflow(t, stub)
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

// 测试启用禁用工作流
func Test_EnableOrDisableWorkflow(t *testing.T) {
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateLinearWorkflow1(t, stub)
	response := MockEnableWorkflow(t, stub)
	if response.Status != shim.ERROR {
		fmt.Println("已启用的工作流不能重复启用，应报错")
		t.FailNow()
	}
	response = MockDisableWorkflow(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		t.FailNow()
	}
	state := stub.State["test_linear_workflow-001"]
	var result WorkflowDef
	json.Unmarshal(state, &result)
	if result.Enabled {
		fmt.Println("应为禁用")
		t.FailNow()
	}
	response = MockEnableWorkflow(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		t.FailNow()
	}
	state = stub.State["test_linear_workflow-001"]
	json.Unmarshal(state, &result)
	if !result.Enabled {
		fmt.Println("应为启用")
		t.FailNow()
	}
}

func Test_ModifyWorkflow(t *testing.T) {
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateLinearWorkflow1(t, stub)
	response := MockModifyWorkflowDef(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		t.FailNow()
	}
	state := stub.State["test_linear_workflow-001"]
	var result WorkflowDef
	json.Unmarshal(state, &result)
	if result.ModifyTime != "2018-03-16 15:54:00" {
		fmt.Println("ModifyTime is incorrect")
		t.FailNow()
	}
	if result.WorkflowName != "测试线性流程001_修改" {
		fmt.Println("WorkflowName is incorrect")
		t.FailNow()
	}
}