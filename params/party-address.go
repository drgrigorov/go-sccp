// Copyright 2019-2023 go-sccp authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package params

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/wmnsk/go-sccp/utils"
)

// PartyAddress is a SCCP parameter that represents a Called/Calling Party Address.
type PartyAddress struct {
	Length             uint8
	Indicator          uint8
	SignalingPointCode uint16
	SubsystemNumber    uint8
	GlobalTitle
}

// GlobalTitle is a GlobalTitle inside the Called/Calling Party Address.
type GlobalTitle struct {
	TranslationType          uint8
	NumberingPlan            int // 1/2 Octet
	EncodingScheme           int // 1/2 Octet
	NatureOfAddressIndicator uint8
	GlobalTitleInfo          []byte
}

// NewPartyAddress creates a new PartyAddress including GlobalTitle.
func NewPartyAddress(gti, spc, ssn, tt, np, es, nai int, gt []byte) *PartyAddress {
	p := &PartyAddress{
		Indicator:          uint8(gti),
		SignalingPointCode: uint16(spc),
		SubsystemNumber:    uint8(ssn),
		GlobalTitle: GlobalTitle{
			TranslationType:          uint8(tt),
			NumberingPlan:            np,
			EncodingScheme:           es,
			NatureOfAddressIndicator: uint8(nai),
			GlobalTitleInfo:          gt,
		},
	}
	p.Length = uint8(p.MarshalLen() - 1)
	return p
}

// MarshalBinary returns the byte sequence generated from a PartyAddress instance.
func (p *PartyAddress) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.MarshalLen())
	if err := p.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
func (p *PartyAddress) MarshalTo(b []byte) error {
	b[0] = p.Length
	b[1] = p.Indicator
	var offset = 2
	if p.HasPC() {
		binary.BigEndian.PutUint16(b[offset:offset+2], p.SignalingPointCode)
		offset += 2
	}
	if p.HasSSN() {
		b[offset] = p.SubsystemNumber
		offset++
	}

	switch p.GTI() {
	case 1:
		b[offset] = p.NatureOfAddressIndicator
		offset++
	case 2:
		b[offset] = p.TranslationType
		offset++
	case 3:
		b[offset] = p.TranslationType
		b[offset+1] = uint8(p.NumberingPlan<<4 | p.EncodingScheme)
		offset += 2
	case 4:
		b[offset] = p.TranslationType
		b[offset+1] = uint8(p.NumberingPlan<<4 | p.EncodingScheme)
		b[offset+2] = p.NatureOfAddressIndicator
		offset += 3
	}

	copy(b[offset:p.MarshalLen()], p.GlobalTitleInfo)
	return nil
}

// ParsePartyAddress decodes given byte sequence as a SCCP common header.
func ParsePartyAddress(b []byte) (*PartyAddress, error) {
	p := new(PartyAddress)
	if err := p.UnmarshalBinary(b); err != nil {
		return nil, err
	}

	return p, nil
}

// UnmarshalBinary sets the values retrieved from byte sequence in a SCCP common header.
func (p *PartyAddress) UnmarshalBinary(b []byte) error {
	if len(b) < 3 {
		return io.ErrUnexpectedEOF
	}
	p.Length = b[0]
	if int(p.Length) >= len(b) {
		return io.ErrUnexpectedEOF
	}
	p.Indicator = b[1]

	var offset = 2
	if p.HasPC() {
		end := offset + 2
		if end >= len(b) {
			return io.ErrUnexpectedEOF
		}
		p.SignalingPointCode = binary.BigEndian.Uint16(b[offset:end])
		offset = end
	}
	if p.HasSSN() {
		p.SubsystemNumber = b[offset]
		offset++
		if offset >= len(b) {
			return io.ErrUnexpectedEOF
		}
	}

	switch p.GTI() {
	case 1:
		p.NatureOfAddressIndicator = b[offset]
		offset++
	case 2:
		p.TranslationType = b[offset]
		offset++
	case 3:
		p.TranslationType = b[offset]
		offset++
		if offset >= len(b) {
			return io.ErrUnexpectedEOF
		}
		p.NumberingPlan = int(b[offset]) >> 4 & 0xf
		p.EncodingScheme = int(b[offset]) & 0xf
		offset++
	case 4:
		p.TranslationType = b[offset]
		offset++
		if offset+1 >= len(b) {
			return io.ErrUnexpectedEOF
		}
		p.NumberingPlan = int(b[offset]) >> 4 & 0xf
		p.EncodingScheme = int(b[offset]) & 0xf
		offset++
		p.NatureOfAddressIndicator = b[offset]
		offset++
	}

	infoLen := 1 + int(p.Length) - offset
	if infoLen < 0 {
		return errors.New("sccp: party address length misfit")
	}
	p.GlobalTitleInfo = make([]byte, infoLen)
	copy(p.GlobalTitleInfo, b[offset:])

	return nil
}

// MarshalLen returns the serial length.
func (p *PartyAddress) MarshalLen() int {
	l := 2 + len(p.GlobalTitleInfo)
	if p.HasPC() {
		l += 2
	}
	if p.HasSSN() {
		l++
	}
	switch p.GTI() {
	case 1:
		l++
	case 2:
		l++
	case 3:
		l += 2
	case 4:
		l += 3
	}

	return l
}

// SetLength sets the length in Length field.
func (p *PartyAddress) SetLength() {
	l := 1 + len(p.GlobalTitleInfo)
	if p.HasPC() {
		l += 2
	}
	if p.HasSSN() {
		l++
	}
	switch p.GTI() {
	case 1:
		l++
	case 2:
		l++
	case 3:
		l += 2
	case 4:
		l += 3
	}

	p.Length = uint8(l)
}

// RouteOnGT reports whether the packet is routed on Global Title or not.
func (p *PartyAddress) RouteOnGT() bool {
	return (int(p.Indicator) >> 6 & 0x1) == 0
}

// GTI returns GlobalTitleIndicator value retrieved from Indicator.
func (p *PartyAddress) GTI() int {
	return (int(p.Indicator) >> 2 & 0xf)
}

// HasSSN reports whether PartyAddress has a Subsystem Number.
func (p *PartyAddress) HasSSN() bool {
	return (int(p.Indicator) >> 1 & 0x1) == 1
}

// HasPC reports whether PartyAddress has a Signaling Point Code.
func (p *PartyAddress) HasPC() bool {
	return (int(p.Indicator) & 0x1) == 1
}

// IsOddDigits reports whether GlobalTitleInfo is odd number or not.
func (p *PartyAddress) IsOddDigits() bool {
	return p.EncodingScheme == 1
}

// GTString returns the GlobalTitleInfo in human readable string.
func (p *PartyAddress) GTString() string {
	if len( p.GlobalTitleInfo > 0 ) {
		return utils.SwappedBytesToStr(p.GlobalTitleInfo, p.IsOddDigits())
	}
	return ""
}
