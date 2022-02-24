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
	Price			int			`json:"Price"`
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

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	failures := []Failure{
		{Name:"engine",Price:100},
		{Name:"Tyre",Price:50},
	}

	failures2 := []Failure{}

	assets := []Asset{
		{ID: "asset7", CarBrand: "blue", CarModel: "proba", CarColor: "Tomoko", OwnerId: "owner1",ProductionYear:1000, Price:100 ,Failures: failures},
		{ID: "asset8", CarBrand: "blue", CarModel: "proba", CarColor: "Tomoko", OwnerId: "owner1",ProductionYear:1000, Price:100 ,Failures: failures2},
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
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, carBrand string, carModel string, carColor string, ownerId string, productionYear int, price int, failures []Failure) error {
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
		Price:					price,
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

func (s *SmartContract) ReadOwner(ctx contractapi.TransactionContextInterface, id string) (*Owner, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Owner
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, carBrand string, carModel string, carColor string, ownerId string, productionYear int, price int ,failures []Failure) error {
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
		Price:					price,
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

	ownerNew, err := s.ReadOwner(ctx, newOwner)
	if err != nil {
		return "", err
	}
	
	oldOwner := asset.OwnerId

	ownerOld, err := s.ReadOwner(ctx, oldOwner)
	if err != nil {
		return "", err
	}
	
	asset.OwnerId = newOwner
	failuresPrice := 0	
	if len(asset.Failures) != 0 {
		for i:= 0;i<len(asset.Failures);i++{
			failuresPrice = failuresPrice + asset.Failures[i].Price
		}
	}

	price := asset.Price - failuresPrice
	ownerNew.Money = ownerNew.Money - price
	ownerOld.Money = ownerOld.Money + price 

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return "", err
	}

	ownerNewJson, err := json.Marshal(ownerNew)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(newOwner, ownerNewJson)
	if err != nil {
		return "", err
	}

	ownerOldJson, err := json.Marshal(ownerOld)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(oldOwner, ownerOldJson)
	if err != nil {
		return "", err
	}

	//s.UpdateAsset(ctx, asset.ID, asset.CarBrand, asset.CarModel, asset.CarColor, asset.OwnerId, asset.ProductionYear, asset.Price, asset.Failures)

	return asset.OwnerId, nil
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

	//oldColor := asset.CarColor
	asset.CarColor = newColor

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return "", err
	}

	//s.UpdateAsset(ctx, asset.ID, asset.CarBrand, asset.CarModel, asset.CarColor, asset.OwnerId, asset.ProductionYear, asset.Price, asset.Failures)

	return asset.CarColor, nil
}

func (s *SmartContract) CreateFailure(ctx contractapi.TransactionContextInterface, id string, failureName string, price int) (string, error) {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return "", err
	}

	failure := Failure{Name:failureName,Price:price}

	asset.Failures = append(asset.Failures, failure)

	failuresPrice := 0	
	if len(asset.Failures) != 0 {
		for i:= 0;i<len(asset.Failures);i++{
			failuresPrice = failuresPrice + asset.Failures[i].Price
		}
	}
	
	if failuresPrice < asset.Price{
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return "", err
		}
	
		err = ctx.GetStub().PutState(id, assetJSON)
		if err != nil {
			return "", err
		}
	}else{
		s.DeleteAsset(ctx,asset.ID)
	}

	return failure.Name, nil

}

func (s *SmartContract) RepairFailures(ctx contractapi.TransactionContextInterface, id string) (string, error) {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return "", err
	}

	asset.Failures = nil

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *SmartContract) FindColor(ctx contractapi.TransactionContextInterface, color string) ([]*Asset, error) {
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
		if color == asset.CarColor {
			assets = append(assets, &asset)
		}
	}

	return assets, nil
}

func (s *SmartContract) FindOwnerctx contractapi.TransactionContextInterface, owner string) ([]*Asset, error) {
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
		if owner == asset.OwnerId {
			assets = append(assets, &asset)
		}
	}

	return assets, nil
}

func (s *SmartContract) FindOwnerColor(ctx contractapi.TransactionContextInterface, color string, owner string) ([]*Asset, error) {
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
		if color == asset.CarColor && owner == asset.OwnerId {
			assets = append(assets, &asset)
		}
	}

	return assets, nil
}

