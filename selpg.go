package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

//selpg的参数
type selpg_args struct {
	start_page  int
	end_page    int
	input_file  string
	destination string
	page_len    int //页长
	page_type   int //换页方式
}

var carg selpg_args   //当前输入参数
var progname string 
var count int    //参数个数

//报错信息
func process_command(args []string) {
	//参数数量少于3
	if len(args) < 3 {
		fmt.Fprintf(os.Stderr, "%s: Arguments number error\n", progname)
		os.Exit(1)
	}
	//第一个参数
	if args[1][0] != '-' || args[1][1] != 's' {
		fmt.Fprintf(os.Stderr, "%s: Please use format: -s number\n", progname)
		os.Exit(1)
	}
	//获取开始页
	start, _ := strconv.Atoi(args[1][2:])
	if start < 1 {
		fmt.Fprintf(os.Stderr, "%s: Start page cannot be %d\n", progname, start)
		os.Exit(1)
	}
	carg.start_page = start
	//第二个参数
	if args[2][0] != '-' || args[2][1] != 'e' {
		fmt.Fprintf(os.Stderr, "%s: Please use format: -e number\n", progname)
		os.Exit(1)
	}
	//获取结束页
	end, _ := strconv.Atoi(args[2][2:])
	if end < 1 || end < start {
		fmt.Fprintf(os.Stderr, "%s: End page cannot be %d\n", progname, end)
		os.Exit(1)
	}
	carg.end_page = end

	//其他参数
	index := 3
	for {
		if index > count-1 || args[index][0] != '-' {
			break
		}
		switch args[index][1] {
		case 'l':
			//获取页长
			pl, _ := strconv.Atoi(args[index][2:])
			if pl < 1 {
				fmt.Fprintf(os.Stderr, "%s: The page length cannot smaller than %d\n", progname, pl)
				os.Exit(1)
			}
			carg.page_len = pl
			index++
		case 'f':
			if len(args[index]) > 2 {
				fmt.Fprintf(os.Stderr, "%s: The option should be \"-f\"\n", progname)
				os.Exit(1)
			}
			carg.page_type = 'f'
			index++
		case 'd':
			if len(args[index]) <= 2 {
				fmt.Fprintf(os.Stderr, "%s: -d option requires a output file\n", progname)
				os.Exit(1)
			}
			carg.destination = args[index][2:]
			index++
		default:
			fmt.Fprintf(os.Stderr, "%s: unknown command", progname)
			os.Exit(1)
		}
	}

	if index <= count-1 {
		carg.input_file = args[index]
	}
}

func process_input() {
	var cmd *exec.Cmd
	var in io.WriteCloser
	var out io.ReadCloser
	if carg.destination != "" {
		cmd = exec.Command("bash", "-c", carg.destination)
		in, _ = cmd.StdinPipe()
		out, _ = cmd.StdoutPipe()
		cmd.Start()
	}
	if carg.input_file != "" {
		inf, err := os.Open(carg.input_file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		line_count := 1
		page_count := 1
		fin := bufio.NewReader(inf)
		for {
			line, _, err := fin.ReadLine() 			//按行读取
			if err != io.EOF && err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if err == io.EOF {
				break
			}
			if page_count >= carg.start_page && page_count <= carg.end_page {
				if carg.destination == "" {
					fmt.Println(string(line))
				} else {
					fmt.Fprintln(in, string(line))
				}
			}
			line_count++
			if carg.page_type == 'l' {
				if line_count > carg.page_len {  //按给定行数换页
					line_count = 1
					page_count++
				}
			} else {
				if string(line) == "\f" { //按分页符换页
					page_count++
				}
			}
		}
		if carg.destination != "" {
			in.Close()
			bytes, err := ioutil.ReadAll(out)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Print(string(bytes))
			//等待退出
			cmd.Wait()
		}
	} else {
		//从标准输入读取内容
		ns := bufio.NewScanner(os.Stdin)
		line_count := 1
		page_count := 1
		out := ""

		for ns.Scan() {
			line := ns.Text()
			line += "\n"
			if page_count >= carg.start_page && page_count <= carg.end_page {
				out += line
			}
			line_count++
			if carg.page_type == 'l' {
				if line_count > carg.page_len {
					line_count = 1
					page_count++
				}
			} else {
				if string(line) == "\f" {
					page_count++
				}
			}
		}
		if carg.destination == "" {
			fmt.Print(out)
		} else {
			fmt.Fprint(in, out)
			in.Close()
			bytes, err := ioutil.ReadAll(out)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Print(string(bytes))
			//等待退出
			cmd.Wait()
		}
	}
}

func main() {
	args := os.Args
	carg.start_page = 1 //默认参数值
	carg.end_page = 1
	carg.page_len = 10 
	carg.page_type = 'l'
	carg.input_file = ""
	carg.destination = ""
	count = len(args)
	process_command(args)
	process_input()
}