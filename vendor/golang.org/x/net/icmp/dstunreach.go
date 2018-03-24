// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icmp

// A DstUnreach represents an ICMP destination unreachable message
// body.
type DstUnreach struct {
	Data       []byte      // data, known as original datagram field
	Extensions []Extension // extensions
}

// Len implements the Len method of MessageBody interface.
func (p *DstUnreach) Len(proto int) int {
	if p == nil {
		return 0
	}
<<<<<<< HEAD
	l, _ := multipartMessageBodyDataLen(proto, p.Data, p.Extensions)
=======
	l, _ := multipartMessageBodyDataLen(proto, true, p.Data, p.Extensions)
>>>>>>> feat(matchers): add more matchers for more fun 🎉
	return 4 + l
}

// Marshal implements the Marshal method of MessageBody interface.
func (p *DstUnreach) Marshal(proto int) ([]byte, error) {
<<<<<<< HEAD
	return marshalMultipartMessageBody(proto, p.Data, p.Extensions)
=======
	return marshalMultipartMessageBody(proto, true, p.Data, p.Extensions)
>>>>>>> feat(matchers): add more matchers for more fun 🎉
}

// parseDstUnreach parses b as an ICMP destination unreachable message
// body.
<<<<<<< HEAD
func parseDstUnreach(proto int, b []byte) (MessageBody, error) {
=======
func parseDstUnreach(proto int, typ Type, b []byte) (MessageBody, error) {
>>>>>>> feat(matchers): add more matchers for more fun 🎉
	if len(b) < 4 {
		return nil, errMessageTooShort
	}
	p := &DstUnreach{}
	var err error
<<<<<<< HEAD
	p.Data, p.Extensions, err = parseMultipartMessageBody(proto, b)
=======
	p.Data, p.Extensions, err = parseMultipartMessageBody(proto, typ, b)
>>>>>>> feat(matchers): add more matchers for more fun 🎉
	if err != nil {
		return nil, err
	}
	return p, nil
}
