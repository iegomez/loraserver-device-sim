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

func main() {

	/*
	 *
	 * Create the mqtt client. Replace vlues where needed.
	 *
	 */
	opts := MQTT.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883")
	opts.SetUsername("lora")
	opts.SetPassword("lora")

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

	gwMac := "00800000a00006cd"

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
	 * Uncomment to perform an OTAA join request.
	 */

	join(client, appKey, appEUI, devEUI, gwMac, false)

	/*
	 * Check MIC
	 */

	/*var testDevEUI ([8]byte)
	  dEUI, _ := hex.DecodeString("0004a30b001a5ae1")
	  copy(testDevEUI[:], dEUI[:])

	  var testAppEUI ([8]byte)
	  aEUI, _ := hex.DecodeString("70b3d57ef0005ed5")
	  copy(testAppEUI[:], aEUI[:])

	  var testAppKey ([16]byte)
	  aKey, _ := hex.DecodeString("9bf8e134a8c6fadb661419759f953a98")
	  copy(testAppKey[:], aKey[:])

	  testMIC(testAppKey, testAppEUI, testDevEUI)*/

	/*
	 * Send a test message with an ABP activated node.
	 */
	var lat float32 = -33.4335625
	var lng float32 = -70.6217137

	for {
		byte0 := []byte{0} //Data to send.

		lat += rand.Float32() / 1000.0
		lng += rand.Float32() / 1000.0

		mPayload := append(byte0[:], generateTemp1byte(int8(rand.Intn(35)))[:]...)
		mPayload = append(mPayload[:], byte0[:]...)
		mPayload = append(mPayload[:], generateLat(lat)[:]...)
		mPayload = append(mPayload[:], generateLng(lng)[:]...)
		mPayload = append(mPayload[:], generateRisk(int8(rand.Intn(10)))[:]...)

		fmt.Println(mPayload)

		err := sendMessage(client, devAddr, appSKey, nwkSKey, gwMac, mPayload)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(5 * time.Second)
	}

}

func createMessage(gwMac string, payload []byte) *Message {

	/*
	 *
	 * Set correct value for your environment.
	 *
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
			MType: lorawan.UnconfirmedDataUp,
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
