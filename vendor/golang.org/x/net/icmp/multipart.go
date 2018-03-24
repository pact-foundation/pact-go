// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icmp

import "golang.org/x/net/internal/iana"

// multipartMessageBodyDataLen takes b as an original datagram and
// exts as extensions, and returns a required length for message body
// and a required length for a padded original datagram in wire
// format.
<<<<<<< HEAD
func multipartMessageBodyDataLen(proto int, b []byte, exts []Extension) (bodyLen, dataLen int) {
=======
func multipartMessageBodyDataLen(proto int, withOrigDgram bool, b []byte, exts []Extension) (bodyLen, dataLen int) {
>>>>>>> feat(matchers): add more matchers for more fun 🎉
	for _, ext := range exts {
		bodyLen += ext.Len(proto)
	}
	if bodyLen > 0 {
<<<<<<< HEAD
		dataLen = multipartMessageOrigDatagramLen(proto, b)
=======
		if withOrigDgram {
			dataLen = multipartMessageOrigDatagramLen(proto, b)
		}
>>>>>>> feat(matchers): add more matchers for more fun 🎉
		bodyLen += 4 // length of extension header
	} else {
		dataLen = len(b)
	}
	bodyLen += dataLen
	return bodyLen, dataLen
}

// multipartMessageOrigDatagramLen takes b as an original datagram,
// and returns a required length for a padded orignal datagram in wire
// format.
func multipartMessageOrigDatagramLen(proto int, b []byte) int {
	roundup := func(b []byte, align int) int {
		// According to RFC 4884, the padded original datagram
		// field must contain at least 128 octets.
		if len(b) < 128 {
			return 128
		}
		r := len(b)
		return (r + align - 1) & ^(align - 1)
	}
	switch proto {
	case iana.ProtocolICMP:
		return roundup(b, 4)
	case iana.ProtocolIPv6ICMP:
		return roundup(b, 8)
	default:
		return len(b)
	}
}

// marshalMultipartMessageBody takes data as an original datagram and
// exts as extesnsions, and returns a binary encoding of message body.
// It can be used for non-multipart message bodies when exts is nil.
<<<<<<< HEAD
func marshalMultipartMessageBody(proto int, data []byte, exts []Extension) ([]byte, error) {
	bodyLen, dataLen := multipartMessageBodyDataLen(proto, data, exts)
=======
func marshalMultipartMessageBody(proto int, withOrigDgram bool, data []byte, exts []Extension) ([]byte, error) {
	bodyLen, dataLen := multipartMessageBodyDataLen(proto, withOrigDgram, data, exts)
>>>>>>> feat(matchers): add more matchers for more fun 🎉
	b := make([]byte, 4+bodyLen)
	copy(b[4:], data)
	off := dataLen + 4
	if len(exts) > 0 {
		b[dataLen+4] = byte(extensionVersion << 4)
		off += 4 // length of object header
		for _, ext := range exts {
			switch ext := ext.(type) {
			case *MPLSLabelStack:
				if err := ext.marshal(proto, b[off:]); err != nil {
					return nil, err
				}
				off += ext.Len(proto)
			case *InterfaceInfo:
				attrs, l := ext.attrsAndLen(proto)
				if err := ext.marshal(proto, b[off:], attrs, l); err != nil {
					return nil, err
				}
				off += ext.Len(proto)
<<<<<<< HEAD
=======
			case *InterfaceIdent:
				if err := ext.marshal(proto, b[off:]); err != nil {
					return nil, err
				}
				off += ext.Len(proto)
>>>>>>> feat(matchers): add more matchers for more fun 🎉
			}
		}
		s := checksum(b[dataLen+4:])
		b[dataLen+4+2] ^= byte(s)
		b[dataLen+4+3] ^= byte(s >> 8)
<<<<<<< HEAD
		switch proto {
		case iana.ProtocolICMP:
			b[1] = byte(dataLen / 4)
		case iana.ProtocolIPv6ICMP:
			b[0] = byte(dataLen / 8)
=======
		if withOrigDgram {
			switch proto {
			case iana.ProtocolICMP:
				b[1] = byte(dataLen / 4)
			case iana.ProtocolIPv6ICMP:
				b[0] = byte(dataLen / 8)
			}
>>>>>>> feat(matchers): add more matchers for more fun 🎉
		}
	}
	return b, nil
}

// parseMultipartMessageBody parses b as either a non-multipart
// message body or a multipart message body.
<<<<<<< HEAD
func parseMultipartMessageBody(proto int, b []byte) ([]byte, []Extension, error) {
=======
func parseMultipartMessageBody(proto int, typ Type, b []byte) ([]byte, []Extension, error) {
>>>>>>> feat(matchers): add more matchers for more fun 🎉
	var l int
	switch proto {
	case iana.ProtocolICMP:
		l = 4 * int(b[1])
	case iana.ProtocolIPv6ICMP:
		l = 8 * int(b[0])
	}
	if len(b) == 4 {
		return nil, nil, nil
	}
<<<<<<< HEAD
	exts, l, err := parseExtensions(b[4:], l)
	if err != nil {
		l = len(b) - 4
	}
	data := make([]byte, l)
	copy(data, b[4:])
=======
	exts, l, err := parseExtensions(typ, b[4:], l)
	if err != nil {
		l = len(b) - 4
	}
	var data []byte
	if l > 0 {
		data = make([]byte, l)
		copy(data, b[4:])
	}
>>>>>>> feat(matchers): add more matchers for more fun 🎉
	return data, exts, nil
}
