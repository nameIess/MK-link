//go:build windows
package linker
import("encoding/binary";"fmt";"os";"syscall";"unsafe";"golang.org/x/sys/windows")
const(fsctlSetReparsePoint=uint32(0x000900A4);ioReparseTagMountPoint=uint32(0xA0000003);mountPointDataHeaderSize=8;reparseHeaderSize=8;reparseMaxDataLength=16*1024)
func createJunction(target,link string)error{
 if err:=os.Mkdir(link,0755);err!=nil{return fmt.Errorf("create junction directory: %w",err)}
 defer func(){if _,err:=os.Lstat(link);err==nil{}}()
 substitute:=utf16Bytes(`\??\`+target);printName:=utf16Bytes(target)
 dataLength:=mountPointDataHeaderSize+len(substitute)+len(printName);if dataLength>reparseMaxDataLength{_ = os.Remove(link);return fmt.Errorf("junction target path is too long")}
 buffer:=make([]byte,reparseHeaderSize+dataLength)
 binary.LittleEndian.PutUint32(buffer[0:4],ioReparseTagMountPoint);binary.LittleEndian.PutUint16(buffer[4:6],uint16(dataLength));binary.LittleEndian.PutUint32(buffer[6:8],0)
 binary.LittleEndian.PutUint16(buffer[8:10],0);binary.LittleEndian.PutUint16(buffer[10:12],uint16(len(substitute)));binary.LittleEndian.PutUint16(buffer[12:14],uint16(len(substitute)));binary.LittleEndian.PutUint16(buffer[14:16],uint16(len(printName)))
 copy(buffer[16:],substitute);copy(buffer[16+len(substitute):],printName)
 handle,err:=openReparsePoint(link);if err!=nil{_ = os.Remove(link);return err};defer func(){_ = windows.CloseHandle(handle)}()
 var returned uint32
 if err:=windows.DeviceIoControl(handle,fsctlSetReparsePoint,(*byte)(unsafe.Pointer(&buffer[0])),uint32(len(buffer)),nil,0,&returned,nil);err!=nil{_ = os.Remove(link);return fmt.Errorf("set junction reparse point: %w",err)}
 return nil
}
func openReparsePoint(path string)(windows.Handle,error){p,err:=windows.UTF16PtrFromString(path);if err!=nil{return 0,fmt.Errorf("encode junction path: %w",err)};h,err:=windows.CreateFile(p,windows.GENERIC_READ|windows.GENERIC_WRITE,windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,nil,windows.OPEN_EXISTING,windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS,0);if err!=nil{return 0,fmt.Errorf("open junction reparse point: %w",err)};return h,nil}
func utf16Bytes(value string)[]byte{encoded,err:=syscall.UTF16FromString(value);if err!=nil{return nil};out:=make([]byte,2*len(encoded));for i,unit:=range encoded{binary.LittleEndian.PutUint16(out[i*2:],unit)};return out}
