package marlen

import (
	"crypto/md5"
	"os"
	"fmt"
	"testing"
)


func TestCache(t *testing.T) {
	const hashsize int = md5.Size
	CACHE_DIR = "test_cache"
	tests := []struct {
		value     string
		signature string
		expFound  bool
	}{
		{"Test hash", "TestCache", true},
		{`Test hash kdsafk sdkfakj j kfdksfkfkdskfkajdfsk jdskafjk ksjdfakdsf
		  dskfajksdfjakdfj dskafkdsjfkasdfk`, "TestCache1", true},
		{"", "TestCache2", true},
	}
	if HASHSIZE != hashsize {
		t.Errorf("Hash size is %d, but it is tested to %d.", HASHSIZE, hashsize)
	}
	for _, test := range tests {
		Store(test.signature, test.value)

		if got, found := Get(test.signature); got.Value != test.value {
			t.Errorf("Expected '%s', but got '%s'", test.value, got.Value)
			if found != test.expFound {
				t.Errorf("Expected '%t', but got '%t'", test.expFound, found)
			}
		}
		if len(test.value) >= MIN_SIZE_FILE_CACHE {
			filename := fmt.Sprintf("%s/file:%x.txt", CACHE_DIR, hash(test.signature))
			if err := os.Remove(filename); err != nil {
				t.Errorf("error %s", err)
			}
		}
	}
}

func TestCleanUpCache(t *testing.T) {
	CACHE_DIR = "test_cache"
	dirRead, _ := os.Open(CACHE_DIR)
	dirFiles, _ := dirRead.Readdir(0)
	for index := range dirFiles {
		file := dirFiles[index]
		filename := file.Name()
		if err := os.Remove(CACHE_DIR + "/" + filename); err != nil {
			t.Errorf("error %s", err)
		}
	}
	if err := os.Remove(CACHE_DIR + "/"); err != nil {
		t.Errorf("error %s", err)
	}
}
