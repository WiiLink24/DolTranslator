package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"
)

type DolHeader struct {
	TextOffsets      [7]uint32
	DataOffsets      [11]uint32
	TextAddresses    [7]uint32
	DataAddresses    [11]uint32
	TextSectionSizes [7]uint32
	DataSectionSizes [11]uint32
	BSSAddress       uint32
	BSSSize          uint32
	EntryPoint       uint32
}

func LoadDol(data []byte) (DolHeader, error) {
	var dol DolHeader
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &dol)
	return dol, err
}

func (d *DolHeader) SetEmptyDataBlock(address, offset, size uint32) {
	var emptyBlock int
	for i, _offset := range d.DataOffsets {
		if _offset == 0 {
			emptyBlock = i
			break
		}
	}

	// Set everything that is required
	d.DataOffsets[emptyBlock] = offset
	d.DataAddresses[emptyBlock] = address
	d.DataSectionSizes[emptyBlock] = size
}

// GetAddressOffset returns the converted offset that we can read from.
func (d *DolHeader) GetAddressOffset(virtualAddress string) (uint32, error) {
	_addr, err := ParseAddress(virtualAddress)
	if err != nil {
		return 0, err
	}

	return d.getAddressOffset(_addr)
}

func (d *DolHeader) getAddressOffset(addr uint32) (uint32, error) {
	for i := 0; i < 11; i++ {
		if d.DataAddresses[i] < addr && addr < d.DataAddresses[i]+d.DataSectionSizes[i] {
			return d.DataOffsets[i] + (addr - d.DataAddresses[i]), nil
		}
	}

	return 0, errors.New("invalid address")
}

func ParseAddress(addr string) (uint32, error) {
	_addr, err := strconv.ParseUint(addr[2:], 16, 64)
	if err != nil {
		return 0, err
	}

	return uint32(_addr), nil
}
