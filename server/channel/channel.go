package channel

import (
	"log"
  "syscall"
  "bytes"
  "errors"

	"github.com/gorpher/gowin32/wrappers"
	"golang.org/x/sys/windows"
)

const WTS_CURRENT_SESSION uint32 = 0xFFFFFFFF
const CHANNEL_CHUNK_LENGTH uint32 = 1600

type channel struct {
  vc_name string
  vc_handle syscall.Handle
  vc_file_handle windows.Handle
}

func New(vc_name string) channel {
  return channel {vc_name, 0, 0}
}

func (c *channel) Init() {
  c.vc_handle = wrappers.WTSVirtualChannelOpenEx(WTS_CURRENT_SESSION, c.vc_name, 0)
  var pbytesreturned uint32
  var pt1 *uint16
  wrappers.WTSVirtualChannelQuery(c.vc_handle, wrappers.WTSVirtualFileHandle, &pt1, &pbytesreturned)

  c.vc_file_handle = windows.Handle(*pt1)
  log.Printf("Obtained file handle %d with %d bytes!", c.vc_file_handle, pbytesreturned)
}

func (c *channel) Read(p []byte) (int, error) {
  hEvent, _ := windows.CreateEvent(nil, 1, 0, nil)

  overlapped := windows.Overlapped {
    Internal: 0,
    InternalHigh: 0,
    Offset: 0,
    OffsetHigh: 0,
    HEvent: hEvent,
  }

  err := windows.ReadFile(c.vc_file_handle, p, nil, &overlapped)
  if err != nil {
    err = windows.GetLastError()
    if err != windows.ERROR_IO_PENDING && err != nil {
      panic(err)
    }
  }

  log.Println("Trying to read")
  event, err := windows.WaitForSingleObject(overlapped.HEvent, windows.INFINITE)
  if err != nil {
    panic(err)
  }

  if event != windows.WAIT_OBJECT_0 {
    panic("wtf")
  }

  var bytes_read uint32
  err = windows.GetOverlappedResult(c.vc_file_handle, &overlapped, &bytes_read, false)
  if err != nil {
    panic(err)
  }

  log.Printf("Read %d bytes", bytes_read)

  return int(bytes_read), nil
}

func (c *channel) Write(buf []byte) (int, error) {
  hEvent, _ := windows.CreateEvent(nil, 1, 0, nil)

  overlapped := windows.Overlapped {
    Internal: 0,
    InternalHigh: 0,
    Offset: 0,
    OffsetHigh: 0,
    HEvent: hEvent,
  }

  err := windows.WriteFile(
    c.vc_file_handle,
    buf,
    nil,
    &overlapped,
  )

  if err != nil {
    err = windows.GetLastError()
    if err != windows.ERROR_IO_PENDING && err != nil {
      panic(err)
    }
  }

  log.Println("Trying to write")
  event, err := windows.WaitForSingleObject(overlapped.HEvent, windows.INFINITE)
  if err != nil {
    panic(err)
  }

  if event != windows.WAIT_OBJECT_0 {
    panic("wtf")
  }

  var bytes_wrote uint32
  err = windows.GetOverlappedResult(c.vc_file_handle, &overlapped, &bytes_wrote, false)
  if err != nil {
    panic(err)
  }
  log.Printf("Wrote %d bytes", bytes_wrote)
  return int(bytes_wrote), nil
}

func (c *channel) Close() error {
  wrappers.WTSVirtualChannelClose(c.vc_handle)  
  return nil
}

func (c *channel) Challenge() error {
  buf := make([]byte, 8)
  buf[0] = 1

  c.Write(buf)

  buf2 := make([]byte, 8)
  _, err := c.Read(buf2)
  if err != nil {
    panic(err)
  }

  if !bytes.Equal(buf, buf2) {
    return errors.New("Challenge failed")
  }
  return nil
}
