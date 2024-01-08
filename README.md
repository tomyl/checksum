# checksum

Simple tool to play around with checksum algorithms available in Go.

```bash
$ go install github.com/tomyl/checksum@main
$ checksum -h
Usage of checksum:
  -a string
    	Checksum algorithm (adler32, crc32, crc32c, crc64, fnv32, fnv64, md5, none, sha1, sha256, xxh64).
  -e string
    	Checksum encoding (base64, hex, raw) (default "hex")
  -stats
    	Log statistics.
$ dd if=/dev/zero bs=1M count=1000 | checksum -a crc32 -e base64 -stats                       
1000+0 records in
1000+0 records out
1048576000 bytes (1,0 GB, 1000 MiB) copied, 0,272765 s, 3,8 GB/s
ZOnQaA==
271.497462ms elapsed, 39.491ms user, 217.763ms system, 94.75% CPU
```
