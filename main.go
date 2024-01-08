package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"hash"
	"hash/adler32"
	"hash/crc32"
	"hash/crc64"
	"hash/fnv"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/cespare/xxhash/v2"
)

type discard struct {
}

func (h discard) Write(p []byte) (int, error) {
	return len(p), nil
}

func (h discard) Sum(b []byte) []byte {
	return nil
}

func (h discard) Reset() {
}

func (h discard) Size() int {
	return 0
}

func (h discard) BlockSize() int {
	return 0
}

func getCPUTimes() (time.Duration, time.Duration, error) {
	var rusage syscall.Rusage

	err := syscall.Getrusage(syscall.RUSAGE_SELF, &rusage)
	if err != nil {
		return 0, 0, err
	}

	userTime := time.Duration(rusage.Utime.Sec)*time.Second + time.Duration(rusage.Utime.Usec)*time.Microsecond
	systemTime := time.Duration(rusage.Stime.Sec)*time.Second + time.Duration(rusage.Stime.Usec)*time.Microsecond

	return userTime, systemTime, nil
}

func run() error {
	flagAlgo := flag.String("a", "", "Checksum algorithm (adler32, crc32, crc32c, crc64, fnv32, fnv64, md5, none, sha1, sha256, xxh64).")
	flagEncoding := flag.String("e", "hex", "Checksum encoding (base64, hex, raw)")
	flagStats := flag.Bool("stats", false, "Log statistics.")
	flag.Parse()

	var hasher hash.Hash

	switch *flagAlgo {
	case "adler32":
		hasher = adler32.New()
	case "crc32":
		hasher = crc32.New(crc32.MakeTable(crc32.IEEE))
	case "crc32c":
		hasher = crc32.New(crc32.MakeTable(crc32.Castagnoli))
	case "crc64":
		hasher = crc64.New(crc64.MakeTable(crc64.ISO))
	case "fnv32":
		hasher = fnv.New32()
	case "fnv64":
		hasher = fnv.New64()
	case "md5":
		hasher = md5.New()
	case "none":
		hasher = discard{}
	case "sha1":
		hasher = sha1.New()
	case "sha256":
		hasher = sha256.New()
	case "xxh64":
		hasher = xxhash.New()
	default:
		return fmt.Errorf("unknown checksum algorithm %q", *flagAlgo)
	}

	t0 := time.Now()

	startUserTime, startSystemTime, err := getCPUTimes()
	if err != nil {
		return err
	}

	if _, err := io.Copy(hasher, os.Stdin); err != nil {
		return err
	}

	checksum := hasher.Sum(nil)

	switch *flagEncoding {
	case "base64":
		fmt.Printf("%s\n", base64.StdEncoding.EncodeToString(checksum))
	case "hex":
		fmt.Printf("%s\n", hex.EncodeToString(checksum))
	case "raw":
		os.Stdout.Write(checksum)
	default:
		return fmt.Errorf("unknown checksum encoding %q", *flagEncoding)
	}

	if *flagStats {
		endUserTime, endSystemTime, err := getCPUTimes()
		if err != nil {
			return err
		}
		userTime := endUserTime - startUserTime
		systemTime := endSystemTime - startSystemTime
		if err != nil {
			return err
		}
		realTime := time.Since(t0)
		cpuUsage := 100 * float64(userTime+systemTime) / float64(realTime)
		fmt.Fprintf(os.Stderr, "%v elapsed, %v user, %v system, %.2f%% CPU\n", realTime, userTime, systemTime, cpuUsage)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
