package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/emicklei/proto"
)

var (
	rootDir       = "../../src/framework/proto/idl"
	outputFile    = "../../src/framework/proto/pb/rpc_msg_define.go"
	importBaseDir = "framework/proto/pb/"
	allMsgMapping = make(map[string]*FileProtoDefinition) // 全部proto文件定义
)

// rpc请求信息
type RpcMsgInfo struct {
	ReqMsgId     int
	ReqMsgIdName string
	ReqMsgName   string
	RespMsgName  string // 当为notify时回复消息为空值
}

// 单个proto文件内的定义
type FileProtoDefinition struct {
	filepath      string
	packageName   string
	importPath    string
	msgIdEnumName string
	msgIdMap      map[string]int
	msgList       []string
	validRpcMsg   []RpcMsgInfo
}

func (d *FileProtoDefinition) genRpcMsgMap() {
	for _, msgname := range d.msgList {
		//"req_"开头的为rpc消息
		if strings.HasPrefix(msgname, "req_") {
			//请求的消息id
			reqMsgId, ok := d.msgIdMap[msgname]
			if !ok {
				fmt.Printf("rpc request message %s not defined a messag id.\n", msgname)
				os.Exit(1)
			}

			//"resp_"开头为rpc应答消息
			respMsgName := "resp_" + msgname[len("req_"):]
			bFindRespMsg := false
			for _, v := range d.msgList {
				if respMsgName == v {
					bFindRespMsg = true
					break
				}
			}
			if !bFindRespMsg {
				fmt.Printf("rpc request message %s not defined a response message in pair.\n", msgname)
				os.Exit(1)
			}

			rpcMsgInfo := RpcMsgInfo{
				ReqMsgId:     reqMsgId,
				ReqMsgIdName: "msg_id_" + msgname,
				ReqMsgName:   msgname,
				RespMsgName:  respMsgName,
			}

			allMsgMapping[d.filepath].validRpcMsg = append(allMsgMapping[d.filepath].validRpcMsg, rpcMsgInfo)
		}
	}

	for _, msgname := range d.msgList {
		//"notify_"开头的为rpc通知消息(one way)
		if strings.HasPrefix(msgname, "notify_") {
			//通知的消息id
			notifyMsgId, ok := d.msgIdMap[msgname]
			if !ok {
				fmt.Printf("rpc notify message %s not defined a messag id.\n", msgname)
				continue
			}

			notifyMsgInfo := RpcMsgInfo{
				ReqMsgId:     notifyMsgId,
				ReqMsgIdName: "msg_id_" + msgname,
				ReqMsgName:   msgname,
				RespMsgName:  "",
			}
			allMsgMapping[d.filepath].validRpcMsg = append(allMsgMapping[d.filepath].validRpcMsg, notifyMsgInfo)
		}
	}
}

// template渲染数据
type TemplateInfo struct {
	ImportPaths []string
	CsRpcMsgs   []*TemplateRpcMsgInfo
	SsRpcMsgs   []*TemplateRpcMsgInfo
}

// rpc请求信息
type TemplateRpcMsgInfo struct {
	ImportPath        string //go文件import的路径
	FileName          string //proto文件名(去掉.proto后缀)
	PackageName       string //包名
	ReqMsgId          int    //消息ID
	TemplReqMsgIdName string //消息id的枚举名(go style)
	TemplReqMsgName   string //(go style)
	TemplRespMsgName  string // 当为notify时回复消息为空值(go style)
}

// 获取目录下的所有proto文件
func getProtoFiles(dirPath string) []string {
	files := []string{}
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".proto") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files
}

// 解析单个proto文件定义
func doParse(filePath string) {
	reader, _ := os.Open(filePath)
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		fmt.Printf("parse file %s error: %s", filePath, err)
		os.Exit(1)
	}

	proto.Walk(definition,
		// 解析包名
		proto.WithPackage(func(p *proto.Package) {
			packageName := strings.ReplaceAll(p.Name, "_", "")
			allMsgMapping[filePath].packageName = packageName
			importPath := filepath.Join(importBaseDir, packageName)
			importPath = strings.ReplaceAll(importPath, "\\", "/")
			allMsgMapping[filePath].importPath = importPath
		}),

		// 解析MessageID
		proto.WithEnum(func(e *proto.Enum) {
			for _, elem := range e.Elements {
				enumField := elem.(*proto.EnumField)
				if !strings.HasPrefix(enumField.Name, "msg_id_") {
					return
				}
				allMsgMapping[filePath].msgIdEnumName = e.Name
				allMsgMapping[filePath].msgIdMap[enumField.Name[len("msg_id_"):]] = enumField.Integer
			}
		}),

		// 解析消息定义
		proto.WithMessage(func(m *proto.Message) {
			allMsgMapping[filePath].msgList = append(allMsgMapping[filePath].msgList, m.Name)
		}))
}

// 打印rpc消息定义
func printAllMsg() {
	fmt.Printf("*******************************************************************\n\n")
	for path, msgMappingInfo := range allMsgMapping {
		if len(msgMappingInfo.validRpcMsg) <= 0 {
			continue
		}
		_, filename := filepath.Split(path)
		fmt.Printf("%-10s: %s\n", "ProtoFile", filename)
		fmt.Printf("%-10s: %s\n", "ImportPath", msgMappingInfo.importPath)

		//fmt.Printf("%-15s:\n", "Rpc Message")
		for _, v := range msgMappingInfo.validRpcMsg {
			if v.RespMsgName != "" {
				fmt.Printf("[%d : %s -> %s]\n", v.ReqMsgId, v.ReqMsgName, v.RespMsgName)
			}
		}
		//fmt.Printf("%-15s:\n", "Notify Message")
		for _, v := range msgMappingInfo.validRpcMsg {
			if v.RespMsgName == "" {
				fmt.Printf("[%d : %s]\n", v.ReqMsgId, v.ReqMsgName)
			}
		}
		fmt.Printf("\n******************************************************************\n\n")
	}
}

