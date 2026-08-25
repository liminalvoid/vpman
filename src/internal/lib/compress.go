package lib

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
)

// Mirrors QT's qCompress, encodes with Base64 and adds prefix
// "vpn"
func CompressConfigForAmnezia(data []byte, level int) (string, error) {
	if len(data) == 0 {
		return string(make([]byte, 4)), nil
	}

	if level < -1 || level > 9 {
		level = -1
	}

	var buf bytes.Buffer
	buf.Grow(len(data) / 2 + 8)
	binary.Write(&buf, binary.BigEndian, uint32(len(data)))

	w, err := zlib.NewWriterLevel(&buf, level)
	if err != nil {
		return "", err
	}

	if _, err := w.Write(data); err != nil {
		w.Close()

		return "", err
	}

	if err := w.Close(); err != nil {
		return "", err
	}

	return "vpn://" + base64.RawURLEncoding.EncodeToString(buf.Bytes()), nil
}
