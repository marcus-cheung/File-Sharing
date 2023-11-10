package client_test

// You MUST NOT change these default imports.  ANY additional imports may
// break the autograder and everyone will be sad.

import (
	// Some imports use an underscore to prevent the compiler from complaining
	// about unused imports.
	_ "encoding/hex"
	_ "errors"
	_ "strconv"
	_ "strings"
	"testing"

	// A "dot" import is used here so that the functions in the ginko and gomega
	// modules can be used without an identifier. For example, Describe() and
	// Expect() instead of ginko.Describe() and gomega.Expect().
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	userlib "github.com/cs161-staff/project2-userlib"

	"github.com/cs161-staff/project2-starter-code/client"
)

func TestSetupAndExecution(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Tests")
}

// ================================================
// Global Variables (feel free to add more!)
// ================================================
const defaultPassword = "password"
const emptyString = ""
const contentOne = "Bitcoin is Nick's favorite "
const contentTwo = "digital "
const contentThree = "cryptocurrency!"

// ================================================
// Describe(...) blocks help you organize your tests
// into functional categories. They can be nested into
// a tree-like structure.
// ================================================

var _ = Describe("Client Tests", func() {

	// A few user declarations that may be used for testing. Remember to initialize these before you
	// attempt to use them!
	var alice *client.User
	var bob *client.User
	var charles *client.User
	var doris *client.User
	var eve *client.User
	// var frank *client.User
	// var grace *client.User
	// var horace *client.User
	// var ira *client.User

	// These declarations may be useful for multi-session testing.
	var alicePhone *client.User
	var aliceLaptop *client.User
	var aliceDesktop *client.User

	var err error

	// A bunch of filenames that may be useful.
	aliceFile := "aliceFile.txt"
	bobFile := "bobFile.txt"
	charlesFile := "charlesFile.txt"
	dorisFile := "dorisFile.txt"
	eveFile := "eveFile.txt"
	// frankFile := "frankFile.txt"
	// graceFile := "graceFile.txt"
	// horaceFile := "horaceFile.txt"
	// iraFile := "iraFile.txt"

	BeforeEach(func() {
		// This runs before each test within this Describe block (including nested tests).
		// Here, we reset the state of Datastore and Keystore so that tests do not interfere with each other.
		// We also initialize
		userlib.DatastoreClear()
		userlib.KeystoreClear()
	})

	Describe("Basic Tests", func() {

		Specify("Basic Test: Testing InitUser/GetUser on a single user.", func() {
			userlib.DebugMsg("Initializing user Alice.")
			alice, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Getting user Alice.")
			aliceLaptop, err = client.GetUser("alice", defaultPassword)
			Expect(err).To(BeNil())
		})

		Specify("Basic Test: Testing Single User Store/Load/Append.", func() {
			userlib.DebugMsg("Initializing user Alice.")
			alice, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Storing file data: %s", contentOne)
			err = alice.StoreFile(aliceFile, []byte(contentOne))
			Expect(err).To(BeNil())

			userlib.DebugMsg("Appending file data: %s", contentTwo)
			err = alice.AppendToFile(aliceFile, []byte(contentTwo))
			Expect(err).To(BeNil())

			userlib.DebugMsg("Appending file data: %s", contentThree)
			err = alice.AppendToFile(aliceFile, []byte(contentThree))
			Expect(err).To(BeNil())

			userlib.DebugMsg("Loading file...")
			data, err := alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))
		})

		Specify("Basic Test: Testing Create/Accept Invite Functionality with multiple users and multiple instances.", func() {
			userlib.DebugMsg("Initializing users Alice (aliceDesktop) and Bob.")
			aliceDesktop, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			bob, err = client.InitUser("bob", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Getting second instance of Alice - aliceLaptop")
			aliceLaptop, err = client.GetUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("aliceDesktop storing file %s with content: %s", aliceFile, contentOne)
			err = aliceDesktop.StoreFile(aliceFile, []byte(contentOne))
			Expect(err).To(BeNil())

			userlib.DebugMsg("aliceLaptop creating invite for Bob.")
			invite, err := aliceLaptop.CreateInvitation(aliceFile, "bob")
			Expect(err).To(BeNil())

			userlib.DebugMsg("Bob accepting invite from Alice under filename %s.", bobFile)
			err = bob.AcceptInvitation("alice", invite, bobFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Bob appending to file %s, content: %s", bobFile, contentTwo)
			err = bob.AppendToFile(bobFile, []byte(contentTwo))
			Expect(err).To(BeNil())

			userlib.DebugMsg("aliceDesktop appending to file %s, content: %s", aliceFile, contentThree)
			err = aliceDesktop.AppendToFile(aliceFile, []byte(contentThree))
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that aliceDesktop sees expected file data.")
			data, err := aliceDesktop.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))

			userlib.DebugMsg("Checking that aliceLaptop sees expected file data.")
			data, err = aliceLaptop.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))

			userlib.DebugMsg("Checking that Bob sees expected file data.")
			data, err = bob.LoadFile(bobFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))

			userlib.DebugMsg("Getting third instance of Alice - alicePhone.")
			alicePhone, err = client.GetUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that alicePhone sees Alice's changes.")
			data, err = alicePhone.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))
		})

		Specify("Basic Test: Testing Revoke Functionality", func() {
			userlib.DebugMsg("Initializing users Alice, Bob, and Charlie.")
			alice, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			bob, err = client.InitUser("bob", defaultPassword)
			Expect(err).To(BeNil())

			charles, err = client.InitUser("charles", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Alice storing file %s with content: %s", aliceFile, contentOne)
			alice.StoreFile(aliceFile, []byte(contentOne))

			userlib.DebugMsg("Alice creating invite for Bob for file %s, and Bob accepting invite under name %s.", aliceFile, bobFile)

			invite, err := alice.CreateInvitation(aliceFile, "bob")
			Expect(err).To(BeNil())

			err = bob.AcceptInvitation("alice", invite, bobFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Alice can still load the file.")
			data, err := alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Bob can load the file.")
			data, err = bob.LoadFile(bobFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Bob creating invite for Charles for file %s, and Charlie accepting invite under name %s.", bobFile, charlesFile)
			invite, err = bob.CreateInvitation(bobFile, "charles")
			Expect(err).To(BeNil())

			err = charles.AcceptInvitation("bob", invite, charlesFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Bob can load the file.")
			data, err = bob.LoadFile(bobFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Charles can load the file.")
			data, err = charles.LoadFile(charlesFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Alice revoking Bob's access from %s.", aliceFile)
			err = alice.RevokeAccess(aliceFile, "bob")
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Alice can still load the file.")
			data, err = alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Bob/Charles lost access to the file.")
			_, err = bob.LoadFile(bobFile)
			Expect(err).ToNot(BeNil())

			_, err = charles.LoadFile(charlesFile)
			Expect(err).ToNot(BeNil())

			userlib.DebugMsg("Checking that the revoked users cannot append to the file.")
			err = bob.AppendToFile(bobFile, []byte(contentTwo))
			Expect(err).ToNot(BeNil())

			err = charles.AppendToFile(charlesFile, []byte(contentTwo))
			Expect(err).ToNot(BeNil())
		})


		Specify("Harder test: Testing if Shared User can edit contents", func() {
			userlib.DebugMsg("Initializing users Alice, Bob, and Charlie.")
			alice, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			bob, err = client.InitUser("bob", defaultPassword)
			Expect(err).To(BeNil())

			charles, err = client.InitUser("charles", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Alice storing file %s with content: %s", aliceFile, contentOne)
			alice.StoreFile(aliceFile, []byte(contentOne))

			userlib.DebugMsg("Alice creating invite for Bob for file %s, and Bob accepting invite under name %s.", aliceFile, bobFile)
			invite, err := alice.CreateInvitation(aliceFile, "bob")
			Expect(err).To(BeNil())

			err = bob.AcceptInvitation("alice", invite, bobFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Alice can still load the file.")
			data, err := alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Bob can load the file.")
			data, err = bob.LoadFile(bobFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Bob creating invite for Charles for file %s, and Charlie accepting invite under name %s.", bobFile, charlesFile)
			invite, err = bob.CreateInvitation(bobFile, "charles")
			Expect(err).To(BeNil())

			err = charles.AcceptInvitation("bob", invite, charlesFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Bob storing to file %s, content: %s", bobFile, contentTwo)
			err = bob.StoreFile(bobFile, []byte(contentTwo))
			Expect(err).To(BeNil())

			data, err = alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentTwo)))
	})

	Specify("Harder test: Testing if Shared User can edit contents", func() {
			userlib.DebugMsg("Initializing users Alice, Bob, and Charlie.")
			alice, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			bob, err = client.InitUser("bob", defaultPassword)
			Expect(err).To(BeNil())

			charles, err = client.InitUser("charles", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Alice storing file %s with content: %s", aliceFile, contentOne)
			alice.StoreFile(aliceFile, []byte(contentOne))

			userlib.DebugMsg("Alice creating invite for Bob for file %s, and Bob accepting invite under name %s.", aliceFile, bobFile)
			invite, err := alice.CreateInvitation(aliceFile, "bob")
			Expect(err).To(BeNil())

			err = bob.AcceptInvitation("alice", invite, bobFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Alice can still load the file.")
			data, err := alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Bob can load the file.")
			data, err = bob.LoadFile(bobFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Bob creating invite for Charles for file %s, and Charlie accepting invite under name %s.", bobFile, charlesFile)
			invite, err = bob.CreateInvitation(bobFile, "charles")
			Expect(err).To(BeNil())

			err = charles.AcceptInvitation("bob", invite, charlesFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Bob storing to file %s, content: %s", bobFile, contentTwo)
			err = bob.StoreFile(bobFile, []byte(contentTwo))
			Expect(err).To(BeNil())

			data, err = alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentTwo)))
	})
	
	Specify("Harder test: Testing if Shared User can edit contents", func() {
		userlib.DebugMsg("Initializing users Alice, Bob, and Charlie.")
		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())

		bob, err = client.InitUser("bob", defaultPassword)
		Expect(err).To(BeNil())

		charles, err = client.InitUser("charles", defaultPassword)
		Expect(err).To(BeNil())

		userlib.DebugMsg("Alice storing file %s with content: %s", aliceFile, contentOne)
		alice.StoreFile(aliceFile, []byte(contentOne))

		userlib.DebugMsg("Alice creating invite for Bob for file %s, and Bob accepting invite under name %s.", aliceFile, bobFile)
		invite, err := alice.CreateInvitation(aliceFile, "bob")
		Expect(err).To(BeNil())

		err = bob.AcceptInvitation("alice", invite, bobFile)
		Expect(err).To(BeNil())

		userlib.DebugMsg("Checking that Alice can still load the file.")
		data, err := alice.LoadFile(aliceFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))

		userlib.DebugMsg("Checking that Bob can load the file.")
		data, err = bob.LoadFile(bobFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))

		userlib.DebugMsg("Bob creating invite for Charles for file %s, and Charlie accepting invite under name %s.", bobFile, charlesFile)
		invite, err = bob.CreateInvitation(bobFile, "charles")
		Expect(err).To(BeNil())

		err = charles.AcceptInvitation("bob", invite, charlesFile)
		Expect(err).To(BeNil())

		userlib.DebugMsg("Bob storing to file %s, content: %s", bobFile, contentTwo)
		err = bob.StoreFile(bobFile, []byte(contentTwo))
		Expect(err).To(BeNil())

		data, err = alice.LoadFile(aliceFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentTwo)))
	})

	Specify("Init User Errors", func() {
		userlib.DebugMsg("Test if username can't be initialized if already exists")
		aliceDesktop, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())
		
		_, err = client.InitUser("alice", defaultPassword)
		Expect(err).NotTo(BeNil())
		
		userlib.DebugMsg("Test if username can't be initialized if empty string")
		_, err = client.InitUser("", defaultPassword)
		Expect(err).NotTo(BeNil())
	})

	Specify("Get User Errors", func() {
		//Error if there is no user with the given name 
		aliceDesktop, err = client.GetUser("alice", defaultPassword)
		Expect(err).NotTo(BeNil())

		//Error if the credentials are invalid 
		bob, err = client.InitUser("bob", defaultPassword)
		bob, err = client.GetUser("bob", emptyString)
		Expect(err).NotTo(BeNil())
		
		//Error if there is malicious activity 
	})

	Specify("Testing two different instances", func() {
		//For example, evanbotlaptop and evanbotphone saving a file or something? 
		aliceDesktop, err = client.InitUser("alice", defaultPassword)
		alicePhone, err = client.GetUser("alice", defaultPassword)
		aliceDesktop.StoreFile(aliceFile, []byte(contentOne))
		data, err := alicePhone.LoadFile(aliceFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))
	})

	Specify("Testing two different users with the same filename", func() {
		//FileNames: 
		//Different users can have the same filename but would refer to different files
		aliceDesktop, err = client.InitUser("alice", defaultPassword)
		bob, err = client.InitUser("bob", defaultPassword)
		sameName := "laskdmf"
		aliceDesktop.StoreFile(sameName, []byte(contentOne))
		bob.StoreFile(sameName, []byte(contentTwo))

		data, err := aliceDesktop.LoadFile(sameName)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))

		data, err = bob.LoadFile(sameName)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentTwo)))
	})


	Specify("Load a file that does not exist", func() {
		userlib.DebugMsg("Initializing users Alice")
		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())

		userlib.DebugMsg("Load non existent file")
		_, err := alice.LoadFile(aliceFile)
		Expect(err).NotTo(BeNil())
	})

	Specify("Can't accept invite for existing file name: Owner", func() {
		userlib.DebugMsg("Initializing users Alice, Bob, and Charlie.")
		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())

		bob, err = client.InitUser("bob", defaultPassword)
		Expect(err).To(BeNil())

		userlib.DebugMsg("Alice storing file %s with content: %s", aliceFile, contentOne)
		alice.StoreFile(aliceFile, []byte(contentOne))
		
		err = bob.StoreFile(bobFile, []byte(contentOne))
		Expect(err).To(BeNil())

		userlib.DebugMsg("Alice invites bob")
		invite, err := alice.CreateInvitation(aliceFile, "bob")
		Expect(err).To(BeNil())

		userlib.DebugMsg("Bob accepting invite")
		err = bob.AcceptInvitation("alice", invite, bobFile)
		Expect(err).NotTo(BeNil())
	})

	Specify("Can't accept invite for existing file name: Shared", func() {
		userlib.DebugMsg("Initializing users Alice, Bob, and Charlie.")
		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())

		bob, err = client.InitUser("bob", defaultPassword)
		Expect(err).To(BeNil())

		charles, err = client.InitUser("charles", defaultPassword)
		Expect(err).To(BeNil())
		
		userlib.DebugMsg("Alice storing file %s with content: %s", aliceFile, contentOne)
		alice.StoreFile(aliceFile, []byte(contentOne))

		err = bob.StoreFile(bobFile, []byte(contentOne))
		Expect(err).To(BeNil())

		userlib.DebugMsg("Alice invites charles")
		invite, err := alice.CreateInvitation(aliceFile, "charles")
		Expect(err).To(BeNil())

		userlib.DebugMsg("Bob accepting invite")
		err = charles.AcceptInvitation("alice", invite, charlesFile)
		Expect(err).To(BeNil())

		userlib.DebugMsg("Bob invites charles")
		invite, err = bob.CreateInvitation(bobFile, "charles")
		Expect(err).To(BeNil())

		userlib.DebugMsg("Bob accepting invite")
		err = charles.AcceptInvitation("bob", invite, charlesFile)
		Expect(err).ToNot(BeNil())
	})

	Specify("Create invitation fails if don't have file or recipient doesn't exist", func() {
		userlib.DebugMsg("Initializing users Alice, Bob, and Charlie.")
		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())

		bob, err = client.InitUser("bob", defaultPassword)
		Expect(err).To(BeNil())

		userlib.DebugMsg("Alice invites bob")
		_, err := alice.CreateInvitation(aliceFile, "bob")
		Expect(err).NotTo(BeNil())

		userlib.DebugMsg("Alice invites non existent charles")
		_, err = alice.CreateInvitation(aliceFile, "charles")
		Expect(err).NotTo(BeNil())
	})



	Specify("Revoke access errors", func() {
		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())

		bob, err = client.InitUser("bob", defaultPassword)
		Expect(err).To(BeNil())

		charles, err = client.InitUser("charles", defaultPassword)
		Expect(err).To(BeNil())

		alice.StoreFile(aliceFile, []byte(contentOne))

		invite, err := alice.CreateInvitation(aliceFile, "bob")
		Expect(err).To(BeNil())

		err = bob.AcceptInvitation("alice", invite, bobFile)
		Expect(err).To(BeNil())

		invite, err = bob.CreateInvitation(bobFile, "charles")
		Expect(err).To(BeNil())

		err = charles.AcceptInvitation("bob", invite, charlesFile)
		Expect(err).To(BeNil())

		//should fail because bob can't revoke charles  
		err = bob.RevokeAccess(bobFile, "charles")
		Expect(err).NotTo(BeNil())

		//should fail because alice can't revoke charles
		err = alice.RevokeAccess(aliceFile, "charles")
		Expect(err).NotTo(BeNil())

		//Bob and Charles should still be able to load the file because they haven't been revoked
		_, err = bob.LoadFile(bobFile)
		Expect(err).To(BeNil())
		_, err = charles.LoadFile(charlesFile)
		Expect(err).To(BeNil())

		//Now alice revokes Bob 
		err = alice.RevokeAccess(aliceFile, "bob")
		Expect(err).To(BeNil())

		//Bob and Charles should not have access to their files anymore
		_, err = bob.LoadFile(bobFile)
		Expect(err).NotTo(BeNil())
		_, err = charles.LoadFile(charlesFile)
		Expect(err).NotTo(BeNil())

	})

	Specify("Alice creates invitation for Bob but revokes before Bob accepts it", func() {
		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())

		bob, err = client.InitUser("bob", defaultPassword)
		Expect(err).To(BeNil())


		alice.StoreFile(aliceFile, []byte(contentOne))

		invite, err := alice.CreateInvitation(aliceFile, "bob")
		Expect(err).To(BeNil())

		err = alice.RevokeAccess(aliceFile, "bob")
		Expect(err).To(BeNil())

		//Bob shouldn't be able to accept the invitation
		err = bob.AcceptInvitation("alice", invite, bobFile)
		Expect(err).NotTo(BeNil())

		//Bob shouldn't be able to read the invitation
		_, err = bob.LoadFile(bobFile)
		Expect(err).NotTo(BeNil())
		_, err = charles.LoadFile(charlesFile)
		Expect(err).NotTo(BeNil())
	})


	Specify("Alice creates invitation for Bob. Alice created invitation for doris. Bob shares with Charles. doris shares with eve. Alice revokes Bob. ", func() {
		//we create alice, bob, charles, doris and eve
		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())

		bob, err = client.InitUser("bob", defaultPassword)
		Expect(err).To(BeNil())

		charles, err = client.InitUser("charles", defaultPassword)
		Expect(err).To(BeNil())

		doris, err = client.InitUser("doris", defaultPassword)
		Expect(err).To(BeNil())
		
		eve, err = client.InitUser("eve", defaultPassword)
		Expect(err).To(BeNil())



		//Alice stores the file contentOne which she will share
		alice.StoreFile(aliceFile, []byte(contentOne))

		//Alice invites both bob and doris and they accept
		invite, err := alice.CreateInvitation(aliceFile, "bob")
		Expect(err).To(BeNil())

		err = bob.AcceptInvitation("alice", invite, bobFile)
		Expect(err).To(BeNil())

		invite, err = alice.CreateInvitation(aliceFile, "doris")
		Expect(err).To(BeNil())

		err = doris.AcceptInvitation("alice", invite, dorisFile)
		Expect(err).To(BeNil())

		
		//We check to see if bob and doris can load the file 
		data, err := bob.LoadFile(bobFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))
		
		data, err = doris.LoadFile(dorisFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))


		//Now bob shares with charles and doris shares with eve 
		invite, err = bob.CreateInvitation(bobFile, "charles")
		Expect(err).To(BeNil())

		err = charles.AcceptInvitation("bob", invite, charlesFile)
		Expect(err).To(BeNil())

		invite, err = doris.CreateInvitation(dorisFile, "eve")
		Expect(err).To(BeNil())

		err = eve.AcceptInvitation("doris", invite, eveFile)
		Expect(err).To(BeNil())



		//charles and eve should see the changes 
		data, err = charles.LoadFile(charlesFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))
		
		data, err = eve.LoadFile(eveFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))


		//Now we revoke Bob
		err = alice.RevokeAccess(aliceFile, "bob")
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))


		//Bob and charles shouldn't be able to read the file
		_, err = bob.LoadFile(bobFile)
		Expect(err).NotTo(BeNil())

		_, err = charles.LoadFile(charlesFile)
		Expect(err).NotTo(BeNil())


		// Doris and Eve should be able to read
		data, err = doris.LoadFile(dorisFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))

		_, err = eve.LoadFile(eveFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))
	})
	

	Specify("Alice creates invitation for Boband doris. Bob shares with Charles. doris shares with eve. Eve makes a change ", func() {
		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())

		bob, err = client.InitUser("bob", defaultPassword)
		Expect(err).To(BeNil())

		charles, err = client.InitUser("charles", defaultPassword)
		Expect(err).To(BeNil())

		doris, err = client.InitUser("doris", defaultPassword)
		Expect(err).To(BeNil())
		
		eve, err = client.InitUser("eve", defaultPassword)
		Expect(err).To(BeNil())



		
		alice.StoreFile(aliceFile, []byte(contentOne))

		invite, err := alice.CreateInvitation(aliceFile, "bob")
		Expect(err).To(BeNil())

		err = bob.AcceptInvitation("alice", invite, bobFile)
		Expect(err).To(BeNil())

		invite, err = alice.CreateInvitation(aliceFile, "doris")
		Expect(err).To(BeNil())

		err = doris.AcceptInvitation("alice", invite, dorisFile)
		Expect(err).To(BeNil())

		

		data, err := bob.LoadFile(bobFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))
		
		data, err = doris.LoadFile(dorisFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))



		invite, err = bob.CreateInvitation(bobFile, "charles")
		Expect(err).To(BeNil())

		err = charles.AcceptInvitation("bob", invite, charlesFile)
		Expect(err).To(BeNil())

		invite, err = doris.CreateInvitation(dorisFile, "eve")
		Expect(err).To(BeNil())

		err = eve.AcceptInvitation("doris", invite, eveFile)
		Expect(err).To(BeNil())



		data, err = charles.LoadFile(charlesFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))
		
		data, err = eve.LoadFile(eveFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))


		eve.StoreFile(eveFile, []byte(contentTwo))

		//Bob and charles shouldn't be able to read the file
		data, err = alice.LoadFile(aliceFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentTwo)))
		
		data, err = bob.LoadFile(bobFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentTwo)))


		data, err = charles.LoadFile(charlesFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentTwo)))

		data, err = doris.LoadFile(dorisFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentTwo)))

		_, err = eve.LoadFile(eveFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentTwo)))
	})

	Specify("Testing errors for init user", func() {
		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())
		userlib.DatastoreClear()
		alice, err = client.GetUser("alice", defaultPassword)
		Expect(err).NotTo(BeNil())

	})

	Specify("Try to init user twice with the same information", func() {
		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())

		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).NotTo(BeNil())

		alice, err = client.InitUser("alice", "wrong_pass")
		Expect(err).NotTo(BeNil())
	})

	Specify("Data store tampered", func() {
		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())

		bob, err = client.InitUser("bob", defaultPassword)
		Expect(err).To(BeNil())

		err = alice.StoreFile(aliceFile, []byte(contentOne))
		Expect(err).To(BeNil())

		invite, err := alice.CreateInvitation(aliceFile, "bob")
		Expect(err).To(BeNil())

		err = bob.AcceptInvitation("alice", invite, bobFile)
		Expect(err).To(BeNil())

		
		

		dsMap := userlib.DatastoreGetMap()
		for k := range dsMap {
			userlib.DatastoreSet(k, []byte("trash"))
		}

		_, err = alice.LoadFile(aliceFile)
		Expect(err).NotTo(BeNil())

		_, err = bob.LoadFile(bobFile)
		Expect(err).NotTo(BeNil())


		alice, err = client.GetUser("alice", defaultPassword)
		Expect(err).NotTo(BeNil())

		bob, err = client.GetUser("bob", defaultPassword)
		Expect(err).NotTo(BeNil())
	})



	Specify("Alice creates invitation for Bob. Bob shares with Charles. Alice appends something. Bob appends something. Charles appends something ", func() {
		//we create alice, bob, charles
		alice, err = client.InitUser("alice", defaultPassword)
		Expect(err).To(BeNil())

		bob, err = client.InitUser("bob", defaultPassword)
		Expect(err).To(BeNil())

		charles, err = client.InitUser("charles", defaultPassword)
		Expect(err).To(BeNil())


		//Alice stores the file contentOne which she will share
		alice.StoreFile(aliceFile, []byte(contentOne))

		//Alice invites bob and they accept
		invite, err := alice.CreateInvitation(aliceFile, "bob")
		Expect(err).To(BeNil())

		err = bob.AcceptInvitation("alice", invite, bobFile)
		Expect(err).To(BeNil())

		
		//We check to see if bob can load the file 
		data, err := bob.LoadFile(bobFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))
		


		//Now bob shares with charles 
		invite, err = bob.CreateInvitation(bobFile, "charles")
		Expect(err).To(BeNil())

		err = charles.AcceptInvitation("bob", invite, charlesFile)
		Expect(err).To(BeNil())


		//charles should see the file 
		data, err = charles.LoadFile(charlesFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne)))
		
		//Alice makes changes by appending 
		alice.AppendToFile(aliceFile, []byte(contentTwo))

		//Bob should see the changes by alice 
		data, err = bob.LoadFile(bobFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne+contentTwo)))

		//Charles should see the changes by alice 
		data, err = charles.LoadFile(charlesFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne+contentTwo)))

		
		//bob makes changes by appending 
		bob.AppendToFile(bobFile, []byte(contentTwo))

		//alice should see the changes by bob 
		data, err = alice.LoadFile(aliceFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne+contentTwo+contentTwo)))

		//charles should see the changes by bob 
		data, err = charles.LoadFile(charlesFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne+contentTwo+contentTwo)))


		//charles makes changes by appending 
		charles.AppendToFile(charlesFile, []byte(contentTwo))

		//alice should see the changes by charles 
		data, err = alice.LoadFile(aliceFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne+contentTwo+contentTwo+contentTwo)))

		//Bob should see the changes by charles 
		data, err = bob.LoadFile(bobFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal([]byte(contentOne+contentTwo+contentTwo+contentTwo)))

	})

	


})
})