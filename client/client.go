package client

// CS 161 Project 2

// Only the following imports are allowed! ANY additional imports
// may break the autograder!
// - bytes
// - encoding/hex
// - encoding/json
// - errors
// - fmt
// - github.com/cs161-staff/project2-userlib
// - github.com/google/uuid
// - strconv
// - strings

import (
	"encoding/json"

	userlib "github.com/cs161-staff/project2-userlib"
	"github.com/google/uuid"

	// hex.EncodeToString(...) is useful for converting []byte to string

	// Useful for string manipulation

	// Useful for formatting strings (e.g. `fmt.Sprintf`).
	"fmt"

	// Useful for creating new error messages to return using errors.New("...")
	"errors"

	// Optional.
	_ "strconv"
)

// This serves two purposes: it shows you a few useful primitives,
// and suppresses warnings for imports not being used. It can be
// safely deleted!
func someUsefulThings() {

	// Creates a random UUID.
	randomUUID := uuid.New()

	// Prints the UUID as a string. %v prints the value in a default format.
	// See https://pkg.go.dev/fmt#hdr-Printing for all Golang format string flags.
	userlib.DebugMsg("Random UUID: %v", randomUUID.String())

	// Creates a UUID deterministically, from a sequence of bytes.
	hash := userlib.Hash([]byte("user-structs/alice"))
	deterministicUUID, err := uuid.FromBytes(hash[:16])
	if err != nil {
		// Normally, we would `return err` here. But, since this function doesn't return anything,
		// we can just panic to terminate execution. ALWAYS, ALWAYS, ALWAYS check for errors! Your
		// code should have hundreds of "if err != nil { return err }" statements by the end of this
		// project. You probably want to avoid using panic statements in your own code.
		panic(errors.New("An error occurred while generating a UUID: " + err.Error()))
	}
	userlib.DebugMsg("Deterministic UUID: %v", deterministicUUID.String())

	// Declares a Course struct type, creates an instance of it, and marshals it into JSON.
	type Course struct {
		name      string
		professor []byte
	}

	course := Course{"CS 161", []byte("Nicholas Weaver")}
	courseBytes, err := json.Marshal(course)
	if err != nil {
		panic(err)
	}

	userlib.DebugMsg("Struct: %v", course)
	userlib.DebugMsg("JSON Data: %v", courseBytes)

	// Generate a random private/public keypair.
	// The "_" indicates that we don't check for the error case here.
	var pk userlib.PKEEncKey
	var sk userlib.PKEDecKey
	pk, sk, _ = userlib.PKEKeyGen()
	userlib.DebugMsg("PKE Key Pair: (%v, %v)", pk, sk)

	// Here's an example of how to use HBKDF to generate a new key from an input key.
	// Tip: generate a new key everywhere you possibly can! It's easier to generate new keys on the fly
	// instead of trying to think about all of the ways a key reuse attack could be performed. It's also easier to
	// store one key and derive multiple keys from that one key, rather than
	originalKey := userlib.RandomBytes(16)
	derivedKey, err := userlib.HashKDF(originalKey, []byte("mac-key"))
	if err != nil {
		panic(err)
	}
	userlib.DebugMsg("Original Key: %v", originalKey)
	userlib.DebugMsg("Derived Key: %v", derivedKey)

	// A couple of tips on converting between string and []byte:
	// To convert from string to []byte, use []byte("some-string-here")
	// To convert from []byte to string for debugging, use fmt.Sprintf("hello world: %s", some_byte_arr).
	// To convert from []byte to string for use in a hashmap, use hex.EncodeToString(some_byte_arr).
	// When frequently converting between []byte and string, just marshal and unmarshal the data.
	//
	// Read more: https://go.dev/blog/strings

	// Here's an example of string interpolation!
	_ = fmt.Sprintf("%s_%d", "file", 1)
}

// This is the type definition for the User struct.
// A Go struct is like a Python or Java class - it can have attributes
// (e.g. like the Username attribute) and methods (e.g. like the StoreFile method below).

type Invitation struct {
	ShareNodeUUID          uuid.UUID
	ShareNodeDecryptionKey []byte
} //Encrypted using RSA and Authenticated using Digital Signatures

type ShareNode struct {
	FileMetadataUUID  uuid.UUID
	FileDecryptionKey []byte
	Dead bool
} //MACed with HashKDF(ShareNodeDecryptionKey), Encrypted using ShareNodeDecryptionKey

type FileOwnerInfo struct {
	FileMetadataUUID  uuid.UUID
	FileDecryptionKey []byte                //Generate hmacKey Deterministically
	SharedMap         map[string]Invitation //key:name, value:Invitation this is unencrypted directly in FileOwnerInfo
} //MACed with File Owner's hmacKey, Encrypted using File Owner's sourceKey

type FileNode struct {
	Content      []byte
	PrevNodeUUID uuid.UUID
} //Encrypted and Maced using the FileDecryptionKey

