package main

import (
	"fmt"
	"strings"
	"github.com/chaosblade-io/chaosblade/exec"
	"flag"
	"context"
	"github.com/chaosblade-io/chaosblade/util"
	"path"
)

var fillDataFile = "chaos_filldisk.log.dat"
var fillDiskMountPoint, fillDiskSize string
var fillDiskStart, fillDiskStop bool

func main() {
	flag.StringVar(&fillDiskMountPoint, "mount-point", "", "mount point of disk")
	flag.StringVar(&fillDiskSize, "size", "", "fill size")
	flag.BoolVar(&fillDiskStart, "start", false, "start fill or not")
	flag.BoolVar(&fillDiskStop, "stop", false, "stop fill or not")

	flag.Parse()

	if fillDiskStart == fillDiskStop {
		printErrAndExit("must specify start or stop operation")
	}
	if fillDiskStart {
		startFill(fillDiskMountPoint, fillDiskSize)
	} else if fillDiskStop {
		stopFill(fillDiskMountPoint)
	} else {
		printErrAndExit("less --start or --stop flag")
	}
}

func startFill(mountPoint, size string) {
	channel := exec.NewLocalChannel()
	ctx := context.Background()
	if mountPoint == "" {
		printErrAndExit("mount-point flag is empty")
	}
	dataFile := path.Join(mountPoint, fillDataFile)
	response := channel.Run(ctx, "dd", fmt.Sprintf(`if=/dev/zero of=%s bs=1b count=1 iflag=fullblock`, dataFile))
	if !response.Success {
		stopFill(mountPoint)
		printErrAndExit(response.Err)
	}
	response = channel.Run(ctx, "nohup",
		fmt.Sprintf(`dd if=/dev/zero of=%s bs=1M count=%s iflag=fullblock >/dev/null 2>&1 &`, dataFile, size))
	if !response.Success {
		stopFill(mountPoint)
		printErrAndExit(response.Err)
	}
	printOutputAndExit(response.Result.(string))
}

func stopFill(mountPoint string) {
	channel := exec.NewLocalChannel()
	ctx := context.Background()
	pids, _ := exec.GetPidsByProcessName(fillDataFile, ctx)

	if pids != nil || len(pids) >= 0 {
		channel.Run(ctx, "kill", fmt.Sprintf("-9 %s", strings.Join(pids, " ")))
	}
	fileName := path.Join(mountPoint, fillDataFile)
	if util.IsExist(fileName) {
		response := channel.Run(ctx, "rm", fmt.Sprintf(`-rf %s`, fileName))
		if !response.Success {
			printErrAndExit(response.Err)
		}
	}
}
