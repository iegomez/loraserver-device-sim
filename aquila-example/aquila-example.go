package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/brocaar/lorawan"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	lds "github.com/iegomez/loraserver-device-sim"
)

func main() {

	//Connect to the broker
	opts := MQTT.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883")
	opts.SetUsername("loraserver_gw")
	opts.SetPassword("ChpP2eeW1T")

	client := MQTT.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Connection error")
		fmt.Println(token.Error())
	}

	fmt.Println("Connection established.")

	//Build your node with known keys (ABP).
	nwsHexKey := "3bc0ddd455d320a6f36ff6f2a25057d0"
	appHexKey := "00de01b45b59a4df9cc2b3fa5eb0fe7c"
	devHexAddr := "07262b83"
	devAddr, err := lds.HexToDevAddress(devHexAddr)
	if err != nil {
		fmt.Printf("dev addr error: %s", err)
	}

	nwkSKey, err := lds.HexToKey(nwsHexKey)
	if err != nil {
		fmt.Printf("nwkskey error: %s", err)
	}

	appSKey, err := lds.HexToKey(appHexKey)
	if err != nil {
		fmt.Printf("appskey error: %s", err)
	}

	appKey := [16]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	appEUI := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
	devEUI := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}

	device := &lds.Device{
		DevEUI:  devEUI,
		DevAddr: devAddr,
		NwkSKey: nwkSKey,
		AppSKey: appSKey,
		AppKey:  appKey,
		AppEUI:  appEUI,
		UlFcnt:  0,
		DlFcnt:  0,
	}

	for {

		/*
			*	Make up some random values.
			*
			*	These should be decoded at lora-app-server with a proper function.
			* 	For this example, the object should look like this:

				obj : {
					"temperature": {
						"value":((bytes[0]*256+bytes[1])/100),"unit":"°C"
					},
					"pressure": {
						"value":((bytes[2]*16*16*16*16+bytes[3]*256+bytes[4])/100),"unit":"hPa"
					},
					"humidity": {
						"value":((bytes[5]*256+bytes[6])/1024),"unit":"%"
					}
				}

			*
		*/

		rand.Seed(time.Now().UnixNano() / 10000)
		tp := float64(rand.Intn(25)) + rand.Float64()*float64(rand.Intn(4))
		lt := -33.4445782 + rand.Float32()/1000.0
		ln := -70.6377106 + rand.Float32()/1000.0
		lat := generateLat(lt)
		lng := generateLng(ln)
		temp := generateTemp2byte(int16(tp))

		fmt.Printf("sending lat: %f and lng: %f\n", lt, ln)

		//Create the payload, data rate and rx info.
		payload := []byte{temp[0], temp[1]}
		payload = append(payload[:], lat...)
		payload = append(payload[:], lng...)

		fmt.Printf("bytes: %v\n", payload)

		//Change to your gateway MAC to build RxInfo.
		gwMac := "b827ebfffe9448d0"

		//Construct DataRate RxInfo with proper values according to your band (example is for US 915).

		dataRate := &lds.DataRate{
			Bandwidth:    500,
			Modulation:   "LORA",
			SpreadFactor: 8,
			BitRate:      0}

		rxInfo := &lds.RxInfo{
			Channel:   0,
			CodeRate:  "4/5",
			CrcStatus: 1,
			DataRate:  dataRate,
			Frequency: 902300000,
			LoRaSNR:   7,
			Mac:       gwMac,
			RfChain:   1,
			Rssi:      -57,
			Size:      23,
			Time:      time.Now().Format(time.RFC3339),
			Timestamp: int32(time.Now().UnixNano() / 1000000000),
		}

		//Now send an uplink
		err = device.Uplink(client, lorawan.UnconfirmedDataUp, 1, rxInfo, payload)
		if err != nil {
			fmt.Printf("couldn't send uplink: %s\n", err)
		}

		time.Sleep(3 * time.Second)

	}

}

func generateRisk(r int8) []byte {
	risk := uint8(r)
	bRep := make([]byte, 1)
	bRep[0] = risk
	return bRep
}

func generateTemp1byte(t int8) []byte {
	temp := uint8(t)
	bRep := make([]byte, 1)
	bRep[0] = temp
	return bRep
}

func generateTemp2byte(t int16) []byte {

	temp := uint16(float32(t/127.0) * float32(math.Pow(2, 15)))
	bRep := make([]byte, 2)
	binary.BigEndian.PutUint16(bRep, temp)
	return bRep
}

func generateLight(l int16) []byte {

	light := uint16(l)
	bRep := make([]byte, 2)
	binary.BigEndian.PutUint16(bRep, light)
	return bRep
}

func generateAltitude(a float32) []byte {

	alt := uint16(float32(a/1200) * float32(math.Pow(2, 15)))
	bRep := make([]byte, 2)
	binary.BigEndian.PutUint16(bRep, alt)
	return bRep
}

func generateLat(l float32) []byte {
	lat := uint32((l / 90.0) * float32(math.Pow(2, 31)))
	bRep := make([]byte, 4)
	binary.BigEndian.PutUint32(bRep, lat)
	return bRep
}

func generateLng(l float32) []byte {
	lng := uint32((l / 180.0) * float32(math.Pow(2, 31)))
	bRep := make([]byte, 4)
	binary.BigEndian.PutUint32(bRep, lng)
	return bRep
}
