package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"os"
)

const (
	AuxFieldRedisVer  = "redis-ver"
	AuxFieldRedisBits = "redis-bits"
	AuxFieldCtime     = "ctime"
	AuxFieldUsedMem   = "used-mem"
)

const (
	OPCodeEOF          = 0xFF
	OPCodeSELECTDB     = 0xFE
	OPCodeEXPIRETIME   = 0xFD
	OPCodeEXPIRETIMEMS = 0xFC
	OPCodeRESIZEDB     = 0xFB
	OPCodeAUX          = 0xFA
)

type RDB struct {
	// Magic
	MagicString [5]byte // 5 bytes SHOULD BE "REDIS"
	RDBVerNum   [4]byte // 4 bytes

	// Auxiliary field
	AuxField  map[string]string
	Databases []Database
}

type Database struct {
	ID       int
	ResizeDB struct {
		HashTableSize   int
		ExpireHashTable int
	}
	Fields []Field
}

type FieldType byte

const (
	FieldTypeString FieldType = 0
)

type Field struct {
	ExpiredTime uint64 // unix ms timestamp
	Type        FieldType
	Key         string
	Value       any
}

type StringValue string

func ParseV2(path string) RDB {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	r := bufio.NewReader(file)

	var rdb RDB
	rdb.AuxField = map[string]string{}

	_, err = r.Read(rdb.MagicString[:])
	if err != nil {
		log.Fatalln(err)
	}

	_, err = r.Read(rdb.RDBVerNum[:])
	if err != nil {
		log.Fatalln(err)
	}

	var curDBID int
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalln(err)
		}

		if b == OPCodeEOF {
			break
		}

		switch b {
		case OPCodeAUX:
			key, value, err := parseAux(r)
			if err != nil {
				log.Fatalln(err)
			}

			if isValidAuxKey(key) {
				rdb.AuxField[key] = value
			}

			continue
		case OPCodeSELECTDB:
			var db Database
			dbID, err := DecodeLength(r)
			if err != nil {
				log.Fatalln(err)
			}

			db.ID = dbID
			curDBID = dbID

			rdb.Databases = append(rdb.Databases, db)
			continue
		case OPCodeRESIZEDB:
			hashTableSize, err := DecodeLength(r)
			if err != nil {
				log.Fatalln(err)
			}
			rdb.Databases[curDBID].ResizeDB.HashTableSize = hashTableSize

			expireHashTableSize, err := DecodeLength(r)
			if err != nil {
				log.Fatalln(err)
			}
			rdb.Databases[curDBID].ResizeDB.ExpireHashTable = expireHashTableSize
			continue
		default:
			var f Field
			switch b {
			case OPCodeEXPIRETIME:
				var data uint32
				binary.Read(r, binary.LittleEndian, &data)
				f.ExpiredTime = uint64(data)
				b, err := r.ReadByte()
				if err != nil {
					log.Fatalln(err)
				}

				f.Type = FieldType(b)
			case OPCodeEXPIRETIMEMS:
				var data uint64
				binary.Read(r, binary.LittleEndian, &data)
				f.ExpiredTime = data
				b, err := r.ReadByte()
				if err != nil {
					log.Fatalln(err)
				}

				f.Type = FieldType(b)
			default:
				f.Type = FieldType(b)
			}

			key, err := DecodeString(r)
			if err != nil {
				log.Fatalln("asdf2", err)
			}

			f.Key = key

			switch f.Type {
			case FieldTypeString:
				val, err := DecodeString(r)
				if err != nil {
					log.Fatalln("asdf3", err)
				}
				f.Value = val
			}

			rdb.Databases[curDBID].Fields = append(rdb.Databases[curDBID].Fields, f)
		}
	}

	return rdb
}

func parseAux(r *bufio.Reader) (string, string, error) {
	var kv [2]string

	for i := 0; i < len(kv); i++ {
		str, err := DecodeString(r)
		if err != nil {
			log.Fatalln(err)
		}

		kv[i] = str
	}

	return kv[0], kv[1], nil
}

func isValidAuxKey(key string) bool {
	switch key {
	case AuxFieldRedisVer,
		AuxFieldRedisBits,
		AuxFieldCtime,
		AuxFieldUsedMem:
		return true
	}

	return false
}
