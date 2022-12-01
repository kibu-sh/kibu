package container

import (
	"context"
	"github.com/discernhq/devx/internal/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/errdefs"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"io"
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
	reader, err := p.client.ImagePull(ctx, params.Image, types.ImagePullOptions{})
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
		err = p.client.ContainerRemove(ctx, value.ID, types.ContainerRemoveOptions{
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
		list, err := p.client.ContainerList(ctx, types.ContainerListOptions{
			Quiet:   true,
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

func (p *Manager) Start(ctx context.Context, params StartParams) (container *Container, err error) {
	container, ok := p.containers.Load(params.Name)
	if !ok {
		err = errors.Wrapf(ErrContainerNotFound, "name: %s", params.Name)
		return
	}

	err = p.client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
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

func (p *Manager) CreateAndStart(ctx context.Context, params CreateParams) (container *Container, err error) {
	container, err = p.Create(ctx, params)
	if err != nil {
		return
	}

	container, err = p.Start(ctx, StartParams{
		Name: params.Name,
	})
	return
}

type StopParams struct {
	ID      string
	Timeout *time.Duration
}

func NewDefaultStopParams(id string) *StopParams {
	return &StopParams{
		ID:      id,
		Timeout: lo.ToPtr(time.Second * 30),
	}
}

func (p *Manager) Stop(ctx context.Context, params StopParams) (err error) {
	return p.client.ContainerStop(ctx, params.ID, params.Timeout)
}
