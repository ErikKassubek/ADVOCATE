package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var basePath, _ = os.Getwd()
var baseDir = filepath.Dir(basePath)

var mode string
var prog string
var test string
var fuzzingMode string

// TODO: add in other progs
var progNames = []string{
	"argo-cd",
	"bleve",
	"bosun",
	"caddy",
	"dns",
	"flannel",
	"frp",
	"gin",
	"fiber",
	"gorums",
	"grpc",
	"hugo",
	"kubernetes",
	"nsq",
	"ocatant",
	"ollama",
	"pholcus",
	"syncthing",
	"terraform",
}

func main() {
	commandLineArgs()

	switch prog {
	case "all":
		err := runGoBench()
		if err != nil {
			panic(err)
		}
		for _, name := range progNames {
			prog = name
			err = runGithubProg()
			if err != nil {
				panic(err)
			}
		}
	case "gobench":
		err := runGoBench()
		if err != nil {
			panic(err)
		}
	default:
		err := runGithubProg()
		if err != nil {
			panic(err)
		}
	}
}

func commandLineArgs() {
	progStrings := ""
	for _, prog := range progNames {
		progStrings += "\n\t- '" + prog + "'"
	}

	flag.StringVar(&mode, "mode", "fuzzing", "Choose mode from 'record', 'analysis', 'fuzzing'. Default: 'fuzzing'.")
	flag.StringVar(&prog, "prog", "gobench", "Select program to run on. Possible values are:"+
		"\n\t- 'all' (run all programs one after another)"+
		"\n\t- 'gobench'"+
		progStrings+
		"Default: 'gobench'")
	flag.StringVar(&test, "test", "", "Select test to run. To run all tests, do not set")
	flag.StringVar(&fuzzingMode, "fuzzingMode", "gfuzzpie", "Select fuzzing mode from 'gfuzz', 'gopie', 'gfuzzpie'. Default: 'gfuzzpie'")

	flag.Parse()

	if prog != "all" && prog != "gobench" && !contains(progNames, prog) {
		panic("Unknown program name: " + prog)
	}

	if mode != "record" && mode != "analysis" && mode != "fuzzing" {
		panic("Unknown mode name: " + mode)
	}

	switch fuzzingMode {
	case "gfuzz":
		fuzzingMode = "GFuzz"
	case "gopie":
		fuzzingMode = "GoPie"
	case "gfuzzpie":
		fuzzingMode = "GoCR"
	default:
		panic("Unknown fuzzing mode name: " + fuzzingMode)
	}
}

func runProg(path string) error {
	fmt.Println("Run: ", path)

	var cmd *exec.Cmd

	timeout := "90"
	if prog == "gobench" {
		timeout = "10"
	}

	if mode == "fuzzing" {
		if test == "" {
			cmd = exec.Command("./gocct", "fuzzing", "-mode", fuzzingMode, "-path", path, "-timeout", timeout, "-stats", "-time")
		} else {
			cmd = exec.Command("./gocct", "fuzzing", "-mode", fuzzingMode, "-path", path, "-exec", test, "-timeout", timeout, "-stats", "-time")
		}
	} else {

		if test == "" {
			cmd = exec.Command("./gocct", mode, "-path", path, "-timeout", timeout, "-stats", "-time")
		} else {
			cmd = exec.Command("./gocct", mode, "-path", path, "-exec", test, "-timeout", timeout, "-stats", "-time")
		}
	}
	cmd.Dir = "../../goCCT"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println(cmd.String())

	return cmd.Run()
}

func runGithubProg() error {
	name, err := cloneRepo(prog)
	if err != nil {
		return err
	}

	path := filepath.Join(baseDir, name)

	err = prepare(path, name)
	if err != nil {
		return err
	}

	return runProg(path)
}

func cloneRepo(prog string) (string, error) {
	link := ""
	branch := ""

	name := prog

	switch prog {
	// TODO: add in other progs
	// TODO: If folder name is different from prog name, set as output
	case "argo-cd":
		link = "git@github.com:argoproj/argo-cd.git"
		branch = "v3.1.0"
	case "bleve":
		link = "git@github.com:blevesearch/bleve"
		branch = "v2.5.3"
	case "bosun":
		link = "git@github.com:bosun-monitor/bosun"
		branch = "0.8.0-preview"
	case "caddy":
		link = "git@github.com:caddyserver/caddy"
		branch = "v2.10.0"
	case "dns":
		link = "git@github.com:miekg/dns.git"
		branch = "v1.1.50"
	case "flannel":
		link = "git@github.com:flannel-io/flannel.git"
		branch = "v0.20.2"
	case "frp":
		link = "git@github.com:fatedier/frp.git"
		branch = "v0.36.0"
	case "gin":
		link = "git@github.com:gin-gonic/gin.git"
		branch = "v1.10.1"
	case "fiber":
		link = "git@github.com:gofiber/fiber.git"
		branch = "v2.40.1"
	case "gorums":
		link = "git@github.com:relab/gorums.git"
		branch = "v0.7.0"
	case "grpc":
		link = "git@github.com:grpc/grpc-go.git"
		branch = "Release 1.51.0"
		name = "grpc-go"
	case "hugo":
		link = "git@github.com:gohugoio/hugo.git"
		branch = "v0.148.2"
	case "kubernetes":
		link = "git@github.com:kubernetes/kubernetes.git"
		branch = "v1.25.5"
	case "nsq":
		link = "git@github.com:nsqio/nsq.git"
		branch = "1.3.0"
	case "octant":
		link = "git@github.com:vmware-archive/octant.git"
		branch = "v0.25.1"
	case "ollama":
		link = "git@github.com:ollama/ollama.git"
		branch = "v0.11.4"
	case "pholcus":
		link = "git@github.com:andeya/pholcus.git"
		branch = "v1.3.4"
	case "syncthing":
		link = "git@github.com:syncthing/syncthing.git"
		branch = "v1.22.1"
	case "terraform":
		link = "git@github.com:hashicorp/terraform.git"
		branch = "v1.12.2"
	case "zinx":
		link = "git@github.com:aceld/zinx.git"
		branch = "v1.2.7"
	default:
		return "", fmt.Errorf("Unknown program name: %s", prog)
	}

	cmd := exec.Command("git", "clone", link, "--branch", branch)
	cmd.Dir = ".."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return name, cmd.Run()
}

func prepare(path, name string) error {
	// create go.mod if not exists
	hasGM, err := hasGoMod(path)
	if err != nil {
		return err
	}
	if !hasGM {
		cmd := exec.Command("go", "mod", "init", name)
		cmd.Dir = path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Run()
	}

	// run tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func hasGoMod(dir string) (bool, error) {
	goMod := filepath.Join(dir, "go.mod")

	info, err := os.Stat(goMod)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !info.IsDir(), nil
}

func runGoBench() error {
	path := baseDir + "/gobench"
	return runProg(path)
}

func contains(l []string, v string) bool {
	for _, e := range l {
		if e == v {
			return true
		}
	}

	return false
}
