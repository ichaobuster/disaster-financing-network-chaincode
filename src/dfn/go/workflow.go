package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type WorkflowDef struct {
	DocType      string `json:"docType"`
	Id           string `json:"id"`
	SubDocType   string `json:"subDocType"`
	WorkflowName string `json:"workflowName"`
	AccessRoles  []string `json:"accessRoles"`
	AccessOrgs   []string `json:"accessOrgs"`
	Enabled      bool   `json:"enabled"`
	Creator      string `json:"creator"`      // 创建人
	LastModifier string `json:"lastModifier"` // 最后修改人
	CreateTime   string `json:"createTime"`         // 创建时间
	ModifyTime   string `json:"modifyTime"`         // 修改时间
}

type WorkflowNode struct {
	DocType     string   `json:"docType"`
	Id          string   `json:"id"`
	WorkflowId  string   `json:"workflowId"`
	NodeName    string   `json:"nodeName"`
	AccessRoles []string `json:"accessRoles"`
	AccessOrgs  []string `json:"accessOrgs"`
	PrevNodeIds []string `json:"prevNodeIds"`
	NextNodeIds []string `json:"nextNodeIds"`
	FirstNode   bool     `json:"firstNode"`
	LastNode    bool     `json:"lastNode"`
}

type Workflow struct {
	WorkflowDef   WorkflowDef    `json:"workflowDef"`
	WorkflowNodes []WorkflowNode `json:"workflowNodes"`
}

// =============================================================================
// 创建线性流程
// 第一个参数为工作流定义
// 之后为线性顺序的节点列表
// =============================================================================
func create_linear_workflow(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	var workflowDef WorkflowDef
	fmt.Println("starting create_linear_workflow")

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting greater than 2")
	}
	creator, err := GetSubmitterName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	// 获取流程定义
	err = json.Unmarshal([]byte(args[0]), &workflowDef)
	if err != nil {
		fmt.Println(err.Error())
		return shim.Error(err.Error())
	}

	//check if already exists
	workflowDefInStore, err := GetWorkflowDefById(stub, workflowDef.Id)
	if err == nil {
		fmt.Println("This workflowDef already exists - " + workflowDef.Id)
		fmt.Println(workflowDefInStore)
		return shim.Error("This workflowDef already exists - " + workflowDef.Id)
	}

	workflowDef.DocType = "workflow"
	workflowDef.SubDocType = "linear"
	workflowDef.Enabled = true
	workflowDef.Creator = creator
	workflowDef.LastModifier = creator
	workflowDef.ModifyTime = workflowDef.CreateTime

	// 获取流程节点
	var workflowNode WorkflowNode
	var workflowNodeAsBytes []byte

	for i := 1; i < len(args); i++ {
		workflowNode = WorkflowNode{}
		err = json.Unmarshal([]byte(args[i]), &workflowNode)
		if err != nil {
			fmt.Println(err.Error())
			return shim.Error(err.Error())
		}
		workflowNode.DocType = "workflowNode"
		workflowNode.WorkflowId = workflowDef.Id
		workflowNode.Id = workflowDef.Id + ":node-" + strconv.Itoa(i)
		if i == 1 {
			workflowNode.FirstNode = true
			workflowDef.AccessRoles = workflowNode.AccessRoles
			workflowDef.AccessOrgs = workflowNode.AccessOrgs
		}
		if i == len(args)-1 {
			workflowNode.LastNode = true
		}
		if !workflowNode.FirstNode {
			workflowNode.PrevNodeIds = []string{workflowDef.Id + ":node-" + strconv.Itoa(i-1)}
		}
		if !workflowNode.LastNode {
			workflowNode.NextNodeIds = []string{workflowDef.Id + ":node-" + strconv.Itoa(i+1)}
		}
		workflowNodeAsBytes, _ = json.Marshal(workflowNode)
		fmt.Println("store node:" + string(workflowNodeAsBytes))
		err = stub.PutState(workflowNode.Id, workflowNodeAsBytes) //store with id as key
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	workflowDefAsBytes, _ := json.Marshal(workflowDef)
	err = stub.PutState(workflowDef.Id, workflowDefAsBytes) //store with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end create_linear_workflow")
	return shim.Success(nil)
}

// =============================================================================
// Get WorkflowDef By id
// =============================================================================
func GetWorkflowDefById(stub shim.ChaincodeStubInterface, id string) (WorkflowDef, error) {
	var data WorkflowDef
	dataAsBytes, err := stub.GetState(id) //getState retreives a key/value from the ledger
	if err != nil {                       //this seems to always succeed, even if key didn't exist
		return data, errors.New("Failed to find workflowDef - " + id)
	}
	json.Unmarshal(dataAsBytes, &data) //un stringify it aka JSON.parse()

	if data.Id != id {
		return data, errors.New("WorkflowDef does not exist - " + id)
	}

	return data, nil
}

// =============================================================================
// Get WorkflowNode By id
// =============================================================================
func GetWorkflowNodeById(stub shim.ChaincodeStubInterface, id string) (WorkflowNode, error) {
	var data WorkflowNode
	dataAsBytes, err := stub.GetState(id) //getState retreives a key/value from the ledger
	if err != nil {                       //this seems to always succeed, even if key didn't exist
		return data, errors.New("Failed to find workflowNode - " + id)
	}
	json.Unmarshal(dataAsBytes, &data) //un stringify it aka JSON.parse()

	if data.Id != id {
		return data, errors.New("WorkflowNode does not exist - " + id)
	}

	return data, nil
}

