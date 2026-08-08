package binary

import (
	"log"
)

type Message struct {
	PacketTimestampMs int64
	PacketSenderId    uint64
	Malformed         bool
	MessageId         string
	Content           string
	SenderNickname    string
	Channel           string
	reply_id          uint32
	message_flags     uint8
	LatitudeI         int32
	LongitudeI        int32
	Altitude          int32
}

const (
	type_unknown  = 0
	type_announce = 0x01
	type_message  = 0x02
)

const (
	packet_flag_has_recipient = 1 << 0
	packet_flag_has_signature = 1 << 1
	packet_flag_is_compressed = 1 << 2
	packet_flag_has_route     = 1 << 3

	// Bump up to avoid overlap with future bitchat flags
	packet_flag_from_meshtastic = 1 << 7
)

const (
	tlv_message_id      = 0x00
	tlv_message_content = 0x01

	// bump up values to avoid clash with upstream bitchat - if they adopt compatible concepts then we can migrate and use their types
	// >> Location - messages will usually not include a location, but might occasionally want to
	tlv_message_location_latitude  = 0x78 // 32bit lat*10000000
	tlv_message_location_longitude = 0x79 // 32bit lon*10000000
	tlv_message_location_altitude  = 0x7a // height above sea level in meters
	// << Location
	tlv_message_flags           = 0x7b // using flags for emoji support in meshtastic messages
	tlv_message_reply_to_id     = 0x7c // indicates that the message is a reply to a former message
	tlv_message_sender_nickname  = 0x7d // it's useful for a Channel to have direct access to the name associated with the message
	tlv_message_channel_content = 0x7e // to avoid regular bitchat clients showing our Channel messages to everyone we omit theregular content
	tlv_message_channel         = 0x7f // the #Channel name
)

func (p *Message) setMalformed(m bool) {
	p.Malformed = m
}

func (parent *Message) ReadIn(reader *BinaryReader) {
	version := reader.read_uint8()
	log.Print("version:", version)
	packet_type := reader.read_uint8()
	log.Print("packet_type:", packet_type)
	reader.read_uint8() // ignore packet_ttl
	parent.PacketTimestampMs = reader.read_int64()
	log.Print("PacketTimestampMs: ", parent.PacketTimestampMs)
	packet_flags := reader.read_uint8()
	log.Print("packet_flags: ", packet_flags)

	var payload_length uint32
	if version == 1 {
		payload_length = uint32(reader.read_uint16())
	} else if version == 2 {
		payload_length = reader.read_uint32()
	}
	log.Print("payload_length: ", payload_length)
	parent.PacketSenderId = reader.read_uint64()
	log.Print("PacketSenderId: ", parent.PacketSenderId)
	if (packet_flags & packet_flag_has_recipient) > 0 {
		packet_recipient_id := reader.read_uint64() // ignore packet_recipient_id
		log.Print("packet_recipient_id: ", packet_recipient_id)
	}
	if version == 2 && (packet_flags&packet_flag_has_route) > 0 {
		routeCount := reader.read_uint8()
		for i := uint8(0); i < routeCount; i++ {
			reader.read_uint64() // ignore hop_id
		}
	}

	parent.setMalformed(false)

	tail := reader.read_remainder_len()
	remainder := reader.read_remainder_len()
	log.Print("pos: ", reader.test_only_current_pos())
	log.Print("remainder: ", remainder)
	log.Print("tail: ", tail)
	for parent.Malformed == false && remainder > tail-payload_length {
		data_type := reader.read_uint8()
		data_length := reader.read_uint8()
		if data_type == tlv_message_flags {
			if data_length == 8 {
				parent.message_flags = reader.read_uint8()
			} else {
				parent.setMalformed(true)
			}
		} else if data_type == tlv_message_reply_to_id {
			if data_length == 32 {
				parent.reply_id = reader.read_uint32()
			} else {
				parent.setMalformed(true)
			}
		} else if data_type == tlv_message_location_latitude {
			if data_length == 32 {
				parent.LatitudeI = reader.read_int32()
			} else {
				parent.setMalformed(true)
			}
		} else if data_type == tlv_message_location_longitude {
			if data_length == 32 {
				parent.LongitudeI = reader.read_int32()
			} else {
				parent.setMalformed(true)
			}
		} else if data_type == tlv_message_location_altitude {
			if data_length == 32 {
				parent.Altitude = reader.read_int32()
			} else {
				parent.setMalformed(true)
			}
		} else {
			data := reader.read_data(uint32(data_length))
			if len(data) > 0 {
				switch data_type {
				case tlv_message_id:
					parent.MessageId = string(data)
					break
				case tlv_message_content:
					parent.Content = string(data)
					break
				case tlv_message_sender_nickname:
					parent.SenderNickname = string(data)
					break
				case tlv_message_channel:
					parent.Channel = string(data)
					break
				case tlv_message_channel_content:
					parent.Content = string(data)
					break
				default:
					break
				}
			} else {
				parent.setMalformed(true)
			}
		}

		remainder = reader.read_remainder_len()
	}
	if parent.Malformed == false && (packet_flags&packet_flag_has_signature) > 0 {
		sig_data := reader.read_data(64) // SIGNATURE_LENGTH=64
		if len(sig_data) != 64 {
			parent.setMalformed(true)
		}
	}
}
