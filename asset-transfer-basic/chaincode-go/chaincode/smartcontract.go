package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
//Insert struct field in alphabetic order => to achieve determinism accross languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	ID				string 		`json:"ID"`
	CarBrand		string 		`json:"CarBrand"`
	CarModel		string 		`json:"CarModel"`
	CarColor		string 		`json:"CarColor"`
	OwnerId			string 		`json:"OwnerId"`
	ProductionYear	int 		`json:"ProductionYear"`
	Failures		[]Failure 	`json:"Failures,omitempty" metadata:"Failures,optional"`
}

type Owner struct {
	ID			string 		`json:"ID"`
	Name		string 		`json:"Name"`
	Surname 	string 		`json:"Surname"`
	Email		string 		`json:"Email"`
	Money		int 		`json:"Money"`
}

type Failure struct {
	Name 		string
	Price		int
}

/*type Car struct {

	
}*/

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	failures := []Failure{
		{Name:"engine",Price:100},
		{Name:"Tyre",Price:50},
	}

	failures2 := []Failure{}

	assets := []Asset{
		{ID: "asset7", CarBrand: "blue", CarModel: "proba", CarColor: "Tomoko", OwnerId: "owner1",ProductionYear:1000, Failures: failures},
		{ID: "asset8", CarBrand: "blue", CarModel: "proba", CarColor: "Tomoko", OwnerId: "owner1",ProductionYear:1000, Failures: failures2},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	owners := []Owner{
		{ID: "owner1", Name: "blue", Surname: "Tomoko",Email:"proba" ,Money: 300},
		{ID: "owner2", Name: "blue", Surname: "Tomoko",Email:"proba" ,Money: 300},
	}

	for _, owner := range owners {
		ownerJSON, err := json.Marshal(owner)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(owner.ID, ownerJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, carBrand string, carModel string, carColor string, ownerId string, productionYear int, failures []Failure) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:             		id,
		CarBrand:          		carBrand,
		CarModel:           	carModel,
		CarColor:          		carColor,
		OwnerId: 				ownerId,
		ProductionYear:			productionYear,
		Failures:				failures,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, carBrand string, carModel string, carColor string, ownerId string, productionYear int, failures []Failure) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
	asset := Asset{
		ID:             		id,
		CarBrand:          		carBrand,
		CarModel:           	carModel,
		CarColor:          		carColor,
		OwnerId: 				ownerId,
		ProductionYear:			productionYear,
		Failures:				failures,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferAsset updates the owner field of asset with given id in world state, and returns the old owner.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) (string, error) {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return "", err
	}

	oldOwner := asset.OwnerId
	if (len(asset.Failures) == 0){
		asset.OwnerId = newOwner
		
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return "", err
		}
	
		err = ctx.GetStub().PutState(id, assetJSON)
		if err != nil {
			return "", err
		}
	}


	return oldOwner, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func (s *SmartContract) GetAllOwners(ctx contractapi.TransactionContextInterface) ([]*Owner, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Owner
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Owner
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			fmt.Println("usao")
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func (s *SmartContract) ChangeColor(ctx contractapi.TransactionContextInterface, id string, newColor string) (string, error) {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return "", err
	}

	oldColor := asset.CarColor
	asset.CarColor = newColor

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return "", err
	}

	return oldColor, nil
}