// ========================================================
// 查询全部流程
// ========================================================
func query_all_workflows(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting query_all_workflows")

	/* 暂时无法处理分页
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	*/
	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Expecting 0")
	}

	result, err := GetAllObjectsByDocType(stub, args, "workflow")
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("end query_all_workflows")

	return shim.Success(result)
}

// ========================================================
// 查询可用流程
// ========================================================
func query_accessable_workflows(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting query_accessable_workflows")
	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Expecting 0")
	}

	submitter, err := GetSubmitterName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	submitterOrgName, err := GetOrgFromCertCommonName(submitter)
	if err != nil {
		return shim.Error(err.Error())
	}

	// 目前只处理了accessOrgs匹配的情况
	// TODO 添加角色匹配的工作流
	var queryBuffer bytes.Buffer
	queryBuffer.WriteString(`{"selector":{"docType":"workflow","enabled":true,"accessOrgs":{"$elemMatch":{"$eq":"`)
	queryBuffer.WriteString(submitterOrgName)
	queryBuffer.WriteString(`"}}}}`)

	result, err := GetQueryResult(stub, queryBuffer.String())
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("end query_accessable_workflows")
	return shim.Success(result)
}

// ========================================================
// 查询流程全部节点
// ========================================================
func GetAllNodesByWorkflowId(stub shim.ChaincodeStubInterface, workflowId string) ([]WorkflowNode, []byte, error) {
	var err error
	var results []WorkflowNode
	fmt.Println("starting queryAllNodesByWorkflowId")

	var queryBuffer bytes.Buffer
	queryBuffer.WriteString(`{"selector":{"docType":"workflowNode", "workflowId":"`)
	queryBuffer.WriteString(workflowId)
	queryBuffer.WriteString(`"}}`)

	resultAsBytes, err := GetQueryResult(stub, queryBuffer.String())
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal(resultAsBytes, &results) //un stringify it aka JSON.parse()
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("end queryAllNodesByWorkflowId")
	return results, resultAsBytes, nil
}

// =============================================================================
// 流程详情
// =============================================================================
func get_workflow_by_id(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	var workflow = Workflow{}
	fmt.Println("starting get_workflow_by_id")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	id := args[0]

	workflowDef, err := GetWorkflowDefById(stub, id)
	if err != nil {
		fmt.Println("This workflow does not exist - " + id)
		return shim.Error("This workflow does not exist - " + id)
	}

	workflowNodes, _, err := GetAllNodesByWorkflowId(stub, id)
	if err != nil {
		return shim.Error(err.Error())
	}

	workflow.WorkflowDef = workflowDef
	workflow.WorkflowNodes = workflowNodes

	workflowAsBytes, _ := json.Marshal(workflow)

	fmt.Println("- end get_workflow_by_id")
	return shim.Success(workflowAsBytes)
}

// =============================================================================
// 禁用/启用流程
// =============================================================================
func enable_or_disable_workflow(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting enable_or_disable_workflow")

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	id := args[0]
	enabled, err := strconv.ParseBool(args[1])
	if err != nil {
		return shim.Error(err.Error())
	}

	modifyTime := args[2]

	submitter, err := GetSubmitterName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	workflowDef, err := GetWorkflowDefById(stub, id)
	if err != nil {
		fmt.Println("This workflow does not exist - " + id)
		return shim.Error("This workflow does not exist - " + id)
	}

	submitterOrg, err := GetOrgFromCertCommonName(submitter)
	if err != nil {
		return shim.Error(err.Error())
	}

	creatorOrg, err := GetOrgFromCertCommonName(workflowDef.Creator)
	if err != nil {
		return shim.Error(err.Error())
	}

	if creatorOrg != submitterOrg {
		fmt.Println("Only creator can remove the workflow - " + id)
		return shim.Error("Only creator can remove the workflow - " + id)
	}

	if workflowDef.Enabled == enabled {
		fmt.Println("This workflow is already enabled/disabled - " + id)
		return shim.Error("This workflow is already enabled/disabled - " + id)
	}

	workflowDef.Enabled = enabled
	workflowDef.ModifyTime = modifyTime

	//store
	workflowDefAsBytes, _ := json.Marshal(workflowDef)
	err = stub.PutState(id, workflowDefAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("end enable_or_disable_workflow")
	return shim.Success(nil)
}

// =============================================================================
// 更新流程信息
// =============================================================================
func modify_workflow_def(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting modify_workflow_def")

	if len(args) < 4 {
		return shim.Error("Incorrect number of arguments. Expecting greater than 4")
	}

	if len(args)%2 != 0 {
		return shim.Error("Incorrect number of arguments. Expecting even number")
	}

	id := args[0]
	modifyTime := args[1]
	submitter, err := GetSubmitterName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	workflowDef, err := GetWorkflowDefById(stub, id)
	if err != nil {
		fmt.Println("This workflow def does not exist - " + id)
		return shim.Error("This workflow def does not exist - " + id)
	}

	for i := 2; i < len(args); i = i + 2 {
		key := args[i]
		value := args[i+1]
		err = UpdateStruct(&workflowDef, key, value)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	// append Modifiers
	workflowDef.LastModifier = submitter
	workflowDef.ModifyTime = modifyTime

	//store
	workflowDefAsBytes, _ := json.Marshal(workflowDef)
	err = stub.PutState(id, workflowDefAsBytes) //store with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("end modify_workflow_def")
	return shim.Success(nil)
}
