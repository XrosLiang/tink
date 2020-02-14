// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
////////////////////////////////////////////////////////////////////////////////

package streamingaead_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"testing"

	"github.com/google/tink/go/subtle/streamingaead"
)

func TestEncryptDecrypt(t *testing.T) {
	ikm, err := hex.DecodeString("000102030405060708090a0b0c0d0e0f00112233445566778899aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}
	aad, err := hex.DecodeString("aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		name               string
		keySizeInBytes     int
		segmentSize        int
		firstSegmentOffset int
		plaintextSize      int
		chunkSize          int
	}{
		{
			name:               "small-1",
			keySizeInBytes:     16,
			segmentSize:        256,
			firstSegmentOffset: 0,
			plaintextSize:      20,
			chunkSize:          64,
		},
		{
			name:               "small-2",
			keySizeInBytes:     16,
			segmentSize:        512,
			firstSegmentOffset: 0,
			plaintextSize:      400,
			chunkSize:          64,
		},
		{
			name:               "small-offset-1",
			keySizeInBytes:     16,
			segmentSize:        256,
			firstSegmentOffset: 8,
			plaintextSize:      20,
			chunkSize:          64,
		},
		{
			name:               "small-offset-2",
			keySizeInBytes:     16,
			segmentSize:        512,
			firstSegmentOffset: 8,
			plaintextSize:      400,
			chunkSize:          64,
		},
		{
			name:               "empty-1",
			keySizeInBytes:     16,
			segmentSize:        256,
			firstSegmentOffset: 0,
			plaintextSize:      0,
			chunkSize:          128,
		},
		{
			name:               "empty-2",
			keySizeInBytes:     16,
			segmentSize:        256,
			firstSegmentOffset: 8,
			plaintextSize:      0,
			chunkSize:          128,
		},
		{
			name:               "medium-1",
			keySizeInBytes:     16,
			segmentSize:        256,
			firstSegmentOffset: 0,
			plaintextSize:      1024,
			chunkSize:          128,
		},
		{
			name:               "medium-2",
			keySizeInBytes:     16,
			segmentSize:        512,
			firstSegmentOffset: 0,
			plaintextSize:      3086,
			chunkSize:          128,
		},
		{
			name:               "medium-3",
			keySizeInBytes:     32,
			segmentSize:        1024,
			firstSegmentOffset: 0,
			plaintextSize:      12345,
			chunkSize:          128,
		},
		{
			name:               "large-chunks-1",
			keySizeInBytes:     16,
			segmentSize:        256,
			firstSegmentOffset: 0,
			plaintextSize:      1024,
			chunkSize:          4096,
		},
		{
			name:               "large-chunks-2",
			keySizeInBytes:     16,
			segmentSize:        512,
			firstSegmentOffset: 0,
			plaintextSize:      5086,
			chunkSize:          4096,
		},
		{
			name:               "large-chunks-3",
			keySizeInBytes:     32,
			segmentSize:        1024,
			firstSegmentOffset: 0,
			plaintextSize:      12345,
			chunkSize:          5000,
		},
		{
			name:               "medium-offset-1",
			keySizeInBytes:     16,
			segmentSize:        256,
			firstSegmentOffset: 8,
			plaintextSize:      1024,
			chunkSize:          64,
		},
		{
			name:               "medium-offset-2",
			keySizeInBytes:     16,
			segmentSize:        512,
			firstSegmentOffset: 20,
			plaintextSize:      3086,
			chunkSize:          256,
		},
		{
			name:               "medium-offset-3",
			keySizeInBytes:     32,
			segmentSize:        1024,
			firstSegmentOffset: 10,
			plaintextSize:      12345,
			chunkSize:          5000,
		},
		{
			name:               "last-segment-full-1",
			keySizeInBytes:     16,
			segmentSize:        256,
			firstSegmentOffset: 0,
			plaintextSize:      216,
			chunkSize:          64,
		},
		{
			name:               "last-segment-full-2",
			keySizeInBytes:     16,
			segmentSize:        256,
			firstSegmentOffset: 16,
			plaintextSize:      200,
			chunkSize:          256,
		},
		{
			name:               "last-segment-full-3",
			keySizeInBytes:     16,
			segmentSize:        256,
			firstSegmentOffset: 16,
			plaintextSize:      440,
			chunkSize:          1024,
		},
		{
			name:               "single-byte-1",
			keySizeInBytes:     16,
			segmentSize:        256,
			firstSegmentOffset: 0,
			plaintextSize:      1024,
			chunkSize:          1,
		},
		{
			name:               "single-byte-2",
			keySizeInBytes:     32,
			segmentSize:        512,
			firstSegmentOffset: 0,
			plaintextSize:      5086,
			chunkSize:          1,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cipher, err := streamingaead.NewAESGCMHKDF(ikm, "SHA256", tc.keySizeInBytes, tc.segmentSize, tc.firstSegmentOffset)
			if err != nil {
				t.Errorf("Cannot create a cipher: %v", err)
			}

			pt, ct, err := encrypt(cipher, aad, tc.plaintextSize)
			if err != nil {
				t.Error(err)
			}

			if err := decrypt(cipher, aad, pt, ct, tc.chunkSize); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestModifiedCiphertext(t *testing.T) {
	ikm, err := hex.DecodeString("000102030405060708090a0b0c0d0e0f00112233445566778899aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}
	aad, err := hex.DecodeString("aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}

	const (
		keySizeInBytes     = 16
		segmentSize        = 256
		firstSegmentOffset = 8
		plaintextSize      = 1024
		chunkSize          = 128
	)

	cipher, err := streamingaead.NewAESGCMHKDF(ikm, "SHA256", keySizeInBytes, segmentSize, firstSegmentOffset)
	if err != nil {
		t.Errorf("Cannot create a cipher: %v", err)
	}

	pt, ct, err := encrypt(cipher, aad, plaintextSize)
	if err != nil {
		t.Error(err)
	}

	t.Run("truncate ciphertext", func(t *testing.T) {
		for i := 0; i < len(ct); i += 8 {
			if err := decrypt(cipher, aad, pt, ct[:i], chunkSize); err == nil {
				t.Errorf("expected error")
			}
		}
	})
	t.Run("append to ciphertext", func(t *testing.T) {
		sizes := []int{1, segmentSize - len(ct)%segmentSize, segmentSize}
		for _, size := range sizes {
			ct2 := append(ct, make([]byte, size)...)
			if err := decrypt(cipher, aad, pt, ct2, chunkSize); err == nil {
				t.Errorf("expected error")
			}
		}
	})
	t.Run("flip bits", func(t *testing.T) {
		for i := range ct {
			ct2 := make([]byte, len(ct))
			copy(ct2, ct)
			ct2[i] ^= byte(1)
			if err := decrypt(cipher, aad, pt, ct2, chunkSize); err == nil {
				t.Errorf("expected error")
			}
		}
	})
	t.Run("delete segments", func(t *testing.T) {
		for i := 0; i < len(ct)/segmentSize+1; i++ {
			start, end := segmentPos(segmentSize, firstSegmentOffset, cipher.HeaderLength(), i)
			if start > len(ct) {
				break
			}
			if end > len(ct) {
				end = len(ct)
			}
			ct2 := append(ct[:start], ct[end:]...)
			if err := decrypt(cipher, aad, pt, ct2, chunkSize); err == nil {
				t.Errorf("expected error")
			}
		}
	})
	t.Run("duplicate segments", func(t *testing.T) {
		for i := 0; i < len(ct)/segmentSize+1; i++ {
			start, end := segmentPos(segmentSize, firstSegmentOffset, cipher.HeaderLength(), i)
			if start > len(ct) {
				break
			}
			if end > len(ct) {
				end = len(ct)
			}
			ct2 := append(ct[:end], ct[start:]...)
			if err := decrypt(cipher, aad, pt, ct2, chunkSize); err == nil {
				t.Errorf("expected error")
			}
		}
	})
	t.Run("modify aad", func(t *testing.T) {
		for i := range aad {
			aad2 := make([]byte, len(aad))
			copy(aad2, aad)
			aad2[i] ^= byte(1)
			if err := decrypt(cipher, aad2, pt, ct, chunkSize); err == nil {
				t.Errorf("expected error")
			}
		}
	})
}

// encrypt generates a random plaintext of size plaintextSize and encrypts it using the cipher.
// Upon success this function returns the actual plaintext and ciphertext bytes.
func encrypt(cipher *streamingaead.AESGCMHKDF, aad []byte, plaintextSize int) ([]byte, []byte, error) {
	pt := make([]byte, plaintextSize)
	for i := range pt {
		pt[i] = byte(i % 253)
	}

	ctBuf := &bytes.Buffer{}
	w, err := cipher.NewEncryptingWriter(ctBuf, aad)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create an encrypt writer: %v", err)
	}
	n, err := w.Write(pt)
	if err != nil {
		return nil, nil, fmt.Errorf("error writing to an encrypt writer: %v", err)
	}
	if n != len(pt) {
		return nil, nil, fmt.Errorf("unexpected number of bytes written. Got=%d;want=%d", n, len(pt))
	}
	if err := w.Close(); err != nil {
		return nil, nil, fmt.Errorf("error closing writer: %v", err)
	}
	return pt, ctBuf.Bytes(), err
}

// decrypt decrypts ciphertext ct using the cipher and validates that it's the same as the original plaintext pt.
func decrypt(cipher *streamingaead.AESGCMHKDF, aad, pt, ct []byte, chunkSize int) error {
	r, err := cipher.NewDecryptingReader(bytes.NewBuffer(ct), aad)
	if err != nil {
		return fmt.Errorf("cannot create an encrypt reader: %v", err)
	}

	var (
		chunk     = make([]byte, chunkSize)
		decrypted = 0
		eof       = false
	)
	for !eof {
		n, err := r.Read(chunk)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading chunk: %v", err)
		}
		eof = err == io.EOF
		got := chunk[:n]
		want := pt[decrypted : decrypted+n]
		if !bytes.Equal(got, want) {
			return fmt.Errorf("decrypted data doesn't match. Got=%s;want=%s", hex.EncodeToString(got), hex.EncodeToString(want))
		}
		decrypted += n
	}
	if decrypted != len(pt) {
		return fmt.Errorf("number of decrypted bytes doesn't match. Got=%d;want=%d", decrypted, len(pt))
	}
	return nil
}

func segmentPos(segmentSize, firstSegmentOffset, headerLen, segmentNr int) (int, int) {
	start := segmentSize * segmentNr
	end := start + segmentSize

	firstSegmentDiff := firstSegmentOffset + headerLen
	if start > 0 {
		start -= firstSegmentDiff
	}
	end -= firstSegmentDiff
	return start + headerLen, end + headerLen
}