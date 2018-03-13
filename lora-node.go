package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/brocaar/lorawan"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Message struct {
	PhyPayload string  `json:"phyPayload"`
	RxInfo     *RxInfo `json:"rxInfo"`
}

type RxInfo struct {
	Channel   int       `json:"channel"`
	CodeRate  string    `json:"codeRate"`
	CrcStatus int       `json:"crcStatus"`
	DataRate  *DataRate `json:"dataRate"`
	Frequency int       `json:"frequency"`
	LoRaSNR   int       `json:"loRaSNR"`
	Mac       string    `json:"mac"`
	RfChain   int       `json:"rfChain"`
	Rssi      int       `json:"rssi"`
	Size      int       `json:"size"`
	Time      string    `json:"time"`
	Timestamp int32     `json:"timestamp"`
}

type DataRate struct {
	Bandwidth    int    `json:"bandwidth"`
	Modulation   string `json:"modulation"`
	SpreadFactor int    `json:"spreadFactor"`
	BitRate      int    `json:"bitrate"`
}

type Location struct {
	Lat       float32
	Lng       float32
	Elevation float32
}

func main() {

	/*
	 *
	 * Create the mqtt client. Replace vlues where needed.
	 *
	 */
	opts := MQTT.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883")
	opts.SetUsername("loraserver_gw")
	opts.SetPassword("ChpP2eeW1Tck")

	client := MQTT.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Connection error")
		fmt.Println(token.Error())
	}

	fmt.Println("Connection established.")

	/*
	 * Define Gateway mac for message publishing.
	 * Replace with correct value.
	 */

	gwMac := "b827ebfffee100b5"

	/*
	 *
	 * To test OTAA activation, send a join request.
	 * Replace appKey, devEUI and appEUI with the correct ones.
	 *
	 */
	appKey := [16]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	appEUI := [8]byte{1, 1, 1, 1, 1, 1, 1, 1}
	devEUI := [8]byte{2, 2, 2, 2, 2, 2, 2, 2}

	/*
	   * For testing purposes, create a node with ABP activation and relaxed frame counter enabled.
	   Generate the keys and devAddr and replace them on nwsHexKey, appHexKey and devHexAddr.
	*/
	nwsHexkey := "3bc0ddd455d320a6f36ff6f2a25057d0"
	appHexKey := "00de01b45b59a4df9cc2b3fa5eb0fe7c"
	devHexAddr := "07262b83"

	rand.Seed(time.Now().UnixNano() / 10000)

	var devAddr ([4]byte)
	da, _ := hex.DecodeString(devHexAddr)
	copy(devAddr[:], da[:])

	var nwkSKey ([16]byte)
	nk, _ := hex.DecodeString(nwsHexkey)
	copy(nwkSKey[:], nk[:]) //{2,2,2,2,2,2,2,2,2,2,2,2,2,2,3,3}
	var appSKey ([16]byte)
	ak, _ := hex.DecodeString(appHexKey)
	copy(appSKey[:], ak[:]) //{2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2}

	/*
	* Generate second testing device
	 */

	/*
	   * For testing purposes, create a node with ABP activation and relaxed frame counter enabled.
	   Generate the keys and devAddr and replace them on nwsHexKey, appHexKey and devHexAddr.
	*/
	nwsHexkey2 := "ae3bca7b09d3def6bcf2552504353665"
	appHexKey2 := "c457759fd07709518793310284aeae82"
	devHexAddr2 := "073de9e6"

	rand.Seed(time.Now().UnixNano() / 10000)

	var devAddr2 ([4]byte)
	da2, _ := hex.DecodeString(devHexAddr2)
	copy(devAddr2[:], da2[:])

	var nwkSKey2 ([16]byte)
	nk2, _ := hex.DecodeString(nwsHexkey2)
	copy(nwkSKey2[:], nk2[:]) //{2,2,2,2,2,2,2,2,2,2,2,2,2,2,3,3}
	var appSKey2 ([16]byte)
	ak2, _ := hex.DecodeString(appHexKey2)
	copy(appSKey2[:], ak2[:]) //{2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2}

	/*
	 * Change last param to true to make an OTAA join request.
	 */

	join(client, appKey, appEUI, devEUI, gwMac, false)

	/*
	 * Send a test message with an ABP activated node.
	 */
	var lat float32 = -33.4335625
	var lng float32 = -70.6217137

	var lat2 float32 = -33.4335625
	var lng2 float32 = -70.6217137

	var temp = 20

	var light int16 = 1345
	var altitude float32 = 600

	/*
	* Preamble carries the message preamble as for our defined protocol, which consists of 1 byte.
	* 7: is gps fixed.
	* 6: does message carry a panic alert.
	* 5: Reserved.
	* 4: Reserved.
	* 3-0: Data group.
	*
	* So, for example 10xx0000 would mean GPS is fixed, there's no panic and group data is 0.
	*
	 */
	preamble := []byte{uint8(128)}
	preamble2 := []byte{uint8(129)}

	for {

		lat += rand.Float32() / 1000.0
		lng += rand.Float32() / 1000.0

		lat2 += rand.Float32() / 1000.0
		lng2 += rand.Float32() / 1000.0

		light += int16(rand.Float32() * 5)
		altitude += (rand.Float32() * 5)

		//lat = location.Lat
		//lng = location.Lng

		byte0 := []byte{0} //Data to send.

		mPayload := append(preamble[:], byte0[:]...)
		mPayload = append(mPayload[:], generateTemp1byte(int8(temp))[:]...)
		mPayload = append(mPayload[:], byte0[:]...)
		mPayload = append(mPayload[:], generateLat(lat)[:]...)
		mPayload = append(mPayload[:], generateLng(lng)[:]...)
		mPayload = append(mPayload[:], generateRisk(int8(rand.Intn(10)))[:]...)

		g2Payload := append(preamble2[:], generateLight(light)[:]...)
		g2Payload = append(g2Payload, generateAltitude(altitude)[:]...)

		mPayload2 := append(preamble[:], byte0[:]...)
		mPayload2 = append(mPayload2[:], generateTemp1byte(int8(temp))[:]...)
		mPayload2 = append(mPayload2[:], byte0[:]...)
		mPayload2 = append(mPayload2[:], generateLat(lat2)[:]...)
		mPayload2 = append(mPayload2[:], generateLng(lng2)[:]...)
		mPayload2 = append(mPayload2[:], generateRisk(int8(rand.Intn(10)))[:]...)

		fmt.Println(mPayload)
		fmt.Println(g2Payload)

		err := sendMessage(client, devAddr, appSKey, nwkSKey, gwMac, mPayload)
		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(2 * time.Second)

		g2Err := sendMessage(client, devAddr, appSKey, nwkSKey, gwMac, g2Payload)
		if g2Err != nil {
			fmt.Println(g2Err)
		}

		/*err2 := sendMessage(client, devAddr2, appSKey2, nwkSKey2, gwMac, mPayload2)
		if err2 != nil {
			fmt.Println(err2)
		}*/

		temp++
		if temp > 50 {
			temp = 20
		}
		time.Sleep(5 * time.Second)
	}

}

