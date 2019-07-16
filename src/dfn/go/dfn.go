package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode - %s", err)
	}
}

// ============================================================================================================================
// Init - initialize the chaincode
//
// Marbles does not require initialization, so let's run a simple test instead.
//
// Shows off PutState() and how to pass an input argument to chaincode.
// Shows off GetFunctionAndParameters() and GetStringArgs()
// Shows off GetTxID() to get the transaction ID of the proposal
//
// Inputs - Array of strings
//  ["314"]
//
// Returns - shim.Success or error
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Simple ABS Contract Is Starting Up")
	funcName, args := stub.GetFunctionAndParameters()
	var number int
	var err error
	txId := stub.GetTxID()

	fmt.Println("Init() is running")
	fmt.Println("Transaction ID:", txId)
	fmt.Println("  GetFunctionAndParameters() function:", funcName)
	fmt.Println("  GetFunctionAndParameters() args count:", len(args))
	fmt.Println("  GetFunctionAndParameters() args found:", args)

	// expecting 1 arg for instantiate or upgrade
	if len(args) == 1 {
		fmt.Println("  GetFunctionAndParameters() arg[0] length", len(args[0]))

		// expecting arg[0] to be length 0 for upgrade
		if len(args[0]) == 0 {
			fmt.Println("  Uh oh, args[0] is empty...")
		} else {
			fmt.Println("  Great news everyone, args[0] is not empty")

			// convert numeric string to integer
			number, err = strconv.Atoi(args[0])
			if err != nil {
				return shim.Error("Expecting a numeric string argument to Init() for instantiate")
			}

			// this is a very simple test. let's write to the ledger and error out on any errors
			// it's handy to read this right away to verify network is healthy if it wrote the correct value
			err = stub.PutState("selftest", []byte(strconv.Itoa(number)))
			if err != nil {
				return shim.Error(err.Error()) //self-test fail
			}
		}
	}

	// showing the alternative argument shim function
	alt := stub.GetStringArgs()
	fmt.Println("  GetStringArgs() args count:", len(alt))
	fmt.Println("  GetStringArgs() args found:", alt)

	// store compatible marbles application version
	/*
		err = stub.PutState("simple_abs_ui", []byte("0.0.1"))
		if err != nil {
			return shim.Error(err.Error())
		}
	*/

	fmt.Println("Ready for action") //self-test pass
	return shim.Success(nil)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	// tMap, _ := stub.GetTransient()
	fmt.Println(" ")
	fmt.Println("starting invoke, for - " + function)

	// Handle different functions
	switch function {
	case "create_project":
		return create_project(stub, args)
	case "get_project_by_id":
		return get_project_by_id(stub, args)
	case "query_all_projects":
		return query_all_projects(stub, args)
	case "query_paging_projects":
		return query_paging_projects(stub, args)
	case "remove_project":
		return remove_project(stub, args)
	case "modify_project":
		return modify_project(stub, args)
	case "create_linear_workflow":
		return create_linear_workflow(stub, args)
	case "get_workflow_by_id":
		return get_workflow_by_id(stub, args)
	case "query_all_workflows":
		return query_all_workflows(stub, args)
	case "enable_or_disable_workflow":
		return enable_or_disable_workflow(stub, args)
	case "modify_workflow_def":
		return modify_workflow_def(stub, args)
	case "query_accessable_workflows":
		return query_accessable_workflows(stub, args)
	case "start_process":
		return start_process(stub, args)
	case "get_process_by_id":
		return get_process_by_id(stub, args)
	case "query_logs_by_process_id":
		return query_logs_by_process_id(stub, args)
	case "transfer_process":
		return transfer_process(stub, args)
	case "return_process":
		return return_process(stub, args)
	case "withdraw_process":
		return withdraw_process(stub, args)
	case "cancel_process":
		return cancel_process(stub, args)
	case "query_todo_process":
		return query_todo_process(stub, args)
	case "query_done_process":
		return query_done_process(stub, args)
	case "save_org_public_key":
		return save_org_public_key(stub, args)
	case "encrypt_data":
		return encrypt_data(stub, args)
	case "decrypt_data":
		return decrypt_data(stub, args)
	default:
		// error out
		fmt.Println("Received unknown invoke function name - " + function)
		return shim.Error("Received unknown invoke function name - '" + function + "'")
	}
}
