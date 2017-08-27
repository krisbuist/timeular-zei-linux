package main

import (
	"github.com/paypal/gatt"
	"log"
)

const ORIENTATION_SERVICE = "c7e70010c84711e681758c89a55d403c"
const ORIENTATION_CHARACTERISTIC = "c7e70012c84711e681758c89a55d403c"

type ZeiManager struct {
	OnOrientationChanged func(side int)
	Done                 chan struct{}
	Device               gatt.Device
}

var ClientOptions = []gatt.Option{
	gatt.LnxMaxConnections(1),
	gatt.LnxDeviceID(-1, true),
}

func (zm *ZeiManager) run() {

	device, err := gatt.NewDevice(ClientOptions...)

	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	zm.Device = device

	zm.Done = make(chan struct{})

	zm.Device.Handle(
		gatt.PeripheralDiscovered(zm.PeripheralDiscovered),
		gatt.PeripheralConnected(zm.PeripheralConnected),
		gatt.PeripheralDisconnected(zm.PeripheralDisconnected),
	)
	zm.Device.Init(zm.StateChanged)
}

func (zm *ZeiManager) PeripheralDiscovered(zei gatt.Peripheral, advertisement *gatt.Advertisement, i int) {
	if zei.Name() != "Timeular ZEI" {
		return
	}
	log.Println("ZEI discovered. Connecting...")

	if !advertisement.Connectable {
		log.Println("ZEI is not connectable")
		return
	}

	zei.Device().StopScanning()
	zei.Device().Connect(zei)
}

func (zm *ZeiManager) PeripheralDisconnected(zei gatt.Peripheral, i error) {
	log.Println("ZEI disconnected")

	if i != nil {
		log.Printf("Error: %s", i)
	}

	close(zm.Done)
	zm.Done = make(chan struct{})
	zm.Device.Scan([]gatt.UUID{}, false)
}

func (zm *ZeiManager) PeripheralConnected(zei gatt.Peripheral, i error) {
	log.Println("ZEI connected")
	defer zei.Device().CancelConnection(zei)

	if err := zei.SetMTU(500); err != nil {
		log.Printf("Failed to set MTU, err: %s\n", err)
	}

	services, err := zei.DiscoverServices(nil)

	if err != nil {
		log.Printf("Failed to discover the ZEI services, err: %s\n", err)
		return
	}

	var service *gatt.Service = nil

	for _, possibleService := range services {
		if possibleService.UUID().String() == ORIENTATION_SERVICE {
			service = possibleService
		}
	}

	if service == nil {
		log.Println("Failed to find the orientation service")
		return
	}

	characteristics, err := zei.DiscoverCharacteristics(nil, service)

	if err != nil {
		log.Printf("Failed to read the orientation characteristic, err: %s\n", err)
		return
	}

	var char *gatt.Characteristic = nil

	for _, possibleChar := range characteristics {
		if possibleChar.UUID().String() == ORIENTATION_CHARACTERISTIC {
			char = possibleChar
		}
	}

	if char == nil {
		log.Println("Failed to find orientation characteristic")
		return
	}

	if (char.Properties() & gatt.CharIndicate) == 0 {
		log.Println("Characteristic does not support indicate")
		return
	}

	_, err = zei.DiscoverDescriptors(nil, char)
	if err != nil {
		log.Printf("Failed to discover descriptors, err: %s\n", err)
		return
	}

	value, err := zei.ReadCharacteristic(char)
	zm.onIndicate(char, value, err)

	if err := zei.SetIndicateValue(char, zm.onIndicate); err != nil {
		log.Printf("Failed to subscribe characteristic, err: %s\n", err)
		return
	}

	<-zm.Done
}

func (zm *ZeiManager) onIndicate(c *gatt.Characteristic, b []byte, err error) {
	if err != nil {
		log.Printf("Failed reading the client characteristic")
		return
	}

	zm.OnOrientationChanged(int(b[0]))
}

func (zm *ZeiManager) StateChanged(device gatt.Device, state gatt.State) {
	log.Printf("State: %s\n", state)
	switch state {
	case gatt.StatePoweredOn:
		log.Println("Scanning...")
		device.Scan([]gatt.UUID{}, false)
		return
	default:
		device.StopScanning()
	}
}
