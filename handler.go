package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/smallnest/rpcx/protocol"
)

func output(key string, msg *protocol.Message) {
	if *withColor {
		fmt.Printf("%s %s %s\n\n", time.Now().Format("15:04:05.000"), key, renderWithColor(msg))
		return
	}

	fmt.Printf("%s %s %s\n", time.Now().Format("15:04:05.000"), key, render(msg))
}

func renderWithColor(msg *protocol.Message) string {
	version := msg.Version()
	messageType := ifelse(msg.MessageType() == protocol.Request, "Request", "Response")
	heartbeat := ifelse(msg.IsHeartbeat(), "true", "false")
	oneway := ifelse(msg.IsOneway(), "true", "false")
	compressType := ifelse(msg.CompressType() == protocol.None, "no", "yes")
	messageStatusType := ifelse(msg.MessageStatusType() == protocol.Normal, "normal", "error")
	serializeType := serializeType(msg.SerializeType())
	seq := strconv.FormatUint(msg.Seq(), 10)

	f := color.GreenString("version") + ": %d, " + color.GreenString("message_type") + ": %s, " + color.GreenString("heartbeat") +
		": %s, " + color.GreenString("oneway") + ": %s, " + color.GreenString("compress_type") + ": %s, " +
		color.GreenString("message_status_type") + ": %s, " + color.GreenString("serialize_type") +
		": %s, " + color.GreenString("seq") + ": %s\n"
	f += color.CyanString("service_path") + ": %s, " + color.CyanString("service_method") + ": %s, " + color.CyanString("metadata") + ": %v"
	if *outputPayload {
		if msg.SerializeType() == protocol.JSON {
			f += ", " + color.CyanString("payload") + ": %s"
		} else {
			f += ", " + color.CyanString("payload") + ": %v"
		}
		return fmt.Sprintf(f, version, messageType, heartbeat, oneway, compressType, messageStatusType, serializeType, seq,
			msg.ServicePath, msg.ServiceMethod, msg.Metadata, msg.Payload)
	}
	return fmt.Sprintf(f, version, messageType, heartbeat, oneway, compressType, messageStatusType, serializeType, seq,
		msg.ServicePath, msg.ServiceMethod, msg.Metadata)
}

func render(msg *protocol.Message) string {
	version := msg.Version()
	messageType := ifelse(msg.MessageType() == protocol.Request, "Request", "Response")
	heartbeat := ifelse(msg.IsHeartbeat(), "true", "false")
	oneway := ifelse(msg.IsOneway(), "true", "false")
	compressType := ifelse(msg.CompressType() == protocol.None, "no", "yes")
	messageStatusType := ifelse(msg.MessageStatusType() == protocol.Normal, "normal", "error")
	serializeType := serializeType(msg.SerializeType())
	seq := strconv.FormatUint(msg.Seq(), 10)

	f := "version: %d, message_type: %s, heartbeat: %s, oneway: %s, compress_type: %s, message_status_type: %s, serialize_type: %s, seq: %s\n"
	f += "service_path: %s, service_method: %s, metadata: %v"
	if *outputPayload {
		if msg.SerializeType() == protocol.JSON {
			f += ", payload: %s"
		} else {
			f += ", payload: %v"
		}
		return fmt.Sprintf(f, version, messageType, heartbeat, oneway, compressType, messageStatusType, serializeType, seq,
			msg.ServicePath, msg.ServiceMethod, msg.Metadata, msg.Payload)
	}
	return fmt.Sprintf(f, version, messageType, heartbeat, oneway, compressType, messageStatusType, serializeType, seq,
		msg.ServicePath, msg.ServiceMethod, msg.Metadata)
}

func ifelse(c bool, a, b string) string {
	if c {
		return a
	}

	return b
}

func serializeType(t protocol.SerializeType) string {
	switch t {
	case protocol.SerializeNone:
		return "none"
	case protocol.JSON:
		return "json"
	case protocol.ProtoBuffer:
		return "protobuffer"
	case protocol.MsgPack:
		return "msgpack"
	case protocol.Thrift:
		return "thrift"
	default:
		return strconv.Itoa(int(t))
	}
}
