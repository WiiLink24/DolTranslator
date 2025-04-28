package main

import "github.com/wii-tools/powerpc"

/*
The Demae Channel is dumb. Text is all over the place. BRLYT's, TPL's, and inside the DOL which is the whole reason
this tool exists. Inside the DOL there are many different ways a string is loaded. There are pointers which are the
most common. However, for some reason some strings are directly referenced by sprintf. We need to adjust the instruction
which calculates the address of the string.

TODO: This is extremely naive. Fix!
*/

type Reference struct {
	Instructions []Instruction `json:"instructions,omitempty"`
}

type Instruction struct {
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
}

func (t *TranslateCtx) EditUTF8Reference(reference Reference, targetAddress uint32) error {
	// Go sequentially down the list of instructions until we are able to produce a set that
	// results in our address
	var currAddress uint32

	for _, instruction := range reference.Instructions {
		if instruction.Mnemonic == "lis" {
			// Lower 16 bits results is what we need.
			toUse := targetAddress >> 16
			iN := powerpc.LIS(powerpc.R4, uint16(toUse))

			iAddress, err := t.dolHeader.GetAddressOffset(instruction.Address)
			if err != nil {
				return err
			}

			for i := 0; i < 4; i++ {
				t.dolData[int(iAddress)+i] = iN[i]
			}

			currAddress = toUse << 16
		} else if instruction.Mnemonic == "addi" {
			iAddress, err := t.dolHeader.GetAddressOffset(instruction.Address)
			if err != nil {
				return err
			}

			if currAddress == targetAddress {
				// Assign a NOP
				iN := powerpc.NOP()

				for i := 0; i < 4; i++ {
					t.dolData[int(iAddress)+i] = iN[i]
				}
				continue
			}

			remaining := targetAddress - currAddress
			iN := powerpc.ADDI(powerpc.R4, powerpc.R4, uint16(remaining))

			for i := 0; i < 4; i++ {
				t.dolData[int(iAddress)+i] = iN[i]
			}

			currAddress += remaining
		}
	}

	return nil
}

func (t *TranslateCtx) EditUTF16Reference(reference Reference, targetAddress uint32) error {
	// EditUTF8Reference but it uses wsnprintf so the string is loaded into r5
	var currAddress uint32

	for _, instruction := range reference.Instructions {
		if instruction.Mnemonic == "lis" {
			// Lower 16 bits results is what we need.
			toUse := targetAddress >> 16
			iN := powerpc.LIS(powerpc.R5, uint16(toUse))

			iAddress, err := t.dolHeader.GetAddressOffset(instruction.Address)
			if err != nil {
				return err
			}

			for i := 0; i < 4; i++ {
				t.dolData[int(iAddress)+i] = iN[i]
			}

			currAddress = toUse << 16
		} else if instruction.Mnemonic == "addi" {
			iAddress, err := t.dolHeader.GetAddressOffset(instruction.Address)
			if err != nil {
				return err
			}

			if currAddress == targetAddress {
				// Assign a NOP
				iN := powerpc.NOP()

				for i := 0; i < 4; i++ {
					t.dolData[int(iAddress)+i] = iN[i]
				}
				continue
			}

			remaining := targetAddress - currAddress
			iN := powerpc.ADDI(powerpc.R5, powerpc.R5, uint16(remaining))

			for i := 0; i < 4; i++ {
				t.dolData[int(iAddress)+i] = iN[i]
			}

			currAddress += remaining
		}
	}

	return nil
}
