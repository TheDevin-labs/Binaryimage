package sbimg

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const (
	Magic   = 0x534249 // "SBI"
	Version = 1
)

var (
	ErrInvalidMagic   = errors.New("sbimg: invalid magic bytes")
	ErrUnsupportedArch = errors.New("sbimg: unsupported architecture")
)

type Arch uint8

const (
	ArchAMD64 Arch = 0x01
	ArchARM64 Arch = 0x02
)

type Header struct {
	Magic   [3]byte
	Version uint8
	Arch    Arch
	_       [3]byte
	Size    uint64
}

type Image struct {
	Header Header
	Code   []byte
}

func Read(path string) (*Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var h Header
	if err := binary.Read(f, binary.LittleEndian, &h); err != nil {
		return nil, err
	}

	magic := uint32(h.Magic[0])<<16 | uint32(h.Magic[1])<<8 | uint32(h.Magic[2])
	if magic != Magic {
		return nil, ErrInvalidMagic
	}

	if h.Arch != ArchAMD64 && h.Arch != ArchARM64 {
		return nil, ErrUnsupportedArch
	}

	code := make([]byte, h.Size)
	if _, err := io.ReadFull(f, code); err != nil {
		return nil, err
	}

	return &Image{Header: h, Code: code}, nil
}

func Write(path string, arch Arch, code []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := Header{
		Magic:   [3]byte{0x53, 0x42, 0x49},
		Version: Version,
		Arch:    arch,
		Size:    uint64(len(code)),
	}

	if err := binary.Write(f, binary.LittleEndian, h); err != nil {
		return err
	}

	_, err = f.Write(code)
	return err
}

func (a Arch) String() string {
	switch a {
	case ArchAMD64:
		return "amd64"
	case ArchARM64:
		return "arm64"
	default:
		return "unknown"
	}
}
