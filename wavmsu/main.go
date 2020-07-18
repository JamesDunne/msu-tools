package main

import (
	"bufio"
	"encoding/binary"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"log"
	"os"
	"path"
	"strings"
)

func FilenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

func main() {
	for i := 1; i < len(os.Args); i++ {
		convertWAV(os.Args[i])
	}
}

func convertWAV(wavFilename string) {
	fi, err := os.Open(wavFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer fi.Close()
	wi := wav.NewDecoder(fi)

	msuFilename := FilenameWithoutExtension(wavFilename) + ".pcm"

	fo, err := os.OpenFile(msuFilename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer fo.Close()

	hdr := []byte{0, 0, 0, 0}
	fo.WriteString("MSU1")
	fo.Write(hdr)

	bfo := bufio.NewWriter(fo)

	buf := &audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: 2,
			SampleRate:  44100,
		},
		Data:           make([]int, 2048),
		SourceBitDepth: 16,
	}
	chunk := make([]byte, 4096)

	err = wi.FwdToPCM()
	if err != nil {
		log.Fatal(err)
	}

	n, err := wi.PCMBuffer(buf)
	if err != nil {
		log.Fatal(err)
	}
	for n > 0 {
		k := 0
		p := 0
		for i := 0; i < n; i++ {
			binary.LittleEndian.PutUint16(chunk[k:k+2], uint16(buf.Data[p]))
			k += 2
			p++
		}

		_, err = bfo.Write(chunk[0:k])
		if err != nil {
			log.Fatal(err)
		}

		n, err = wi.PCMBuffer(buf)
		if err != nil {
			log.Fatal(err)
		}
	}
	bfo.Flush()
}
