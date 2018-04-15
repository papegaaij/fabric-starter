
package main

import (
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/pem"
	"crypto/x509"
	"strings"
)

var logger = shim.NewLogger("TravelChaincode")

// TravelChaincode example simple Chaincode implementation
type TravelChaincode struct {
}

func (t *TravelChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")

	return shim.Success(nil)
}

func (t *TravelChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(err.Error())
	}

	name, org := getCreator(creatorBytes)

	logger.Debug("transaction creator " + name + "@" + org)

	function, args := stub.GetFunctionAndParameters()
	if function == "swipe" {
		return t.swipe(stub, name, args)
	} else if function == "query" {
		return t.query(stub, args)
	}

	return pb.Response{Status:400, Message:"Invalid invoke function name."}
}

func (t *TravelChaincode) swipe(stub shim.ChaincodeStubInterface, name string, args []string) pb.Response {

	if len(args) != 1 {
		return pb.Response{Status:400, Message:"Incorrect number of arguments"}
	}

	cardNumber := args[0]

	cardKey, _ := stub.CreateCompositeKey("Card", []string{cardNumber})

	locBytes, err := stub.GetState(cardKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	if locBytes == nil {
		// checking in
		err := stub.PutState(cardKey, []byte(name))
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		// checking out
		from := string(locBytes)
		to := name

		var veoliaShare, nsShare int

		if from == "amsterdam" && to == "amsterdam" {
			nsShare = 100
			veoliaShare = 0
		} else if (from == "amsterdam" && to == "rotterdam") || (from == "rotterdam" && to == "amsterdam") {
			nsShare = 40
			veoliaShare = 60
		} else if from == "rotterdam" && to == "rotterdam" {
			nsShare = 0
			veoliaShare = 100
		}

		if nsShare > 0 {
			stub.SetEvent("nsPay", []byte(strconv.Itoa(nsShare)))
		}

		if veoliaShare > 0 {
			stub.SetEvent("veoliaPay", []byte(strconv.Itoa(veoliaShare)))
		}

		err = stub.DelState(cardKey)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	return shim.Success(nil)
}

func (t *TravelChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	err := shim.Start(new(TravelChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}
