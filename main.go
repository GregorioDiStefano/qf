package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
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
	chunk     = 3 * 1024 * 1024
	objkey    = 2
	cryptokey = 8
)

func randomCode(l int) string {
	possible := "23456789abcdefghjkmnpqrstuvwxyz"
	str := ""

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < l; i++ {
		loc := rand.Int31n(int32(len(possible)))
		str += string(possible[loc])
	}

	return str
}

func sendfile(gs clouds.Cloud) {
	obj := randomCode(objkey)
	key := randomCode(cryptokey)

	fmt.Printf("ID: %s\n", obj+key)

	endOfFile := 0
	count := 1
	payload := []byte{}

	r := bufio.NewReader(os.Stdin)
	data := make([]byte, 1<<20)

	for {
		n, err := r.Read(data)

		if err == io.EOF {
			endOfFile = 1
		} else if err == nil {
			payload = append(payload, data[0:n]...)
		}

		if endOfFile != 1 && len(payload) < chunk {
			continue
		}

		fn := fmt.Sprintf("%s_%d_%d", obj, count, endOfFile)
		reader, writer := io.Pipe()

		go func() {
			writer.Write(payload)
			writer.Close()
		}()

		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)

		var buf2 []byte
		writer2 := bytes.NewBuffer(buf2)

		crypto.EncryptFile(buf, writer2, []byte(key))
		gs.Upload(fn, writer2)

		payload = []byte{}
		count++

		if endOfFile > 0 {
			break
		}
	}
}

func recvfile(obj string, gs clouds.Cloud) {
	id := obj[0:objkey]
	key := obj[objkey:]

	next := 1
	fails := 0

	for {
		endOfFile := 0

		if fails == 1 || fails == 5 {
			endOfFile = 1
		} else if fails >= 10 {
			panic("failed to find file part")
		}

		obj = fmt.Sprintf("%s_%d_%d", id, next, endOfFile)
		r, err := gs.Download(obj)

		if err != nil {
			fails++

			if fails > 1 {
				time.Sleep(time.Second * 2 * time.Duration(fails))
			}

			continue
		}

		fails = 0
		crypto.DecryptFile(r, os.Stdout, []byte(key))
		gs.Remove(obj)

		if endOfFile == 1 {
			os.Exit(0)
		}

		next++
	}
}

func main() {
	var cloud clouds.Cloud

	if googleCreds != "" {
		data, err := base64.StdEncoding.DecodeString(googleCreds)

		if err != nil {
			panic("failed to read base64'ed google credentials: " + err.Error())
		}

		ioutil.WriteFile("creds.json", []byte(data), 0400)
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "./creds.json")
		cloud = clouds.NewGoogleStorage(bucket)

	} else if awsKey != "" {

		os.Setenv("AWS_ACCESS_KEY_ID", awsKey)
		os.Setenv("AWS_SECRET_ACCESS_KEY", awsSecret)
		os.Setenv("AWS_REGION", awsRegion)

		cloud = clouds.NewAWSStorage(bucket)
	}

	stat, _ := os.Stdin.Stat()

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		sendfile(cloud)
	} else {
		if len(os.Args) == 0 || len(os.Args) > 2 {
			fmt.Println("either pipe file into command, or pass ID as single argument to retrieve file")
		} else {
			obj := os.Args[1]
			recvfile(obj, cloud)
		}
	}
}