type FileMetadata struct {
	LatestNodeUUID uuid.UUID
} //MACed with HashKDF(FileDecryptionKey, "HMAC"), Encrypted using HashKDF(FileDecryptionKey, "encryption")

type User struct {
	Username         string
	sourceKey        []byte //Argon2Key(hash(username) + hash(password), username, 16)
	PublicRSA        userlib.PKEEncKey
	PrivateRSA       userlib.PKEDecKey //= PKEKeyGen(); // keystore.add(username + “RSA”, PublicRSA);
	PublicSign       userlib.DSVerifyKey
	PrivateSign      userlib.DSSignKey   //DSKeyGen();	// keystore.add(username + “Verify”, PublicSign);
	EncPrivateRSA    userlib.DSVerifyKey //SymEnc(sourceKey, iv1, privateRSA);	//iv1 = RandomBytes(16);
	EncPrivateSign   userlib.DSSignKey   //SynEnc(sourceKey, iv2, privateSign);  //iv2 = RandomBytes(16);
	FileMapUUID      uuid.UUID           //UUID for make(map[string][]byte)	//HashMap<key:filename, value:UUID for FileOwnerInfo>
	SharedWithMeUUID uuid.UUID           //UUID for make(map[string][]byte)   //HashMap<key:filename, value:Invitation>
}

// You can add other attributes here if you want! But note that in order for attributes to
// be included when this struct is serialized to/from JSON, they must be capitalized.
// On the flipside, if you have an attribute that you want to be able to access from
// this struct's methods, but you DON'T want that value to be included in the serialized value
// of this struct that's stored in datastore, then you can use a "private" variable (e.g. one that
// begins with a lowercase letter).

// NOTE: The following methods have toy (insecure!) implementations.

func RetrieveUserdata(username string) (macEncSerUserdata []byte, err error) {
	if username == "" {
		return nil, errors.New("username too short")
	}
	usernameUUID, err := uuid.FromBytes(userlib.Hash([]byte(username))[:16])
	if err != nil {
		return nil, errors.New("usernameUUID Generation failed")
	}
	macEncSerUserdata, InUse := userlib.DatastoreGet(usernameUUID)
	if !InUse {
		return nil, nil
	}

	return macEncSerUserdata, nil
}

func GenerateSourceKey(username string, password string) (sourceKey []byte) {
	SALT := "QT0TXdIVGQrPJzQR"
	sourceKeySalt := username + SALT
	sourceKey = userlib.Argon2Key([]byte(password), []byte(sourceKeySalt), 16)
	return sourceKey
}

func GenerateHMACKey(sourceKey []byte) (hmac []byte, err error) {
	hmacKey, err := userlib.HashKDF(sourceKey, []byte("Hmac_Key"))
	if err != nil {
		return nil, err
	}
	return hmacKey[:16], err
}

func SerEncryptMacStore(uuid uuid.UUID, data any, encKey []byte) error {

	hmacKey, err := GenerateHMACKey(encKey)
	if err != nil {
		return errors.New("hmac key generationn failed")
	}

	serData, err := json.Marshal(data)
	if err != nil {
		return errors.New("serialize failed")
	}

	iv := userlib.RandomBytes(16)
	encSerdata := userlib.SymEnc(encKey, iv, serData)

	HMACtag, err := userlib.HMACEval(hmacKey, encSerdata)
	if err != nil {
		return errors.New("couldn't MAC the Encrypted, Serialized data")
	}

	macEncSerData := append(encSerdata, HMACtag...)

	userlib.DatastoreSet(uuid, macEncSerData)

	return nil
}

func CheckMacDecryptDeserialize(macEncSerData []byte, pointer any, encKey []byte) error {

	hmacKey, err := GenerateHMACKey(encKey)
	if err != nil {
		return errors.New("hmac key generationn failed")
	}

	//Split data to get Mac Tag
	if len(macEncSerData) - 64 <= 0 {
		return errors.New("data tampered")
	}
	encSerData := macEncSerData[:len(macEncSerData)-64]
	validationTag := macEncSerData[len(macEncSerData)-64:]

	// Check HMAC tag for integrity
	hmacTag, err := userlib.HMACEval(hmacKey, encSerData)
	if err != nil {
		return errors.New("hmac tag generation failed")
	}

	//Checks if the MAC is equal
	if !userlib.HMACEqual(hmacTag, validationTag) {
		return errors.New("data tampered with")
	}

	//Decrypt the Encrypted, Serialized Data so now it is just serialized
	serData := userlib.SymDec(encKey, encSerData)

	//Deserialized the data
	err = json.Unmarshal(serData, pointer)
	if err != nil {
		return errors.New("deserialization failed")
	}
	return nil
}

