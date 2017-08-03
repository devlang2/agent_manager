package collectors

import (
	"crypto/aes"
	"crypto/cipher"
	"log"
	"net"

	"github.com/davecgh/go-spew/spew"
	"github.com/iwondory/agent_manager/event"
	//	"github.com/nanobox-io/golang-syslogparser/rfc5424"
)

const (
	msgBufSize = 1024
)

var (
	iv  = []byte("2981eeca66b5c3cd")
	key = []byte("c43ac86d84469030f28c0a9656b1c533")
)

type UDPCollector struct {
	format string
	addr   *net.UDPAddr
}

func (s *UDPCollector) Start(c chan<- *event.Event) error {
	spew.Dump()
	conn, err := net.ListenUDP("udp", s.addr)
	if err != nil {
		return err
	}

	go func() {
		buf := make([]byte, msgBufSize)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("Read error: " + err.Error())
				continue
			}

			block, err := aes.NewCipher(key)
			if err != nil {
				panic(err)
			}

			//    // The IV needs to be unique, but not secure. Therefore it's common to
			//    // include it at the beginning of the ciphertext.
			//    if len(ciphertext) < aes.BlockSize {
			//        panic("ciphertext too short")
			//    }

			//    // CBC mode always works in whole blocks.
			//    if len(ciphertext)%aes.BlockSize != 0 {
			//        panic("ciphertext is not a multiple of the block size")
			//    }

			//    mode := cipher.NewCBCDecrypter(block, iv)

			//    // CryptBlocks can work in-place if the two arguments are the same.
			//    mode.CryptBlocks(ciphertext, ciphertext)

			//    // If the original plaintext lengths are not a multiple of the block
			//    // size, padding would have to be added when encrypting, which would be
			//    // removed at this point. For an example, see
			//    // https://tools.ietf.org/html/rfc5246#section-6.2.3.2. However, it's
			//    // critical to note that ciphertexts must be authenticated (i.e. by
			//    // using crypto/hmac) before being decrypted in order to avoid creating
			//    // a padding oracle.

			//    fmt.Printf("%s\n", ciphertext)

			//			spew.Dump(buf[:n])
			//            text := decrypt(buf[:n])

			//			p := rfc5424.NewParser(buf[:n])
			//			err = p.Parse()
			//			if err != nil {
			//				log.Printf("Parse error: " + err.Error())
			//				continue
			//			}

			//			event := event.NewEvent()
			//			event.Data = p.Dump()
			//			event.Addr = addr
			//			event.Origin = string(buf[:n]) // Original message

			//			c <- event
		}
	}()
	return nil
}

func (s *UDPCollector) Addr() net.Addr {
	return s.addr
}
