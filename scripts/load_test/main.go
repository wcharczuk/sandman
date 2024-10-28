package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v3"
	"go.charczuk.com/sdk/cliutil"
	"go.charczuk.com/sdk/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "sandman/proto/v1"
	"sandman/sandctl/viewmodel"
)

func main() {
	if err := command.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

var command = &cli.Command{
	Name:  "load_test",
	Usage: "Load test the sandman server",
	Flags: DefaultClientFlags(
		&cli.IntFlag{
			Name:  "count",
			Usage: "The number of timers to create",
			Value: 1024,
		},
		&cli.StringMapFlag{
			Name:    "label",
			Aliases: []string{"l"},
		},
		&cli.UintFlag{
			Name:    "priority",
			Aliases: []string{"p"},
			Value:   1,
		},
		&cli.DurationFlag{
			Name:  "due-in",
			Value: 5 * time.Minute,
		},
		&cli.StringFlag{
			Name:  "hook-url",
			Value: "http://localhost:8081",
		},
		&cli.StringFlag{
			Name:  "hook-method",
			Value: "POST",
		},
		&cli.StringMapFlag{
			Name: "hook-header",
		},
		&cli.StringFlag{
			Name: "hook-body",
		},
	),
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var count = int(cmd.Int("count"))
		c, err := createClient(cmd)
		if err != nil {
			return fmt.Errorf("load test; create client: %w", err)
		}

		var hookBodyData string
		if hookBodyPath := cmd.String("hook-body"); hookBodyPath != "" {
			rawHookBodyData, err := cliutil.FileOrStdin(hookBodyPath)
			if err != nil {
				return fmt.Errorf("cannot read args data; %w", err)
			}
			hookBodyData = base64.StdEncoding.EncodeToString(rawHookBodyData)
		}
		timer := viewmodel.Timer{
			Labels:   cmd.StringMap("label"),
			Priority: uint32(cmd.Uint("priority")),
			Hook: viewmodel.Hook{
				URL:     cmd.String("hook-url"),
				Method:  cmd.String("hook-method"),
				Headers: cmd.StringMap("hook-headers"),
				Body:    hookBodyData,
			},
		}
		if dueIn := cmd.Duration("due-in"); dueIn > 0 {
			timer.DueUTC = time.Now().UTC().Add(dueIn)
		} else {
			timer.DueUTC = time.Now().UTC().Add(time.Minute)
		}
		for x := 0; x < count; x++ {
			timer.Name = uuid.V4().String()
			timer.ShardKey = fmt.Sprintf("uid_%04d", x)
			_, err := c.CreateTimer(ctx, timer.ToProto())
			if err != nil {
				return fmt.Errorf("load test; create timer failed: %w", err)
			}
		}
		return nil
	},
}

func DefaultFlags(moreFlags ...cli.Flag) []cli.Flag {
	return append([]cli.Flag{
		&cli.BoolFlag{
			Name: "debug",
		},
	}, moreFlags...)
}

func DefaultClientFlags(moreFlags ...cli.Flag) []cli.Flag {
	return DefaultFlags(append([]cli.Flag{
		&cli.StringFlag{
			Name:  "address",
			Value: "localhost:8833",
		},
		&cli.StringFlag{
			Name:  "authority",
			Value: "sandman-srv.local",
		},
	}, moreFlags...)...)
}

func createClient(cmd *cli.Command) (v1.TimersClient, error) {
	addr := cmd.String("address")
	authority := cmd.String("authority")
	c, err := grpc.NewClient(addr, grpc.WithAuthority(authority), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return v1.NewTimersClient(c), nil
}
