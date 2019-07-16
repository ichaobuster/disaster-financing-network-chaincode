package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type Process struct {
	DocType         string   `json:"docType"`
	Id              string   `json:"id"`
	AttachDocType   string   `json:"attachDocType"`
	AttachDocId     string   `json:"attachDocId"`
	AttachDocName   string   `json:"attachDocName"`
	WorkflowId      string   `json:"workflowId"`
	WorkflowName    string   `json:"workflowName"`
	CurrentNodeId   string   `json:"currentNodeId"`
	CurrentNodeName string   `json:"currentNodeName"`
	CurrentOwner    string   `json:"currentOwner"`
	Participants    []string `json:"participants"`
	Finished        bool     `json:"finished"`
	Canceled        bool     `json:"canceled"`
	Creator         string   `json:"creator"`      // 创建人
	LastModifier    string   `json:"lastModifier"` // 最后修改人
	CreateTime      string   `json:"createTime"`
	ModifyTime      string   `json:"modifyTime"`
}

type ProcessLog struct {
	DocType      string `json:"docType"`
	Id           string `json:"id"`
	ProcessId    string `json:"processId"`
	FromNodeId   string `json:"fromNodeId"`
	FromNodeName string `json:"fromNodeName"`
	FromOrg      string `json:"fromOrg"`
	ToNodeId     string `json:"toNodeId"`
	ToNodeName   string `json:"toNodeName"`
	ToOrg        string `json:"toOrg"`
	Operation    string `json:"operation"`
	Remark       string `json:"remark"`
	CreateTime   string `json:"createTime"`
}

// =============================================================================
// 启用流程实例
// =============================================================================
func start_process(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	var process Process
	fmt.Println("starting process")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	creator, err := GetSubmitterName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	// 获取流程实例
	err = json.Unmarshal([]byte(args[0]), &process)
	if err != nil {
		fmt.Println(err.Error())
		return shim.Error(err.Error())
	}

	//check if process already exists
	processInStore, err := GetProcessById(stub, process.Id)
	if err == nil {
		fmt.Println("This process already exists - " + process.Id)
		fmt.Println(processInStore)
		return shim.Error("This process already exists - " + process.Id)
	}

	//check if workflow exists and is enabled
	workflowDef, err := GetWorkflowDefById(stub, process.WorkflowId)
	if err != nil {
		fmt.Println("Workflow does not exist - " + process.WorkflowId)
		return shim.Error("Workflow does not exist - " + process.WorkflowId)
	}
	if !workflowDef.Enabled {
		fmt.Println("Workflow is disabled - " + process.WorkflowId)
		return shim.Error("Workflow is disabled - " + process.WorkflowId)
	}

	//check if attachDoc exists and docType is correct
	attachDocName, err := GetDocNameByDocTypeAndId(stub, process.AttachDocType, process.AttachDocId)
	if err != nil {
		fmt.Println("AttachDocId or attachDocType is incorrect")
		return shim.Error("AttachDocId or attachDocType is incorrect")
	}

	// check submitter's org and role
	creatorOrgName, err := GetOrgFromCertCommonName(creator)
	if err != nil {
		return shim.Error(err.Error())
	}

	workflowNodes, _, err := GetAllNodesByWorkflowId(stub, process.WorkflowId)
	if err != nil {
		return shim.Error(err.Error())
	}

	firstNode := workflowNodes[0]
	// check org or role
	if firstNode.AccessOrgs != nil {
		if !ContainsString(firstNode.AccessOrgs, creatorOrgName) {
			fmt.Println("Submitter's org are not allowed to start process.")
			return shim.Error("Submitter's org are not allowed to start process.")
		}
	}

	// TODO 补充检查角色的逻辑

	// store process
	process.DocType = "process"
	process.AttachDocName = attachDocName
	process.WorkflowName = workflowDef.WorkflowName
	process.CurrentNodeId = firstNode.Id
	process.CurrentNodeName = firstNode.NodeName
	process.CurrentOwner = creatorOrgName
	process.Canceled = false
	process.Creator = creator
	process.LastModifier = creator
	process.Participants = []string{creatorOrgName}
	process.ModifyTime = process.CreateTime

	processAsBytes, _ := json.Marshal(process)
	err = stub.PutState(process.Id, processAsBytes) //store with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	// store log
	err = StoreProcessLog(stub, true, process.Id, "Init", "开始", "", process.CurrentNodeId, process.CurrentNodeName, process.CurrentOwner, "InitProcess", "", process.CreateTime)

	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- process started: " + process.Id)
	return shim.Success(nil)
}