// 生成胶水代码
func genFileCode() {
	text := `// Code generated by proto_parser tool. DO NOT EDIT.
package pb

import (
{{- range $i, $v := $.ImportPaths}}
{{printf "\t\"%s\"" $v -}}
{{- end}}
)
{{$filename := ""}}
var CSRpcMsg [][]any = [][]any{
{{- range $i, $v := $.CsRpcMsgs}}
{{- if ne $filename $v.FileName}}
{{printf "\t//%s" $v.FileName -}}
{{$filename = $v.FileName}}
{{- end}}
{{- if eq $v.TemplRespMsgName ""}}
{{printf "\t{int(%s), &%s{}}," $v.TemplReqMsgIdName $v.TemplReqMsgName}}
{{- else}}
{{printf "\t{int(%s), &%s{}, &%s{}}," $v.TemplReqMsgIdName $v.TemplReqMsgName $v.TemplRespMsgName}}
{{- end}}
{{- end}}
}
{{$filename = ""}}
var SSRpcMsg [][]any = [][]any{
{{- range $i, $v := $.SsRpcMsgs}}
{{- if ne $filename $v.FileName}}
{{printf "\t//%s" $v.FileName -}}
{{$filename = $v.FileName}}
{{- end}}
{{- if eq $v.TemplRespMsgName ""}}
{{printf "\t{int(%s), &%s{}}," $v.TemplReqMsgIdName $v.TemplReqMsgName}}
{{- else}}
{{printf "\t{int(%s), &%s{}, &%s{}}," $v.TemplReqMsgIdName $v.TemplReqMsgName $v.TemplRespMsgName}}
{{- end}}
{{- end}}
}
`
	tmpl, err := template.New("proto").Parse(text)
	if err != nil {
		fmt.Printf("create template error: %s\n", err)
		return
	}

	codeFile, err := os.OpenFile(outputFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("create go file error: %s\n", err)
		return
	}

	templInfo := &TemplateInfo{}
	// import list
	for _, fileRpcInfo := range allMsgMapping {
		if len(fileRpcInfo.validRpcMsg) == 0 {
			continue
		}
		templInfo.ImportPaths = append(templInfo.ImportPaths, fileRpcInfo.importPath)
	}
	sort.SliceStable(templInfo.ImportPaths, func(i, j int) bool {
		return templInfo.ImportPaths[i] < templInfo.ImportPaths[j]
	})

	for path, fileRpcInfo := range allMsgMapping {
		_, filename := filepath.Split(path)
		filename = strings.TrimSuffix(filename, ".proto")
		packageName := fileRpcInfo.packageName
		msgIdEnumName := fileRpcInfo.msgIdEnumName
		for _, rpcInfo := range fileRpcInfo.validRpcMsg {
			data := TemplateRpcMsgInfo{}
			data.FileName = filename
			data.PackageName = packageName
			data.ReqMsgId = rpcInfo.ReqMsgId
			data.TemplReqMsgIdName = packageName + "." + toGoStyle(msgIdEnumName) + "_" + rpcInfo.ReqMsgIdName
			data.TemplReqMsgName = packageName + "." + toGoStyle(rpcInfo.ReqMsgName)
			if rpcInfo.RespMsgName != "" {
				data.TemplRespMsgName = packageName + "." + toGoStyle(rpcInfo.RespMsgName)
			}

			if strings.HasPrefix(packageName, "cs") {
				templInfo.CsRpcMsgs = append(templInfo.CsRpcMsgs, &data)
			} else if strings.HasPrefix(packageName, "ss") {
				templInfo.SsRpcMsgs = append(templInfo.SsRpcMsgs, &data)
			}
		}
	}
	//由于每个proto文件的消息ID人为约束范围，直接用消息ID排序
	sort.SliceStable(templInfo.CsRpcMsgs, func(i, j int) bool {
		return templInfo.CsRpcMsgs[i].ReqMsgId < templInfo.CsRpcMsgs[j].ReqMsgId
	})
	sort.SliceStable(templInfo.SsRpcMsgs, func(i, j int) bool {
		return templInfo.SsRpcMsgs[i].ReqMsgId < templInfo.SsRpcMsgs[j].ReqMsgId
	})

	err = tmpl.Execute(codeFile, templInfo)
	if err != nil {
		fmt.Printf("template executed error: %s\n", err)
		return
	}
}

func toGoStyle(src string) string {
	dest := strings.Builder{}
	grow := 'A' - 'a'
	flag := true
	for i := 0; i < len(src); i++ {
		if src[i] == '_' {
			flag = true
			continue
		}

		if flag {
			if src[i] >= 'a' && src[i] <= 'z' {
				dest.WriteByte(src[i] + byte(grow))
			} else {
				dest.WriteByte(src[i])
			}
			flag = false
		} else {
			dest.WriteByte(src[i])
		}
	}
	return dest.String()
}

func main() {
	protoFiles := getProtoFiles(rootDir)
	for _, path := range protoFiles {
		allMsgMapping[path] = &FileProtoDefinition{
			filepath: path,
			msgIdMap: make(map[string]int),
		}
		doParse(path)
		allMsgMapping[path].genRpcMsgMap()
	}
	//控制台打印
	printAllMsg()

	//生成胶水文件
	genFileCode()
}
