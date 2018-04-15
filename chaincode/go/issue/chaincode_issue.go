
package main

import (
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/pem"
	"crypto/x509"
	"strings"
)

var logger = shim.NewLogger("IssueChaincode")

// IssueChaincode example simple Chaincode implementation
type IssueChaincode struct {
}

func (t *IssueChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")

	err := stub.PutState("accountCounter", []byte(strconv.Itoa(1000)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState("cardCounter", []byte(strconv.Itoa(0)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *IssueChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(err.Error())
	}

	name, org := getCreator(creatorBytes)

	logger.Debug("transaction creator " + name + "@" + org)

	function, args := stub.GetFunctionAndParameters()
	if function == "create" {
		return t.create(stub, name)
	} else if function == "query" {
		return t.query(stub, args)
	}

	return pb.Response{Status:400, Message:"Invalid invoke function name."}
}

func (t *IssueChaincode) create(stub shim.ChaincodeStubInterface, name string) pb.Response {

	cardCounterBytes, err := stub.GetState("cardCounter")
	if err != nil {
		return shim.Error(err.Error())
	}
	cardCounterVal, _ := strconv.Atoi(string(cardCounterBytes))
	cardCounterVal = cardCounterVal + 1

	err = stub.PutState("cardCounter", []byte(strconv.Itoa(cardCounterVal)))
	if err != nil {
		return shim.Error(err.Error())
	}

	accountCounterBytes, err := stub.GetState("accountCounter")
	if err != nil {
		return shim.Error(err.Error())
	}
	accountCounterVal, _ := strconv.Atoi(string(accountCounterBytes))
	accountCounterVal = accountCounterVal + 1

	err = stub.PutState("accountCounter", []byte(strconv.Itoa(accountCounterVal)))
	if err != nil {
		return shim.Error(err.Error())
	}

	cardKey, _ := stub.CreateCompositeKey("Card", []string{strconv.Itoa(cardCounterVal)})

	err = stub.PutState(cardKey, []byte(strconv.Itoa(accountCounterVal)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(cardKey))
}

func (t *IssueChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return pb.Response{Status:400, Message:"Incorrect number of arguments"}
	}

	cardNumber := args[0]

	cardKey, _ := stub.CreateCompositeKey("Card", []string{cardNumber})

	// Get the state from the ledger
	valBytes, err := stub.GetState(cardKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	if valBytes == nil {
		return pb.Response{Status:404, Message:"Entity not found"}
	}

	return shim.Success(valBytes)
}

var getCreator = func (certificate []byte) (string, string) {
	data := certificate[strings.Index(string(certificate), "-----"): strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	commonName := cert.Subject.CommonName
	logger.Debug("commonName: " + commonName + ", organization: " + organization)

	organizationShort := strings.Split(organization, ".")[0]

	return commonName, organizationShort
}

func main() {
	err := shim.Start(new(IssueChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}