// =============================================================================
// Get Process By id
// =============================================================================
func GetProcessById(stub shim.ChaincodeStubInterface, id string) (Process, error) {
	var data Process
	dataAsBytes, err := stub.GetState(id) //getState retreives a key/value from the ledger
	if err != nil {                       //this seems to always succeed, even if key didn't exist
		return data, errors.New("Failed to find process - " + id)
	}
	json.Unmarshal(dataAsBytes, &data) //un stringify it aka JSON.parse()

	if data.Id != id {
		return data, errors.New("Process does not exist - " + id)
	}

	return data, nil
}

// =============================================================================
// 流程实例详情
// =============================================================================
func get_process_by_id(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting get_process_by_id")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	id := args[0]

	process, err := GetProcessById(stub, id)
	if err != nil {
		fmt.Println("This workflow does not exist - " + id)
		return shim.Error("This workflow does not exist - " + id)
	}

	processAsBytes, _ := json.Marshal(process)

	fmt.Println("- end get_process_by_id")
	return shim.Success(processAsBytes)
}

// ========================================================
// 查询全部日志
// ========================================================
func query_logs_by_process_id(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting query_logs_by_process_id")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	processId := args[0]

	_, result, err := GetLogsByProcessId(stub, processId)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("end query_logs_by_process_id")

	return shim.Success(result)
}

// ========================================================
// 查询全部日志
// ========================================================
func GetLogsByProcessId(stub shim.ChaincodeStubInterface, processId string) ([]ProcessLog, []byte, error) {
	var err error
	var results []ProcessLog
	fmt.Println("starting GetLogsByProcessId")

	var queryBuffer bytes.Buffer
	queryBuffer.WriteString(`{"selector":{"docType":"processLog", "processId":"`)
	queryBuffer.WriteString(processId)
	queryBuffer.WriteString(`"}}`)

	resultAsBytes, err := GetQueryResult(stub, queryBuffer.String())
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal(resultAsBytes, &results) //un stringify it aka JSON.parse()
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("end GetLogsByProcessId")
	return results, resultAsBytes, nil
}

// ========================================================
// 存储日志
// ========================================================
func StoreProcessLog(stub shim.ChaincodeStubInterface, isInit bool, processId string, fromNodeId string, fromNodeName string, fromOrg string, toNodeId string, toNodeName string, toOrg string, operation string, remark string, createTime string) error {
	var err error
	logsLens := "0"
	if !isInit {
		// query logs
		logs, _, err := GetLogsByProcessId(stub, processId)
		if err != nil {
			return err
		}
		logsLens = strconv.Itoa(len(logs))
	}

	// store log
	var log = ProcessLog{}
	log.Id = "processLog-" + processId + "-" + logsLens
	log.DocType = "processLog"
	log.ProcessId = processId
	log.FromNodeId = fromNodeId
	log.FromNodeName = fromNodeName
	log.FromOrg = fromOrg
	log.ToNodeId = toNodeId
	log.ToNodeName = toNodeName
	log.ToOrg = toOrg
	log.Operation = operation
	log.Remark = remark
	log.CreateTime = createTime

	logAsBytes, _ := json.Marshal(log)
	err = stub.PutState(log.Id, logAsBytes) //store with id as key
	if err != nil {
		return err
	}

	SendProcessLogEvent(stub, log, logAsBytes)

	return nil
}

