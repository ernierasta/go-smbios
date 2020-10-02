package hwinfo

import (
	"encoding/binary"
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

const (
	// headerLen is the length of the Header structure.
	headerLen = 4

	// typeEndOfTable indicates the end of a stream of Structures.
	typeEndOfTable = 127

	// sizes of data structure determines layout and
	// and SMBIOS spec version
	sizeAsPer2_1 = 17
	sizeAsPer2_3 = 23
	sizeAsPer2_6 = 24
	sizeAsPer2_7 = 30
	sizeAsPer2_8 = 36
	sizeAsPer3_2 = 80
	sizeAsPer3_3 = 88
)

type MemoryFormFactor int

func (mff MemoryFormFactor) String() string {
	return [...]string{
		"Undefined",
		"Other",
		"Unknown",
		"SIMM",
		"SIP",
		"Chip",
		"DIP",
		"ZIP",
		"Proprietary Card",
		"DIMM",
		"TSOP",
		"RowOfChips",
		"RIMM0",
		"SODIMM",
		"SRIMM",
		"FBDIMM",
		"Die",
	}[mff]
}

type MemoryType int

func (mt MemoryType) String() string {
	return [...]string{
		"Undefined",
		"Other",
		"Unknown",
		"DRAM",
		"EDRAM",
		"VRAM",
		"SRAM",
		"RAM",
		"ROM",
		"FLASH",
		"EEPROM",
		"FEPROM",
		"EPROM",
		"CDRAM",
		"3DRAM",
		"SDRAM",
		"SGRAM",
		"RDRAM",
		"DDR",
		"DDR2",
		"DDR2 FB-DIMM",
		"Reserved",
		"Reserved",
		"Reserved",
		"DDR3",
		"FBD2",
		"DDR4",
		"LPDDR",
		"LPDDR2",
		"LPDDR3",
		"LPDDR4",
		"Logical non-volatile device",
		"HBM (High Bandwidth Memory)",
		"HBM2 (High Bandwidth Memory Generation 2)",
		"DDR5",
		"LPDDR5",
	}[mt]
}

type MemoryTypeDetail int

func (mtd MemoryTypeDetail) String() string {
	return [...]string{
		"Reserved",
		"Other",
		"Unknown",
		"Fast-paged",
		"Static column",
		"Pseudo-static",
		"RAMBUS",
		"Synchronous",
		"CMOS",
		"EDO",
		"Window DRAM",
		"Cache DRAM",
		"Non-volatile",
		"Registered (Buffered)",
		"Unbuffered (Unregistered)",
		"LRDIMM",
	}[mtd]
}

// TODO: move that to separate file
type BitField []byte

// NewFromUint16 returns a new BitField of 4 bytes, with n initial value
func BitFieldFromUint16(n uint16) BitField {
	b := BitField(make([]byte, 2))
	b[0] = byte(n)
	b[1] = byte(n >> 8)
	return b
}

// Test returns true/false on bit i value
func (b BitField) Test(i uint32) bool {
	idx, offset := (i / 8), (i % 8)
	return (b[idx] & (1 << uint(offset))) != 0
}

type MemoryRaw struct {
	MemArrayHandle                          uint16
	MemErrorInfoHandle                      uint16
	TotalWidth                              uint16
	DataWidth                               uint16
	Size                                    uint16
	FormFactor                              byte
	DeviceSet                               byte
	DeviceLocator                           byte
	BankLocator                             byte
	MemType                                 byte
	TypeDetail                              uint16
	Speed                                   uint16
	Manufacturer                            byte
	SerialNumber                            byte
	AssetTag                                byte
	PartNumber                              byte
	Attribute                               byte
	ExtendedSize                            uint32
	ConfiguredMemClockSpeed                 uint16
	MinVoltage                              uint16
	MaxVoltage                              uint16
	ConfiguredVoltage                       uint16
	MemoryTechnology                        byte
	MemoryOperatingModeCapability           uint16
	FirmwareVersion                         byte
	ModuleManufacturerID                    uint16
	ModuleProductID                         uint16
	MemorySubsystemControllerManufacturerID uint16
	MemorySubsystemControllerProductID      uint16
	NonVolatileSize                         uint64
	VolatileSize                            uint64
	CacheSize                               uint64
	LogicalSize                             uint64
	ExtendedSpeed                           uint32
	ExtendedConfiguredMemorySpeed           uint32
}

type Memory struct {
	// TotalWidth is in bits
	TotalWidth uint16
	// DataWidth is in bits
	DataWidth uint16
	// SizeKB ram size in KB
	SizeKB uint64
	// SizeMB the same size but in MB
	SizeMB        uint64
	FormFactor    string
	DeviceSet     byte
	DeviceLocator string
	BankLocator   string
	MemType       string
	TypeDetail    []string
	Speed         uint32
	Manufacturer  string
	SerialNumber  string
	AssetTag      string
}

func GetMemory(fb []byte, ss []string) error {
	var mem Memory

	// guards
	if len(fb) < sizeAsPer2_1 {
		return fmt.Errorf("formated byte array is too short, len: %d", len(fb))
	}

	raw := getMemoryRaw(fb, ss)
	arrSize := byte(len(ss))

	// Copy data width
	if raw.TotalWidth > 0 {
		mem.TotalWidth = raw.TotalWidth
	}
	if raw.DataWidth > 0 {
		mem.DataWidth = raw.DataWidth
	}

	// memory size
	memSize := uint64(raw.Size)
	//If the raw.Size size is 32GB or greater, we need to parse the extended field.
	// Spec says 0x7fff in regular size field means we should parse the extended.
	if raw.Size == 0x7FFF && len(fb) >= sizeAsPer2_7 {
		memSize = uint64(raw.ExtendedSize)
	}
	// The granularity in which the value is specified
	// depends on the setting of the most-significant bit (bit
	// 15). If the bit is 0, the value is specified in megabyte
	// units; if the bit is 1, the value is specified in kilobyte
	// units.
	if fb[9]&0x80 == 0 { // MB
		mem.SizeKB = memSize * 1024
		mem.SizeMB = memSize
	} else { // kB
		mem.SizeKB = memSize
		mem.SizeMB = mem.SizeKB / 1024
	}

	// Check for the size, if size is zero that means Empty DIMM Slot
	// TODO: what with that? should we return something?
	if mem.SizeKB != 0 {
		//systemInfo.PhyMemory = append(systemInfo.PhyMemory, mem)
	}

	mem.FormFactor = MemoryFormFactor(raw.FormFactor).String()

	mem.DeviceSet = raw.DeviceSet

	if raw.DeviceLocator > 0 {
		mem.DeviceLocator = ss[raw.DeviceLocator-1]
	}

	if raw.BankLocator > 0 {
		mem.BankLocator = ss[raw.BankLocator-1]
	}

	mem.MemType = MemoryType(raw.MemType).String()

	// TypeDetail slice
	detailBitField := BitFieldFromUint16(raw.TypeDetail)
	for i := 0; i <= 15; i++ {
		if detailBitField.Test(uint32(i)) {
			mem.TypeDetail = append(mem.TypeDetail, MemoryTypeDetail(i).String())
		}
	}

	// Speed MT/s
	// 3.3 says:
	// If the value is 0, no memory device is installed in the socket.
	// If the size is unknown, the field value is FFFFh.
	// If the size is 32GB-1MB or greater, the field value is 7FFFh
	// and the actual size is stored in the Extended Sizefield.
	// But 3.4alfa (not final!):
	// 0000h = the speed is unknown
	// FFFFh = the speed is 65,535 MT/s or greater, and the actual
	// speed is stored in the Extended Speed field.
	if (raw.Speed == 0x7FFF || raw.Speed == 0xFFFF) && len(fb) >= sizeAsPer3_3 {
		mem.Speed = raw.ExtendedSpeed
	} else {
		mem.Speed = uint32(raw.Speed)
	}

	// In case of virtual machine Manufacturer may not be reterived, Reason: NOT RETURNED FROM SMBIOS INFO
	if raw.Manufacturer > 0 && raw.Manufacturer <= arrSize {
		index := raw.Manufacturer - 1
		if index >= 0 {
			mem.Manufacturer = ss[index]
		}
	}

	// In case of virtual machine SerialNumber may not be reterived, Reason: NOT RETURNED FROM SMBIOS INFO
	if raw.SerialNumber > 0 && raw.SerialNumber <= arrSize {
		index := raw.SerialNumber - 1
		if index >= 0 {
			memSerNo := ss[index]
			//systemInfo.Memory = memSerNo
			mem.SerialNumber = memSerNo
		}
	}

	spew.Dump(raw)
	spew.Dump(mem)
	return nil
}

func getMemoryRaw(fb []byte, ss []string) *MemoryRaw {
	fbLen := len(fb)
	var raw MemoryRaw
	//Note:- For description of each field please refer the pdf 'DSP0134_3.0.0.pdf', which is available in the same repo.
	//Checking size as per SMBIOS 2.1 spec
	if fbLen >= sizeAsPer2_1 {
		//This will copy 2 Bytes related to 'Physical Memory Array Handle'
		raw.MemArrayHandle = binary.LittleEndian.Uint16(fb[0:2])
		//This will copy 2 Bytes related to 'Memory Error Information Handle'
		raw.MemErrorInfoHandle = binary.LittleEndian.Uint16(fb[2:4])
		//This will copy 2 Bytes related to 'Total Width'
		raw.TotalWidth = binary.LittleEndian.Uint16(fb[4:6])
		//This will copy 2 Bytes related to 'Data Width'
		raw.DataWidth = binary.LittleEndian.Uint16(fb[6:8])
		//This will copy 2 Bytes related to 'Size'
		raw.Size = binary.LittleEndian.Uint16(fb[8:10])
		//This will copy 1 Byte related to 'Form Factor'
		raw.FormFactor = fb[10]
		//This will copy 1 Byte related to 'Device Set'
		raw.DeviceSet = fb[11]
		//This will copy 1 Byte related to 'Device Locator'
		raw.DeviceLocator = fb[12]
		//This will copy 1 Byte related to 'Bank Locator'
		raw.BankLocator = fb[13]
		//This will copy 1 Byte related to 'Memory Type'
		raw.MemType = fb[14]
		//This will copy 2 Bytes related to 'Type Detail'
		raw.TypeDetail = binary.LittleEndian.Uint16(fb[15:17])
	}

	//Checking size as per SMBIOS 2.3 spec
	if fbLen >= sizeAsPer2_3 {
		//This will copy 2 Bytes related to 'Speed'
		raw.Speed = binary.LittleEndian.Uint16(fb[17:19])
		//This will copy 1 Byte related to 'Index of the Manufacturer string'
		raw.Manufacturer = fb[19]
		//This will copy 1 Byte related to 'Index of the SerialNumber string'
		raw.SerialNumber = fb[20]
		//This will copy 1 Byte related to 'Index of the AssetTag string'
		raw.AssetTag = fb[21]
		//This will copy 1 Byte related to 'Index of the PartNumber string'
		raw.PartNumber = fb[22]
	}

	//Checking size as per SMBIOS 2.6 spec
	if fbLen >= sizeAsPer2_6 {
		//This will copy 1 Byte related to 'Attribute'
		raw.Attribute = fb[23]
	}

	//Checking size as per SMBIOS 2.7 spec
	if fbLen >= sizeAsPer2_7 {
		//This will copy 4 Bytes related to 'Extended Memory Size'
		raw.ExtendedSize = binary.LittleEndian.Uint32(fb[24:28])
		//This will copy 2 Bytes related to 'Configured Memory Clock Speed'
		raw.ConfiguredMemClockSpeed = binary.LittleEndian.Uint16(fb[28:30])
	}

	//Checking size as per SMBIOS 2.8 spec
	if fbLen >= sizeAsPer2_8 {
		//This will copy 2 Bytes related to 'Minimum voltage'
		raw.MinVoltage = binary.LittleEndian.Uint16(fb[30:32])
		//This will copy 2 Bytes related to 'Maximum voltage'
		raw.MaxVoltage = binary.LittleEndian.Uint16(fb[32:34])
		//This will copy 2 Bytes related to 'Configured voltage'
		raw.ConfiguredVoltage = binary.LittleEndian.Uint16(fb[34:36])
	}

	if fbLen >= sizeAsPer3_2 {
		raw.MemoryTechnology = fb[36]
		raw.MemoryOperatingModeCapability = binary.LittleEndian.Uint16(fb[37:39])
		raw.FirmwareVersion = fb[39]
		raw.ModuleManufacturerID = binary.LittleEndian.Uint16(fb[40:42])
		raw.ModuleProductID = binary.LittleEndian.Uint16(fb[42:44])
		raw.MemorySubsystemControllerManufacturerID = binary.LittleEndian.Uint16(fb[44:46])
		raw.MemorySubsystemControllerProductID = binary.LittleEndian.Uint16(fb[46:48])
		raw.NonVolatileSize = binary.LittleEndian.Uint64(fb[48:56])
		raw.VolatileSize = binary.LittleEndian.Uint64(fb[56:64])
		raw.CacheSize = binary.LittleEndian.Uint64(fb[64:72])
		raw.LogicalSize = binary.LittleEndian.Uint64(fb[72:80])
	}

	if fbLen >= sizeAsPer3_3 {
		raw.ExtendedSpeed = binary.LittleEndian.Uint32(fb[80:84])
		raw.ExtendedConfiguredMemorySpeed = binary.LittleEndian.Uint32(fb[84:88])
	}

	return &raw
}
