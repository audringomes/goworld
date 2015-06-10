package goworld

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

// Acp holds I/O buffer to communicate with Magik ACP
type Acp struct {
	Name string
	io   *bufio.ReadWriter
}

// NewAcp creates and init new Acp with name
func NewAcp(name string) *Acp {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	inout := bufio.NewReadWriter(reader, writer)
	return &Acp{Name: name, io: inout}
}

// Flush send buffer data
func (a *Acp) Flush() {
	a.io.Flush()
}

// Write writes buffer to Acp output
func (a *Acp) Write(buf []byte) {
	a.io.Write(buf)
	a.Flush()
}

// PutBool sends boolean value to Acp output
func (a *Acp) PutBool(b bool) {
	var ival byte
	if !b {
		ival = 1
	}
	a.io.WriteByte(ival)
	a.Flush()
}

// PutUshort sends unsigned short value to Acp output
func (a *Acp) PutUshort(short uint16) {
	buf := make([]byte, 2, 2)
	binary.LittleEndian.PutUint16(buf, short)
	a.Write(buf)
}

// PutString sends string value to Acp output
func (a *Acp) PutString(s string) {
	bytes := []byte(s)
	l := len(bytes)
	a.PutUshort(uint16(l))
	a.Write(bytes)
}

// read bytes from Acp input
func (a *Acp) readBytes(n int) (buf []byte, err error) {
	buf = make([]byte, 0, n)
	_, err = a.io.Read(buf[:cap(buf)])
	if err != nil && err != io.EOF {
		err = fmt.Errorf("Error reading short: %v\n", err)
		return
	}
	buf = buf[:n]
	return
}

// ReadNumber reads number from Acp input
func (a *Acp) ReadNumber(data interface{}) {
	if err := binary.Read(a.io, binary.LittleEndian, data); err != nil {
		log.Fatal(err)
	}
}

// GetUshort reads unsigned short from Acp input
func (a *Acp) GetUshort() int {
	var (
		res uint16
	)
	a.ReadNumber(&res)
	return int(res)
}

// GetShort reads short from Acp input
func (a *Acp) GetShort() int {
	var (
		res int16
	)
	a.ReadNumber(&res)
	return int(res)
}

// GetUint reads unsigned int from Acp input
func (a *Acp) GetUint() int {
	var (
		res uint32
	)
	a.ReadNumber(&res)
	return int(res)
}

// GetString reads string from Acp input
func (a *Acp) GetString() string {
	b := a.GetUshort()
	log.Printf("get string - bytes to read %v\n", b)
	buf, err := a.readBytes(b)
	if err != nil {
		log.Fatal(err)
	}
	return string(buf)
}

// VerifyConnection verify Acp process name
func (a *Acp) VerifyConnection(name string) bool {
	acpName := a.GetString()
	res := acpName == name
	log.Printf("Name: %s; From SW: %s; Res: %v\n", name, acpName, res)
	a.PutBool(res)
	return res
}

// EstablishProtocol checks Acp protocol
func (a *Acp) EstablishProtocol(minProtocol, maxProtocol int) bool {
	min, max := minProtocol, maxProtocol
	for {
		in := a.GetUshort()
		log.Printf("Protocol from SW: %d\n", in)
		if in < min || in > max {
			a.PutBool(false)
			a.PutUshort(uint16(min))
			a.PutUshort(uint16(max))
			return false
		}
		break
	}
	a.PutBool(true)
	return true
}

// Connect verify connection and protocol to Acp
func (a *Acp) Connect(processName string, protocolMin, protocolMax int) (err error) {
	log.Printf("ACP started name: %s\n", processName)
	if res := a.VerifyConnection(processName); !res {
		err = errors.New("Error verify connection")
		return
	}
	log.Println("Connection verified")
	if res := a.EstablishProtocol(protocolMin, protocolMax); !res {
		err = errors.New("Error establish protocol")
		return
	}
	log.Println("Protocol established")
	log.Println("Connected")
	return
}