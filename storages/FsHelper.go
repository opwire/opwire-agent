package storages

import (
	"github.com/spf13/afero"
)

var _fs_ afero.Fs

func SetFs(newFs afero.Fs) {
	if newFs != nil {
		_fs_ = newFs
	}
}

func GetFs() afero.Fs {
	if _fs_ == nil {
		_fs_ = afero.NewOsFs()
	}
	return _fs_
}
