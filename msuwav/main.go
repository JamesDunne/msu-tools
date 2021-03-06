package main

import (
	"bufio"
	"encoding/binary"
	"encoding/csv"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

func FilenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

func main() {
	cw := csv.NewWriter(os.Stdout)
	for i := 1; i < len(os.Args); i++ {
		filename := os.Args[i]
		loopPoint := convertMSU(filename)

		_ = cw.Write([]string{filename, strconv.FormatUint(uint64(loopPoint), 10)})
		cw.Flush()
	}
}

func convertMSU(msuFilename string) (loopPoint uint32) {
	fi, err := os.Open(msuFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer fi.Close()

	wavFilename := FilenameWithoutExtension(msuFilename) + ".wav"

	fo, err := os.OpenFile(wavFilename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer fo.Close()

	wo := wav.NewEncoder(fo, 44100, 16, 2, 1)
	defer wo.Close()

	hdr := make([]byte, 4)
	_, _ = fi.Read(hdr)
	if string(hdr) != "MSU1" {
		log.Fatal("MSU file header is not MSU1")
	}

	// read 4x 0 bytes:
	_, _ = fi.Read(hdr)
	loopPoint = binary.LittleEndian.Uint32(hdr)

	bfi := bufio.NewReader(fi)
	buf := &audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: 2,
			SampleRate:  44100,
		},
		Data:           make([]int, 2048),
		SourceBitDepth: 0,
	}
	chunk := make([]byte, 4096)
	n, err := bfi.Read(chunk)
	written := 0
	for n > 0 {
		k := 0
		j := 0
		for i := 0; i < n/2; i++ {
			buf.Data[k] = int(binary.LittleEndian.Uint16(chunk[j : j+2]))
			j += 2
			k++
			written += 2
		}
		buf.Data = buf.Data[0:k]
		err = wo.Write(buf)
		if err != nil {
			log.Fatal(err)
		}

		n, err = bfi.Read(chunk)
	}


	// Write a "smpl" chunk:
	// chunkID
	wo.AddLE([]byte("smpl"))
	// chunkSize
	wo.AddLE(uint32(9*4 + 6*4))
	// dwManufacturer
	wo.AddLE(uint32(0))
	// dwProduct
	wo.AddLE(uint32(0))
	// dwSamplePeriod
	wo.AddLE(uint32(22675)) // 44.1 kHz in nanoseconds
	// dwMIDIUnityNote
	wo.AddLE(uint32(60))
	// dwMIDIPitchFraction
	wo.AddLE(uint32(0))
	// dwSMPTEFormat
	wo.AddLE(uint32(0))
	// dwSMPTEOffset
	wo.AddLE(uint32(0))
	// cSampleLoops
	wo.AddLE(uint32(1))
	// cbSamplerData
	wo.AddLE(uint32(0))

	// dwIdentifier
	wo.AddLE(uint32(0))
	// dwType
	wo.AddLE(uint32(0)) // loop forward
	// dwStart
	wo.AddLE(uint32(loopPoint))
	// dwEnd
	wo.AddLE(uint32(written))
	// dwFraction
	wo.AddLE(uint32(0))
	// dwPlayCount
	wo.AddLE(uint32(0))

	return
}
