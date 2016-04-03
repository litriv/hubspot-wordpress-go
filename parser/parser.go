package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"code.litriv.com/southerly/migrate/global"
)

func Parse(target interface{}, t string) interface{} {
	b, err := ioutil.ReadFile(filepath.Join(global.BasePath, t, fmt.Sprint(t, ".json")))
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, target)
	if err != nil {
		panic(err)
	}
	return target
}
