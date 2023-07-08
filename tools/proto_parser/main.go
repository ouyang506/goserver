package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
)

//全部proto文件定义
type MsgMapping map[string]*FileProtoDefinition

//单个proto文件内的定义
type FileProtoDefinition struct {
	filename          string
	packageName       string
	msgIdMap          map[string]int
	msgList           []string
	validRpcMsg       []RpcMsgInfo
	validRpcNotifyMsg []RpcNotifyMsgInfo
}

// func (d *FileProtoDefinition) String() string {
// 	return fmt.Sprintf("filename: %s, package: %s, msgIdMap: %v, msgList: %v",
// 		d.filename, d.packageName, d.msgIdMap, d.msgList)
// }

func (d *FileProtoDefinition) genRpcMsgMap() {
	for _, msgname := range d.msgList {
		//req_开头的为rpc消息
		if strings.HasPrefix(msgname, "req_") {
			//请求的消息id
			reqMsgId, ok := d.msgIdMap[msgname]
			if !ok {
				fmt.Printf("rpc request message %s not defined a messag id.\n", msgname)
				os.Exit(1)
			}

			//resp_开头为rpc应答消息
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

			//应答的消息id
			// respMsgId, ok := d.msgIdMap[respMsgName]
			// if !ok {
			// 	fmt.Printf("rpc response message %s not defined a messag id.\n", respMsgName)
			// 	os.Exit(1)
			// }

			rpcMsgInfo := RpcMsgInfo{
				reqMsgId:    reqMsgId,
				reqMsgName:  msgname,
				respMsgName: respMsgName,
			}

			allMsgMapping[currentFile].validRpcMsg = append(allMsgMapping[currentFile].validRpcMsg, rpcMsgInfo)
		} else if strings.HasPrefix(msgname, "notify_") {
			//通知的消息id
			notifyMsgId, ok := d.msgIdMap[msgname]
			if !ok {
				fmt.Printf("rpc notify message %s not defined a messag id.\n", msgname)
				continue
			}

			notifyMsgInfo := RpcNotifyMsgInfo{
				msgId:   notifyMsgId,
				msgName: msgname,
			}

			allMsgMapping[currentFile].validRpcNotifyMsg = append(allMsgMapping[currentFile].validRpcNotifyMsg, notifyMsgInfo)
		} else {
			//ignore, normal message definition
		}
	}
}

//rpc请求信息
type RpcMsgInfo struct {
	reqMsgId    int
	reqMsgName  string
	respMsgName string
}

//rpc通知信息(one way)
type RpcNotifyMsgInfo struct {
	msgId   int
	msgName string
}

var (
	rootDir       = "C:\\Users\\ouyang\\Desktop\\workspace\\goserver\\common\\pbmsg\\idl"
	allMsgMapping = make(MsgMapping)
	currentFile   string
)

func main() {
	protoFiles := getProtoFiles(rootDir)
	for _, filename := range protoFiles {
		allMsgMapping[filename] = &FileProtoDefinition{
			filename: filename,
			msgIdMap: make(map[string]int),
		}
		currentFile = filename
		doParse(filename)
		allMsgMapping[filename].genRpcMsgMap()
	}

	//打印rpc消息定义
	fmt.Printf("*******************************************\n")
	for filename, msgMappingInfo := range allMsgMapping {
		fmt.Printf("%-15s: %s\n", "Proto File", filename)
		if len(msgMappingInfo.validRpcMsg) > 0 {
			fmt.Printf("%-15s:\n", "Rpc Message")
			for _, v := range msgMappingInfo.validRpcMsg {
				fmt.Printf("[%d : %s -> %s]\n", v.reqMsgId, v.reqMsgName, v.respMsgName)
			}
		}

		if len(msgMappingInfo.validRpcNotifyMsg) > 0 {
			fmt.Printf("%-15s:\n", "Notify Message")
			for _, v := range msgMappingInfo.validRpcNotifyMsg {
				fmt.Printf("[%d : %s]\n", v.msgId, v.msgName)
			}
		}

		fmt.Printf("*******************************************\n")
	}
}

//获取目录下的所有proto文件
func getProtoFiles(dirPath string) []string {
	files := []string{}
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".proto") {
			return nil
		}
		files = append(files, info.Name())
		return nil
	})
	return files
}

//解析单个proto文件定义
func doParse(filename string) {
	filepath := rootDir + "/" + filename
	reader, _ := os.Open(filepath)
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, _ := parser.Parse()

	proto.Walk(definition,
		proto.WithPackage(handlePackage),
		proto.WithEnum(handleEnum),
		proto.WithMessage(handleMessage))
}

//解析包名
func handlePackage(p *proto.Package) {
	allMsgMapping[currentFile].packageName = p.Name
}

//解析MessageID
func handleEnum(e *proto.Enum) {
	if e.Name != "MessageID" {
		return
	}
	visitor := new(enumVisitor)
	for _, each := range e.Elements {
		each.Accept(visitor)
	}
}

//enum遍历
type enumVisitor struct {
	proto.NoopVisitor
}

func (v enumVisitor) VisitEnumField(i *proto.EnumField) {
	allMsgMapping[currentFile].msgIdMap[i.Name] = i.Integer
}

//解析消息定义
func handleMessage(m *proto.Message) {
	allMsgMapping[currentFile].msgList = append(allMsgMapping[currentFile].msgList, m.Name)
}
