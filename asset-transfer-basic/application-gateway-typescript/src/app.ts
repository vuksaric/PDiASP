/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { connect, Contract, Identity, Signer, signers } from '@hyperledger/fabric-gateway';
import * as crypto from 'crypto';
import { promises as fs } from 'fs';
import * as path from 'path';
import { TextDecoder } from 'util';
import * as readline from 'node:readline';
const prompt = require('prompt-sync')();

const channelName = envOrDefault('CHANNEL_NAME', 'mychannel');
const chaincodeName = envOrDefault('CHAINCODE_NAME', 'basic');
const mspId = envOrDefault('MSP_ID', 'Org1MSP');

// Path to crypto materials.
const cryptoPath = envOrDefault('CRYPTO_PATH', path.resolve(__dirname, '..', '..', '..', 'test-network', 'organizations', 'peerOrganizations', 'org1.example.com'));

// Path to user private key directory.
const keyDirectoryPath = envOrDefault('KEY_DIRECTORY_PATH', path.resolve(cryptoPath, 'users', 'User1@org1.example.com', 'msp', 'keystore'));

// Path to user certificate.
const certPath = envOrDefault('CERT_PATH', path.resolve(cryptoPath, 'users', 'User1@org1.example.com', 'msp', 'signcerts', 'cert.pem'));

// Path to peer tls certificate.
const tlsCertPath = envOrDefault('TLS_CERT_PATH', path.resolve(cryptoPath, 'peers', 'peer0.org1.example.com', 'tls', 'ca.crt'));

// Gateway peer endpoint.
const peerEndpoint = envOrDefault('PEER_ENDPOINT', 'localhost:7051');

// Gateway peer SSL host name override.
const peerHostAlias = envOrDefault('PEER_HOST_ALIAS', 'peer0.org1.example.com');

const utf8Decoder = new TextDecoder();
const assetId = `asset${Date.now()}`;

async function main(): Promise<void> {

    await displayInputParameters();

    // The gRPC client connection should be shared by all Gateway connections to this endpoint.
    const client = await newGrpcConnection();

    const gateway = connect({
        client,
        identity: await newIdentity(),
        signer: await newSigner(),
        // Default timeouts for different gRPC calls
        evaluateOptions: () => {
            return { deadline: Date.now() + 5000 }; // 5 seconds
        },
        endorseOptions: () => {
            return { deadline: Date.now() + 15000 }; // 15 seconds
        },
        submitOptions: () => {
            return { deadline: Date.now() + 5000 }; // 5 seconds
        },
        commitStatusOptions: () => {
            return { deadline: Date.now() + 60000 }; // 1 minute
        },
    });

    try {
        // Get a network instance representing the channel where the smart contract is deployed.
        const network = gateway.getNetwork(channelName);

        // Get the smart contract from the network.
        const contract = network.getContract(chaincodeName);

        // Initialize a set of asset data on the ledger using the chaincode 'InitLedger' function.
        //await initLedger(contract);


        do{
            var input = prompt("\n 1.) GetAllAssets \n 2.) GetAllOwners \n 3.) TransferAsset \n 4.) ChangeColor \n 5.) CreateFailure \n 6.) RepairFailures \n 7.) FindColor \n 8.) FindOwner \n 9.) FindColorOwner");
            if(input == 1){
                await getAllAssets(contract);
                var input2 = prompt("");
            }
            else if(input == 2){
                await getAllOwners(contract);
                var input2 = prompt("");
            }
            else if(input == 3){
                const assetId = prompt("AssetId:");
                const newOwner = prompt("NewOwnerId:");
                const buyWithFailure = prompt("BuyWithFailure:");
                await transferAssetAsync(contract, assetId, newOwner, buyWithFailure);
                var input2 = prompt("");
            }
            else if(input == 4){
                const assetId = prompt("AssetId:");
                const color = prompt("Color:"); 
                await ChangeColorAsync(contract, assetId, color)
                var input2 = prompt("");
            }
            else if(input == 5){
                const assetId = prompt("AssetId:");
                const failureName = prompt("Failure name:");
                const price = prompt("Price:");
                await CreateFailureAsync(contract, assetId, failureName, price);
                var input2 = prompt("");
            }
            else if(input == 6){
                const assetId = prompt("AssetId:");
                await RepairFailuresAsync(contract, assetId);
                var input2 = prompt("");
            }
            else if(input == 7){
                const color = prompt("Color:"); 
                await FindColorAsync(contract, color)
                var input2 = prompt("");
            }
            else if(input == 8){
                const ownerId = prompt("OwnerId:");
                await FindOwnerAsync(contract, ownerId);
                var input2 = prompt("");
            }
            else if(input == 9){
                const ownerId = prompt("OwnerId:");
                const color = prompt("Color:"); 
                await FindColorOwnerAsync(contract, color, ownerId);
                var input2 = prompt("");
            }else{
                console.log("You choosed wrong value")
            }
        }
        while(input != 0)
        /*rl.question("Welcome", function(input){
            console.log(input);
            rl2.question("Choose option: /n 1.) Proba1 /n 2.) Proba2 /n 3.) Proba3",function(input2){
                    while((parseInt(input) != 4)){
                    console.log(input2);
                    }
            rl.close
            console.log("izasao");
            })
        })*/
        // Return all the current assets on the ledger.
        //await getAllAssets(contract);

        // Create a new asset on the ledger.
        //await createAsset(contract);

        // Update an existing asset asynchronously.
        //await transferAssetAsync(contract);

        // Get the asset details by assetID.
        //await readAssetByID(contract);

        // Update an asset which does not exist.
        //await updateNonExistentAsset(contract)
    } finally {
        gateway.close();
        client.close();
    }
}

