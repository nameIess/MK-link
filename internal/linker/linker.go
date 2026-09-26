package linker
import("errors";"fmt";"os";"path/filepath";"strings";"golang.org/x/sys/windows")
type Type string
const(TypeFileSymlink Type="file-symlink";TypeDirSymlink Type="directory-symlink";TypeJunction Type="junction";TypeHardlink Type="hardlink")
type Request struct{Type Type;Target string;Link string}
var(ErrInvalidType=errors.New("unsupported link type");ErrTargetMissing=errors.New("target does not exist");ErrLinkExists=errors.New("link path already exists");ErrTargetNotFile=errors.New("target must be a file");ErrTargetNotDirectory=errors.New("target must be a directory");ErrCrossVolume=errors.New("hard links require the same volume");ErrSamePath=errors.New("target and link paths must differ"))
func Validate(r Request)error{
 if strings.TrimSpace(r.Target)==""||strings.TrimSpace(r.Link)==""{return errors.New("target and link paths are required")}
 switch r.Type{case TypeFileSymlink,TypeDirSymlink,TypeJunction,TypeHardlink:default:return ErrInvalidType}
 target,err:=absolutePath(r.Target);if err!=nil{return fmt.Errorf("resolve target path: %w",err)}
 link,err:=absolutePath(r.Link);if err!=nil{return fmt.Errorf("resolve link path: %w",err)}
 if strings.EqualFold(target,link){return ErrSamePath}
 info,err:=os.Stat(target);if err!=nil{if errors.Is(err,os.ErrNotExist){return ErrTargetMissing};return fmt.Errorf("inspect target: %w",err)}
 switch r.Type{case TypeFileSymlink,TypeHardlink:if info.IsDir(){return ErrTargetNotFile};case TypeDirSymlink,TypeJunction:if !info.IsDir(){return ErrTargetNotDirectory}}
 if _,err:=os.Lstat(link);err==nil{return ErrLinkExists}else if !errors.Is(err,os.ErrNotExist){return fmt.Errorf("inspect link path: %w",err)}
 if r.Type==TypeHardlink{same,err:=sameVolume(target,filepath.Dir(link));if err!=nil{return err};if !same{return ErrCrossVolume}}
 return nil
}
type Service struct{}
func NewService()*Service{return &Service{}}
func(s *Service)Create(r Request)error{
 if err:=Validate(r);err!=nil{return err}
 target,err:=absolutePath(r.Target);if err!=nil{return fmt.Errorf("resolve target path: %w",err)}
 link,err:=absolutePath(r.Link);if err!=nil{return fmt.Errorf("resolve link path: %w",err)}
 if err:=os.MkdirAll(filepath.Dir(link),0755);err!=nil{return fmt.Errorf("create link parent directory: %w",err)}
 switch r.Type{case TypeFileSymlink:return createSymlink(target,link,false);case TypeDirSymlink:return createSymlink(target,link,true);case TypeJunction:return createJunction(target,link);case TypeHardlink:return createHardLink(target,link);default:return ErrInvalidType}
}
func(s *Service)RemoveExistingLink(path string)error{
 clean,err:=absolutePath(path);if err!=nil{return fmt.Errorf("resolve existing link path: %w",err)}
 info,err:=os.Lstat(clean);if errors.Is(err,os.ErrNotExist){return nil};if err!=nil{return fmt.Errorf("inspect existing link: %w",err)}
 if info.IsDir(){if err:=os.Remove(clean);err!=nil{return fmt.Errorf("remove existing link directory: %w",err)};return nil}
 if err:=os.Remove(clean);err!=nil{return fmt.Errorf("remove existing link: %w",err)};return nil
}
func createSymlink(target,link string,directory bool)error{
 flags:=uint32(windows.SYMBOLIC_LINK_FLAG_ALLOW_UNPRIVILEGED_CREATE);if directory{flags|=windows.SYMBOLIC_LINK_FLAG_DIRECTORY}
 tp,err:=windows.UTF16PtrFromString(target);if err!=nil{return fmt.Errorf("encode target path: %w",err)}
 lp,err:=windows.UTF16PtrFromString(link);if err!=nil{return fmt.Errorf("encode link path: %w",err)}
 if err:=windows.CreateSymbolicLink(lp,tp,flags);err!=nil{return fmt.Errorf("create symbolic link: %w",err)};return nil
}
func createHardLink(target,link string)error{
 tp,err:=windows.UTF16PtrFromString(target);if err!=nil{return fmt.Errorf("encode target path: %w",err)}
 lp,err:=windows.UTF16PtrFromString(link);if err!=nil{return fmt.Errorf("encode link path: %w",err)}
 if err:=windows.CreateHardLink(lp,tp,0);err!=nil{return fmt.Errorf("create hard link: %w",err)};return nil
}
func absolutePath(path string)(string,error){clean:=filepath.Clean(strings.TrimSpace(path));if clean=="."{return "",errors.New("path is empty")};return filepath.Abs(clean)}
func sameVolume(target,parent string)(bool,error){a,err:=volumeRoot(target);if err!=nil{return false,err};b,err:=volumeRoot(parent);if err!=nil{return false,err};return strings.EqualFold(a,b),nil}
func volumeRoot(path string)(string,error){a,err:=filepath.Abs(path);if err!=nil{return "",fmt.Errorf("resolve volume path: %w",err)};v:=filepath.VolumeName(a);if v==""{return "",errors.New("path has no Windows volume")};return v,nil}
