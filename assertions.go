package pmset

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

void get_system_assertions();
void get_pid_assertions();
void subscribe_assertions();
void run_subscribed_assertions();

*/
import "C"
import (
	"log"
	"sync"
)

// PidAssertion represents one process that has an assertion
type PidAssertion struct {
	PID  int
	Name string
}

// AssertionChange represents a process making a change to an assertion
type AssertionChange struct {
	Action string
	Type   string
	Pid    PidAssertion
}

// GetAssertions returns a map of assertion keys to values
func GetAssertions() map[string]int {
	C.get_system_assertions()
	<-systemDone
	return systemAssertions
}

// GetPIDAssertions returns a map of assertion keys to procceses holding those assertions
func GetPIDAssertions() map[string][]PidAssertion {
	C.get_pid_assertions()
	<-pidDone
	return pidAssertions
}

// SubscribeAssertionChangesAndRun does not return, changes come through the supplied channel
func SubscribeAssertionChangesAndRun(channel chan<- AssertionChange) {
	SubscribeAssertionChanges(channel)
	C.run_subscribed_assertions()
}

// SubscribeAssertionChanges return, changes come through the supplied channel once dispatch_main or nsapplication is run
func SubscribeAssertionChanges(channel chan<- AssertionChange) {
	go func() {
		for range subscriptionReady {
			channel <- assertionChange
		}
	}()
	C.subscribe_assertions()
}

var systemMutex = &sync.Mutex{}
var systemAssertions map[string]int
var systemDone = make(chan bool, 1)

//export startSystemAssertions
func startSystemAssertions() {
	systemMutex.Lock()
	systemAssertions = make(map[string]int)
}

//export systemAssertion
func systemAssertion(nameCStr *C.char, val int) {
	name := C.GoString(nameCStr)
	systemAssertions[name] = val
}

//export doneSystemAssertions
func doneSystemAssertions() {
	systemMutex.Unlock()
	systemDone <- true
}

var pidMutex = &sync.Mutex{}
var pidAssertions map[string][]PidAssertion
var pidDone = make(chan bool, 1)

//export startPidAssertions
func startPidAssertions() {
	pidMutex.Lock()
	pidAssertions = make(map[string][]PidAssertion)
}

//export pidAssertion
func pidAssertion(pid int, keyCStr *C.char, val int, nameCStr *C.char, timedoutCStr *C.char) {
	key := C.GoString(keyCStr)
	name := C.GoString(nameCStr)
	timedout := C.GoString(timedoutCStr)
	if timedout != "" {
		log.Printf("Getting pids timed out: %s\n", timedout)
	}
	pidAssertions[key] = append(pidAssertions[key], PidAssertion{
		PID:  pid,
		Name: name,
	})
}

//export donePidAssertions
func donePidAssertions() {
	pidMutex.Unlock()
	pidDone <- true
}

var subscriptionMutex = &sync.Mutex{}
var assertionChange AssertionChange
var subscriptionReady = make(chan bool, 1)

//export assertionChangeStart
func assertionChangeStart() {
	subscriptionMutex.Lock()
	assertionChange = AssertionChange{}
}

//export subscriptionAction
func subscriptionAction(actionCStr *C.char) {
	action := C.GoString(actionCStr)
	assertionChange.Action = action
}

//export subscriptionType
func subscriptionType(typeCStr *C.char) {
	subType := C.GoString(typeCStr)
	assertionChange.Type = subType
}

//export subscriptionPid
func subscriptionPid(pid int) {
	assertionChange.Pid.PID = pid
}

//export subscriptionProcessName
func subscriptionProcessName(processNameCStr *C.char) {
	processName := C.GoString(processNameCStr)
	assertionChange.Pid.Name = processName
}

//export assertionChangeReady
func assertionChangeReady() {
	subscriptionReady <- true
	subscriptionMutex.Unlock()
}
