package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"unicode/utf16"
)

type Translations struct {
	Translations []Translation `json:"translations"`
}

type Translation struct {
	Address    string      `json:"address"`
	Text       string      `json:"text"`
	Encoding   *string     `json:"encoding,omitempty"`
	References []Reference `json:"refs,omitempty"`
}

type TranslateCtx struct {
	buf          *bytes.Buffer
	dolData      []byte
	dolHeader    DolHeader
	translations *Translations
	newAddress   uint32
}

func NewTranslateCtx(dol []byte, trans []byte, newAddress uint32) (*TranslateCtx, error) {
	dolHeader, err := LoadDol(dol)
	if err != nil {
		return nil, err
	}

	var translations Translations
	err = json.Unmarshal(trans, &translations)
	if err != nil {
		return nil, err
	}

	return &TranslateCtx{
		buf:          new(bytes.Buffer),
		dolData:      dol,
		dolHeader:    dolHeader,
		translations: &translations,
		newAddress:   newAddress,
	}, nil
}

func (t *TranslateCtx) ApplyTranslations() error {
	for _, translation := range t.translations.Translations {
		err := t.ApplyTranslation(&translation)
		if err != nil {
			return err
		}
	}

	// We now append the strings in the new block to the dol and fix the header
	t.dolHeader.SetEmptyDataBlock(t.newAddress, uint32(len(t.dolData)), uint32(t.buf.Len()))
	t.dolData = append(t.dolData, t.buf.Bytes()...)

	return nil
}

func (t *TranslateCtx) ApplyUtf8Translation(translation *Translation) error {
	virtualAddress := translation.Address
	offset, err := t.dolHeader.GetAddressOffset(virtualAddress)
	if err != nil {
		return err
	}

	// UTF-8 / Shift-JIS strings have a singular byte as null terminator.
	// Read until the null terminator to determine the size of the string.
	size := 0
	workingOffset := offset
	for {
		if t.dolData[workingOffset] == 0 {
			break
		}

		size++
		workingOffset++
	}

	// If the translated text is larger than the original text then we must go into the new data block.
	if len(translation.Text) > size {
		// Before writing the string, we need to get the address at which it will be at in the Wii's memory.
		newAddress := t.newAddress + uint32(t.buf.Len())
		t.buf.WriteString(translation.Text)
		t.buf.WriteByte(0)

		// Now we need to replace occurrences of the old address with the new.
		oldAddress, err := ParseAddress(translation.Address)
		if err != nil {
			return err
		}

		err = t.ReplaceAllOccurrences(translation, oldAddress, newAddress)
		if err != nil {
			return err
		}
	} else {
		workingOffset = offset
		for _, c := range translation.Text {
			t.dolData[workingOffset] = byte(c)
			workingOffset++
		}

		// Nullify the leftover string
		for i := workingOffset; i < offset+uint32(size); i++ {
			t.dolData[i] = 0
		}
	}

	return nil
}

func (t *TranslateCtx) ApplyTranslation(translation *Translation) error {
	if translation.Encoding != nil && *translation.Encoding == "utf8" {
		err := t.ApplyUtf8Translation(translation)
		if err != nil {
			return err
		}

		return nil
	}

	virtualAddress := translation.Address
	offset, err := t.dolHeader.GetAddressOffset(virtualAddress)
	if err != nil {
		return err
	}

	// Read until the null terminator to determine the size of the string.
	size := 0
	workingOffset := offset
	for {
		if binary.BigEndian.Uint16(t.dolData[workingOffset:]) == 0 {
			break
		}

		size += 2
		workingOffset += 2
	}

	// If the translated text is larger than the original text then we must go into the new data block.
	utf16Encoded := utf16.Encode([]rune(translation.Text))
	if len(utf16Encoded)*2 > size {
		// Before writing the string, we need to get the address at which it will be at in the Wii's memory.
		newAddress := t.newAddress + uint32(t.buf.Len())

		for _, u := range utf16Encoded {
			err = binary.Write(t.buf, binary.BigEndian, u)
			if err != nil {
				return err
			}
		}

		// Ensure there is a null terminator
		if t.buf.Len()%4 == 0 {
			t.buf.Write(make([]byte, 4))
		}

		// Pad to 4 bytes
		for t.buf.Len()%4 != 0 {
			t.buf.WriteByte(0)
		}

		// Now we need to replace occurrences of the old address with the new.
		oldAddress, err := ParseAddress(translation.Address)
		if err != nil {
			return err
		}

		err = t.ReplaceAllOccurrences(translation, oldAddress, newAddress)
		if err != nil {
			return err
		}
	} else {
		workingOffset = offset
		for _, u := range utf16Encoded {
			binary.BigEndian.PutUint16(t.dolData[workingOffset:], u)
			workingOffset += 2
		}

		// Nullify the leftover string
		for i := workingOffset; i < offset+uint32(size); i++ {
			t.dolData[i] = 0
		}
	}

	return nil
}

func (t *TranslateCtx) ReplaceAllOccurrences(translation *Translation, toFind uint32, toReplace uint32) error {
	if translation.References != nil {
		for _, reference := range translation.References {
			if translation.Encoding != nil {
				err := t.EditUTF8Reference(reference, toReplace)
				if err != nil {
					return err
				}
			} else {
				err := t.EditUTF16Reference(reference, toReplace)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	for i := 0; i < len(t.dolData); i += 4 {
		if binary.BigEndian.Uint32(t.dolData[i:]) == toFind {
			binary.BigEndian.PutUint32(t.dolData[i:], toReplace)
		}
	}

	return nil
}

func (t *TranslateCtx) WriteTranslatedDol() error {
	// We first need to replace the old dol header with the new
	outBuf := new(bytes.Buffer)
	err := binary.Write(outBuf, binary.BigEndian, t.dolHeader)
	if err != nil {
		return err
	}

	// Append data to new header
	outBuf.Write(t.dolData[binary.Size(DolHeader{}):])

	// Now write the dol data
	return os.WriteFile("translated.dol", outBuf.Bytes(), 0666)
}
