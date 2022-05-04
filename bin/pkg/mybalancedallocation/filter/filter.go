package filter

import (
	"github.com/thedevsaddam/gojsonq/v2"
	"strconv"
)

const NodeDataJsonPath = "data.result.[0].value.[1]"

//解析HTTP API返回来的数据
func ParseDataToInt(responseString string) (int64, error) { //输入是json数据
	r := gojsonq.New().FromString(responseString).Find(NodeDataJsonPath)
	//获取json中第一个的结果，第二个值，这个要根据自己去curl 一下获得的json数据来看
	return strconv.ParseInt(r.(string), 10, 64) //两个数字代表10进制，int64的数据
}
func ParseDataToFloat(responseString string) (float64, error) { //输入是json数据
	r := gojsonq.New().FromString(responseString).Find(NodeDataJsonPath)
	//获取json中第一个的结果，第二个值，这个要根据自己去curl 一下获得的json数据来看
	return strconv.ParseFloat(r.(string), 64)
}
