package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"strconv"
)

const STARTROWS = 6

type ShopCfg struct {
	Cfgs map[int32][]string
}

func main() {

	f, err := excelize.OpenFile("C:\\MyAllProject\\golang_note\\game\\excelize\\16 商店配置表.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	// 获取工作表中指定单元格的值
	//cell, err := f.GetCellValue("q_shop_goods", "B2")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(cell)

	ShopCfg := ShopCfg{Cfgs: make(map[int32][]string)}

	// 获取 Sheet1 上所有单元格
	rows, err := f.GetRows("q_shop_goods")
	if err != nil {
		fmt.Println(err)
		return
	}
	for i, row := range rows {
		if i+1 < STARTROWS {
			continue
		}
		k, _ := strconv.Atoi(row[0])
		ShopCfg.Cfgs[int32(k)]= row
	}

	fmt.Println(ShopCfg)
}
