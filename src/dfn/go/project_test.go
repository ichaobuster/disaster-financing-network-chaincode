package main

import (
	"fmt"
	"testing"
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// mock 创建一个project
func MockCreateProject1(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("create_project"), 
		[]byte(`{"id":"project-bankcomm-000003","projectName":"测试交行项目000003号","scale":"10亿元人民币","basicAssets":"小微企业","initiator":"交通银行","trustee":"交银国信","depositary":"兴业银行","agent":"中债登","assetService":"上海融孚律师事务所","assessor":"深圳市世联资产评估有限公司","creditRater":"中债资信","liquiditySupporter":"中证信用增进股份有限公司","underwriter":"招商证券","lawyer":"北京市金杜律师事务所","accountant":"普华永道","createTime":"2018-3-14 09:06:03"}`),
	})
	return response
}

// mock 创建一个project
func MockCreateProject2(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("create_project"), 
		[]byte(`{"id":"project-bankcomm-000002","projectName":"测试交行项目000002号","scale":"500万元人民币","basicAssets":"个人按揭贷款","initiator":"交通银行","trustee":"交银国信","depositary":"兴业银行","agent":"中债登","assetService":"上海融孚律师事务所","assessor":"深圳市世联资产评估有限公司","creditRater":"中债资信","liquiditySupporter":"中证信用增进股份有限公司","underwriter":"招商证券","lawyer":"北京市金杜律师事务所","accountant":"普华永道","createTime":"2018-3-16 09:03:45"}`),
	})
	return response
}

// mock 根据id获取project
func MockGetProjectByID(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("get_project_by_id"), 
		[]byte("project-bankcomm-000003"),
	})
	return response
}

// mock 获取全部project
func MockQueryAllProject(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{[]byte("query_all_projects")})
	return response
}

// mock 根据id删除project
func MockRemoveProject(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), [][]byte{
		[]byte("remove_project"), 
		[]byte("project-bankcomm-000003"),
	})
	return response
}

// mock 修改一个project
func MockModifyProject(t *testing.T, stub *shim.MockStub) pb.Response {
	response := stub.MockInvoke(GetTestTxID(), 
		[][]byte{
			[]byte("modify_project"), 
			[]byte("project-bankcomm-000003"),
			[]byte("2018-03-16 15:54:00"),
			[]byte("projectName"),
			[]byte("测试交行项目000003号_新品种"),
			[]byte("scale"),
			[]byte("100亿元人民币"),
		})
	return response
}

// 测试创建project
func Test_CreateProject(t *testing.T) {
	stub := GetMockStub()
	MockInit(t, stub)
	response := MockCreateProject1(t, stub)
	if response.Status != shim.OK {
		fmt.Println(string(response.Message))
		t.FailNow()
	}
	response = MockCreateProject2(t, stub)
	if response.Status != shim.OK {
		fmt.Println(string(response.Message))
		t.FailNow()
	}
	// 重复id将报错
	response = MockCreateProject2(t, stub)
	if response.Status != shim.ERROR {
		fmt.Println("应该不可重复添加ID")
		t.FailNow()
	}
}

// 测试查询project
func Test_GetProjectById(t *testing.T) {
	stub := GetMockStub()
	MockInit(t, stub)
	response := MockGetProjectByID(t, stub)
	if response.Status != shim.ERROR {
		fmt.Println("未添加的project的情况下不应能正确查询")
		t.FailNow()
	}

	MockCreateProject1(t, stub)
	response = MockGetProjectByID(t, stub)
	if response.Status != shim.OK {
		fmt.Println(string(response.Message))
		t.FailNow()
	}
}

func Test_QueryAllProject(t *testing.T) {
	// 由于mock引擎还没有实现GetQueryResult，测试将直接失败
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateProject1(t, stub)
	MockCreateProject2(t, stub)
	response := MockQueryAllProject(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		// t.FailNow()
	}
    var result []interface{}
	json.Unmarshal(response.Payload, &result)
	if len(result) != 2 {
		fmt.Println("应有2条")
		// t.FailNow()
	}
}

func Test_RemoveProject(t *testing.T) {
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateProject1(t, stub)
	response := MockRemoveProject(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		t.FailNow()
	}
	response = MockGetProjectByID(t, stub)
	if response.Status != shim.ERROR {
		fmt.Println("已删除的project不应该能被查出")
		t.FailNow()
	}
}

func Test_ModifyProject(t *testing.T) {
	stub := GetMockStub()
	MockInit(t, stub)
	MockCreateProject1(t, stub)
	response := MockModifyProject(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		t.FailNow()
	}
	response = MockGetProjectByID(t, stub)
	if response.Status != shim.OK {
		fmt.Println(response.GetMessage())
		t.FailNow()
	}
	var result Project
	json.Unmarshal(response.Payload, &result)
	if result.ModifyTime != "2018-03-16 15:54:00" {
		fmt.Println("ModifyTime is incorrect")
		t.FailNow()
	}
	if result.ProjectName != "测试交行项目000003号_新品种" {
		fmt.Println("ProjectName is incorrect")
		t.FailNow()
	}
	if result.Scale != "100亿元人民币" {
		fmt.Println("Scale is incorrect")
		t.FailNow()
	}
}
