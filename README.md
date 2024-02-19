# DolTranslator
Inserts strings into a Nintendo Wii/GC DOL.

## Usage
You will need a Nintendo Wii/GC DOL which will be translated, as well as a JSON file similar to the following.
```json
{
  "translations": [
    {
      "address": "0x802d7740",
      "text": "This will be inserted!"
    }
  ]
}
```
The `address` key is the memory address where this string can be found. You can find this in Ghidra or in Dolphin Emulator Memory view.

You will also need to designate a new region of memory to place translations in the case that an inserted string overflows the original.

Running from CMD:
```
./DolTranslator [path/to/dol] [path/to/json] [new address]
```