//go:build !windows
package linker
import "errors"
func createJunction(target,link string)error{return errors.New("junctions are supported only on Windows")}
