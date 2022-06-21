package util

import (
	"bytes"

	"go.etcd.io/etcd/mvcc/mvccpb"
)

func UnitToSlice(kvs []*mvccpb.KeyValue) []byte {
	first := true
	var buffer bytes.Buffer
	buffer.WriteByte('[')
	for _, kv := range kvs {

		if !first {
			buffer.WriteByte(',')
		} else {
			first = false
		}
		buffer.Write(kv.Value)

	}
	buffer.WriteByte(']')
	return buffer.Bytes()
}

func ToMap(kvs []*mvccpb.KeyValue) []byte {
	//template := "{%s}"
	first := true
	var buffer bytes.Buffer
	buffer.WriteByte('{')
	for _, kv := range kvs {
		if !first {
			buffer.WriteByte(',')
		} else {
			first = false
		}
		buffer.WriteByte('"')
		buffer.Write(kv.Key[bytes.LastIndex(kv.Key, []byte{'/'})+1:])
		buffer.WriteByte('"')
		buffer.WriteByte(':')
		buffer.Write(kv.Value)
	}
	buffer.WriteByte('}')
	return buffer.Bytes()
}
