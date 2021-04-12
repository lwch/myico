package convert

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"mime"
	"net/http"
	"path"

	"golang.org/x/image/draw"
)

const (
	typeICO = 1
	typeCUR = 2
)

// Generate generate ico from jpeg or png file
// file_format: https://en.wikipedia.org/wiki/ICO_(file_format)
func Generate(src, dst string) error {
	srcData, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	ctype := mime.TypeByExtension(path.Ext(src))
	if ctype == "" {
		if len(srcData) > 512 {
			ctype = http.DetectContentType(srcData[:512])
		} else {
			ctype = http.DetectContentType(srcData)
		}
	}

	var srcImg image.Image
	switch ctype {
	case "image/jpeg":
		srcImg, err = jpeg.Decode(bytes.NewReader(srcData))
	case "image/png":
		srcImg, err = png.Decode(bytes.NewReader(srcData))
	default:
		return fmt.Errorf("unsupported file type: %s", ctype)
	}
	if err != nil {
		return err
	}

	return build(srcImg, dst)
}

func build(src image.Image, dst string) error {
	sizes := []int{256, 128, 64, 48, 32, 16}

	var buf bytes.Buffer

	// write header
	var hdr struct {
		Reserved uint16
		Type     uint16
		Count    uint16
	}
	hdr.Type = typeICO
	hdr.Count = uint16(len(sizes))
	err := binary.Write(&buf, binary.LittleEndian, hdr)
	if err != nil {
		return err
	}

	// write entry header
	type entryHdr struct {
		Width    uint8
		Height   uint8
		Color    uint8
		Reserved uint8
		Planes   uint16
		Bits     uint16
		Size     uint32
		Offset   uint32
	}
	type entry struct {
		hdr  entryHdr
		data []byte
	}
	var entryList []entry
	for _, size := range sizes {
		rect := image.Rect(0, 0, int(size), int(size))
		dst := image.NewRGBA(rect)
		draw.CatmullRom.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
		var dstBuf bytes.Buffer
		err = png.Encode(&dstBuf, dst)
		if err != nil {
			return err
		}
		data := dstBuf.Bytes()
		entryList = append(entryList, entry{
			hdr: entryHdr{
				Width:  uint8(size),
				Height: uint8(size),
				Planes: 1,
				Bits:   32,
				Size:   uint32(len(data)),
			},
			data: data,
		})
	}

	offset := uint32(6 + 16*len(sizes))
	for _, entry := range entryList {
		entry.hdr.Offset = offset
		err = binary.Write(&buf, binary.LittleEndian, entry.hdr)
		if err != nil {
			return err
		}
		offset += entry.hdr.Size
	}
	for _, entry := range entryList {
		_, err = buf.Write(entry.data)
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(dst, buf.Bytes(), 0644)
}
