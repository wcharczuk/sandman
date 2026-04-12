package control

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.charczuk.com/sdk/log"
)

//
// Kubernetes REST client
//

const (
	serviceAccountTokenPath     = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	serviceAccountCAPath        = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	serviceAccountNamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

// K8sClient is a minimal Kubernetes REST API client for in-cluster use.
type K8sClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewK8sClientInCluster creates a K8sClient using the in-cluster service account.
func NewK8sClientInCluster() (*K8sClient, error) {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" || port == "" {
		return nil, fmt.Errorf("not running in a kubernetes cluster (KUBERNETES_SERVICE_HOST/PORT not set)")
	}
	tokenBytes, err := os.ReadFile(serviceAccountTokenPath)
	if err != nil {
		return nil, fmt.Errorf("reading service account token: %w", err)
	}
	caCert, err := os.ReadFile(serviceAccountCAPath)
	if err != nil {
		return nil, fmt.Errorf("reading service account ca cert: %w", err)
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)
	return &K8sClient{
		baseURL: fmt.Sprintf("https://%s:%s", host, port),
		token:   string(tokenBytes),
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: pool,
				},
			},
		},
	}, nil
}

// InClusterNamespace returns the namespace this pod is running in.
func InClusterNamespace() (string, error) {
	data, err := os.ReadFile(serviceAccountNamespacePath)
	if err != nil {
		return "", fmt.Errorf("reading service account namespace: %w", err)
	}
	return string(data), nil
}

func (c *K8sClient) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return c.client.Do(req)
}

//
// Kubernetes types
//

