package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"

	"trs-test/entity"
)

func main() {
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "")
	}

	// 送信先のアドレスを指定
	// net.ResolveUDPAddrは、UDPのエンドポイントアドレスを解決します。
	// 第1引数にネットワーク種別("udp"), 第2引数にアドレスとポートを指定します。
	// DefaultPort 4503
	udpAddr, err := net.ResolveUDPAddr("udp", "10.0.3.102:4503")
	if err != nil {
		log.Fatalf("アドレスの解決に失敗しました: %v", err)
	}

	// UDPコネクションを作成
	// net.DialUDPは指定したアドレスに接続するためのUDPコネクションを作成します。
	// この時点ではパケットは送信されません。
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("コネクションの作成に失敗しました: %v", err)
	}
	// 関数終了時にコネクションを閉じる
	defer conn.Close()

	// 自分のアドレスとポートを表示
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	log.Printf("ローカルアドレス: %s\n", localAddr.String())

	var payload []byte
	commonKey, _ := hex.DecodeString("c7e7f58bb3c74429a462932926270a44")

	endKey, err := entity.GenerateCommonKey()
	if err != nil {
		log.Fatal("keyの生成に失敗しました: ", err)
	}
	log.Println("生成したend key: ", string(endKey), "長さ: ", len(endKey))

	switch os.Args[1] {
	case "in":
		// TunnelRequest(From Initiator)
		payload = []byte{
			//  End Key Length(16), Padding(16)
			0x00, 0x00, 0x00, 0x00,
			// End Key(Variable) 32Bytes
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
		}

		copy(payload[4:36], endKey)
		endKeyLength := len(endKey)
		payload[0] = byte(endKeyLength >> 8)   // 上位バイト
		payload[1] = byte(endKeyLength & 0xFF) // 下位バイト

		payloadLength := len(payload)

		payload, err = entity.EncryptPacket(payload, commonKey, 0, uint16(len(commonKey)))
		if err != nil {
			log.Fatal("暗号化に失敗しました: ", err)
		}

		log.Println("InitiatorとしてTunnelRequestを送信します")
		SendPacket(payload, "sample1", 16, conn, udpAddr, payloadLength)
		fmt.Println("--------------------------------")

		// Receive TunnelResponse(From Responder)
		log.Println("受信を待機")
		ReceivePacket(conn)
		fmt.Println("--------------------------------")

		// Send Capsule Message(From Initiator)
		payload = []byte("Hello, this is a Capsule Message from Initiator.")
		log.Println("InitiatorとしてCapsule Messageを送信します")
		SendPacket(payload, "sample1", 20, conn, udpAddr, len(payload))
		fmt.Println("--------------------------------")

		// Receive Capsule Message
		log.Println("受信を待機")
		ReceivePacket(conn)
		fmt.Println("--------------------------------")

	case "rn":
		// HolePunching(From Responder)
		log.Println("ResponderとしてUDP Hole Punchingを送信します")
		SendPacket(payload, "sample1", 3, conn, udpAddr, len(payload))
		fmt.Println("--------------------------------")

		// Receive TunnelRequest(From Initiator)
		log.Println("受信を待機")
		ReceivePacket(conn)
		fmt.Println("--------------------------------")

		// TunnelResponse(From Responder)
		log.Println("ResponderとしてTunnelResponseを送信します")
		SendPacket(payload, "sample1", 17, conn, udpAddr, len(payload))
		fmt.Println("--------------------------------")

		// Receive Capsule Message
		log.Println("受信を待機")
		ReceivePacket(conn)
		fmt.Println("--------------------------------")

		// Send Capsule Message(From Responder)
		payload = []byte("Hello, this is a Capsule Message from Responder.")
		log.Println("ResponderとしてCapsule Messageを送信します")
		SendPacket(payload, "sample1", 21, conn, udpAddr, len(payload))
		fmt.Println("--------------------------------")

	default:
		litetest()
	}
}

func ReceivePacket(conn *net.UDPConn) ([]byte, *net.UDPAddr, error) {
	buffer := make([]byte, 1024) // 受信用のバッファを作成
	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		log.Printf("メッセージの受信に失敗しました: %v", err)
		return nil, nil, err
	}

	receivedMessage := buffer[:n]
	log.Printf("'%s' からメッセージを受信しました。\n受信データ (byte配列): %v\n受信データ (String): %s\n", addr.String(), receivedMessage, string(receivedMessage))
	return receivedMessage, addr, nil
}

func SendPacket(payload []byte, pathID string, packetType int, conn *net.UDPConn, addr *net.UDPAddr, payloadLength int) error {
	message := []byte{
		// TransactionID
		0x00, 0x00, 0x00, 0x00,
		// Version(8), Type(8), Status(8), Count(8)
		0x00, 0x00, 0x00, 0x00,
		// Sequence Number(32)
		0x00, 0x00, 0x00, 0x00,
		// Message Length(16), Next Opt(8), Token(8)
		0x00, 0x00, 0x00, 0x00,
		// ID(128)
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	// IDの部分にコピー
	pathIDBytes := []byte(pathID)
	copy(message[16:32], pathIDBytes)

	// Typeの部分に設定
	packetTypeByte := byte(packetType)
	message[5] = packetTypeByte

	hmacPayload := []byte{
		// HMAC(128)
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
	payload = append(payload, hmacPayload...)

	// メッセージの長さを設定
	messageLength := payloadLength
	message[12] = byte(messageLength >> 8)   // 上位バイト
	message[13] = byte(messageLength & 0xFF) // 下位バイト

	// payloadをメッセージに追加
	message = append(message, payload...)

	// メッセージを送信
	_, err := conn.Write(message)
	if err != nil {
		log.Printf("メッセージの送信に失敗しました: %v", err)
		return err
	}

	log.Printf("'%s' 宛にメッセージを送信しました。\n送信データ (byte配列): %v\n送信データ (String): %s\n", addr.String(), message, string(message))
	return nil
}

func byteToHexString(data []byte) string {
	hexStr := ""
	for _, b := range data {
		hexStr += string("0123456789abcdef"[b>>4])
		hexStr += string("0123456789abcdef"[b&0x0F])
	}
	return hexStr
}
