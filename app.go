package main
import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"mklink/internal/linker"
)
type App struct{ ctx context.Context; service *linker.Service }
func NewApp(service *linker.Service)*App{return &App{service:service}}
type LinkRequest struct{Type string `json:"type"`;Target string `json:"target"`;Link string `json:"link"`;Overwrite bool `json:"overwrite"`}
type LinkResult struct{LinkType string `json:"linkType"`;Target string `json:"target"`;Link string `json:"link"`;Message string `json:"message"`}
func(a *App)startup(ctx context.Context){a.ctx=ctx}
func(a *App)CreateLink(r LinkRequest)(LinkResult,error){
 if a.service==nil{return LinkResult{},errors.New("link service is unavailable")}
 if r.Overwrite{if err:=a.service.RemoveExistingLink(r.Link);err!=nil{return LinkResult{},err}}
 if err:=a.service.Create(linker.Request{Type:linker.Type(r.Type),Target:r.Target,Link:r.Link});err!=nil{return LinkResult{},err}
	target:=strings.TrimSpace(r.Target);link:=strings.TrimSpace(r.Link)
	t,err:=filepath.Abs(filepath.Clean(target));if err!=nil{return LinkResult{},fmt.Errorf("resolve target path: %w",err)}
	l,err:=filepath.Abs(filepath.Clean(link));if err!=nil{return LinkResult{},fmt.Errorf("resolve link path: %w",err)}
	return LinkResult{LinkType:r.Type,Target:t,Link:l,Message:"Link created successfully."},nil
}
func(a *App)Validate(r LinkRequest)error{return linker.Validate(linker.Request{Type:linker.Type(r.Type),Target:r.Target,Link:r.Link})}