func createMessage(gwMac string, payload []byte) *Message {

	/*
	 * Set correct values for your environment.
	 */

	dataRate := &DataRate{
		Bandwidth:    500,
		Modulation:   "LORA",
		SpreadFactor: 8,
		BitRate:      0}

	rxInfo := &RxInfo{
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
		Timestamp: int32(time.Now().UnixNano() / 1000000000)}

	message := &Message{
		PhyPayload: string(payload),
		RxInfo:     rxInfo}

	return message

}

func publish(client MQTT.Client, topic string, v interface{}) error {
	bytes, err := json.Marshal(v)
	fmt.Println("Marshaled")
	fmt.Println(string(bytes))
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Publishing")
	if token := client.Publish(topic, 0, false, bytes); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return token.Error()
	}

	fmt.Println("Plugin debug publishing")
	if dToken := client.Publish("application/1/node/0004a30b001abe98/rx", 0, false, []byte("0101220501f20132ff4d")); dToken.Wait() && dToken.Error() != nil {
		fmt.Println(dToken.Error())
		return dToken.Error()
	}

	return nil
}

func join(client MQTT.Client, appKey [16]byte, appEUI, devEUI [8]byte, gwMac string, send bool) error {

	//Send a join only when set to true.
	if !send {
		return nil
	}

	joinPhy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinRequestPayload{
			AppEUI:   appEUI,
			DevEUI:   devEUI,
			DevNonce: [2]byte{byte(rand.Intn(255)), byte(rand.Intn(255))},
		},
	}

	if err := joinPhy.SetMIC(appKey); err != nil {
		panic(err)
	}

	joinStr, err := joinPhy.MarshalText()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(joinStr))

	fmt.Println("Printing MIC")
	fmt.Println(hex.EncodeToString(joinPhy.MIC[:]))

	message := createMessage(gwMac, joinStr)

	pErr := publish(client, "gateway/"+gwMac+"/rx", message)

	return pErr

}

func testMIC(appKey [16]byte, appEUI, devEUI [8]byte) error {
	joinPhy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinRequestPayload{
			AppEUI:   appEUI,
			DevEUI:   devEUI,
			DevNonce: [2]byte{byte(rand.Intn(255)), byte(rand.Intn(255))},
		},
	}

	if err := joinPhy.SetMIC(appKey); err != nil {
		panic(err)
	}

	fmt.Println("Printing MIC")
	fmt.Println(hex.EncodeToString(joinPhy.MIC[:]))

	joinStr, err := joinPhy.MarshalText()
	if err != nil {
		panic(err)
	}
	fmt.Println(joinStr)

	return nil
}

func sendMessage(client MQTT.Client, devAddr [4]byte, appSKey, nwkSKey [16]byte, gwMac string, payload []byte) error {

	fPort := uint8(1)

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.ConfirmedDataUp,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.MACPayload{
			FHDR: lorawan.FHDR{

				DevAddr: lorawan.DevAddr(devAddr),
				FCtrl: lorawan.FCtrl{
					ADR:       false,
					ADRACKReq: false,
					ACK:       false,
				},
				FCnt:  0,
				FOpts: []lorawan.MACCommand{}, // you can leave this out when there is no MAC command to send
			},
			FPort:      &fPort,
			FRMPayload: []lorawan.Payload{&lorawan.DataPayload{Bytes: payload}},
		},
	}

	if err := phy.EncryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	if err := phy.SetMIC(nwkSKey); err != nil {
		panic(err)
	}

	upDataStr, err := phy.MarshalText()
	if err != nil {
		panic(err)
	}

	message := createMessage(gwMac, upDataStr)

	pErr := publish(client, "gateway/"+gwMac+"/rx", message)

	return pErr

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
