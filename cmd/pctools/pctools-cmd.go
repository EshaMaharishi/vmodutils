package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/pointcloud"

	"github.com/erh/vmodutils"
)

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}

func realMain() error {
	logger := logging.NewLogger("cmd-wave")
	ctx := context.Background()

	host := flag.String("host", "", "hostname")
	cmd := flag.String("cmd", "", "command")
	cameraName := flag.String("camera", "", "camera to use")
	out := flag.String("out", "", "output file")
	in := flag.String("in", "", "input file")

	flag.Parse()

	if *cmd == "" {
		return fmt.Errorf("need a cmd")
	}

	if *cmd == "download" {
		if *out == "" {
			return fmt.Errorf("need an 'out'")
		}

		machine, err := vmodutils.ConnectToHostFromCLIToken(ctx, *host, logger)
		if err != nil {
			return err
		}
		defer machine.Close(ctx)

		myCamera, err := camera.FromRobot(machine, *cameraName)
		if err != nil {
			return err
		}

		pc, err := myCamera.NextPointCloud(ctx)
		if err != nil {
			return err
		}

		f, err := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()

		return pointcloud.ToPCD(pc, f, pointcloud.PCDBinary)
	}

	if *cmd == "size" {
		in, err := pointcloud.NewFromFile(*in, "")
		if err != nil {
			return err
		}
		logger.Infof("size: %d", in.Size())
		return nil
	}

	return fmt.Errorf("invalid command [%s]", *cmd)

}
