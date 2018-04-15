
package main

import (
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/pem"
	"crypto/x509"
	"strings"
)

var logger = shim.NewLogger("PaymentChaincode")

// PaymentChaincode example simple Chaincode implementation
type PaymentChaincode struct {
}

func (t *PaymentChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")

	return shim.Success(nil)
}

func (t *PaymentChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(err.Error())
	}

	name, org := getCreator(creatorBytes)

	logger.Debug("transaction creator " + name + "@" + org)

	function, args := stub.GetFunctionAndParameters()
	if function == "request" {
		// Make payment of x units from a to b
		return t.request(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemented in invoke
		return t.query(stub, args)
	}

	return pb.Response{Status:403, Message:"Invalid invoke function name."}
}

// Transaction makes payment of x units from a to b
func (t *PaymentChaincode) request(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	cardNumber := args[0]
	company := args[2]


	cardKey, _ := stub.CreateCompositeKey("Card", []string{cardNumber, company})
	amount, _ := strconv.Atoi(args[1])

	// Get the state from the ledger
	amountBytes, err := stub.GetState(cardKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	
	if amountBytes != nil {
		amountOldVal, _ := strconv.Atoi(string(amountBytes))
		amount = amount + amountOldVal
	}
	

	// Write the state back to the ledger
	err = stub.PutState(cardKey, []byte(strconv.Itoa(amount)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(strconv.Itoa(amount)))
}


// read value
func (t *PaymentChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var err error

	if len(args) != 2 {
		return pb.Response{Status:403, Message:"Incorrect number of arguments"}
	}

	cardNumber := args[0]
	company := args[1]

	cardKey, _ := stub.CreateCompositeKey("Card", []string{cardNumber, company})

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
	err := shim.Start(new(PaymentChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}
