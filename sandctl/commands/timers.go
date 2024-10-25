package commands

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	v1 "sandman/proto/v1"
	"sandman/sandctl/viewmodel"
	"time"

	"github.com/urfave/cli/v3"
	"go.charczuk.com/sdk/cliutil"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

func Timers() *cli.Command {
	timers := &cli.Command{
		Name:    "timers",
		Aliases: []string{"timer"},
		Usage:   "Control sandman timers",
		Commands: []*cli.Command{
			timerGenerate(),
			timerCreate(),
			timerList(),
			timerGet(),
			timerDelete(),
		},
	}
	return timers
}

func timerGenerate() *cli.Command {
	return &cli.Command{
		Name:    "generate",
		Aliases: []string{"gen", "g"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Aliases:  []string{"n"},
				Required: true,
			},
			&cli.TimestampFlag{
				Name: "due-utc",
			},
			&cli.DurationFlag{
				Name: "due-in",
			},
			&cli.StringMapFlag{
				Name:    "label",
				Aliases: []string{"l"},
			},
			&cli.StringFlag{
				Name:     "hook-url",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "hook-method",
				Required: true,
			},
			&cli.StringMapFlag{
				Name: "hook-header",
			},
			&cli.StringFlag{
				Name:     "hook-body",
				Required: true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			rawHookBodyData, err := cliutil.FileOrStdin(cmd.String("hook-body"))
			if err != nil {
				return fmt.Errorf("cannot read args data; %w", err)
			}

			hookBodyData := base64.StdEncoding.EncodeToString(rawHookBodyData)

			t := viewmodel.Timer{
				Name:   cmd.String("name"),
				Labels: cmd.StringMap("label"),
				Hook: viewmodel.Hook{
					URL:     cmd.String("hook-url"),
					Method:  cmd.String("hook-method"),
					Headers: cmd.StringMap("hook-headers"),
					Body:    hookBodyData,
				},
			}
			if dueUTC := cmd.Timestamp("due-utc"); !dueUTC.IsZero() {
				t.DueUTC = dueUTC
			} else if dueIn := cmd.Duration("due-in"); dueIn > 0 {
				t.DueUTC = time.Now().UTC().Add(dueIn)
			}
			_ = yaml.NewEncoder(os.Stdout).Encode(t)
			return nil
		},
	}
}

func timerCreate() *cli.Command {
	return &cli.Command{
		Name: "create",
		Flags: DefaultClientFlags(
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
			},
		),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			data, err := cliutil.FileOrStdin(cmd.String("file"))
			if err != nil {
				return fmt.Errorf("ticker create; could not read file: %w", err)
			}
			var timer viewmodel.Timer
			if err := yaml.Unmarshal(data, &timer); err != nil {
				return fmt.Errorf("ticker create; could not unmarshal file: %w", err)
			}
			c, err := createClient(cmd)
			if err != nil {
				return fmt.Errorf("ticker create; create client: %w", err)
			}
			res, err := c.CreateTimer(ctx, timer.ToProto())
			if err != nil {
				return fmt.Errorf("ticker create; failed: %w", err)
			}
			fmt.Printf("created timer %s!\n", res.GetId())
			return nil
		},
	}
}

func createClient(cmd *cli.Command) (v1.TimersClient, error) {
	addr := cmd.String("address")
	authority := cmd.String("authority")
	c, err := grpc.NewClient(addr, grpc.WithAuthority(authority), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return v1.NewTimersClient(c), nil
}

func timerList() *cli.Command {
	return &cli.Command{
		Name: "list",
		Flags: DefaultClientFlags(
			&cli.TimestampFlag{
				Name: "after",
			},
			&cli.TimestampFlag{
				Name: "before",
			},
			&cli.StringFlag{
				Name:    "label",
				Aliases: []string{"l"},
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "json",
			},
		),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c, err := createClient(cmd)
			if err != nil {
				return fmt.Errorf("ticker create; create client: %w", err)
			}
			res, err := c.ListTimers(ctx, &v1.ListTimersArgs{
				After:    timestamppb.New(cmd.Timestamp("after")),
				Before:   timestamppb.New(cmd.Timestamp("before")),
				Selector: cmd.String("label"),
			})
			if err != nil {
				return err
			}
			switch cmd.String("output") {
			default:
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "\t")
				return enc.Encode(res.GetTimers())
			case "yaml":
				enc := yaml.NewEncoder(os.Stdout)
				return enc.Encode(res.GetTimers())
			}
			return nil
		},
	}
}

func timerGet() *cli.Command {
	return &cli.Command{
		Name: "get",
		Flags: DefaultClientFlags(
			&cli.StringFlag{
				Name: "id",
			},
			&cli.StringFlag{
				Name: "name",
			},
		),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c, err := createClient(cmd)
			if err != nil {
				return fmt.Errorf("ticker create; create client: %w", err)
			}
			res, err := c.GetTimer(ctx, &v1.GetTimerArgs{
				Id:   cmd.String("id"),
				Name: cmd.String("name"),
			})
			if err != nil {
				return err
			}
			switch cmd.String("output") {
			default:
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "\t")
				return enc.Encode(res)
			case "yaml":
				enc := yaml.NewEncoder(os.Stdout)
				return enc.Encode(res)
			}
			return nil
		},
	}
}

func timerDelete() *cli.Command {
	return &cli.Command{
		Name: "delete",
		Flags: DefaultClientFlags(
			&cli.StringFlag{
				Name: "id",
			},
			&cli.StringFlag{
				Name: "name",
			},
		),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c, err := createClient(cmd)
			if err != nil {
				return fmt.Errorf("ticker create; create client: %w", err)
			}
			_, err = c.DeleteTimer(ctx, &v1.DeleteTimerArgs{
				Id:   cmd.String("id"),
				Name: cmd.String("name"),
			})
			if err != nil {
				return err
			}
			fmt.Println("ok!")
			return nil
		},
	}
}
