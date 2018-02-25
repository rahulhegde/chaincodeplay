package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"fmt"
	"encoding/json"
)

// asset definition
type PersistentCoin uint64
type PersistentCoinAccountId string

// key-value store - asset profile definition
type PersistentCoinAccount struct {
	AccountID 	PersistentCoinAccountId
	Balance		PersistentCoin
}

// contracts from business function - creating an account to hold asset
type PersistentCoinAccountCreateContract struct{
	AccountId 	PersistentCoinAccountId
	Balance   	PersistentCoin
}

// contracts from business function - transferring asset
type CoinTransferContract struct{
	FromAccountId PersistentCoinAccountId
	ToAccountId   PersistentCoinAccountId
	CoinCount     PersistentCoin
}

// chaincode or smart contract application
type PersistentCoinApplication struct {
}

// initialization function for chaincode
func (app *PersistentCoinApplication) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// business invoke function for chaincode
func (app *PersistentCoinApplication) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "CreatePersistentCoinAccount" {
		return app.CreatePersistentCoinAccount(stub, args[0])
	} else if function == "TransferPersistentCoin" {
		return 	app.TransferPersistentCoin(stub, args[0])
	} else {
		shim.Error("rejecting endorsement - unknown business invoke function")
	}
	return shim.Success(nil)
}

func (app *PersistentCoinApplication) CreatePersistentCoinAccount (stub shim.ChaincodeStubInterface,
serializedContract string ) pb.Response {

	var contract PersistentCoinAccountCreateContract

	json.Unmarshal(serializedContract, contract)
	fmt.Println("printing contract: ", contract)

	// check account already exist
	_, err := stub.GetState(contract.AccountId)
	if err != nil  {
		return shim.Error("rejecting endorsement: Persistent Coin AccountId: " +
			contract.AccountId + " already exist!")
	}

	// since account does not exist
	err = stub.PutState(contract.AccountId, serializedContract)
	if err != nil {
		return shim.Error("rejecting endorsement: Problem creating Persistent Account Id: " +
			contract.AccountId + " " + err.Error())
	}

	// mark the endorsement as success
	return shim.Success(nil)
}

func (app *PersistentCoinApplication) TransferPersistentCoin (stub shim.ChaincodeStubInterface,
serializedContract string ) (pb.Response) {
	var fromAccount, toAccount PersistentCoinAccount
	var contract CoinTransferContract

	// un marshal the contract for execution
	json.Unmarshal(serializedContract, contract)

	// fetch the account information and un marshal to a go-struct for use
	serializedAccountBytes, err := stub.GetState(contract.FromAccountId)
	if err != nil {
		return shim.Error("rejecting endorsement: Persistent Coin AccountId: " +
			contract.FromAccountId + " does not exist!")
	}
	if json.Unmarshal(serializedAccountBytes, fromAccount) != nil {
		return shim.Error("rejecting endorsement: json unmarshaling failed for from account")
	}

	// fetch the account information and un marshal to a go-struct for use
	serializedAccountBytes, err = stub.GetState(contract.ToAccountId)
	if err != nil {
		return shim.Error("rejecting endorsement: Persistent Coin AccountId: " +
			contract.ToAccountId + " does not exist!")
	}
	if json.Unmarshal(serializedAccountBytes, toAccount) != nil {
		return shim.Error("rejecting endorsement: json unmarshaling failed for to account")
	}

	// check if enough coins is available in the account for transfer
	if contract.CoinCount > fromAccount.Balance {
		return shim.Error("rejecting endorsement: not enough coins available")
	}

	// transfer persistent coin after successful initial persistent coin account, balance verification
	fromAccount.Balance -= contract.CoinCount
	toAccount.Balance += contract.CoinCount

	// propose ledger update for persistent coin account transfer
	serializedAccountBytes, err = json.Marshal(fromAccount)
	err = stub.PutState(fromAccount.AccountID, serializedAccountBytes)
	if err != nil {
		return shim.Error("rejecting endorsement: error saving from account: " + fromAccount.AccountID + " " + err.Error())
	}

	serializedAccountBytes, err = json.Marshal(toAccount)
	err = stub.PutState(toAccount.AccountID, serializedAccountBytes)
	if err != nil {
		return shim.Error("rejecting endorsement: error saving from account: " + toAccount.AccountID + " " + err.Error())
	}

	// mark the endorsement as success
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(PersistentCoinApplication))
	if err != nil {
		fmt.Println("error starting lottery application" + err.Error())
	}
}