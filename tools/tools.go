package tools

import (
	"os"
	"path/filepath"

	pkgerrors "github.com/pkg/errors"

	"github.com/symflower/eval-dev-quality/log"
)

// InstallPathDefault returns the default installation path for tools.
func InstallPathDefault() (installPath string, err error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", pkgerrors.WithStack(pkgerrors.WithMessage(err, "cannot query home directory"))
	}

	return filepath.Join(homePath, ".eval-dev-quality", "bin"), nil
}

// Install install all basic evaluation tools.
func Install(logger *log.Logger, installPath string) (err error) {
	if err := SymflowerInstall(logger, installPath); err != nil {
		return pkgerrors.WithStack(pkgerrors.WithMessage(err, "cannot install Symflower"))
	}

	return nil
}