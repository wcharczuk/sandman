package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.charczuk.com/sdk/apputil"
	"go.charczuk.com/sdk/configutil"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/db/migration"
	"go.charczuk.com/sdk/log"
	"go.charczuk.com/sdk/slant"

	"sandman/pkg/config"
	"sandman/pkg/control"
	"sandman/pkg/model"
)

var (
	flagMode               = flag.String("mode", "", "Controller mode: log, fork, k8s")
	flagEvaluationInterval = flag.Duration("evaluation-interval", 0, "How often to evaluate desired scale")
	flagMinReplicas        = flag.Int("min-replicas", 0, "Minimum number of worker replicas")
	flagMaxReplicas        = flag.Int("max-replicas", 0, "Maximum number of worker replicas (0 = unlimited)")
	flagCullInterval       = flag.Duration("cull-interval", 0, "How often to sweep delivered/exhausted timers")
	flagCullRetention      = flag.Duration("cull-retention", 0, "How long past due_utc a delivered timer is kept before cull")
	flagNamespace          = flag.String("namespace", "", "Kubernetes namespace (k8s mode)")
	flagDeployment         = flag.String("deployment", "", "Deployment name to scale (k8s mode)")
	flagLeaseName          = flag.String("lease-name", "", "Lease name for leader election (k8s mode)")
	flagPodName            = flag.String("pod-name", "", "Pod identity for leader election (k8s mode)")
)

type controlConfig struct {
	config.Config `yaml:",inline"`

	Mode               string            `yaml:"mode"`
	EvaluationInterval time.Duration     `yaml:"evaluation_interval"`
	MinReplicas        int               `yaml:"min_replicas"`
	MaxReplicas        int               `yaml:"max_replicas"`
	CullInterval       time.Duration     `yaml:"cull_interval"`
	CullRetention      time.Duration     `yaml:"cull_retention"`
	K8s                control.K8sConfig `yaml:"k8s"`
}

func (c *controlConfig) Resolve(ctx context.Context) error {
	return configutil.Resolve(ctx,
		(&c.Config).Resolve,
		configutil.Set(&c.Mode, configutil.Lazy(flagMode), configutil.Env[string]("CONTROLLER_MODE"), configutil.Const("log")),
		configutil.Set(&c.EvaluationInterval, configutil.Lazy(flagEvaluationInterval), configutil.Env[time.Duration]("EVALUATION_INTERVAL")),
		configutil.Set(&c.MinReplicas, configutil.Lazy(flagMinReplicas), configutil.Env[int]("MIN_REPLICAS")),
		configutil.Set(&c.MaxReplicas, configutil.Lazy(flagMaxReplicas), configutil.Env[int]("MAX_REPLICAS")),
		configutil.Set(&c.CullInterval, configutil.Lazy(flagCullInterval), configutil.Env[time.Duration]("CULL_INTERVAL")),
		configutil.Set(&c.CullRetention, configutil.Lazy(flagCullRetention), configutil.Env[time.Duration]("CULL_RETENTION")),
		configutil.Set(&c.K8s.Namespace, configutil.Lazy(flagNamespace), configutil.Env[string]("POD_NAMESPACE")),
		configutil.Set(&c.K8s.Deployment, configutil.Lazy(flagDeployment), configutil.Env[string]("DEPLOYMENT_NAME")),
		configutil.Set(&c.K8s.LeaseName, configutil.Lazy(flagLeaseName), configutil.Env[string]("LEASE_NAME"), configutil.Const("sandman-control")),
		configutil.Set(&c.K8s.PodName, configutil.Lazy(flagPodName), configutil.Env[string]("POD_NAME"), configutil.Env[string]("HOSTNAME")),
	)
}

var entrypoint = apputil.DBEntryPoint[controlConfig]{
	Setup: func(_ context.Context, _ controlConfig) error {
		return nil
	},
	Migrate: func(ctx context.Context, _ controlConfig, dbc *db.Connection) error {
		return model.Migrations(
			migration.OptLog(log.GetLogger(ctx)),
		).Apply(ctx, dbc)
	},
	Start: func(ctx context.Context, cfg controlConfig, dbc *db.Connection) error {
		slant.Print(os.Stdout, "sandman-control")

		ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()

		modelMgr := &model.Manager{
			BaseManager: dbutil.NewBaseManager(dbc),
		}
		if err := modelMgr.Initialize(ctx); err != nil {
			return err
		}

		ctrlCfg := control.Config{
			EvaluationInterval:    cfg.EvaluationInterval,
			MinReplicas:           int32(cfg.MinReplicas),
			MaxReplicas:           int32(cfg.MaxReplicas),
			WorkerBatchSize:       cfg.Worker.BatchSizeOrDefault(),
			WorkerPollingInterval: cfg.Worker.PollingIntervalOrDefault(),
			CullInterval:          cfg.CullInterval,
			CullRetention:         cfg.CullRetention,
		}

		logger := log.GetLogger(ctx)

		var scaler control.Scaler
		var k8sClient *control.K8sClient

		switch cfg.Mode {
		case "log":
			scaler = &control.LogScaler{}
		case "fork":
			fs := &control.ForkScaler{}
			defer fs.Close()
			scaler = fs
		case "k8s":
			if cfg.K8s.Deployment == "" {
				return fmt.Errorf("--deployment is required for k8s mode")
			}
			if cfg.K8s.Namespace == "" {
				ns, err := control.InClusterNamespace()
				if err != nil {
					return fmt.Errorf("--namespace is required (or must run in-cluster): %w", err)
				}
				cfg.K8s.Namespace = ns
			}
			if cfg.K8s.PodName == "" {
				return fmt.Errorf("--pod-name is required for k8s mode")
			}
			var err error
			k8sClient, err = control.NewK8sClientInCluster()
			if err != nil {
				return fmt.Errorf("creating k8s client: %w", err)
			}
			scaler = &control.K8sScaler{
				Client:         k8sClient,
				Namespace:      cfg.K8s.Namespace,
				DeploymentName: cfg.K8s.Deployment,
			}
		default:
			return fmt.Errorf("unknown mode: %q (must be log, fork, or k8s)", cfg.Mode)
		}

		ctrl := &control.Controller{
			Config: ctrlCfg,
			Model:  modelMgr,
			Scaler: scaler,
		}

		logger.Info("starting controller", log.String("mode", cfg.Mode))

		if cfg.Mode == "k8s" {
			le := &control.LeaderElector{
				Client:    k8sClient,
				Namespace: cfg.K8s.Namespace,
				LeaseName: cfg.K8s.LeaseName,
				Identity:  cfg.K8s.PodName,
			}
			return le.Run(ctx, func(leaderCtx context.Context) {
				if err := ctrl.Run(leaderCtx); err != nil {
					logger.Error("controller error", log.Any("err", err))
				}
			})
		}
		return ctrl.Run(ctx)
	},
}

func init() {
	entrypoint.Init()
}

func main() {
	entrypoint.Main()
}
