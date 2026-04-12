package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	v1 "sandman/proto/v1"
	"time"

	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

func Workers() *cli.Command {
	workers := &cli.Command{
		Name:    "workers",
		Aliases: []string{"worker"},
		Usage:   "Control sandman workers",
		Commands: []*cli.Command{
			workerList(),
		},
	}
	return workers
}

func createWorkersClient(cmd *cli.Command) (v1.WorkersClient, error) {
	addr := cmd.String("address")
	authority := cmd.String("authority")
	c, err := grpc.NewClient(addr, grpc.WithAuthority(authority), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return v1.NewWorkersClient(c), nil
}

func workerList() *cli.Command {
	return &cli.Command{
		Name: "list",
		Flags: DefaultClientFlags(
			&cli.TimestampFlag{
				Name:  "after",
				Value: time.Now().UTC().Add(-5 * time.Minute),
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "json",
			},
		),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c, err := createWorkersClient(cmd)
			if err != nil {
				return fmt.Errorf("workers list; create client: %w", err)
			}
			res, err := c.ListWorkers(ctx, &v1.ListWorkersArgs{
				LastSeenAfter: timestamppb.New(cmd.Timestamp("after")),
			})
			if err != nil {
				return err
			}
			switch cmd.String("output") {
			default:
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "\t")
				return enc.Encode(res.GetWorkers())
			case "yaml":
				enc := yaml.NewEncoder(os.Stdout)
				return enc.Encode(res.GetWorkers())
			}
			return nil
		},
	}
}