func InitUser(username string, password string) (userdataptr *User, err error) {
	macEncSerUserdata, err := RetrieveUserdata(username)
	if err != nil {
		return nil, err
	} else if macEncSerUserdata != nil {
		return nil, errors.New("user already exists")
	}
	// Setup new struct
	var userdata User
	userdataptr = &userdata

	userdata.Username = username

	//Generating sourceKey and hmacKey
	userdata.sourceKey = GenerateSourceKey(username, password)

	//Generating RSA Keys
	userdata.PublicRSA, userdata.PrivateRSA, err = userlib.PKEKeyGen()
	if err != nil {
		return nil, errors.New("failed to generate RSA Public and Private Keys")
	}
	err = userlib.KeystoreSet("rsa Public Key for: "+username, userdata.PublicRSA)
	if err != nil {
		return nil, errors.New("failed to put RSA public key into Keystore")
	}

	//Generating Signature Keys
	userdata.PrivateSign, userdata.PublicSign, err = userlib.DSKeyGen()
	if err != nil {
		return nil, errors.New("failed to generate Signature Public and Private Keys")
	}
	err = userlib.KeystoreSet("signature for: "+username, userdata.PublicSign)
	if err != nil {
		return nil, errors.New("failed to store public key")
	}

	//Initialize empty HashMap for FileMap, then encrypted, macing and storing in DataStore
	userdata.FileMapUUID = uuid.New()
	err = SerEncryptMacStore(userdata.FileMapUUID, make(map[string]uuid.UUID), userdata.sourceKey)
	if err != nil {
		return nil, err
	}

	//Initialize empty HashMap for SharedWithMe, then encrypted, macing and storing in DataStore
	userdata.SharedWithMeUUID = uuid.New()
	err = SerEncryptMacStore(userdata.SharedWithMeUUID, make(map[string]Invitation), userdata.sourceKey)
	if err != nil {
		return nil, err
	}

	userUUID, err := uuid.FromBytes(userlib.Hash([]byte(username))[:16])
	if err != nil {
		return nil, errors.New("user uuid generation failed")
	}

	err = SerEncryptMacStore(userUUID, userdata, userdata.sourceKey)
	if err != nil {
		return nil, err
	}

	return userdataptr, nil
}

func GetUser(username string, password string) (userdataptr *User, err error) {
	macEncSerUserdata, err := RetrieveUserdata(username)
	if err != nil {
		return nil, err
	} else if macEncSerUserdata == nil {
		return nil, errors.New("user doesn't exist")
	}

	sourceKey := GenerateSourceKey(username, password)
	if err != nil {
		return nil, errors.New("generating source key failed")
	}

	var userdata User

	err = CheckMacDecryptDeserialize(macEncSerUserdata, &userdata, sourceKey)
	if err != nil {
		return nil, err
	}

	userdata.sourceKey = sourceKey

	return &userdata, nil
}

func (userdata *User) StoreFile(filename string, content []byte) (err error) {
	// issue: if file is in shared with me and alice attempts to store file, may have multiple instances of same file name
	fileMetadataUUID, fileEncryptionKey, owner, err := RetrieveFileMetadataWithKey(userdata, filename)
	// It is a file shared with user
	if err == nil && !owner {
		var fileMetadata FileMetadata
		macEncSerFileMetadata, exists := userlib.DatastoreGet(userdata.FileMapUUID)
		if !exists {
			return errors.New("file list retrieval failed")
		}
		err = CheckMacDecryptDeserialize(macEncSerFileMetadata, &fileMetadata, userdata.sourceKey)
		if err != nil {
			return err
		}

		//Create new FileNode
		var newFileNode FileNode
		newFileNode.Content = content
		newFileNode.PrevNodeUUID = uuid.Nil
		newFileNodeUUID := uuid.New()

		//Storing FileNode in DataStore
		err = SerEncryptMacStore(newFileNodeUUID, newFileNode, fileEncryptionKey)
		if err != nil {
			return errors.New("datastore new file failed")
		}
		fileMetadata.LatestNodeUUID = newFileNodeUUID
		//Storing FileMetaData in DataStore
		err = SerEncryptMacStore(fileMetadataUUID, fileMetadata, fileEncryptionKey)
		if err != nil {
			return errors.New("datastore new file metadata failed")
		}
		return nil
	}
	//Checking if the FileMap for the user exists
	macEncSerFileMap, exists := userlib.DatastoreGet(userdata.FileMapUUID)
	if !exists {
		return errors.New("file list retrieval failed")
	}

	//Getting FileMap for the user by checking the Mac, decrypting it and deserializing it
	var fileMap map[string]uuid.UUID
	err = CheckMacDecryptDeserialize(macEncSerFileMap, &fileMap, userdata.sourceKey)
	if err != nil {
		return err
	}

	//Set up keys
	fileEncryptionKey = userlib.RandomBytes(16)

	//Create new FileNode
	var newFileNode FileNode
	newFileNode.Content = content
	newFileNode.PrevNodeUUID = uuid.Nil
	newFileNodeUUID := uuid.New()

	//Storing FileNode in DataStore
	err = SerEncryptMacStore(newFileNodeUUID, newFileNode, fileEncryptionKey)
	if err != nil {
		return errors.New("datastore new file failed")
	}

	//Create new file metadata
	var newFileMetadata FileMetadata
	newFileMetadata.LatestNodeUUID = newFileNodeUUID
	newFileMetadataUUID := uuid.New()
	if err != nil {
		return errors.New("meta data uuid creation failed")
	}

	//Storing FileMetaData in DataStore
	err = SerEncryptMacStore(newFileMetadataUUID, newFileMetadata, fileEncryptionKey)
	if err != nil {
		return errors.New("datastore new file metadata failed")
	}

	//Create new fileowner info
	var newFileOwnerInfo FileOwnerInfo
	newFileOwnerInfo.FileMetadataUUID = newFileMetadataUUID
	newFileOwnerInfo.FileDecryptionKey = fileEncryptionKey
	newFileOwnerInfo.SharedMap = make(map[string]Invitation)
	newFileOwnerInfoUUID := uuid.New()

	//Storing FileOwnerInfo in DataStore
	err = SerEncryptMacStore(newFileOwnerInfoUUID, newFileOwnerInfo, userdata.sourceKey)
	if err != nil {
		return errors.New("datastore new file owner info failed")
	}

	//Add filemetadata to file map and store it back
	fileMap[filename] = newFileOwnerInfoUUID
	err = SerEncryptMacStore(userdata.FileMapUUID, fileMap, userdata.sourceKey)
	if err != nil {
		return errors.New("datastore new file metadata failed")
	}

	return nil
}

