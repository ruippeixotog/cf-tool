package cmd

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/ruippeixotog/cf-tool/client"
	"github.com/ruippeixotog/cf-tool/config"
)

// Parse command
func Parse() (err error) {
	cfg := config.Instance
	cln := client.Instance
	info := Args.Info
	tpls := applyInfo(cfg.FileTemplates, info)

	source := ""
	ext := ""
	if cfg.GenAfterParse {
		if len(cfg.Template) == 0 {
			return errors.New("You have to add at least one code template by `cf config`")
		}
		path := cfg.Template[cfg.Default].Path
		ext = filepath.Ext(path)
		if source, err = readTemplateSource(path, cln); err != nil {
			return
		}
	}
	work := func() error {
		problems, err := cln.Parse(info, tpls.Input, tpls.Answer)
		if err != nil {
			return err
		}
		if cfg.GenAfterParse {
			for _, prob := range problems {
				path := strings.ReplaceAll(tpls.Code, "$%prob%$", prob)
				gen(source, path, ext)
			}
		}
		return nil
	}
	if err = work(); err != nil {
		if err = loginAgain(cln, err); err == nil {
			err = work()
		}
	}
	return
}
