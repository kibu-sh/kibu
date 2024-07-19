package container

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/discernhq/devx/pkg/ctxutil"
	"github.com/discernhq/devx/pkg/netx"
	"github.com/discernhq/devx/pkg/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"io"
	"net"
	"time"

	containerapi "github.com/docker/docker/api/types/container"
	networkapi "github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	imagev1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type Contract interface {
	Start() error
	Stop() error
	Restart() error
	Reset() error
}

type Migrations interface {
	Up() error
	Down() error
	Reset() error
	UpTo(version int) error
	DownTo(version int) error
}

type Manager struct {
	client     *dockerclient.Client
	containers *utils.SyncMap[Container]
}

type Container struct {
	ID     string
	Name   string
	Config CreateParams
	Info   types.ContainerJSON
}

type Address struct {
	p  nat.Port
	pb nat.PortBinding
}

func (a Address) Network() string {
	return a.p.Proto()
}

func (a Address) Host() string {
	return a.pb.HostIP
}

func (a Address) Port() string {
	return a.pb.HostPort
}

func (a Address) String() string {
	return fmt.Sprintf("%s://%s", a.Network(), a.HostPort())
}

func (a Address) HostPort() string {
	return net.JoinHostPort(a.Host(), a.Port())
}

var ErrPortNotFound = errors.New("port not found")

func (c *Container) Address(port string) (address netx.Address, err error) {
	for p, bindings := range c.Info.HostConfig.PortBindings {
		if string(p) == port {
			for _, binding := range bindings {
				address = Address{
					p:  p,
					pb: binding,
				}
				return
			}
		}
	}
	err = errors.Wrapf(ErrPortNotFound, "port: %s", port)
	return
}

var DefaultPostgresImage = "postgres:14"

func NewManager() (p *Manager, err error) {
	p = &Manager{
		containers: utils.NewSyncMap[Container](),
	}

	p.client, err = dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return
	}

	return
}

type CreateParams struct {
	Name      string
	Platform  *imagev1.Platform
	Container *containerapi.Config
	Host      *containerapi.HostConfig
	Network   *networkapi.NetworkingConfig
}

type ImagePullParams struct {
	Image string
	Out   io.Writer
}

func (p *Manager) PullImage(ctx context.Context, params ImagePullParams) (err error) {
	reader, err := p.client.ImagePull(ctx, params.Image, image.PullOptions{})
	if err != nil {
		return
	}
	// this API is wonky
	// it forces us to read the entire response body
	// this data is intended to show image pull progress in the terminal
	defer reader.Close()
	_, err = io.Copy(params.Out, reader)
	return
}

func (p *Manager) Cleanup(ctx context.Context) (err error) {
	p.containers.Range(func(key string, value *Container) bool {
		err = p.client.ContainerRemove(ctx, value.ID, containerapi.RemoveOptions{
			Force: true,
		})
		return err != nil
	})
	return
}

func (p *Manager) Create(ctx context.Context, params CreateParams) (container *Container, err error) {
	err = p.PullImage(ctx, ImagePullParams{
		Image: params.Container.Image,
		// ignore the output of the image pull it's required
		Out: io.Discard,
	})
	if err != nil {
		return
	}

	body, err := p.client.ContainerCreate(ctx, params.Container, params.Host, params.Network, params.Platform, params.Name)
	// bail on non-conflict errors
	if err != nil && !errdefs.IsConflict(err) {
		return
	}

	// guard on existing container
	if errdefs.IsConflict(err) {
		list, err := p.client.ContainerList(ctx, containerapi.ListOptions{
			Size:    false,
			All:     true,
			Filters: filters.NewArgs(filters.Arg("name", params.Name)),
		})

		if err != nil {
			return container, err
		}

		body.ID = list[0].ID
	}

	info, err := p.client.ContainerInspect(ctx, body.ID)
	if err != nil {
		return
	}

	container = &Container{
		ID:     body.ID,
		Name:   params.Name,
		Config: params,
		Info:   info,
	}

	p.containers.Store(params.Name, container)
	return
}

type RefreshParams struct {
	Name string
}

var ErrContainerNotFound = errors.New("container not found")

func (p *Manager) RefreshState(ctx context.Context, params RefreshParams) (container *Container, err error) {
	container, ok := p.containers.Load(params.Name)
	if !ok {
		err = errors.Wrapf(ErrContainerNotFound, "name: %s", params.Name)
		return
	}

	container.Info, err = p.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return
	}

	p.containers.Store(params.Name, container)
	return
}

type StartParams struct {
	Name    string
	Timeout *time.Duration
}

var ErrStartTimeout = errors.New("timeout waiting for container to start")

type StartOption func(container *Container) error

func (p *Manager) Start(ctx context.Context, params StartParams, opts ...StartOption) (container *Container, err error) {

	defer func() {
		if err != nil {
			return
		}
		for _, opt := range opts {
			err = opt(container)
			if err != nil {
				return
			}
		}
	}()

	container, ok := p.containers.Load(params.Name)
	if !ok {
		err = errors.Wrapf(ErrContainerNotFound, "name: %s", params.Name)
		return
	}

	err = p.client.ContainerStart(ctx, container.ID, containerapi.StartOptions{})
	if err != nil {
		return
	}

	container, err = p.RefreshState(ctx, RefreshParams{
		Name: params.Name,
	})

	if container.Info.State.Running {
		return
	}

	if params.Timeout == nil {
		params.Timeout = lo.ToPtr(time.Second * 30)
	}

	statusChan, errChan := p.client.ContainerWait(ctx, container.ID, containerapi.WaitConditionNotRunning)
	select {
	// error waiting for container
	case err = <-errChan:
		return
	// timeout
	case <-time.After(*params.Timeout):
		err = errors.Wrapf(ErrStartTimeout, "name: %s, id: %s", params.Name, container.ID)
		return
	// canceled
	case <-ctx.Done():
		err = ctx.Err()
		return
	// success
	case <-statusChan:
		container, err = p.RefreshState(ctx, RefreshParams{
			Name: params.Name,
		})
		return
	}
}

func waitForPortUsingBackoff(addr string) error {
	return backoff.Retry(func() (err error) {
		conn, err := net.Dial("tcp", addr)
		if conn != nil {
			_ = conn.Close()
		}
		return
	}, backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 30))
}

func WaitForPort(port string) StartOption {
	return func(container *Container) (err error) {
		for p, bindings := range container.Info.HostConfig.PortBindings {
			if string(p) == port {
				for _, binding := range bindings {
					err = waitForPortUsingBackoff(binding.HostIP + ":" + binding.HostPort)
					if err != nil {
						return
					}
				}
			}
		}
		return
	}
}

func (p *Manager) CreateAndStart(ctx context.Context, params CreateParams, opts ...StartOption) (container *Container, err error) {
	container, err = p.Create(ctx, params)
	if err != nil {
		return
	}

	container, err = p.Start(ctx, StartParams{
		Name: params.Name,
	}, opts...)
	return
}

type StopParams struct {
	ID      string
	Timeout *int
}

func NewDefaultStopParams(id string) *StopParams {
	return &StopParams{
		ID:      id,
		Timeout: lo.ToPtr(30),
	}
}

func (p *Manager) Stop(ctx context.Context, params StopParams) (err error) {
	return p.client.ContainerStop(ctx, params.ID, containerapi.StopOptions{
		Timeout: params.Timeout,
	})
}

type containerManagerCtxKey struct{}

var ManagerContextStore = ctxutil.NewStore[*Manager, containerManagerCtxKey]()