func RetrieveFileMetadataWithKey(userdata *User, filename string) (fileMetadataUUID uuid.UUID, fileDecryptionKey []byte, owner bool, err error) {
	//First getting the user's FileMap in case the user is the owner of this file
	macEncSerFileMap, exists := userlib.DatastoreGet(userdata.FileMapUUID)
	if !exists {
		return uuid.Nil, nil, false, errors.New("file list retrieval failed")
	}

	//Creating the HashMap for Filemap by checking mac, decrypting it and deserializing it
	var fileMap map[string]uuid.UUID
	err = CheckMacDecryptDeserialize(macEncSerFileMap, &fileMap, userdata.sourceKey)
	if err != nil {
		return uuid.Nil, nil, false, err
	}

	//Checks if this user is the owner of the file and the fileMetadata is in the FileMap HashMap
	fileOwnerInfoUUID, exists := fileMap[filename]

	if exists {
		macEncSerFileOwnerInfo, exists := userlib.DatastoreGet(fileOwnerInfoUUID)
		if exists {
			//We get the HMACED, encrypted, serialized FileOwnerInfoStruct
			var fileOwnerInfo FileOwnerInfo
			err := CheckMacDecryptDeserialize(macEncSerFileOwnerInfo, &fileOwnerInfo, userdata.sourceKey)
			if err != nil {
				return uuid.Nil, nil, false, err
			}
			//Setting the fileMetadataUUID and fileDecryptionKey from the FileOwnerInfoStruct
			fileMetadataUUID = fileOwnerInfo.FileMetadataUUID
			fileDecryptionKey = fileOwnerInfo.FileDecryptionKey
			return fileMetadataUUID, fileDecryptionKey, true, nil
		}
	}
	//The user isn't the owner of the file, so we must check the SharedWithMe
	macEncSerSharedWithMe, exists := userlib.DatastoreGet(userdata.SharedWithMeUUID)
	if !exists {
		return uuid.Nil, nil, false, errors.New("shared with me retrieval failed")
	}

	//Getting the Maced, Encrypted, Serialized SharedWithMe HashMap from DataStore
	var sharedWithMe map[string]Invitation
	err = CheckMacDecryptDeserialize(macEncSerSharedWithMe, &sharedWithMe, userdata.sourceKey)
	if err != nil {
		return uuid.Nil, nil, false, err
	}

	//Getting the corresponding invitation for a given filename
	invitation, exists := sharedWithMe[filename]
	//If the invitation exists, this means we have an invitation for the given filename
	if exists {
		var shareNode ShareNode
		//Getting the ShareNode for the given filename from DataStore
		macEncSerShareNode, exists := userlib.DatastoreGet(invitation.ShareNodeUUID)
		if exists {
			//Retrieving the clean ShareNode struct after checking the mac, decrypting and deserializing using the ShareNode hmac/decryption key
			err = CheckMacDecryptDeserialize(macEncSerShareNode, &shareNode, invitation.ShareNodeDecryptionKey)
			if err != nil {
				return uuid.Nil, nil, false, err
			}
			if shareNode.Dead {
				delete(sharedWithMe, filename)
				// Persist sharedWithMe
				err = SerEncryptMacStore(userdata.SharedWithMeUUID, sharedWithMe, userdata.sourceKey)
				if err != nil {
					return uuid.Nil, nil, false, errors.New("failed to persist sharedwithme")
				}
				fmt.Printf("check")
			} else {
				return shareNode.FileMetadataUUID, shareNode.FileDecryptionKey, false, nil
			}
		}
	}

	return uuid.Nil, nil, false, errors.New("file doesn't exist")

}

