package BlueTooth

import (
	"github.com/paypal/gatt"
	"log"
)

const (
	orientationService        = "c7e70010c84711e681758c89a55d403c"
	orientationCharacteristic = "c7e70012c84711e681758c89a55d403c"
)

type ZeiManager struct {
	OnOrientationChanged func(side int)

	done                 chan struct{}
	device               gatt.Device
}


func (zm *ZeiManager) Run() {
	clientOptions := []gatt.Option{
		gatt.LnxMaxConnections(1),
		gatt.LnxDeviceID(-1, true),
	}

	device, err := gatt.NewDevice(clientOptions...)

	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	zm.device = device

	zm.done = make(chan struct{})

	zm.device.Handle(
		gatt.PeripheralDiscovered(zm.peripheralDiscovered),
		gatt.PeripheralConnected(zm.peripheralConnected),
		gatt.PeripheralDisconnected(zm.peripheralDisconnected),
	)
	zm.device.Init(zm.stateChanged)
}

func (zm *ZeiManager) peripheralDiscovered(zei gatt.Peripheral, advertisement *gatt.Advertisement, i int) {
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

func (zm *ZeiManager) peripheralDisconnected(zei gatt.Peripheral, i error) {
	log.Println("ZEI disconnected")

	if i != nil {
		log.Printf("Error: %s", i)
	}

	close(zm.done)
	zm.done = make(chan struct{})
	zm.device.Scan([]gatt.UUID{}, false)
}

func (zm *ZeiManager) peripheralConnected(zei gatt.Peripheral, i error) {
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
		if possibleService.UUID().String() == orientationService {
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
		if possibleChar.UUID().String() == orientationCharacteristic {
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

	<-zm.done
}

func (zm *ZeiManager) onIndicate(c *gatt.Characteristic, b []byte, err error) {
	if err != nil {
		log.Printf("Failed reading the client characteristic")
		return
	}

	go zm.OnOrientationChanged(int(b[0]))
}

func (zm *ZeiManager) stateChanged(device gatt.Device, state gatt.State) {
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
