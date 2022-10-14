/*
Copyright 2022 The KubeBlocks Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helm

import (
	"context"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/containers/common/pkg/retry"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/apecloud/kubeblocks/internal/dbctl/util"
)

const (
	helmUser   = "yimeisun"
	helmPasswd = "8V+PmX1oSDv4pumDvZp6m7LS8iPgbY3A"
	helmURL    = "yimeisun.azurecr.io"
)

type InstallOpts struct {
	Name      string
	Chart     string
	Namespace string
	Sets      []string
	Wait      bool
	Version   string
	TryTimes  int
	Login     bool
}

type LoginOpts struct {
	User   string
	Passwd string
	URL    string
}

// AddRepo will add a repo
func AddRepo(r *repo.Entry) error {
	settings := cli.New()
	repoFile := settings.RepositoryConfig
	b, err := os.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}

	// Check if the repo Name is legal
	if strings.Contains(r.Name, "/") {
		return errors.Errorf("repository Name (%s) contains '/', please specify a different Name without '/'", r.Name)
	}

	if f.Has(r.Name) {
		existing := f.Get(r.Name)
		if *r != *existing {

			// The input coming in for the Name is different from what is already
			// configured. Return an error.
			return errors.Errorf("repository Name (%s) already exists, please specify a different Name", r.Name)
		}

		// The add is idempotent so do nothing
		return nil
	}

	cp, err := repo.NewChartRepository(r, getter.All(settings))
	if err != nil {
		return err
	}

	if _, err := cp.DownloadIndexFile(); err != nil {
		return errors.Wrapf(err, "looks like %q is not a valid Chart repository or cannot be reached", r.URL)
	}

	f.Update(r)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		return err
	}
	util.Infof("%s has been added to your repositories\n", r.Name)
	return nil
}

// RemoveRepo will remove a repo
func RemoveRepo(r *repo.Entry) error {
	settings := cli.New()
	repoFile := settings.RepositoryConfig
	b, err := os.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}

	if f.Has(r.Name) {
		f.Remove(r.Name)
		if err := f.WriteFile(repoFile, 0644); err != nil {
			return err
		}
	}
	util.Infof("%s has been remove to your repositories\n", r.Name)
	return nil
}

// getInstalled get helm package if installed.
func (i *InstallOpts) getInstalled(cfg *action.Configuration) (*release.Release, error) {
	getClient := action.NewGet(cfg)
	res, err := getClient.Run(i.Name)
	if err != nil {
		if strings.Contains(err.Error(), "release: not found") {
			return nil, nil
		}
		util.Infof("Failed check %s installed\n", i.Name)
		return nil, err
	}
	return res, nil
}

// Install will install a Chart
func (i *InstallOpts) Install(cfg *action.Configuration) error {
	ctx := context.Background()
	opts := retry.Options{
		MaxRetry: 1 + i.TryTimes,
	}

	if err := retry.IfNecessary(ctx, func() error {
		if err := i.tryInstall(cfg); err != nil {
			return err
		}
		return nil
	}, &opts); err != nil {
		return errors.Errorf("Install chart %s error: %s", i.Name, err.Error())
	}

	return nil
}

func (i *InstallOpts) tryInstall(cfg *action.Configuration) error {
	util.InfoP(1, "Install "+i.Chart+"...")
	s := spinner.New(spinner.CharSets[rand.Intn(44)], 100*time.Millisecond)
	if err := s.Color("green"); err != nil {
		return err
	}
	s.Start()
	defer s.Stop()

	res, _ := i.getInstalled(cfg)
	if res != nil {
		return nil
	}

	settings := cli.New()
	err := i.tryLogin(cfg)
	if err != nil {
		return err
	}

	// TODO: Does not work now
	// If a release does not exist, install it.
	histClient := action.NewHistory(cfg)
	histClient.Max = 1
	if _, err := histClient.Run(i.Name); err != nil && err != driver.ErrReleaseNotFound {
		return err
	}

	client := action.NewInstall(cfg)
	client.ReleaseName = i.Name
	client.Namespace = i.Namespace
	client.CreateNamespace = true
	client.Wait = i.Wait
	client.Timeout = time.Second * 300
	client.Version = i.Version

	cp, err := client.ChartPathOptions.LocateChart(i.Chart, settings)
	if err != nil {
		return err
	}

	setOpts := values.Options{
		Values: i.Sets,
	}

	p := getter.All(settings)
	vals, err := setOpts.MergeValues(p)
	if err != nil {
		return err
	}

	// Check Chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return err
	}

	// Create context and prepare the handle of SIGTERM
	ctx := context.Background()
	_, cancel := context.WithCancel(ctx)

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	cSignal := make(chan os.Signal, 2)
	signal.Notify(cSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cSignal
		util.Infof("Install has been cancelled.\n")
		cancel()
	}()

	_, err = client.RunWithContext(ctx, chartRequested, vals)
	if err != nil && err.Error() != "cannot re-use a name that is still in use" {
		return err
	}
	return nil
}

// UnInstall will uninstall a Chart
func (i *InstallOpts) UnInstall(cfg *action.Configuration) error {
	ctx := context.Background()
	opts := retry.Options{
		MaxRetry: 1 + i.TryTimes,
	}

	if err := retry.IfNecessary(ctx, func() error {
		if err := i.tryUnInstall(cfg); err != nil {
			return err
		}
		return nil
	}, &opts); err != nil {
		return errors.Errorf("UnInstall chart %s error: %s", i.Name, err.Error())
	}

	return nil
}

func (i *InstallOpts) tryUnInstall(cfg *action.Configuration) error {
	util.InfoP(1, "uninstall "+i.Name+"...")
	s := spinner.New(spinner.CharSets[rand.Intn(44)], 100*time.Millisecond)
	if err := s.Color("green"); err != nil {
		return err
	}
	s.Start()
	defer s.Stop()

	err := i.tryLogin(cfg)
	if err != nil {
		return err
	}

	client := action.NewUninstall(cfg)
	client.Wait = i.Wait
	client.Timeout = time.Second * 300

	// Create context and prepare the handle of SIGTERM
	ctx := context.Background()
	_, cancel := context.WithCancel(ctx)

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	cSignal := make(chan os.Signal, 2)
	signal.Notify(cSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cSignal
		util.Infof("Install has been cancelled.\n")
		cancel()
	}()

	_, err = client.Run(i.Name)
	if err != nil {
		return err
	}
	return nil
}

func (i *InstallOpts) tryLogin(cfg *action.Configuration) error {
	if !i.Login {
		return nil
	}

	return cfg.RegistryClient.Login(helmURL, registry.LoginOptBasicAuth(helmUser, helmPasswd),
		registry.LoginOptInsecure(false))
}

func NewActionConfig(ns string, config string) (*action.Configuration, error) {
	settings := cli.New()
	cfg := new(action.Configuration)

	settings.SetNamespace(ns)
	settings.KubeConfig = config
	registryClient, err := registry.NewClient(
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(io.Discard),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	)
	if err != nil {
		return nil, err
	}
	cfg.RegistryClient = registryClient
	err = cfg.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {})
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func FakeActionConfig() *action.Configuration {
	registryClient, err := registry.NewClient()
	if err != nil {
		return nil
	}

	return &action.Configuration{
		Releases:       storage.Init(driver.NewMemory()),
		KubeClient:     &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: io.Discard}},
		Capabilities:   chartutil.DefaultCapabilities,
		RegistryClient: registryClient,
		Log: func(format string, v ...interface{}) {
		},
	}
}