func (userdata *User) AppendToFile(filename string, content []byte) error {
	//Retrieving the fileMetadataUUID and the fileEncryptionKey for both the FileMetadata and the FileNodeleEncryptionKey from either the
	fileMetadataUUID, fileEncryptionKey, _, err := RetrieveFileMetadataWithKey(userdata, filename)
	if err != nil {
		return err
	}

	//Get the hmaced, encrypted, serialized FileMetadata using the UUID that we generated from the helper function
	macEncSerFileMetadata, exists := userlib.DatastoreGet(fileMetadataUUID)
	if !exists {
		return errors.New("file metadata retrieval failed")
	}
	var fileMetadata FileMetadata
	err = CheckMacDecryptDeserialize(macEncSerFileMetadata, &fileMetadata, fileEncryptionKey)
	if err != nil {
		return err
	}

	//Create new FileNode for data to append
	var newFileNode FileNode
	newFileNode.Content = content
	newFileNode.PrevNodeUUID = fileMetadata.LatestNodeUUID
	newFileNodeUUID := uuid.New()

	//Storing new FileNode in DataStore
	err = SerEncryptMacStore(newFileNodeUUID, newFileNode, fileEncryptionKey)
	if err != nil {
		return errors.New("datastore new file failed")
	}

	// Update the FileMetadata with newest node, and store it back into Datastore by serializing, encrypting and macing.
	fileMetadata.LatestNodeUUID = newFileNodeUUID
	err = SerEncryptMacStore(fileMetadataUUID, fileMetadata, fileEncryptionKey)
	if err != nil {
		return errors.New("datastore new file failed")
	}

	return nil
}

func (userdata *User) LoadFile(filename string) (content []byte, err error) {
	//Retrieving the fileMetadataUUID and the fileEncryptionKey for both the FileMetadata and the FileNodeEncryptionKey from either the
	fileMetadataUUID, fileEncryptionKey, _, err := RetrieveFileMetadataWithKey(userdata, filename)
	if err != nil {
		return nil, err
	}

	//Get the hmaced, encrypted, serialized FileMetadata using the UUID that we generated from the helper function
	macEncSerFileMetadata, exists := userlib.DatastoreGet(fileMetadataUUID)
	if !exists {
		return nil, errors.New("file metadata retrieval failed")
	}
	var fileMetadata FileMetadata
	//Put the clean FileMetadata struct into fileMetadata
	err = CheckMacDecryptDeserialize(macEncSerFileMetadata, &fileMetadata, fileEncryptionKey)
	if err != nil {
		return nil, err
	}

	//Get the most updated FileNode for the given filename
	currentNodeUUID := fileMetadata.LatestNodeUUID
	var fileNode FileNode
	//If the currentNodeUUID isn't null, get the MACed, Encrypted, Serialized FileNode from DataStore
	for currentNodeUUID != uuid.Nil {
		macEncSerFileNode, exists := userlib.DatastoreGet(currentNodeUUID)
		if !exists {
			return nil, errors.New("file node retrieval failed")
		}
		err = CheckMacDecryptDeserialize(macEncSerFileNode, &fileNode, fileEncryptionKey)
		if err != nil {
			return nil, err
		}

		content = append(fileNode.Content, content...)
		currentNodeUUID = fileNode.PrevNodeUUID
	}
	return content, nil
}

