package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/docopt/docopt-go"

	"github.com/fatih/color"
	"github.com/ruippeixotog/cf-tool/client"
	"github.com/ruippeixotog/cf-tool/config"
	"github.com/ruippeixotog/cf-tool/util"
)

// Eval opts
func Eval(opts docopt.Opts) error {
	Args = &ParsedArgs{}
	opts.Bind(Args)
	if err := parseArgs(opts); err != nil {
		return err
	}
	if Args.Config {
		return Config()
	} else if Args.Submit {
		return Submit()
	} else if Args.List {
		return List()
	} else if Args.Parse {
		return Parse()
	} else if Args.Gen {
		return Gen()
	} else if Args.Test {
		return Test()
	} else if Args.Watch {
		return Watch()
	} else if Args.Open {
		return Open()
	} else if Args.Stand {
		return Stand()
	} else if Args.Sid {
		return Sid()
	} else if Args.Race {
		return Race()
	} else if Args.Pull {
		return Pull()
	} else if Args.Clone {
		return Clone()
	} else if Args.Upgrade {
		return Upgrade()
	}
	return nil
}

func applyInfo(tpls config.FileTemplates, info client.Info) config.FileTemplates {
	basePath := info.Path()

	apply := func(str string) string {
		if info.ProblemID != "" {
			str = strings.ReplaceAll(str, "$%prob%$", info.ProblemID)
		}
		return filepath.Join(basePath, str)
	}

	return config.FileTemplates{
		Input:  apply(tpls.Input),
		Answer: apply(tpls.Answer),
		Code:   apply(tpls.Code),
	}
}

func applyReplacement(tpls config.FileTemplates, key, value string) config.FileTemplates {
	return config.FileTemplates{
		Input:  strings.ReplaceAll(tpls.Input, "$%"+key+"%$", value),
		Answer: strings.ReplaceAll(tpls.Answer, "$%"+key+"%$", value),
		Code:   strings.ReplaceAll(tpls.Code, "$%"+key+"%$", value),
	}
}

func getSampleID(tpls config.FileTemplates) (samples []string) {
	paths, err := ioutil.ReadDir(filepath.Dir(tpls.Input))
	if err != nil {
		return
	}
	reg := regexp.MustCompile(strings.ReplaceAll(filepath.Base(tpls.Input), "$%i%$", `(\d+)`))
	for _, path := range paths {
		name := path.Name()
		tmp := reg.FindSubmatch([]byte(name))
		if tmp != nil {
			idx := string(tmp[1])
			ans := strings.ReplaceAll(tpls.Answer, "$%i%$", fmt.Sprint(idx))
			if _, err := os.Stat(ans); err == nil {
				samples = append(samples, idx)
			}
		}
	}
	return
}

// CodeList Name matches some template suffix, index are template array indexes
type CodeList struct {
	Name  string
	Index []int
}

func getCode(filename string, templates []config.CodeTemplate) (codes []CodeList, err error) {
	mp := make(map[string][]int)
	for i, temp := range templates {
		suffixMap := map[string]bool{}
		for _, suffix := range temp.Suffix {
			if _, ok := suffixMap[suffix]; !ok {
				suffixMap[suffix] = true
				sf := "." + suffix
				mp[sf] = append(mp[sf], i)
			}
		}
	}

	if info, _ := os.Stat(filename); info != nil && !info.IsDir() {
		ext := filepath.Ext(filename)
		if idx, ok := mp[ext]; ok {
			return []CodeList{{filename, idx}}, nil
		}
		return nil, fmt.Errorf("%v can not match any template. You could add a new template by `cf config`", filename)
	} else {
		for ext, idx := range mp {
			fullPath := filename + ext
			if info, _ := os.Stat(fullPath); info != nil && !info.IsDir() {
				codes = append(codes, CodeList{fullPath, idx})
			}
		}
	}
	return codes, nil
}

func getOneCode(filename string, templates []config.CodeTemplate) (name string, index int, err error) {
	codes, err := getCode(filename, templates)
	if err != nil {
		return
	}
	if len(codes) < 1 {
		return "", 0, errors.New("Cannot find any code.\nMaybe you should add a new template by `cf config`")
	}
	if len(codes) > 1 {
		color.Cyan("There are multiple files can be selected.")
		for i, code := range codes {
			fmt.Printf("%3v: %v\n", i, code.Name)
		}
		i := util.ChooseIndex(len(codes))
		codes[0] = codes[i]
	}
	if len(codes[0].Index) > 1 {
		color.Cyan("There are multiple languages match the file.")
		for i, idx := range codes[0].Index {
			fmt.Printf("%3v: %v\n", i, client.Langs[templates[idx].Lang])
		}
		i := util.ChooseIndex(len(codes[0].Index))
		codes[0].Index[0] = codes[0].Index[i]
	}
	return codes[0].Name, codes[0].Index[0], nil
}

func loginAgain(cln *client.Client, err error) error {
	if err != nil && err.Error() == client.ErrorNotLogged {
		color.Red("Not logged. Try to login\n")
		err = cln.Login()
	}
	return err
}
