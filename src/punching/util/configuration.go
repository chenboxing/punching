package util

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
)

// LoadTomlFile 加载配置文件
func LoadTomlFile(fileName string) (sections map[string]toml.Primitive, m toml.MetaData, err error) {
	// 判断配置文件是否存在
	if _, err = os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("configuration file %s does not exist.\r\n", fileName)
		} else {
			err = fmt.Errorf("configuration file %s execption:%s\r\n", fileName, err.Error())
		}
		return
	}

	// 加载配置文件

	var file toml.Primitive
	var meta toml.MetaData

	if meta, err = toml.DecodeFile(fileName, &file); err != nil {
		err = fmt.Errorf("load configuration file %s failed:%s", fileName, err.Error())
	} else {

		err = meta.PrimitiveDecode(file, &sections)
	}
	m = meta
	return
}

// DecodeSection 解码一个节点的配置信息
func DecodeSection(filename, name string, v interface{}) (err error) {

	sections, meta, err := LoadTomlFile(filename)
	if err != nil {
		return
	}

	if section, ok := sections[name]; ok {
		return meta.PrimitiveDecode(section, v)
	}

	return nil
}