func (userdata *User) CreateInvitation(filename string, recipientUsername string) (
	invitationPtr uuid.UUID, err error) {

	var invitation Invitation

	invitationUUID := uuid.New()

	// Check file list
	//First getting the user's FileMap in case the user is the owner of this file
	macEncSerFileMap, exists := userlib.DatastoreGet(userdata.FileMapUUID)
	if !exists {
		return uuid.Nil, errors.New("file list retrieval failed")
	}

	//Creating the HashMap for Filemap by checking mac, decrypting it and deserializing it
	var fileMap map[string]uuid.UUID
	err = CheckMacDecryptDeserialize(macEncSerFileMap, &fileMap, userdata.sourceKey)
	if err != nil {
		return uuid.Nil, err
	}

	//Checks if this user is the owner of the file and the fileMetadata is in the FileMap HashMap
	fileOwnerInfoUUID, exists := fileMap[filename]

	if exists {
		macEncSerFileOwnerInfo, exists := userlib.DatastoreGet(fileOwnerInfoUUID)
		if exists {
			//We get the HMACED, encrypted, serialized FileOwnerInfoStruct
			var fileOwnerInfo FileOwnerInfo
			err := CheckMacDecryptDeserialize(macEncSerFileOwnerInfo, &fileOwnerInfo, userdata.sourceKey)
			if err != nil {
				return uuid.Nil, err
			}

			// create new share node
			newShareNodeUUID := uuid.New()

			var newShareNode ShareNode

			newShareNode.FileMetadataUUID = fileOwnerInfo.FileMetadataUUID

			newShareNode.FileDecryptionKey = fileOwnerInfo.FileDecryptionKey

			newShareNode.Dead = false

			// store share node in data store
			shareNodeDecryptionKey := userlib.RandomBytes(16)

			err = SerEncryptMacStore(newShareNodeUUID, newShareNode, shareNodeDecryptionKey)
			if err != nil {
				return uuid.Nil, errors.New("datastore new file failed")
			}

			invitation.ShareNodeUUID = newShareNodeUUID
			invitation.ShareNodeDecryptionKey = shareNodeDecryptionKey

			//Add new invitation to SharedMap
			fileOwnerInfo.SharedMap[recipientUsername] = invitation

			//  Persist fileOwnerInfo
			//Storing FileOwnerInfo in DataStore
			err = SerEncryptMacStore(fileOwnerInfoUUID, fileOwnerInfo, userdata.sourceKey)
			if err != nil {
				return uuid.Nil, errors.New("datastore new file owner info failed")
			}
		}
	} else {
		//The user isn't the owner of the file, so we must check the SharedWithMe
		macEncSerSharedWithMe, exists := userlib.DatastoreGet(userdata.SharedWithMeUUID)
		if !exists {
			return uuid.Nil, errors.New("shared with me retrieval failed")
		}

		//Getting the Maced, Encrypted, Serialized SharedWithMe HashMap from DataStore
		var sharedWithMe map[string]Invitation
		err = CheckMacDecryptDeserialize(macEncSerSharedWithMe, &sharedWithMe, userdata.sourceKey)
		if err != nil {
			return uuid.Nil, err
		}

		//Getting the corresponding invitation for a given filename
		tmpInvitation, exists := sharedWithMe[filename]
		if !exists {
			return uuid.Nil, errors.New("file doesn't exist for user")
		}

		// check if share node is still valid
		var shareNode ShareNode
		//Getting the ShareNode for the given filename from DataStore
		macEncSerShareNode, exists := userlib.DatastoreGet(tmpInvitation.ShareNodeUUID)
		if !exists {
			return uuid.Nil, errors.New("share node doesn't exist")
		}
		//Retrieving the clean ShareNode struct after checking the mac, decrypting and deserializing using the ShareNode hmac/decryption key
		err = CheckMacDecryptDeserialize(macEncSerShareNode, &shareNode, tmpInvitation.ShareNodeDecryptionKey)
		if err != nil {
			return uuid.Nil, err
		}

		if shareNode.Dead {
			delete(sharedWithMe, filename)
			// Persist sharedWithMe
			err = SerEncryptMacStore(userdata.SharedWithMeUUID, sharedWithMe, userdata.sourceKey)
			if err != nil {
				return uuid.Nil, errors.New("failed to persist sharedwithme")
			}
			return uuid.Nil, errors.New("file got revoked")
		}

		invitation = tmpInvitation
		
	}

	// Securely store invitation
	serInvitation, err := json.Marshal(invitation)
	if err != nil {
		return uuid.Nil, errors.New("serialize failed")
	}

	//Encrypt the serialized invitation using the recipient's public Key
	recipientPublicKey, userExists := userlib.KeystoreGet("rsa Public Key for: " + recipientUsername)
	if !userExists {
		return uuid.Nil, errors.New("receipient public key doesn't exist")
	}
	rsaEncryptedInvitation, err := userlib.PKEEnc(recipientPublicKey, serInvitation)
	if err != nil {
		return uuid.Nil, errors.New("public key encryption failed")
	}
	//512 bytes messagersaencryptedandsigned -> 256 bytes of RSAmessage, 256 bytes of signedMessage
	//check if message^recipientpublickey^recipientprivatekey	and 	message^senderprivateSignature^senderPublicSign are equal
	//Create digital signature using the sender's signature
	signatureOnInvitation, err := userlib.DSSign(userdata.PrivateSign, serInvitation)
	if err != nil {
		return uuid.Nil, errors.New("signing invitation failed")
	}

	//combine both the RSA encrypted invitation and the signature
	invitationBytesEncryptedandSigned := append(rsaEncryptedInvitation, signatureOnInvitation...)

	//store the RSA encrypted invitation w/ signature into Datastore
	userlib.DatastoreSet(invitationUUID, invitationBytesEncryptedandSigned)
	return invitationUUID, nil
}