// =============================================================================
// 将流转日志作为event发送
// =============================================================================
func SendProcessLogEvent(stub shim.ChaincodeStubInterface, log ProcessLog, logAsBytes []byte) {
	SendEvent(stub, log.Operation, logAsBytes);
}

// =============================================================================
// 流程运行
// =============================================================================
func transfer_process(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting transfer_process")

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	submitter, err := GetSubmitterName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	// check submitter's org and role
	submitterOrgName, err := GetOrgFromCertCommonName(submitter)
	if err != nil {
		return shim.Error(err.Error())
	}

	processId := args[0]
	nextNodeId := args[1]
	nextOwner := args[2]
	modifyTime := args[3]

	//check if process already exists
	process, err := GetProcessById(stub, processId)
	if err != nil {
		fmt.Println("This process does not exists - " + processId)
		return shim.Error("This process does not exists - " + processId)
	}

	if process.Canceled {
		fmt.Println("This process has been canceled - " + processId)
		return shim.Error("This process has been canceled - " + processId)
	}

	if process.Finished {
		fmt.Println("This process has been finished - " + processId)
		return shim.Error("This process has been finished - " + processId)
	}

	// check if submitter's org is current owner's org
	if submitterOrgName != process.CurrentOwner {
		fmt.Println("You are not allowed to transfer the process - " + submitterOrgName)
		return shim.Error("You are not allowed to transfer the process - " + submitterOrgName)
	}

	// transfer to next node
	currentNode, err := GetWorkflowNodeById(stub, process.CurrentNodeId)
	if err != nil {
		return shim.Error(err.Error())
	}

	if !currentNode.LastNode {
		// check if can transfer to the node
		if !ContainsString(currentNode.NextNodeIds, nextNodeId) {
			fmt.Println("You are not allowed to transfer to the node - " + nextNodeId)
			return shim.Error("You are not allowed to transfer to the node - " + nextNodeId)
		}

		nextNode, err := GetWorkflowNodeById(stub, nextNodeId)
		if err != nil {
			return shim.Error(err.Error())
		}

		// check org or role
		if nextNode.AccessOrgs != nil {
			if !ContainsString(nextNode.AccessOrgs, nextOwner) {
				fmt.Println("You are not allowed to transfer to next owner - " + nextOwner)
				return shim.Error("You are not allowed to transfer to next owner - " + nextOwner)
			}
		}

		// TODO 补充检查角色的逻辑

		// store process
		process.CurrentNodeId = nextNode.Id
		process.CurrentNodeName = nextNode.NodeName
		process.CurrentOwner = nextOwner

	} else {
		// finish the process
		// store process
		process.Finished = true
		process.CurrentNodeId = "Finish"
		process.CurrentNodeName = "结束"
		process.CurrentOwner = ""
	}

	process.LastModifier = submitter
	process.ModifyTime = modifyTime
	if !ContainsString(process.Participants, submitterOrgName) {
		process.Participants = append(process.Participants, submitterOrgName)
	}

	processAsBytes, _ := json.Marshal(process)
	err = stub.PutState(process.Id, processAsBytes) //store with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	// store log
	err = StoreProcessLog(stub, false, processId, currentNode.Id, currentNode.NodeName, submitterOrgName, process.CurrentNodeId, process.CurrentNodeName, process.CurrentOwner, "TransferProcess", "", modifyTime)

	fmt.Println("- end transfer_process")
	return shim.Success(nil)
}

