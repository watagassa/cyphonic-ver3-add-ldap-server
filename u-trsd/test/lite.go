package main

import (
	"encoding/hex"
	"log"

	"trs-test/entity"
)

func litetest() {
	var payload []byte
	commonKey, _ := hex.DecodeString("c7e7f58bb3c74429a462932926270a44")
	log.Println("common key: ", byteToHexString(commonKey), "length: ", len(commonKey))

	endKey, err := entity.GenerateCommonKey()
	if err != nil {
		log.Fatal("keyの生成に失敗しました: ", err)
	}
	log.Println("end key: ", byteToHexString(endKey), "length: ", len(endKey))

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

	log.Printf("payload: %s\n", byteToHexString(payload))

	encrypted, err := entity.EncryptPacket(payload, commonKey, 0, uint16(len(commonKey)))
	if err != nil {
		log.Fatal("暗号化に失敗しました: ", err)
	}

	log.Printf("payload: %s\n", byteToHexString(encrypted))

	decrypted, err := entity.DecryptPacket(encrypted, commonKey, 0, uint16(len(commonKey)))
	if err != nil {
		log.Fatal("復号に失敗しました: ", err)
	}
	log.Printf("payload: %s\n", byteToHexString(decrypted))

	if string(decrypted) != string(payload) {
		log.Fatal("復号結果が一致しません")
	} else {
		log.Println("復号結果が一致しました")
	}
}
