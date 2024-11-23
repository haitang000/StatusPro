package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

// SystemStatus 定义了系统状态的结构体
type SystemStatus struct {
	CPUModel  string  `json:"cpu_model"`
	CPUUsage  float64 `json:"cpu_usage"`
	RAMUsage  float64 `json:"ram_usage"`
	DiskUsage float64 `json:"disk_usage"`
}

var server *http.Server

func main() {
	// 打印提示信息
	fmt.Println("Welcome to StatusPro. Type 'help' for a list of commands.")

	// 启动HTTP服务器
	go startHTTPServer()

	// 创建一个扫描器来读取命令行输入
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		scanner.Scan()
		input := scanner.Text()

		switch strings.ToLower(input) {
		case "status":
			printSystemStatusToConsole()
		case "stop":
			log.Println("Stopping the program...")
			if server != nil {
				// 关闭HTTP服务器
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := server.Shutdown(ctx); err != nil {
					log.Println("Error shutting down HTTP server:", err)
				} else {
					log.Println("HTTP server shut down gracefully.")
				}
			}
			return
		case "help":
			printHelp()
		case "about":
			printAbout()
		default:
			fmt.Println("Unknown command. Please enter 'status', 'stop', 'help', or 'about'.")
		}
	}
}

func openBrowser(url string) {
	var err error
	switch os := runtime.GOOS; os {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux", "freebsd":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		// 使用start命令打开浏览器
		err = exec.Command("cmd", "/c", "start", url).Start()
	default:
		fmt.Printf("OS %s not supported\n", os)
	}
	if err != nil {
		log.Fatal("Error opening browser: ", err)
	}
}

func printSystemStatusToConsole() {
	var status SystemStatus

	// 获取CPU信息
	cpuInfo, err := cpu.Info()
	if err != nil {
		log.Fatal(err)
	}
	status.CPUModel = cpuInfo[0].ModelName

	// 获取CPU占用率
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		log.Fatal(err)
	}
	status.CPUUsage = cpuPercent[0]

	// 获取RAM使用率
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal(err)
	}
	status.RAMUsage = memInfo.UsedPercent

	// 获取磁盘使用率
	diskInfo, err := disk.Usage("/")
	if err != nil {
		log.Fatal(err)
	}
	status.DiskUsage = diskInfo.UsedPercent

	// 直接打印系统状态信息
	fmt.Printf("CPU Model: %s\n", status.CPUModel)
	fmt.Printf("CPU Usage: %.2f%%\n", status.CPUUsage)
	fmt.Printf("RAM Usage: %.2f%%\n", status.RAMUsage)
	fmt.Printf("Disk Usage: %.2f%%\n", status.DiskUsage)
}

func startHTTPServer() {
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		var status SystemStatus

		// 获取CPU占用率
		cpuPercent, err := cpu.Percent(0, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		status.CPUUsage = cpuPercent[0]

		// 获取RAM使用率
		memInfo, err := mem.VirtualMemory()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		status.RAMUsage = memInfo.UsedPercent

		// 获取磁盘使用率
		diskInfo, err := disk.Usage("/")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		status.DiskUsage = diskInfo.UsedPercent

		// 获取CPU型号
		cpuInfo, err := cpu.Info()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		status.CPUModel = cpuInfo[0].ModelName

		// 将系统状态编码为JSON并返回
		jsonStatus, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonStatus)
	})

	// Serve index.html for the root URL
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	server = &http.Server{Addr: ":8080"}
	log.Println("Server starting on port 8080...")
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server ListenAndServe error: ", err)
		}
	}()
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  status  - Print system status")
	fmt.Println("  stop    - Stop the program")
	fmt.Println("  help    - Show this help message")
	fmt.Println("  about   - Show information about the program")
}

func printAbout() {
	fmt.Println("StatusPro v1.0")
	fmt.Println("A simple system status monitoring tool.")
}