// =============================================================================
// 流程回退
// =============================================================================
func return_process(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting return_process")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	submitter, err := GetSubmitterName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	// check submitter's org and role
	submitterOrgName, err := GetOrgFromCertCommonName(submitter)
	if err != nil {
		return shim.Error(err.Error())
	}

	processId := args[0]
	modifyTime := args[1]

	//check if process already exists
	process, err := GetProcessById(stub, processId)
	if err != nil {
		fmt.Println("This process does not exists - " + processId)
		return shim.Error("This process does not exists - " + processId)
	}

	if process.Canceled {
		fmt.Println("This process has been canceled - " + processId)
		return shim.Error("This process has been canceled - " + processId)
	}

	if process.Finished {
		fmt.Println("This process has been finished - " + processId)
		return shim.Error("This process has been finished - " + processId)
	}

	// check if submitter's org is current owner's org
	if submitterOrgName != process.CurrentOwner {
		fmt.Println("You are not allowed to return the process - " + submitterOrgName)
		return shim.Error("You are not allowed to return the process - " + submitterOrgName)
	}

	// get current node
	currentNode, err := GetWorkflowNodeById(stub, process.CurrentNodeId)
	if err != nil {
		return shim.Error(err.Error())
	}

	if currentNode.FirstNode {
		fmt.Println("The process can not be returned again - " + processId)
		return shim.Error("The process can not be returned again - " + processId)
	}

	// 根据流转日志对流程进行回退
	var queryBuffer bytes.Buffer
	queryBuffer.WriteString(`{"selector":{"docType":"processLog","processId":"`)
	queryBuffer.WriteString(processId)
	queryBuffer.WriteString(`","operation":"TransferProcess","toOrg":"`)
	queryBuffer.WriteString(submitterOrgName)
	queryBuffer.WriteString(`","toNodeId":"`)
	queryBuffer.WriteString(currentNode.Id)
	queryBuffer.WriteString(`"}}`)
	resultAsBytes, err := GetQueryResult(stub, queryBuffer.String())
	if err != nil {
		fmt.Println(err.Error())
		return shim.Error(err.Error())
	}

	var logs []ProcessLog
	err = json.Unmarshal(resultAsBytes, &logs) //un stringify it aka JSON.parse()
	if err != nil {
		fmt.Println(err.Error())
	}

	if logs == nil {
		fmt.Println("Can not find process logs to return -" + processId)
		return shim.Error("Can not find process logs to return -" + processId)
	}

	targetLog := logs[0]
	idPrefix := "processLog-" + processId + "-"
	logId, _ := strconv.Atoi(strings.TrimPrefix(targetLog.Id, idPrefix))
	if len(logs) > 1 {
		for i := 1; i < len(logs); i++ {
			newLogId, _ := strconv.Atoi(strings.TrimPrefix(logs[i].Id, idPrefix))
			if newLogId > logId {
				targetLog = logs[i]
				logId = newLogId
			}
		}
	}

	// store process
	process.CurrentNodeId = targetLog.FromNodeId
	process.CurrentNodeName = targetLog.FromNodeName
	process.CurrentOwner = targetLog.FromOrg
	process.LastModifier = submitter
	process.ModifyTime = modifyTime

	processAsBytes, _ := json.Marshal(process)
	err = stub.PutState(process.Id, processAsBytes) //store with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	// store log
	err = StoreProcessLog(stub, false, processId, currentNode.Id, currentNode.NodeName, submitterOrgName, process.CurrentNodeId, process.CurrentNodeName, process.CurrentOwner, "ReturnProcess", "", modifyTime)

	fmt.Println("- end return_process")
	return shim.Success(nil)
}

