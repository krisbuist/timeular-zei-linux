package BlueTooth

import (
	"github.com/currantlabs/ble"
	"log"
	"strings"
	"golang.org/x/net/context"
	"github.com/currantlabs/ble/linux"
)

const (
	orientationService        = "c7e70010c84711e681758c89a55d403c"
	orientationCharacteristic = "c7e70012c84711e681758c89a55d403c"
)

type ZeiManager struct {
	OnOrientationChanged func(side int)
}

func (zm *ZeiManager) Run() {
	d, err := linux.NewDevice()
	if err != nil {
		log.Fatalf("Can't create new device : %service", err)
	}
	ble.SetDefaultDevice(d)

	zm.connectAndRun()
}

func (zm *ZeiManager) connectAndRun() {
	log.Println("Trying to connect to the ZEI")

	cln, err := ble.Connect(context.Background(), func(a ble.Advertisement) bool {
		return strings.ToUpper(a.LocalName()) == strings.ToUpper("Timeular ZEI")
	})

	if err != nil {
		log.Fatalf("Can't connect : %service", err)
	}

	log.Println("Connected to the ZEI")

	defer cln.CancelConnection()

	done := make(chan struct{})
	go func() {
		<-cln.Disconnected()
		log.Println("ZEI disconnected")
		close(done)
	}()

	profile, err := cln.DiscoverProfile(true)

	if err != nil {
		log.Fatalf("Can't discover the profile: %s", err)
	}

	for _, service := range profile.Services {
		if !service.UUID.Equal(ble.MustParse(orientationService)) {
			continue
		}
		for _, char := range service.Characteristics {
			if !char.UUID.Equal(ble.MustParse(orientationCharacteristic)) {
				continue
			}

			// Do an initial read on the device side to know the starting position
			val, err := cln.ReadCharacteristic(char)
			if err != nil {
				log.Fatalf("Failed to read characteristic: %s\n", err)
			}
			go zm.OnOrientationChanged(int(val[0]))

			err = cln.Subscribe(char, true, func(val []byte) {
				go zm.OnOrientationChanged(int(val[0]))
			})
			if err != nil {
				log.Fatalf("Subscribing failed: %s\n", err)
			}
			log.Println("Subscribed to ZEI side changes")
		}
	}

	<-done
	zm.connectAndRun()
}
