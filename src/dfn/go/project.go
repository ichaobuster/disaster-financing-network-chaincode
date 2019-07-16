package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// ----- Project ----- //
type Project struct {
	DocType            string `json:"docType"`
	Id                 string `json:"id"`
	ProjectName        string `json:"projectName"`
	Scale              string `json:"scale"`              // 发行规模
	BasicAssets        string `json:"basicAssets"`        // 基础资产
	Initiator          string `json:"initiator"`          // 发起机构
	Trustee            string `json:"trustee"`            // 受托机构
	Depositary         string `json:"depositary"`         // 资金保管机构
	Agent              string `json:"agent"`              // 登记/支付代理机构
	AssetService       string `json:"assetService"`       // 资产服务机构
	Assessor           string `json:"assessor"`           // 评估机构
	CreditRater        string `json:"creditRater"`        // 信用评级机构
	LiquiditySupporter string `json:"liquiditySupporter"` // 流动性支持机构
	Underwriter        string `json:"underwriter"`        // 承销商/薄记管理人机构
	Lawyer             string `json:"lawyer"`             // 律师
	Accountant         string `json:"accountant"`         // 会计师
	Creator            string `json:"creator"`            // 创建人
	LastModifier       string `json:"lastModifier"`       // 最后修改人
	CreateTime         string `json:"createTime"`         // 创建时间
	ModifyTime         string `json:"modifyTime"`         // 修改时间
}

// =============================================================================
// 创建项目
// =============================================================================
func create_project(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	var project Project
	fmt.Println("starting create_project")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	err = json.Unmarshal([]byte(args[0]), &project)
	if err != nil {
		fmt.Println(err.Error())
		return shim.Error(err.Error())
	}

	creator, err := GetSubmitterName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	//check if project id already exists
	projectInStore, err := GetProjectById(stub, project.Id)
	if err == nil {
		fmt.Println("This project already exists - " + project.Id)
		fmt.Println(projectInStore)
		return shim.Error("This project already exists - " + project.Id)
	}

	project.DocType = "project"
	project.Creator = creator
	project.LastModifier = creator
	project.ModifyTime = project.CreateTime

	fmt.Println(project)

	//store project
	projectAsBytes, _ := json.Marshal(project)

	err = PutState(stub, project.Id, projectAsBytes) //store with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end create_project")
	return shim.Success(nil)
}

// =============================================================================
// 项目详情
// =============================================================================
func get_project_by_id(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting get_project_by_id")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	id := args[0]

	project, err := GetProjectById(stub, id)
	if err != nil {
		fmt.Println("This project does not exist - " + id)
		return shim.Error("This project does not exist - " + id)
	}

	projectAsBytes, _ := json.Marshal(project)

	fmt.Println("- end get_project_by_id")
	return shim.Success(projectAsBytes)
}

// =============================================================================
// Get Project By id
// =============================================================================
func GetProjectById(stub shim.ChaincodeStubInterface, id string) (Project, error) {
	var data Project
	dataAsBytes, err := stub.GetState(id) //getState retreives a key/value from the ledger
	if err != nil {                       //this seems to always succeed, even if key didn't exist
		return data, errors.New("Failed to find project - " + id)
	}
	json.Unmarshal(dataAsBytes, &data) //un stringify it aka JSON.parse()

	if data.Id != id {
		return data, errors.New("Project does not exist - " + id)
	}

	return data, nil
}

// ========================================================
// 查询全部项目
// ========================================================
func query_all_projects(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting query_all_projects")

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Expecting 0")
	}

	result, err := GetAllObjectsByDocType(stub, args, "project")
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("end query_all_projects")

	return shim.Success(result)
}

// ========================================================
// 分页查询项目
// ========================================================
func query_paging_projects(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting query_paging_projects")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	result, err := GetPagingObjectsByDocType(stub, args, "project")
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("end query_paging_projects")

	return shim.Success(result)
}

// =============================================================================
// 删除项目
// =============================================================================
func remove_project(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting remove_project")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	id := args[0]
	submitter, err := GetSubmitterName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	project, err := GetProjectById(stub, id)
	if err != nil {
		fmt.Println("This project does not exist - " + id)
		return shim.Error("This project does not exist - " + id)
	}

	if project.Creator != submitter {
		fmt.Println("Only creator can remove the project - " + id)
		return shim.Error("Only creator can remove the project - " + id)
	}

	err = DelState(stub, id, "project")
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("end remove_project")
	return shim.Success(nil)
}

// =============================================================================
// 更新项目
// =============================================================================
func modify_project(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting modify_project")

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

	project, err := GetProjectById(stub, id)
	if err != nil {
		fmt.Println("This project does not exist - " + id)
		return shim.Error("This project does not exist - " + id)
	}

	for i := 2; i < len(args); i = i + 2 {
		key := args[i]
		value := args[i+1]
		err = UpdateStruct(&project, key, value)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	// append Modifiers
	project.LastModifier = submitter
	project.ModifyTime = modifyTime

	//store project
	projectAsBytes, _ := json.Marshal(project)
	err = PutState(stub, id, projectAsBytes) //store with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("end modify_project")
	return shim.Success(nil)
}