// =============================================================================
// 流程撤回
// =============================================================================
func withdraw_process(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting withdraw_process")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	submitter, err := GetSubmitterName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	// check submitter's org and role
	submitterOrgName, err := GetOrgFromCertCommonName(submitter)
	if err != nil {
		return shim.Error(err.Error())
	}

	processId := args[0]
	modifyTime := args[1]

	//check if process already exists
	process, err := GetProcessById(stub, processId)
	if err != nil {
		fmt.Println("This process does not exists - " + processId)
		return shim.Error("This process does not exists - " + processId)
	}

	if process.Canceled {
		fmt.Println("This process has been canceled - " + processId)
		return shim.Error("This process has been canceled - " + processId)
	}

	// TODO 添加其他检查条件

	// get current node
	currentNode, err := GetWorkflowNodeById(stub, process.CurrentNodeId)
	if err != nil {
		return shim.Error(err.Error())
	}

	if currentNode.FirstNode {
		fmt.Println("The process can not be withdrawed again - " + processId)
		return shim.Error("The process can not be withdrawed again - " + processId)
	}

	// 根据流转日志对流程进行回退
	var queryBuffer bytes.Buffer
	queryBuffer.WriteString(`{"selector":{"docType":"processLog","processId":"`)
	queryBuffer.WriteString(processId)
	queryBuffer.WriteString(`","operation":"TransferProcess","toOrg":"`)
	queryBuffer.WriteString(process.CurrentOwner)
	queryBuffer.WriteString(`","toNodeId":"`)
	queryBuffer.WriteString(currentNode.Id)
	queryBuffer.WriteString(`"}}`)
	resultAsBytes, err := GetQueryResult(stub, queryBuffer.String())
	if err != nil {
		fmt.Println(err.Error())
		return shim.Error(err.Error())
	}

	var logs []ProcessLog
	err = json.Unmarshal(resultAsBytes, &logs) //un stringify it aka JSON.parse()
	if err != nil {
		fmt.Println(err.Error())
	}

	if logs == nil {
		fmt.Println("Can not find process logs to return -" + processId)
		return shim.Error("Can not find process logs to return -" + processId)
	}

	targetLog := logs[0]
	idPrefix := "processLog-" + processId + "-"
	logId, _ := strconv.Atoi(strings.TrimPrefix(targetLog.Id, idPrefix))
	if len(logs) > 1 {
		for i := 1; i < len(logs); i++ {
			newLogId, _ := strconv.Atoi(strings.TrimPrefix(logs[i].Id, idPrefix))
			if newLogId > logId {
				targetLog = logs[i]
				logId = newLogId
			}
		}
	}

	// check if submitter's org can withdraw the process
	if targetLog.FromOrg != submitterOrgName {
		fmt.Println("You are not allowed to withdraw the process - " + submitterOrgName)
		return shim.Error("You are not allowed to withdraw the process - " + submitterOrgName)
	}

	// store process
	if process.Finished {
		process.Finished = false
	}
	process.CurrentNodeId = targetLog.FromNodeId
	process.CurrentNodeName = targetLog.FromNodeName
	process.CurrentOwner = submitterOrgName
	process.LastModifier = submitter
	process.ModifyTime = modifyTime

	processAsBytes, _ := json.Marshal(process)
	err = stub.PutState(process.Id, processAsBytes) //store with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	// store log
	err = StoreProcessLog(stub, false, processId, currentNode.Id, currentNode.NodeName, targetLog.ToOrg, process.CurrentNodeId, process.CurrentNodeName, process.CurrentOwner, "WithdrawProcess", "", modifyTime)

	fmt.Println("- end withdraw_process")
	return shim.Success(nil)
}

