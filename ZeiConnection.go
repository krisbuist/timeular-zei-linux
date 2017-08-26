package main

import (
	"github.com/paypal/gatt"
	"log"
)

const ORIENTATION_SERVICE = "c7e70010c84711e681758c89a55d403c"
const ORIENTATION_CHARACTERISTIC = "c7e70012c84711e681758c89a55d403c"

var ClientOptions = []gatt.Option{
	gatt.LnxMaxConnections(1),
	gatt.LnxDeviceID(-1, true),
}
var done = make(chan struct{})

func main() {
	d, err := gatt.NewDevice(ClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	d.Handle(
		gatt.PeripheralDiscovered(onPeripheralDiscovered),
		gatt.PeripheralConnected(onPeripheralConnected),
		gatt.PeripheralDisconnected(onPeripheralDisconnected),
	)

	d.Init(onStateChanged)
	<-done
	log.Println("Done")
}

func onPeripheralDiscovered(zei gatt.Peripheral, advertisement *gatt.Advertisement, i int) {
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

func onPeripheralDisconnected(zei gatt.Peripheral, i error) {
	log.Println("ZEI disconnected")
	close(done)
}

func onPeripheralConnected(zei gatt.Peripheral, i error) {
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

	log.Printf("Found orientation service: %s\n", service.UUID().String())

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

	log.Printf("Found characteristic: %s\n", char.UUID().String())

	if (char.Properties() & gatt.CharIndicate) == 0 {
		log.Println("Characteristic does not support indicate")
		return
	}

	ds, err := zei.DiscoverDescriptors(nil, char)
	if err != nil {
		log.Printf("Failed to discover descriptors, err: %s\n", err)
		return
	}

	log.Printf("%d descriptors found\n", len(ds))

	if err := zei.SetIndicateValue(char, onIndicate); err != nil {
		log.Printf("Failed to subscribe characteristic, err: %s\n", err)
		return
	}

	<-done
}

func onIndicate(c *gatt.Characteristic, b []byte, err error) {
	if err != nil {
		log.Printf("Failed reading the client characteristic")
		return
	}

	log.Printf("Orientation changed %s: %x\n", c.Name(), b)
}

func onStateChanged(device gatt.Device, state gatt.State) {
	log.Println("State:", state)
	switch state {
	case gatt.StatePoweredOn:
		log.Println("Scanning...")
		device.Scan([]gatt.UUID{}, false)
		return
	default:
		device.StopScanning()
	}
}