main().catch(error => {
    console.error('******** FAILED to run the application:', error);
    process.exitCode = 1;
});

async function newGrpcConnection(): Promise<grpc.Client> {
    const tlsRootCert = await fs.readFile(tlsCertPath);
    const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);
    return new grpc.Client(peerEndpoint, tlsCredentials, {
        'grpc.ssl_target_name_override': peerHostAlias,
    });
}

async function newIdentity(): Promise<Identity> {
    const credentials = await fs.readFile(certPath);
    return { mspId, credentials };
}

async function newSigner(): Promise<Signer> {
    const files = await fs.readdir(keyDirectoryPath);
    const keyPath = path.resolve(keyDirectoryPath, files[0]);
    const privateKeyPem = await fs.readFile(keyPath);
    const privateKey = crypto.createPrivateKey(privateKeyPem);
    return signers.newPrivateKeySigner(privateKey);
}

/**
 * This type of transaction would typically only be run once by an application the first time it was started after its
 * initial deployment. A new version of the chaincode deployed later would likely not need to run an "init" function.
 */
async function initLedger(contract: Contract): Promise<void> {
    console.log('\n--> Submit Transaction: InitLedger, function creates the initial set of assets on the ledger');

    await contract.submitTransaction('InitLedger');

    console.log('*** Transaction committed successfully');
}

/**
 * Evaluate a transaction to query ledger state.
 */
async function getAllAssets(contract: Contract): Promise<void> {
    console.log('\n--> Evaluate Transaction: GetAllAssets, function returns all the current assets on the ledger');

    const resultBytes = await contract.evaluateTransaction('GetAllAssets');

    const resultJson = utf8Decoder.decode(resultBytes);
    const result = JSON.parse(resultJson);
    console.log('*** Result:', result);
}

async function FindColorAsync(contract: Contract, color: string): Promise<void> {
    console.log('\n--> Evaluate Transaction: getAllOwners, function returns all the current assets on the ledger');

    const resultBytes = await contract.evaluateTransaction('FindColor',color);

    const resultJson = utf8Decoder.decode(resultBytes);
    const result = JSON.parse(resultJson);
    console.log('*** Result:', result);
}

async function FindOwnerAsync(contract: Contract, ownerId: string): Promise<void> {
    console.log('\n--> Evaluate Transaction: getAllOwners, function returns all the current assets on the ledger');

    const resultBytes = await contract.evaluateTransaction('FindOwner',ownerId);

    const resultJson = utf8Decoder.decode(resultBytes);
    const result = JSON.parse(resultJson);
    console.log('*** Result:', result);
}

async function FindColorOwnerAsync(contract: Contract, color: string, ownerId: string): Promise<void> {
    console.log('\n--> Evaluate Transaction: getAllOwners, function returns all the current assets on the ledger');

    const resultBytes = await contract.evaluateTransaction('FindOwnerColor', color, ownerId);

    const resultJson = utf8Decoder.decode(resultBytes);
    const result = JSON.parse(resultJson);
    console.log('*** Result:', result);
}

async function getAllOwners(contract: Contract): Promise<void> {
    console.log('\n--> Evaluate Transaction: getAllOwners, function returns all the current assets on the ledger');

    const resultBytes = await contract.evaluateTransaction('getAllOwners');

    const resultJson = utf8Decoder.decode(resultBytes);
    const result = JSON.parse(resultJson);
    console.log('*** Result:', result);
}

/**
 * Submit a transaction synchronously, blocking until it has been committed to the ledger.
 */
async function createAsset(contract: Contract): Promise<void> {
    console.log('\n--> Submit Transaction: CreateAsset, creates new asset with ID, Color, Size, Owner and AppraisedValue arguments');

    const array: string[] = []
    await contract.submitTransaction(
        'CreateAsset',
        assetId,
        'BMW',
        'Black',
        'owner1',
        '2010',
        '1000',
        '0',
        '[]'
        
    );

    console.log('*** Transaction committed successfully');
}

/**
 * Submit transaction asynchronously, allowing the application to process the smart contract response (e.g. update a UI)
 * while waiting for the commit notification.
 */
