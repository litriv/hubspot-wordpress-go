package writer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"code.litriv.com/southerly/migrate/global"
)

const wrapper = `<?php
	include_once(plugin_dir_path( __FILE__ ) . 'upload.php');
	function {{funcName}}() {
		{{if globals}}global {{globals}};{{end}}
    return array(
			{{range .}}
				{{template "inner" .}}		
			{{end}}
    );
	}
?>
`

func Execute(name string, content interface{}, funcMap template.FuncMap) {

	// Setup func map that will apply to all templates
	initFuncMap := map[string]interface{}{
		"globals": func() string { return "" },
		"escapeApos": func(s string) string {
			return strings.Replace(s, "'", "\\'", -1)
		},
		"formatDate": func(ms int64) string {
			msToTime := func(ms string) (time.Time, error) {
				msInt, err := strconv.ParseInt(ms, 10, 64)
				if err != nil {
					return time.Time{}, err
				}
				return time.Unix(0, msInt*int64(time.Millisecond)), nil
			}
			t, err := msToTime(strconv.FormatInt(ms, 10))
			if err != nil {
				panic(err)
			}
			return t.Format("2006-01-02 15:04:05")
		},
	}

	// Read template file
	tf, err := ioutil.ReadFile(filepath.Join(global.BasePath, name, name+".tmpl"))

	// Parse template file, creating template
	// The order that the func maps gets added are important and should be preserved (globals first)
	t := template.Must(template.New("wrapper").Funcs(initFuncMap).Funcs(funcMap).Parse(wrapper))
	_, err = t.Parse(string(tf))
	if err != nil {
		panic(err)
	}
	fp := filepath.Join(global.TargetPath, name+".php")
	// Open file
	f, err := os.Create(fp)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	// Execute template
	err = t.Execute(f, content)
	if err != nil {
		panic(err)
	}
}
