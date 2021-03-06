package rpm

import (
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/lebauce/nikos/types"
)

type OpenSUSEBackend struct {
	target     *types.Target
	dnfBackend *DnfBackend
}

func (b *OpenSUSEBackend) GetKernelHeaders(directory string) error {
	// For OpenSUSE Leap, we first try with only the repo-oss and repo-update repositories
	// If we don't find it, we use the full list of repositories

	kernelRelease := b.target.Uname.Kernel

	pkgNevra := "kernel"
	flavourIndex := strings.LastIndex(kernelRelease, "-")
	if flavourIndex != -1 {
		pkgNevra += kernelRelease[flavourIndex:]
		kernelRelease = kernelRelease[:flavourIndex]
	}
	pkgNevra += "-devel-" + kernelRelease

	var disabledRepositories []*Repository
	for _, repo := range b.dnfBackend.GetEnabledRepositories() {
		if repo.id != "repo-oss" && repo.id != "repo-update" &&
			repo.id != "openSUSE-Leap-15.2-Oss" && repo.id != "openSUSE-Leap-15.2-Update" {
			b.dnfBackend.DisableRepository(repo)
			disabledRepositories = append(disabledRepositories, repo)
		}
	}

	if err := b.dnfBackend.GetKernelHeaders(pkgNevra, directory); err != nil {
		log.Infof("Retrying with the full set of repositories")
		for _, repo := range disabledRepositories {
			b.dnfBackend.EnableRepository(repo)
		}
		return b.dnfBackend.GetKernelHeaders(pkgNevra, directory)
	}

	return nil
}

func NewOpenSUSEBackend(target *types.Target) (types.Backend, error) {
	dnfBackend, err := NewDnfBackend(target.Distro.Release)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create DNF backend")
	}

	return &OpenSUSEBackend{
		target:     target,
		dnfBackend: dnfBackend,
	}, nil
}
