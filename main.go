package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/GregorioDiStefano/qf/clouds"
	"github.com/GregorioDiStefano/qf/crypto"
)

var bucket string

var googleCreds string
var awsKey, awsSecret, awsRegion string

const (
	firstChunkSize  = 1 * 1024 * 1024
	chunkSize       = 5 * 1024 * 1024
	objKeyLength    = 2
	cryptoKeyLength = 10
	objKeyPrefix    = "_qf_ft_"
)

func randomCode(l int) string {
	possible := "123456789abcdefghjkmnpqrstuvwxyzABCDEFGHIJKLMNPQRSTUVWXYV"
	str := ""

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < l; i++ {
		loc := rand.Int31n(int32(len(possible)))
		str += string(possible[loc])
	}

	return str
}

func storageObjectName(objKey string, count, endOfFile int) string {
	return fmt.Sprintf("%s_%s_%d_%d", objKeyPrefix, objKey, count, endOfFile)
}

func sendfile(gs clouds.Cloud) {
	obj := randomCode(objKeyLength)
	key := randomCode(cryptoKeyLength)

	fmt.Printf("ID: %s\n", obj+key)

	endOfFile := 0
	count := 1
	payload := []byte{}

	r := bufio.NewReader(os.Stdin)
	data := make([]byte, 1<<20)
	bufferSize := chunkSize

	for {
		n, err := r.Read(data)

		if count == 1 {
			bufferSize = firstChunkSize
		}

		if err == io.EOF {
			endOfFile = 1
		} else if err == nil {
			payload = append(payload, data[0:n]...)
		}

		if endOfFile != 1 && len(payload) < bufferSize {
			continue
		}

		reader, writer := io.Pipe()

		go func() {
			writer.Write(payload)
			writer.Close()
		}()

		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)

		var uploadBuffer []byte
		uploadWriter := bytes.NewBuffer(uploadBuffer)

		crypto.EncryptFile(buf, uploadWriter, []byte(key))

		fn := storageObjectName(obj, count, endOfFile)
		if err := gs.Upload(fn, uploadWriter); err != nil {
			panic(err)
		}

		payload = []byte{}
		count++

		if endOfFile > 0 {
			break
		}
	}
}

func recvfile(obj string, cloud clouds.Cloud, keep bool) {
	id := obj[0:objKeyLength]
	key := obj[objKeyLength:]

	count := 1
	fails := 0

	for {
		endOfFile := 0

		if fails == 1 || fails == 5 {
			endOfFile = 1
		} else if fails >= 10 {
			panic("failed to find file part")
		}

		obj := storageObjectName(id, count, endOfFile)
		r, err := cloud.Download(obj)

		if err != nil {
			fails++

			if fails > 1 {
				time.Sleep(time.Second * 3 * time.Duration(fails))
			}

			continue
		}

		fails = 0
		crypto.DecryptFile(r, os.Stdout, []byte(key))

		if !keep {
			cloud.Remove(obj)
		}

		if endOfFile == 1 {
			os.Exit(0)
		}

		count++
	}
}

func removeAll(cloud clouds.Cloud) error {
	var err error
	var files []string

	if files, err = cloud.ListObjectsWithPrefix(objKeyPrefix); err != nil {
		return err
	}

	var count int
	for _, obj := range files {
		if err = cloud.Remove(obj); err != nil {
			return err
		} else {
			count += 1
		}
	}

	log.Printf("%d objects removed from bucket.", count)

	return nil
}

func parseCmdline() (help, keep, deleteAll bool) {
	flag.BoolVar(&help, "h", false, "print help screen")
	flag.BoolVar(&keep, "k", false, "don't remove files after they are downloaded")
	flag.BoolVar(&deleteAll, "d", false, "delete all stored files and quit")
	flag.Parse()

	return
}

func printHelp() {
	flag.PrintDefaults()
}

func main() {
	var cloud clouds.Cloud
	help, keep, deleteAll := parseCmdline()

	stat, _ := os.Stdin.Stat()

	if (stat.Mode()&os.ModeCharDevice != 0) && (len(os.Args) == 1 || help) {
		flag.Usage()
		fmt.Println("\n You need to pipe data into qf (cat file | qf) to upload a file or pass an ID as single argument to retrieve a file")
		os.Exit(1)
	}

	if googleCreds != "" {
		cloud = clouds.NewGoogleStorage(bucket, googleCreds)
	} else if awsKey != "" {
		cloud = clouds.NewAWSStorage(bucket, awsKey, awsSecret, awsRegion)
	} else {
		panic("Amazon Web Service/Google cloud credentials not built into binary! Rebuild.")
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		sendfile(cloud)
	} else if deleteAll {
		if err := removeAll(cloud); err != nil {
			panic(err)
		}
	} else {
		obj := os.Args[len(os.Args)-1]
		recvfile(obj, cloud, keep)
	}
}
