package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"github.com/google/uuid"
	"net"
	"strings"
	"sync"
	"time"
)

type SUID struct {
	value []byte // [16]byte
}

// Difference in 100-nanosecond intervals between
// UUID epoch (October 15, 1582) and Unix epoch (January 1, 1970).
const epochStart = 122192928000000000
const radix = 62

var digitalAry62 = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// UUID v1/v2 store.
var (
	storageMutex  sync.Mutex
	storageOnce   sync.Once
	clockSequence uint16
	lastTime      uint64
	hardwareAddr  [6]byte
)

// SetVersion sets version bits.
func (this *SUID) SetVersion(v byte) {
	this.value[6] = (this.value[6] & 0x0f) | (v << 4)
}

// SetVariant sets variant bits as described in RFC 4122.
func (this *SUID) SetVariant() {
	this.value[8] = (this.value[8] & 0xbf) | 0x80
}

func NewSUID() *SUID {
	value := make([]byte, 16)
	this := SUID{value}

	t, q, h := getStorage()

	binary.BigEndian.PutUint32(value[0:], uint32(t))
	binary.BigEndian.PutUint16(value[4:], uint16(t>>32))
	binary.BigEndian.PutUint16(value[6:], uint16(t>>48))
	binary.BigEndian.PutUint16(value[8:], q)

	copy(this.value[10:], h)

	this.SetVersion(1)
	this.SetVariant()

	return &this
}

func initClockSequence() {
	buf := make([]byte, 2)
	safeRandom(buf)
	clockSequence = binary.BigEndian.Uint16(buf)
}

func safeRandom(dest []byte) {
	if _, err := rand.Read(dest); err != nil {
		panic(err)
	}
}

func initHardwareAddr() {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			if len(iface.HardwareAddr) >= 6 {
				copy(hardwareAddr[:], iface.HardwareAddr)
				return
			}
		}
	}

	// Initialize hardwareAddr randomly in case
	// of real network interfaces absence
	safeRandom(hardwareAddr[:])

	// Set multicast bit as recommended in RFC 4122
	hardwareAddr[0] |= 0x01
}

func initStorage() {
	initClockSequence()
	initHardwareAddr()
}

// Returns UUID v1/v2 store state.
// Returns epoch timestamp, clock sequence, and hardware address.
func getStorage() (uint64, uint16, []byte) {
	storageOnce.Do(initStorage)

	storageMutex.Lock()
	defer storageMutex.Unlock()

	timeNow := unixTimeFunc()
	// Clock changed backwards since last UUID generation.
	// Should increase clock sequence.
	if timeNow <= lastTime {
		clockSequence++
	}
	lastTime = timeNow

	return timeNow, clockSequence, hardwareAddr[:]
}

// Returns difference in 100-nanosecond intervals between
// UUID epoch (October 15, 1582) and current time.
// This is default epoch calculation function.
func unixTimeFunc() uint64 {
	return epochStart + uint64(time.Now().UnixNano()/100)
}

func (this *SUID) String() string {
	return suidToShortS(this.value)
}

func suidToShortS(data []byte) string {
	// [16]byte
	buf := make([]byte, 22)
	sb := bytes.NewBuffer(buf)
	sb.Reset()

	var msb int64
	for i := 0; i < 8; i++ {
		msb = msb<<8 | int64(data[i])
	}

	var lsb int64
	for i := 8; i < 16; i++ {
		lsb = lsb<<8 | int64(data[i])
	}

	digTo62(msb>>12, 8, sb)
	digTo62(msb>>16, 4, sb)
	digTo62(msb, 4, sb)
	digTo62(lsb>>48, 4, sb)
	digTo62(lsb, 12, sb)

	return sb.String()
}

func digTo62(_val int64, _digs byte, _sb *bytes.Buffer) {
	hi := int64(1) << (_digs * 4)
	i := hi | (_val & (hi - 1))

	negative := i < 0
	if !negative {
		i = -i
	}

	skip := true
	for i <= -radix {
		if skip {
			skip = false
		} else {
			offset := -(i % radix)
			_sb.WriteByte(digitalAry62[int(offset)])
		}
		i = i / radix
	}
	_sb.WriteByte(digitalAry62[int(-i)])

	if negative {
		_sb.WriteByte('-')
	}
}

func GoogleUUID() string {
	uuidStr := uuid.New().String()
	uuidList := strings.Split(uuidStr, "-")
	uuidStr = strings.Join(uuidList, "")
	return uuidStr
}
