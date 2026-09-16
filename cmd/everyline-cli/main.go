package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"git.qtech.cn/ai/everyline-cli/internal/app"
	"git.qtech.cn/ai/everyline-cli/internal/update"
)

// main 启动 everyline-cli，并将应用层返回的稳定退出码交给操作系统。
// 入参：无。
// 返回值：无；进程通过 os.Exit 返回退出码。
func main() {
	if len(os.Args) > 1 && os.Args[1] == "__everyline-cli-replace" {
		if err := update.RunDeferredReplacement(os.Args[2:]); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		_, _ = fmt.Fprintln(os.Stderr, "Windows 更新助手已完成二进制替换。")
		return
	}
	// 将 Ctrl+C 传递给请求和探测 helper，让 app.Run 完成子进程回收后退出。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-ctx.Done()
		stop() // stdin 等不可取消的等待仍可用第二次 Ctrl+C 强制退出。
	}()
	exitCode := app.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	stop()
	os.Exit(exitCode)
}
