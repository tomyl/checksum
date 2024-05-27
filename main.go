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
	"io/fs"
	"os"
	"path/filepath"
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

type stats struct {
	startTime       time.Time
	startUserTime   time.Duration
	startSystemTime time.Duration
}

func newStats() (stats, error) {
	startUserTime, startSystemTime, err := getCPUTimes()
	if err != nil {
		return stats{}, err
	}
	return stats{time.Now(), startUserTime, startSystemTime}, nil
}

func (s stats) dump() error {
	endUserTime, endSystemTime, err := getCPUTimes()
	if err != nil {
		return err
	}
	userTime := endUserTime - s.startUserTime
	systemTime := endSystemTime - s.startSystemTime
	if err != nil {
		return err
	}
	realTime := time.Since(s.startTime)
	cpuUsage := 100 * float64(userTime+systemTime) / float64(realTime)
	fmt.Fprintf(os.Stderr, "%v elapsed, %v user, %v system, %.2f%% CPU\n", realTime, userTime, systemTime, cpuUsage)
	return nil
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

func runFile(name, algo, encoding string, r io.Reader) error {
	var hasher hash.Hash

	switch algo {
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
		return fmt.Errorf("unknown checksum algorithm %q", algo)
	}

	if _, err := io.Copy(hasher, r); err != nil {
		return err
	}

	checksum := hasher.Sum(nil)
	suffix := "\n"
	if name != "" {
		suffix = " " + name + "\n"
	}

	switch encoding {
	case "base64":
		fmt.Printf("%s%s", base64.StdEncoding.EncodeToString(checksum), suffix)
	case "hex":
		fmt.Printf("%s%s", hex.EncodeToString(checksum), suffix)
	case "raw":
		os.Stdout.Write(checksum)
		os.Stdout.WriteString(suffix)
	default:
		return fmt.Errorf("unknown checksum encoding %q", encoding)
	}

	return nil
}

func run() error {
	flagAlgo := flag.String("a", "", "Checksum algorithm (adler32, crc32, crc32c, crc64, fnv32, fnv64, md5, none, sha1, sha256, xxh64).")
	flagEncoding := flag.String("e", "hex", "Checksum encoding (base64, hex, raw)")
	flagStats := flag.Bool("stats", false, "Log statistics.")
	flag.Parse()

	names := flag.Args()

	stats, err := newStats()
	if err != nil {
		return err
	}

	if len(names) > 0 {
		for _, name := range names {
			if err := filepath.WalkDir(name, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				if d.Type() != 0 {
					fmt.Fprintf(os.Stderr, "%s: skipping because of mode %v\n", path, d.Type())
					return nil
				}
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				if err := runFile(path, *flagAlgo, *flagEncoding, file); err != nil {
					return err
				}
				return file.Close()
			}); err != nil {
				return err
			}
		}
		if *flagStats {
			return stats.dump()
		}
		return nil
	}

	if err := runFile("", *flagAlgo, *flagEncoding, os.Stdin); err != nil {
		return err
	}

	if *flagStats {
		return stats.dump()
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