func (userdata *User) AcceptInvitation(senderUsername string, invitationPtr uuid.UUID, filename string) error {
	_, _, _, err := RetrieveFileMetadataWithKey(userdata, filename)
	if err == nil {
		return errors.New("file already exists")
	}

	//get the Invitation from Datastore
	invitationBytesEncryptedSigned, exists := userlib.DatastoreGet(invitationPtr)
	if !exists {
		return errors.New("the encrypted, signed Invitation doesn't exist")
	}

	//split the Invitation into the Signature on the serialized Invitation and RSA encrypted serialized Invitation
	rsaEncryptedInvitation := invitationBytesEncryptedSigned[:len(invitationBytesEncryptedSigned)-256]
	signatureOnInvitation := invitationBytesEncryptedSigned[len(invitationBytesEncryptedSigned)-256:]

	//Get the sender's SignaturePublicKey
	senderSignature, senderExists := userlib.KeystoreGet("signature for: " + senderUsername)
	if !senderExists {
		return errors.New("the Sender's signature doesn't exist")
	}

	//Decrypt the RSA encrypted, serialized invitation
	serializedInvitation, err := userlib.PKEDec(userdata.PrivateRSA, rsaEncryptedInvitation)
	if err != nil {
		return errors.New("invitation RSA decryption failed")
	}

	//Verify the signature on the serialized invitation	vitation
	err = userlib.DSVerify(senderSignature, serializedInvitation, signatureOnInvitation)
	if err != nil {
		return errors.New("signature check failed")
	}
	var invitation Invitation
	err = json.Unmarshal(serializedInvitation, &invitation)
	if err != nil {
		return errors.New("couldn't deserialize the invitation")
	}

	// if share node is dead fail
	var shareNode ShareNode
	//Getting the ShareNode for the given filename from DataStore
	macEncSerShareNode, exists := userlib.DatastoreGet(invitation.ShareNodeUUID)
	if !exists {
		return errors.New("share node doesn't exist")
	}
	//Retrieving the clean ShareNode struct after checking the mac, decrypting and deserializing using the ShareNode hmac/decryption key
	err = CheckMacDecryptDeserialize(macEncSerShareNode, &shareNode, invitation.ShareNodeDecryptionKey)
	if err != nil {
		return err
	}

	if shareNode.Dead {
		return errors.New("Invitation from user whose access got revoked already")
	}

	// Put invitation into sharedwithme
	macEncSerSharedWithMe, exists := userlib.DatastoreGet(userdata.SharedWithMeUUID)
	if !exists {
		return errors.New("shared with me retrieval failed")
	}

	var sharedWithMe map[string]Invitation
	err = CheckMacDecryptDeserialize(macEncSerSharedWithMe, &sharedWithMe, userdata.sourceKey)
	if err != nil {
		return err
	}

	sharedWithMe[filename] = invitation

	// persist sharedwithme
	err = SerEncryptMacStore(userdata.SharedWithMeUUID, sharedWithMe, userdata.sourceKey)
	if err != nil {
		return err
	}

	return nil
}

func GetFileOwnerInfo(userdata *User, pointer *FileOwnerInfo, filename string) (fileOwnerInfoUUID uuid.UUID, err error) {
	// check if file exists and that user owns file
	//First getting the user's FileMap in case the user is the owner of this file
	macEncSerFileMap, exists := userlib.DatastoreGet(userdata.FileMapUUID)
	if !exists {
		return uuid.Nil, errors.New("file list retrieval failed")
	}

	//Creating the HashMap for Filemap by checking mac, decrypting it and deserializing it
	var fileMap map[string]uuid.UUID
	err = CheckMacDecryptDeserialize(macEncSerFileMap, &fileMap, userdata.sourceKey)
	if err != nil {
		return uuid.Nil, err
	}

	//Checks if this user is the owner of the file and the fileMetadata is in their FileMap
	fileOwnerInfoUUID, exists = fileMap[filename]

	if !exists {
		return uuid.Nil, errors.New("user doesn't own file")
	}

	//Get the fileOwnerInfo struct from Datastore
	macEncSerFileOwnerInfo, exists := userlib.DatastoreGet(fileOwnerInfoUUID)
	if !exists {
		return uuid.Nil, errors.New("fileownerinfo couldn't be found in data store")
	}

	//We get the HMACED, encrypted, serialized FileOwnerInfoStruct and get the clean struct
	err = CheckMacDecryptDeserialize(macEncSerFileOwnerInfo, pointer, userdata.sourceKey)
	if err != nil {
		return uuid.Nil, err
	}


	return fileOwnerInfoUUID, nil

}