// =============================================================================
// 取消流程
// =============================================================================
func cancel_process(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting cancel_process")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	submitter, err := GetSubmitterName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	// check submitter's org and role
	submitterOrgName, err := GetOrgFromCertCommonName(submitter)
	if err != nil {
		return shim.Error(err.Error())
	}

	processId := args[0]
	modifyTime := args[1]

	//check if process already exists
	process, err := GetProcessById(stub, processId)
	if err != nil {
		fmt.Println("This process does not exists - " + processId)
		return shim.Error("This process does not exists - " + processId)
	}

	if process.Canceled {
		fmt.Println("This process has been canceled - " + processId)
		return shim.Error("This process has been canceled - " + processId)
	}

	if process.Finished {
		fmt.Println("This process has been finished - " + processId)
		return shim.Error("This process has been finished - " + processId)
	}

	// TODO 其他约束条件

	// check if submitter's org is creator's org
	creatorOrgName, err := GetOrgFromCertCommonName(process.Creator)
	if err != nil {
		return shim.Error(err.Error())
	}

	if submitterOrgName != creatorOrgName {
		fmt.Println("You are not allowed to cancel the process - " + submitterOrgName)
		return shim.Error("You are not allowed to cancel the process - " + submitterOrgName)
	}

	// cancel process
	process.Canceled = true
	process.LastModifier = submitter
	process.ModifyTime = modifyTime

	processAsBytes, _ := json.Marshal(process)
	err = stub.PutState(process.Id, processAsBytes) //store with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	// store logs
	err = StoreProcessLog(stub, false, processId, process.CurrentNodeId, process.CurrentNodeName, process.CurrentOwner, "Canceled", "取消", "", "CancelProcess", "", modifyTime)

	fmt.Println("- end cancel_process")
	return shim.Success(nil)
}

// =============================================================================
// 查询待办流程
// =============================================================================
func query_todo_process(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var result []byte
	var err error
	fmt.Println("starting query_todo_process")

	/* 暂时无法处理分页条件
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	*/
	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Expecting 0")
	}

	/*
		limit, skip, err := SanitizePagingArgument(args[0:2])
		if err != nil {
			return shim.Error(err.Error())
		}
	*/

	submitter, err := GetSubmitterName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	submitterOrgName, err := GetOrgFromCertCommonName(submitter)
	if err != nil {
		return shim.Error(err.Error())
	}

	var queryBuffer bytes.Buffer
	/*
		queryBuffer.WriteString(`{"selector":{"docType":"process","finished":false,"canceled":false,"currentOwner":"`)
		queryBuffer.WriteString(submitterOrgName)
		queryBuffer.WriteString(`"},"limit":`)
		queryBuffer.WriteString(limit)
		queryBuffer.WriteString(`,"skip":`)
		queryBuffer.WriteString(skip)
		queryBuffer.WriteString(`"}`)
	*/
	queryBuffer.WriteString(`{"selector":{"docType":"process","finished":false,"canceled":false,"currentOwner":"`)
	queryBuffer.WriteString(submitterOrgName)
	queryBuffer.WriteString(`"}}`)

	result, err = GetQueryResult(stub, queryBuffer.String())
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("end query_todo_process")

	return shim.Success(result)
}

// =============================================================================
// 查询已办流程
// =============================================================================
func query_done_process(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var result []byte
	var err error
	fmt.Println("starting query_done_process")

	/* 暂时无法处理分页条件
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	*/
	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Expecting 0")
	}

	/*
		limit, skip, err := SanitizePagingArgument(args[0:2])
		if err != nil {
			return shim.Error(err.Error())
		}
	*/

	submitter, err := GetSubmitterName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	submitterOrgName, err := GetOrgFromCertCommonName(submitter)
	if err != nil {
		return shim.Error(err.Error())
	}

	var queryBuffer bytes.Buffer
	/*
		queryBuffer.WriteString(`{"selector":{"docType":"process","participants":{"$elemMatch":{"$eq":"`)
		queryBuffer.WriteString(submitterOrgName)
		queryBuffer.WriteString(`"}}},"limit":`)
		queryBuffer.WriteString(limit)
		queryBuffer.WriteString(`,"skip":`)
		queryBuffer.WriteString(skip)
		queryBuffer.WriteString(`}`)
	*/
	queryBuffer.WriteString(`{"selector":{"docType":"process","participants":{"$elemMatch":{"$eq":"`)
	queryBuffer.WriteString(submitterOrgName)
	queryBuffer.WriteString(`"}}}}`)

	result, err = GetQueryResult(stub, queryBuffer.String())
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("end query_done_process")

	return shim.Success(result)
}
