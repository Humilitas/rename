package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	CloseDelay int
	ReplaceMap map[string]string
	Log        bool
}

func main() {

	err, config := readConfig()
	handleErr(err)

	rename(config)

	autoClose(config)
}

func autoClose(config Config) {
	delay := config.CloseDelay

	// 提示信息
	fmt.Printf("操作完成，%d 秒后程序自动关闭。\n", delay)
	go func() {
		for delay > 0 {
			fmt.Printf("%d..\r", delay)
			delay--
			time.Sleep(time.Second)
		}
	}()
	time.Sleep(time.Second * time.Duration(delay))
}

func rename(config Config) {
	replaceMap := config.ReplaceMap
	shouldLog := config.Log
	// 读取当前文件夹
	dir, err := ioutil.ReadDir(".")
	handleErr(err)

	// 创建 log 文件
	var logFile *os.File
	if shouldLog {
		logFile, err = os.Create("replacement-record.txt")
		handleErr(err)
		defer func() {
			err = logFile.Close()
			handleErr(err)
		}()
	}

	// 改名
	for _, info := range dir {

		if !info.IsDir() {
			destFileName := info.Name()
			for oldStr, newStr := range replaceMap {
				destFileName = strings.Replace(destFileName, oldStr, newStr, 1)
			}

			// 若文件名未发生替换，则表明无需改名
			if destFileName != info.Name() {
				file, err := os.Open(info.Name())
				handleErr(err)

				// 使用临时文件转存原文件
				tempFile, err := ioutil.TempFile(".", "*-"+info.Name())
				handleErr(err)
				_, err = io.Copy(tempFile, file)
				handleErr(err)
				// 先关闭临时文件，以保存其内容。
				err = tempFile.Close()
				handleErr(err)
				tempFile, err = os.Open(tempFile.Name())
				handleErr(err)


				// 删除原文件 | 由于文件名不区分大小写，所以为处理大小写转换的情况，必须将原文件删除
				err = file.Close()
				handleErr(err)
				err = os.Remove(info.Name())
				handleErr(err)
				fmt.Printf("%s 已删除\n", info.Name())

				// 创建新文件
				create, err := os.Create(destFileName)
				handleErr(err)
				_, err = io.Copy(create, tempFile)
				handleErr(err)

				err = create.Close()
				handleErr(err)
				err = tempFile.Close()
				handleErr(err)
				err = os.Remove(tempFile.Name())
				handleErr(err)

				fmt.Printf("%s -> %s\n", info.Name(), destFileName)

				// log
				if shouldLog {
					logString := fmt.Sprintf("%s | %s 已删除, %s -> %s\n", time.Now().Format("2006-01-02 15:04:05"), info.Name(), info.Name(), destFileName)
					_, err := logFile.WriteString(logString)
					handleErr(err)
				}
			}

		}

	}
}

func readConfig() (error, Config) {
	// 读取配置
	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println("没有读取到配置文件，将在当前目录下自动生成默认配置文件 config.json")

		//	无配置则生成默认配置
		configFile, err := os.Create("config.json")
		handleErr(err)
		if err != nil {
			return err, Config{}
		}
		defaultConfig := Config{
			CloseDelay: 5,
			ReplaceMap: map[string]string{"old": "new"},
			Log:        true,
		}
		jsonString, err := json.MarshalIndent(defaultConfig, "", "\t")
		if err != nil {
			return err, Config{}
		}
		_, err = configFile.Write(jsonString)
		if err != nil {
			return err, Config{}
		}

		fmt.Println("请修改配置后再次执行程序。程序将在 5 秒后自动关闭")
		time.Sleep(time.Second * 5)
		os.Exit(1)
	}

	defer func() {
		err = configFile.Close()
		handleErr(err)
	}()

	byteValue, err := ioutil.ReadAll(configFile)
	handleErr(err)

	var config Config
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return err, Config{}
	}

	return err, config
}

func handleErr(err error) {
	if err != nil {
		log.Println("发生错误：" + err.Error())
		log.Println("5 秒后自动关闭")
		time.Sleep(time.Second * 5)

		os.Exit(1)
	}
}