func DeleteFile(userdata *User, filename string) error {
	fileMetadataUUID, fileEncryptionKey, owner, err := RetrieveFileMetadataWithKey(userdata, filename)
	if !owner {
		return errors.New("only owner can delete file")
	} else if err != nil {
		return err
	}

	//Get the hmaced, encrypted, serialized FileMetadata using the UUID that we generated from the helper function
	macEncSerFileMetadata, exists := userlib.DatastoreGet(fileMetadataUUID)
	if !exists {
		return errors.New("file metadata retrieval failed")
	}
	var fileMetadata FileMetadata
	//Put the clean FileMetadata struct into fileMetadata
	err = CheckMacDecryptDeserialize(macEncSerFileMetadata, &fileMetadata, fileEncryptionKey)
	if err != nil {
		return err
	}

	//Get the most updated FileNode for the given filename
	currentNodeUUID := fileMetadata.LatestNodeUUID
	var fileNode FileNode
	var oldNodeUUID uuid.UUID
	//If the currentNodeUUID isn't null, get the MACed, Encrypted, Serialized FileNode from DataStore
	for currentNodeUUID != uuid.Nil {
		macEncSerFileNode, exists := userlib.DatastoreGet(currentNodeUUID)
		if !exists {
			return errors.New("file node retrieval failed")
		}
		err = CheckMacDecryptDeserialize(macEncSerFileNode, &fileNode, fileEncryptionKey)
		if err != nil {
			return err
		}
		oldNodeUUID = currentNodeUUID
		currentNodeUUID = fileNode.PrevNodeUUID
		userlib.DatastoreDelete(oldNodeUUID)
	}
	userlib.DatastoreDelete(oldNodeUUID)
	userlib.DatastoreDelete(fileMetadataUUID)
	return nil
}
func (userdata *User) RevokeAccess(filename string, recipientUsername string) error {
	// checks if the user is the owner and gets the FileOwnerInfoStruct if the user is.
	var oldFileOwnerInfo FileOwnerInfo
	oldFileOwnerInfoUUID, err := GetFileOwnerInfo(userdata, &oldFileOwnerInfo, filename)
	if err != nil {
		return err
	}
	fmt.Printf("\nOld: %+v\n", oldFileOwnerInfo)

	//Save the old SharedMap for restoration
	oldSharedMap := oldFileOwnerInfo.SharedMap

	_, exists := oldSharedMap[recipientUsername]

	if !exists {
		return errors.New("user to revoke not in share map")
	}

	// Rencrypt and move file Metadata to generate a new FileOwnerInfo w/ a new FileMetadataUUID and FileEncryptionKey
	// move filemetadata
	content, err := userdata.LoadFile(filename)
	if err != nil {
		return err
	}

	err = DeleteFile(userdata, filename)
	if err != nil {
		return err
	}

	//Restore content w/ same filename so that it creates a new FileOwnerInfo w/ a new FileMetadataUUID and FileEncryptionKey
	err = userdata.StoreFile(filename, content)
	if err != nil {
		return err
	}

	//Delete the oldFileOwnerInfoUUID and old FileMetadata since it will be regenerated when we store file
	userlib.DatastoreDelete(oldFileOwnerInfoUUID)
	userlib.DatastoreDelete(oldFileOwnerInfo.FileMetadataUUID)

	//Get the new FileOwnerInfo for the filename
	var newFileOwnerInfo FileOwnerInfo
	newFileOwnerInfoUUID, err := GetFileOwnerInfo(userdata, &newFileOwnerInfo, filename)
	if err != nil {
		return err
	}

	// get new info
	newFileMetadataUUID := newFileOwnerInfo.FileMetadataUUID
	newFileEncryptionKey := newFileOwnerInfo.FileDecryptionKey

	// distribute filemetadata to everyone in shared with me
	for name, invitation := range oldFileOwnerInfo.SharedMap {
		var shareNode ShareNode
		//Getting the ShareNode for the given filename from DataStore
		macEncSerShareNode, exists := userlib.DatastoreGet(invitation.ShareNodeUUID)
		if !exists {
			return errors.New("share node doesn't exist")
		}
		//Retrieving the clean ShareNode struct after checking the mac, decrypting and deserializing using the ShareNode hmac/decryption key
		err = CheckMacDecryptDeserialize(macEncSerShareNode, &shareNode, invitation.ShareNodeDecryptionKey)
		if err != nil {
			return err
		}

		if name == recipientUsername {
			shareNode.Dead = true
		} else {
			shareNode.FileMetadataUUID = newFileMetadataUUID
			shareNode.FileDecryptionKey = newFileEncryptionKey
		}

		// persist share Node
		err = SerEncryptMacStore(invitation.ShareNodeUUID, shareNode, invitation.ShareNodeDecryptionKey)
		if err != nil {
			return err
		}
	}

	delete(oldSharedMap, recipientUsername)

	newFileOwnerInfo.SharedMap = oldSharedMap

	// persist new file owner info
	err = SerEncryptMacStore(newFileOwnerInfoUUID, newFileOwnerInfo, userdata.sourceKey)
	if err != nil {
		return err
	}

	//user has to be the owner of the file, so the filename must be in the fileList
	//Then go to FileOwnerInfo for that file
	//generate a new FileNodeKey and change it inside fileOwnerInfo
	//Then, for all user's except the one you're revoking, go into the ShareMap to access the Invitation
	//using the Invitation's information, get the ShareNode from DataStore and decrypt it using the password in Invitation
	//Once you have access to every ShareNode except the revoked User's update the ShareNode.FileDecryptionKey to be the new key!
	return nil
}
