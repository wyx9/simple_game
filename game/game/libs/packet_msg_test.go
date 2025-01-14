package libs

import (
	"fmt"
	"testing"
)

func TestPackMsg(t *testing.T) {
	p := "test"
	msg := Pack2Msg(p)
	fmt.Println(msg)

	//pack := EnCodePack(msg)
	//fmt.Println(pack)
	//fmt.Println(string(pack))
	//
	//codePack := DeCodePack(pack)
	//fmt.Println(*codePack)

}