type k8sObjectMeta struct {
	Name            string `json:"name"`
	Namespace       string `json:"namespace,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

type k8sMicroTime struct {
	time.Time
}

func (t k8sMicroTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.UTC().Format("2006-01-02T15:04:05.000000Z"))
}

func (t *k8sMicroTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

func ptrMicroTime(t time.Time) *k8sMicroTime { return &k8sMicroTime{Time: t} }

//
// Lease types and operations
//

type k8sLease struct {
	APIVersion string        `json:"apiVersion"`
	Kind       string        `json:"kind"`
	Metadata   k8sObjectMeta `json:"metadata"`
	Spec       k8sLeaseSpec  `json:"spec"`
}

type k8sLeaseSpec struct {
	HolderIdentity       *string       `json:"holderIdentity,omitempty"`
	LeaseDurationSeconds *int32        `json:"leaseDurationSeconds,omitempty"`
	AcquireTime          *k8sMicroTime `json:"acquireTime,omitempty"`
	RenewTime            *k8sMicroTime `json:"renewTime,omitempty"`
	LeaseTransitions     *int32        `json:"leaseTransitions,omitempty"`
}

func (c *K8sClient) GetLease(ctx context.Context, namespace, name string) (*k8sLease, error) {
	resp, err := c.do(ctx, http.MethodGet,
		fmt.Sprintf("/apis/coordination.k8s.io/v1/namespaces/%s/leases/%s", namespace, name), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get lease: %d: %s", resp.StatusCode, body)
	}
	var l k8sLease
	return &l, json.NewDecoder(resp.Body).Decode(&l)
}

func (c *K8sClient) CreateLease(ctx context.Context, l *k8sLease) error {
	resp, err := c.do(ctx, http.MethodPost,
		fmt.Sprintf("/apis/coordination.k8s.io/v1/namespaces/%s/leases", l.Metadata.Namespace), l)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create lease: %d: %s", resp.StatusCode, body)
	}
	return json.NewDecoder(resp.Body).Decode(l)
}

func (c *K8sClient) UpdateLease(ctx context.Context, l *k8sLease) error {
	resp, err := c.do(ctx, http.MethodPut,
		fmt.Sprintf("/apis/coordination.k8s.io/v1/namespaces/%s/leases/%s", l.Metadata.Namespace, l.Metadata.Name), l)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update lease: %d: %s", resp.StatusCode, body)
	}
	return json.NewDecoder(resp.Body).Decode(l)
}

//
// Scale types and operations
//

type k8sScale struct {
	APIVersion string        `json:"apiVersion"`
	Kind       string        `json:"kind"`
	Metadata   k8sObjectMeta `json:"metadata"`
	Spec       k8sScaleSpec  `json:"spec"`
}

type k8sScaleSpec struct {
	Replicas int32 `json:"replicas"`
}

func (c *K8sClient) GetDeploymentScale(ctx context.Context, namespace, name string) (*k8sScale, error) {
	resp, err := c.do(ctx, http.MethodGet,
		fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments/%s/scale", namespace, name), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get deployment scale: %d: %s", resp.StatusCode, body)
	}
	var s k8sScale
	return &s, json.NewDecoder(resp.Body).Decode(&s)
}

func (c *K8sClient) UpdateDeploymentScale(ctx context.Context, scale *k8sScale) error {
	resp, err := c.do(ctx, http.MethodPut,
		fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments/%s/scale",
			scale.Metadata.Namespace, scale.Metadata.Name), scale)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update deployment scale: %d: %s", resp.StatusCode, body)
	}
	return nil
}

//
// K8sScaler implements Scaler by adjusting a Kubernetes Deployment's replica count.
//

// K8sScaler scales a Kubernetes Deployment.
type K8sScaler struct {
	Client         *K8sClient
	Namespace      string
	DeploymentName string
}

func (s *K8sScaler) SetDesiredScale(ctx context.Context, desired int32) error {
	logger := log.GetLogger(ctx)
	scale, err := s.Client.GetDeploymentScale(ctx, s.Namespace, s.DeploymentName)
	if err != nil {
		return fmt.Errorf("getting current scale: %w", err)
	}
	if scale.Spec.Replicas == desired {
		return nil
	}
	logger.Info("k8s scaler; scaling deployment",
		log.String("deployment", s.DeploymentName),
		log.Int("from", int(scale.Spec.Replicas)),
		log.Int("to", int(desired)),
	)
	scale.Spec.Replicas = desired
	return s.Client.UpdateDeploymentScale(ctx, scale)
}

//
// LeaderElector implements leader election using Kubernetes Lease objects.
//

// LeaderElector manages leader election using Kubernetes Lease objects.
// It is designed for a triplet of pods where only one actively controls scale.
type LeaderElector struct {
	Client    *K8sClient
	Namespace string
	LeaseName string
	Identity  string

	LeaseDuration time.Duration
	RenewInterval time.Duration
	RetryInterval time.Duration

	currentLease *k8sLease
}

const (
	DefaultLeaseDuration = 15 * time.Second
	DefaultRenewInterval = 10 * time.Second
	DefaultRetryInterval = 5 * time.Second
)

func (le *LeaderElector) leaseDuration() time.Duration {
	if le.LeaseDuration > 0 {
		return le.LeaseDuration
	}
	return DefaultLeaseDuration
}

func (le *LeaderElector) renewInterval() time.Duration {
	if le.RenewInterval > 0 {
		return le.RenewInterval
	}
	return DefaultRenewInterval
}

func (le *LeaderElector) retryInterval() time.Duration {
	if le.RetryInterval > 0 {
		return le.RetryInterval
	}
	return DefaultRetryInterval
}

// Run blocks and manages leader election. When this instance becomes the leader,
// onElected is called with a context that will be cancelled if leadership is lost.
// The function loops, re-contesting the election after each leadership loss.
func (le *LeaderElector) Run(ctx context.Context, onElected func(ctx context.Context)) error {
	logger := log.GetLogger(ctx)
	for {
		if ctx.Err() != nil {
			return nil
		}
		acquired, err := le.tryAcquire(ctx)
		if err != nil {
			logger.Error("leader election; acquire error", log.Any("err", err))
			sleepContext(ctx, le.retryInterval())
			continue
		}
		if !acquired {
			logger.Info("leader election; not leader, retrying",
				log.Duration("retry_in", le.retryInterval()),
			)
			sleepContext(ctx, le.retryInterval())
			continue
		}

		logger.Info("leader election; acquired leadership",
			log.String("identity", le.Identity),
		)
		leaderCtx, cancel := context.WithCancel(ctx)
		renewDone := make(chan struct{})
		go func() {
			defer close(renewDone)
			le.renewLoop(leaderCtx, cancel)
		}()
		onElected(leaderCtx)
		cancel()
		<-renewDone
		le.currentLease = nil
		logger.Info("leader election; stepped down")
	}
}

func (le *LeaderElector) tryAcquire(ctx context.Context) (bool, error) {
	now := time.Now().UTC()
	durationSec := int32(le.leaseDuration().Seconds())

	existing, err := le.Client.GetLease(ctx, le.Namespace, le.LeaseName)
	if err != nil {
		return false, err
	}

	if existing == nil {
		// Lease doesn't exist, try to create it
		l := &k8sLease{
			APIVersion: "coordination.k8s.io/v1",
			Kind:       "Lease",
			Metadata: k8sObjectMeta{
				Name:      le.LeaseName,
				Namespace: le.Namespace,
			},
			Spec: k8sLeaseSpec{
				HolderIdentity:       &le.Identity,
				LeaseDurationSeconds: &durationSec,
				AcquireTime:          ptrMicroTime(now),
				RenewTime:            ptrMicroTime(now),
				LeaseTransitions:     ptrInt32(0),
			},
		}
		if err := le.Client.CreateLease(ctx, l); err != nil {
			return false, nil // someone else won the race
		}
		le.currentLease = l
		return true, nil
	}

	// We already hold the lease; renew it
	if existing.Spec.HolderIdentity != nil && *existing.Spec.HolderIdentity == le.Identity {
		existing.Spec.RenewTime = ptrMicroTime(now)
		if err := le.Client.UpdateLease(ctx, existing); err != nil {
			return false, err
		}
		le.currentLease = existing
		return true, nil
	}

	// Check if the current holder's lease has expired
	if existing.Spec.RenewTime != nil && existing.Spec.LeaseDurationSeconds != nil {
		expiry := existing.Spec.RenewTime.Add(time.Duration(*existing.Spec.LeaseDurationSeconds) * time.Second)
		if now.After(expiry) {
			transitions := int32(0)
			if existing.Spec.LeaseTransitions != nil {
				transitions = *existing.Spec.LeaseTransitions
			}
			transitions++
			existing.Spec.HolderIdentity = &le.Identity
			existing.Spec.LeaseDurationSeconds = &durationSec
			existing.Spec.AcquireTime = ptrMicroTime(now)
			existing.Spec.RenewTime = ptrMicroTime(now)
			existing.Spec.LeaseTransitions = &transitions
			if err := le.Client.UpdateLease(ctx, existing); err != nil {
				return false, nil // conflict, another pod acquired it
			}
			le.currentLease = existing
			return true, nil
		}
	}
	return false, nil
}

func (le *LeaderElector) renewLoop(ctx context.Context, cancel context.CancelFunc) {
	tick := time.NewTicker(le.renewInterval())
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if le.currentLease == nil {
				cancel()
				return
			}
			le.currentLease.Spec.RenewTime = ptrMicroTime(time.Now().UTC())
			if err := le.Client.UpdateLease(ctx, le.currentLease); err != nil {
				log.GetLogger(ctx).Error("leader election; renew failed, stepping down",
					log.Any("err", err),
				)
				cancel()
				return
			}
		}
	}
}

func sleepContext(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

func ptrInt32(v int32) *int32 { return &v }
