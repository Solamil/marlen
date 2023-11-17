package marlen

import (
	"fmt"
	"strings"
	"github.com/hashicorp/golang-lru/v2"
	"os"
	"path/filepath"
	"crypto/md5"
	"time"
)

type cacheRecord struct {
	Value  string
	Expiry time.Time
}

const CACHESIZE int = 10000
const MIN_SIZE_FILE_CACHE int = 80

var CACHE_DIR string = "cache"

const HASHSIZE int = md5.Size

var cache, _ = lru.New[[HASHSIZE]byte, cacheRecord](CACHESIZE)

func Store(signature, value string) string {
	cacheSignature := hash(signature)
	if len(value) >= MIN_SIZE_FILE_CACHE {
		value = storeInFile(signature, value)
	}
	cache.Add(cacheSignature, cacheRecord{value, time.Now()})
	return fmt.Sprintf("%x", cacheSignature)
}

func storeInFile(signature, value string) string {
	if _, err := os.Stat(CACHE_DIR); os.IsNotExist(err) {
		err = os.Mkdir(CACHE_DIR, 0755)
		if err != nil {
			fmt.Printf("error %s", err)
		}
	}
	filename := fmt.Sprintf("file:%x.txt", hash(signature))
	err := os.WriteFile(filepath.Join(CACHE_DIR, filename), []byte(value), 0644)
	if err != nil {
		fmt.Printf("error %s", err)
	}
	return filename
}

func Get(signature string) (cacheRecord, bool) {
	cacheSignature := hash(signature)
	record, found := cache.Get(cacheSignature)
	if found && record.Value != "" {
		if strings.Compare(record.Value, fmt.Sprintf("file:%x.txt", cacheSignature)) == 0 {
			// filename := fmt.Sprintf("%s/file:%x.txt", CACHE_DIR, cacheSignature)
			filename := filepath.Join(CACHE_DIR, fmt.Sprintf("file:%x.txt", cacheSignature))
			record.Value = readAllFile(filename)
		}
	}
	return record, found
}

func hash(signature string) [HASHSIZE]byte {
	return md5.Sum([]byte(signature))
}

func readAllFile(filename string) string {
	result, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		fmt.Printf("error %s", err)
		return ""
	}
	return string(result)
}

