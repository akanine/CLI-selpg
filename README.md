> CLI（Command Line Interface）实用程序是Linux下应用开发的基础。Linux提供了cat、ls、copy等命令与操作系统交互；go语言提供一组实用程序完成从编码、编译、库管理、产品发布全过程支持。在开发领域，CLI在编程、调试、运维、管理中提供了图形化程序不可替代的灵活性与效率。
## Selpg功能概述
- selpg 是从文本输入选择页范围的实用程序。它从标准输入或从作为命令行参数给出的文件名读取文本输入，允许用户指定来自该输入并随后将被输出的页面范围。在很多情况下，可以避免打印浪费，节省了资源。
- 如用如下的命令，将需要的页面打印出来




```csharp
$ selpg -s 1 -e 2 input.txt //将input.txt的第一页到第二页输出到屏幕
```

- selpg 是以在 Linux 中创建命令的事实上的约定为模型创建的，这些约定包括：

	- 独立工作
	- 在命令管道中作为组件工作（通过读取标准输入或文件名参数，	以及写至标准输出和标准错误）
	- 接受修改其行为的命令行选项

## Selpg程序逻辑
**“-sNumber”和“-eNumber”强制选项：**

selpg 要求用户用两个命令行参数“-sNumber”（例如，“-s10”表示从第 10 页开始）和“-eNumber”（例如，“-e20”表示在第 20 页结束）指定要抽取的页面范围的起始页和结束页。selpg 对所给的页号进行合理性检查；换句话说，它会检查两个数字是否为有效的正整数以及结束页是否不小于起始页。这两个选项，“-sNumber”和“-eNumber”是强制性的，而且必须是命令行上在命令名 selpg 之后的头两个参数
- 程序实现：

```go
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
```

**“-lNumber”和“-f”可选选项：**
selpg 可以处理两种输入文本：

类型 1：该类文本的页行数固定。这是缺省类型，因此不必给出选项进行说明。也就是说，如果既没有给出“-lNumber”也没有给出“-f”选项，则 selpg 会理解为页有固定的长度（每页 72 行）。

```go
$ selpg -s10 -e20 -l66
```
类型 2：该类型文本的页由 ASCII 换页字符（十进制数值为 12，在 C 中用“\f”表示）定界。该格式与“每页行数固定”格式相比的好处在于，当每页的行数有很大不同而且文件有很多页时，该格式可以节省磁盘空间。在含有文本的行后面，类型 2 的页只需要一个字符 ― 换页 ― 就可以表示该页的结束。打印机会识别换页符并自动根据在新的页开始新行所需的行数移动打印头。

```go
$ selpg -s10 -e20 -f
```
- 程序实现：

```go
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
```

```go
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
```
**“-dDestination”可选选项：**

selpg 还允许用户使用“-dDestination”选项将选定的页直接发送至打印机。这里，“Destination”应该是 lp 命令“-d”选项（请参阅“man lp”）可接受的打印目的地名称。该目的地应该存在 ― selpg 不检查这一点。在运行了带“-d”选项的 selpg 命令后，若要验证该选项是否已生效，请运行命令“lpstat -t”。该命令应该显示添加到“Destination”打印队列的一项打印作业。如果当前有打印机连接至该目的地并且是启用的，则打印机应打印该输出。这一特性是用 popen() 系统调用实现的，该系统调用允许一个进程打开到另一个进程的管道，将管道用于输出或输入。

```go
selpg -s10 -e20 -dlp1
```
该命令将选定的页作为打印作业发送至 lp1 打印目的地

- 程序实现：

```go
case 'd':
			if len(args[index]) <= 2 {
				fmt.Fprintf(os.Stderr, "%s: -d option requires a output file\n", progname)
				os.Exit(1)
			}
			carg.destination = args[index][2:]
			index++
```

```go
	if carg.destination != "" {
		cmd = exec.Command("bash", "-c", carg.destination)
		in, _ = cmd.StdinPipe()
		out, _ = cmd.StdoutPipe()
		cmd.Start()
	}
```
- 程序的参数结构体如下：

```go
//selpg的参数
type selpg_args struct {
	start_page  int
	end_page    int
	input_file  string
	destination string
	page_len    int //页长
	page_type   int //换页方式
}
```
具体实现见完整版代码吧

## 测试Selpg
1. 将input.txt文件第一页到第二页的内容输出到屏幕

```go
$ selpg -s1 -e2 input.txt
```

![在这里插入图片描述](https://img-blog.csdnimg.cn/20191007202659837.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L2FrYW5pbmU=,size_16,color_FFFFFF,t_70)
2.将input.txtx第一页的内容输出到屏幕
```go
$ selpg -s1 -e1 < input.txt
```
![在这里插入图片描述](https://img-blog.csdnimg.cn/20191007202438945.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L2FrYW5pbmU=,size_16,color_FFFFFF,t_70)
 3. 将input.txt中第二页到第四页的内容输出到output.txt

```go
$ selpg -s2 -e4 input.txt > output.txt
```
 ![在这里插入图片描述](https://img-blog.csdnimg.cn/20191007203729174.png)
![在这里插入图片描述](https://img-blog.csdnimg.cn/20191007202902540.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L2FrYW5pbmU=,size_16,color_FFFFFF,t_70)
4. 由换页符决定换页，将input.txt第一页的内容输出到屏幕

```go
$ selpg -s1 -e1 -f input.txt
```
![在这里插入图片描述](https://img-blog.csdnimg.cn/20191007202835449.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L2FrYW5pbmU=,size_16,color_FFFFFF,t_70)
5. 将第一页发送至命令cat，将input.txt第一页的前五行输出到屏幕

```go
$ selpg -s1 -e1 -l5 -dcat input.txt
```

![在这里插入图片描述](https://img-blog.csdnimg.cn/20191007203509932.png)
6. 将第二页至第三页的前四行写入output.txt和error.txt
![在这里插入图片描述](https://img-blog.csdnimg.cn/2019100720380230.png)![在这里插入图片描述](https://img-blog.csdnimg.cn/20191007203923998.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L2FrYW5pbmU=,size_16,color_FFFFFF,t_70)
7. 将错误消息显示至error.txt
![在这里插入图片描述](https://img-blog.csdnimg.cn/20191007204046839.png)
![在这里插入图片描述](https://img-blog.csdnimg.cn/20191007204518369.png)

代码[传送门]()
