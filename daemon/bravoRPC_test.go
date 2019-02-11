package main

import (
	"io/ioutil"
	"net/rpc"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

func Test_getIPFromConfig(t *testing.T) {
	os.Remove("/tmp/Test_getIPFromConfig")

	const testConfig = `
[Interface]
PrivateKey = someprivatekey
Address = 255.255.255.1/24,0:0:0:0:0:ffff:ffff:ff01/48
DNS =  255.255.255.255

[Peer]
PublicKey = somepublickey
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = 255.255.255.255:65536
PersistentKeepalive = 25
`

	err := ioutil.WriteFile("/tmp/Test_getIPFromConfig", []byte(testConfig), 0644)
	if err != nil {
		t.Error(err)
	}

	expected := "255.255.255.255:65536"
	result := getIPFromConfig("/tmp/Test_getIPFromConfig")

	if expected != result {
		t.Errorf("getIPFromConfig() = %v, want %v", result, expected)
	}
}

func TestBravoRPC_ListAll(t *testing.T) {
	exec.Command("rm", "-rf", "/tmp/TestBravoRPC_ListAll").Run() // the quickest option available

	path := "/tmp/TestBravoRPC_ListAll"

	expected := []Interface{
		{"abc", "255.255.255.255:65536"},
		{"def", "255.255.255.255:65536"},
		{"ghi", "255.255.255.255:65536"},
	}

	if err := os.Mkdir("/tmp/TestBravoRPC_ListAll", os.ModePerm); err != nil {
		t.Error(err)
	}

	for _, i := range expected {
		if err := ioutil.WriteFile(path + "/" + i.Name + ".conf", []byte("Endpoint = " + i.IP), 0644); err != nil {
			t.Error(err)
		}
	}

	go startRPC(make(chan interface{}))

	// a bit hacky:
	// - wait for RPC server to init
	// - wait for tester to punch a button
	//   to allow RPC to receive incoming connections (macOS)
	time.Sleep(5 * time.Second)

	client, err := rpc.DialHTTP("tcp", ":" + PORT)
	if err != nil {
		t.Error("Error dialing:", err)
	}
	defer client.Close()

	reply := new([]Interface)

	if err := client.Call("BravoRPC.ListAll", &path, &reply); err != nil {
		t.Error("BravoRPC error:", err)
	}

	if !reflect.DeepEqual(expected, *reply) {
		t.Errorf("BravoRPC.ListAll() = %v, want %v", reply, expected)
	}
}