async function transferAssetAsync(contract: Contract, assetId:string, newOwner: string, buyWithFailure: string ): Promise<void> {
    console.log('\n--> Async Submit Transaction: TransferAsset, updates existing asset owner');

    const commit = await contract.submitTransaction('TransferAsset', assetId,newOwner,buyWithFailure);
    //const oldOwner = utf8Decoder.decode(commit.getResult());

    //console.log(`*** Successfully submitted transaction to transfer ownership from ${oldOwner} to Saptha`);
    console.log('*** Waiting for transaction commit');

    /*const status = await commit.getStatus();
    if (!status.successful) {
        throw new Error(`Transaction ${status.transactionId} failed to commit with status code ${status.code}`);
    }*/

    console.log('*** Transaction committed successfully');
}

async function ChangeColorAsync(contract: Contract, assetId:string, color: string ): Promise<void> {
    console.log('\n--> Async Submit Transaction: TransferAsset, updates existing asset owner');

    const commit = await contract.submitTransaction('ChangeColor',assetId,color);
    //const oldOwner = utf8Decoder.decode(commit.getResult());

    //console.log(`*** Successfully submitted transaction to transfer ownership from ${oldOwner} to Saptha`);
    console.log('*** Waiting for transaction commit');

    /*const status = await commit.getStatus();
    if (!status.successful) {
        throw new Error(`Transaction ${status.transactionId} failed to commit with status code ${status.code}`);
    }*/

    console.log('*** Transaction committed successfully');
}

async function CreateFailureAsync(contract: Contract, assetId:string, failure: string, price: string ): Promise<void> {
    console.log('\n--> Async Submit Transaction: TransferAsset, updates existing asset owner');

    const commit = await contract.submitTransaction('CreateFailure',assetId,failure,price);
    //const oldOwner = utf8Decoder.decode(commit.getResult());

    //console.log(`*** Successfully submitted transaction to transfer ownership from ${oldOwner} to Saptha`);
    console.log('*** Waiting for transaction commit');

    /*const status = await commit.getStatus();
    if (!status.successful) {
        throw new Error(`Transaction ${status.transactionId} failed to commit with status code ${status.code}`);
    }*/

    console.log('*** Transaction committed successfully');
}

async function RepairFailuresAsync(contract: Contract, assetId:string): Promise<void> {
    console.log('\n--> Async Submit Transaction: TransferAsset, updates existing asset owner');

    const commit = await contract.submitTransaction('RepairFailures',assetId);
    //const oldOwner = utf8Decoder.decode(commit.getResult());

    //console.log(`*** Successfully submitted transaction to transfer ownership from ${oldOwner} to Saptha`);
    console.log('*** Waiting for transaction commit');

    /*const status = await commit.getStatus();
    if (!status.successful) {
        throw new Error(`Transaction ${status.transactionId} failed to commit with status code ${status.code}`);
    }*/

    console.log('*** Transaction committed successfully');
}

async function readAssetByID(contract: Contract): Promise<void> {
    console.log('\n--> Evaluate Transaction: ReadAsset, function returns asset attributes');

    const resultBytes = await contract.evaluateTransaction('ReadAsset', assetId);

    const resultJson = utf8Decoder.decode(resultBytes);
    const result = JSON.parse(resultJson);
    console.log('*** Result:', result);
}

/**
 * submitTransaction() will throw an error containing details of any error responses from the smart contract.
 */
async function updateNonExistentAsset(contract: Contract): Promise<void>{
    console.log('\n--> Submit Transaction: UpdateAsset asset70, asset70 does not exist and should return an error');

    try {
        await contract.submitTransaction(
            'UpdateAsset',
            'asset7',
            'BMW',
            'Black',
            'owner1',
            '2010',
            '1000',
            '0',
            '[]'
        );
        console.log('******** FAILED to return an error');
    } catch (error) {
        console.log('*** Successfully caught the error: \n', error);
    }
}

/**
 * envOrDefault() will return the value of an environment variable, or a default value if the variable is undefined.
 */
function envOrDefault(key: string, defaultValue: string): string {
    return process.env[key] || defaultValue;
}

/**
 * displayInputParameters() will print the global scope parameters used by the main driver routine.
 */
async function displayInputParameters(): Promise<void> {
    console.log(`channelName:       ${channelName}`);
    console.log(`chaincodeName:     ${chaincodeName}`);
    console.log(`mspId:             ${mspId}`);
    console.log(`cryptoPath:        ${cryptoPath}`);
    console.log(`keyDirectoryPath:  ${keyDirectoryPath}`);
    console.log(`certPath:          ${certPath}`);
    console.log(`tlsCertPath:       ${tlsCertPath}`);
    console.log(`peerEndpoint:      ${peerEndpoint}`);
    console.log(`peerHostAlias:     ${peerHostAlias}`);
}