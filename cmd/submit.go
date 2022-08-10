package cmd

import (
	"io/ioutil"

	"github.com/ruippeixotog/cf-tool/client"
	"github.com/ruippeixotog/cf-tool/config"
)

// Submit command
func Submit() (err error) {
	cln := client.Instance
	cfg := config.Instance
	info := Args.Info
	tpls := applyInfo(cfg.FileTemplates, info)

	basePath := Args.File
	if Args.File == "" {
		basePath = tpls.Code
	}
	filename, index, err := getOneCode(basePath, cfg.Template)
	if err != nil {
		return
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	source := string(bytes)

	lang := cfg.Template[index].Lang
	if err = cln.Submit(info, lang, source); err != nil {
		if err = loginAgain(cln, err); err == nil {
			err = cln.Submit(info, lang, source)
		}
	}
	return
}
